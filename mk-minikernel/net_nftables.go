package main

import (
	"net"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
)

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

func (n *rulesNFTables) makeNAT(id string, guestAddr net.IP) error {
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
			// payload load 4b @ network header + 12 => reg 1
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       12,
				Len:          4,
			},
			// cmp eq reg 1 0x0245a8c0
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     guestAddr.To4(),
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
			// payload load 4b @ network header + 12 => reg 1
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       12,
				Len:          4,
			},
			// cmp eq reg 1 0x0245a8c0
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     guestAddr.To4(),
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
			// payload load 4b @ network header + 16 => reg 1
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       16,
				Len:          4,
			},
			// cmp eq reg 1 0x0245a8c0
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     guestAddr.To4(),
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
