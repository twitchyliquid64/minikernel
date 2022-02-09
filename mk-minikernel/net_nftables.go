package main

import (
	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"inet.af/netaddr"
)

type firewall struct {
	allowICMP bool
	allowTCP  bool
	allowUDP  bool

	deny netaddr.IPSet
}

func (fw *firewall) CreateSets(n *rulesNFTables) (*nftables.Set, *nftables.Set, error) {
	pSet := &nftables.Set{
		Name:    "permitted_protos",
		Table:   n.table,
		KeyType: nftables.TypeInetProto,
	}

	var pElements []nftables.SetElement

	if fw.allowTCP {
		pElements = append(pElements, nftables.SetElement{Key: []byte{unix.IPPROTO_TCP}})
	}
	if fw.allowUDP {
		pElements = append(pElements, nftables.SetElement{Key: []byte{unix.IPPROTO_UDP}})
	}
	if fw.allowICMP {
		pElements = append(pElements, nftables.SetElement{Key: []byte{unix.IPPROTO_ICMP}})
	}
	if err := n.nft.AddSet(pSet, pElements); err != nil {
		return nil, nil, err
	}

	ipSet := &nftables.Set{
		Name:     "ip_denylist",
		Table:    n.table,
		Interval: true,
		KeyType:  nftables.TypeIPAddr,  // prefix
		DataType: nftables.TypeInteger, // mask
	}
	var iElements []nftables.SetElement
	for _, r := range fw.deny.Ranges() {
		from, err := r.From().MarshalBinary()
		if err != nil {
			return nil, nil, err
		}
		to, err := r.To().MarshalBinary()
		if err != nil {
			return nil, nil, err
		}

		iElements = append(iElements, nftables.SetElement{
			Key:         from,
			IntervalEnd: false,
		}, nftables.SetElement{
			Key:         to,
			IntervalEnd: true,
		})
	}

	if err := n.nft.AddSet(ipSet, iElements); err != nil {
		return nil, nil, err
	}
	return pSet, ipSet, nil
}

type rulesNFTables struct {
	nft   *nftables.Conn
	table *nftables.Table

	prerouteChain  *nftables.Chain
	postrouteChain *nftables.Chain
	natCounter     *nftables.CounterObj

	filterChain   *nftables.Chain
	fwdOutCounter *nftables.CounterObj
	fwdInCounter  *nftables.CounterObj
	dropCounter   *nftables.CounterObj
}

func (n *rulesNFTables) initTable(id string) error {
	n.table = &nftables.Table{
		Family: nftables.TableFamilyIPv4,
		Name:   "minikernel-" + id,
	}
	n.table = n.nft.AddTable(n.table)

	n.prerouteChain = n.nft.AddChain(&nftables.Chain{
		Name:     "prerouting",
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityFilter,
		Table:    n.table,
		Type:     nftables.ChainTypeNAT,
	})
	n.postrouteChain = n.nft.AddChain(&nftables.Chain{
		Name:     "postrouting",
		Hooknum:  nftables.ChainHookPostrouting,
		Priority: nftables.ChainPriorityNATSource,
		Table:    n.table,
		Type:     nftables.ChainTypeNAT,
	})
	n.natCounter = n.nft.AddObj(&nftables.CounterObj{
		Table: n.table,
		Name:  "natted",
	}).(*nftables.CounterObj)
	n.fwdOutCounter = n.nft.AddObj(&nftables.CounterObj{
		Table: n.table,
		Name:  "fwd-out",
	}).(*nftables.CounterObj)
	n.fwdInCounter = n.nft.AddObj(&nftables.CounterObj{
		Table: n.table,
		Name:  "fwd-in",
	}).(*nftables.CounterObj)
	n.dropCounter = n.nft.AddObj(&nftables.CounterObj{
		Table: n.table,
		Name:  "drop",
	}).(*nftables.CounterObj)

	n.filterChain = n.nft.AddChain(&nftables.Chain{
		Name:     "forward",
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
		Table:    n.table,
		Type:     nftables.ChainTypeFilter,
	})
	return nil
}

func (n *rulesNFTables) makeFirewall(id string, tap netlink.Link, fw firewall) error {
	if n.table == nil {
		if err := n.initTable(id); err != nil {
			return err
		}
	}

	protoSet, denySet, err := fw.CreateSets(n)
	if err != nil {
		return err
	}

	// Filter by protocol
	n.nft.AddRule(&nftables.Rule{
		Table: n.table,
		Chain: n.filterChain,
		Exprs: []expr.Any{
			// Load meta-information 'output-interface ID' => reg 1
			&expr.Meta{
				Key:      expr.MetaKeyOIF,
				Register: 1,
			},
			// cmp eq reg 1 0x0245a8c0
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     binaryutil.NativeEndian.PutUint32(uint32(tap.Attrs().Index)),
			},
			// [ meta load l4proto => reg 1 ]
			&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
			&expr.Lookup{
				SourceRegister: 1,
				SetName:        protoSet.Name,
				SetID:          protoSet.ID,
				Invert:         true,
			},
			&expr.Objref{
				Type: 1,
				Name: n.dropCounter.Name,
			},
			//[ immediate reg 0 drop ]
			&expr.Verdict{
				Kind: expr.VerdictDrop,
			},
		}})

	// Filter by destination IP
	n.nft.AddRule(&nftables.Rule{
		Table: n.table,
		Chain: n.filterChain,
		Exprs: []expr.Any{
			// Load meta-information 'output-interface ID' => reg 1
			&expr.Meta{
				Key:      expr.MetaKeyOIF,
				Register: 1,
			},
			// cmp eq reg 1 0x0245a8c0
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     binaryutil.NativeEndian.PutUint32(uint32(tap.Attrs().Index)),
			},
			// payload load 4b @ network header + 12 => reg 1
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       12,
				Len:          4,
			},
			// Match from target set
			&expr.Lookup{
				SourceRegister: 1,
				SetName:        denySet.Name,
				SetID:          denySet.ID,
			},
			&expr.Objref{
				Type: 1,
				Name: n.dropCounter.Name,
			},
			//[ immediate reg 0 drop ]
			&expr.Verdict{
				Kind: expr.VerdictDrop,
			},
		}})

	return nil
}

func (n *rulesNFTables) makeNAT(id string, tap netlink.Link) error {
	if n.table == nil {
		if err := n.initTable(id); err != nil {
			return err
		}
	}

	// Add rule to masquerade connections from our box IP.
	n.nft.AddRule(&nftables.Rule{
		Table: n.table,
		Chain: n.postrouteChain,
		Exprs: []expr.Any{
			// Load meta-information 'output-interface ID' => reg 1
			&expr.Meta{
				Key:      expr.MetaKeyIIF,
				Register: 1,
			},
			// cmp eq reg 1 0x0245a8c0
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     binaryutil.NativeEndian.PutUint32(uint32(tap.Attrs().Index)),
			},
			// counter 'natted'
			&expr.Objref{
				Type: 1,
				Name: n.natCounter.Name,
			},
			// masq
			&expr.Masq{},
		},
	})

	// Add rule to count packets from our box IP.
	n.nft.AddRule(&nftables.Rule{
		Table: n.table,
		Chain: n.filterChain,
		Exprs: []expr.Any{
			// Load meta-information 'output-interface ID' => reg 1
			&expr.Meta{
				Key:      expr.MetaKeyIIF,
				Register: 1,
			},
			// cmp eq reg 1 0x0245a8c0
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     binaryutil.NativeEndian.PutUint32(uint32(tap.Attrs().Index)),
			},
			// counter 'fwded'
			&expr.Objref{
				Type: 1,
				Name: n.fwdOutCounter.Name,
			},
		},
	})

	// Add rule to count packets to our box IP.
	n.nft.AddRule(&nftables.Rule{
		Table: n.table,
		Chain: n.filterChain,
		Exprs: []expr.Any{
			// Load meta-information 'output-interface ID' => reg 1
			&expr.Meta{
				Key:      expr.MetaKeyOIF,
				Register: 1,
			},
			// cmp eq reg 1 0x0245a8c0
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     binaryutil.NativeEndian.PutUint32(uint32(tap.Attrs().Index)),
			},
			// counter 'fwded'
			&expr.Objref{
				Type: 1,
				Name: n.fwdInCounter.Name,
			},
		},
	})

	return n.nft.Flush()
}

func (n *rulesNFTables) teardown() error {
	if n.table != nil {
		n.nft.DelTable(n.table)
		return n.nft.Flush()
	}
	return nil
}
