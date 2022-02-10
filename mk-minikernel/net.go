package main

import (
	"fmt"
	"io/ioutil"
	"net"

	"github.com/google/nftables"
	"github.com/vishvananda/netlink"
)

type netCtlr struct {
	up       bool
	link     netlink.Link
	hostAddr net.IPNet

	nft *rulesNFTables
}

func (nc *netCtlr) Close() error {
	if err := nc.nft.teardown(); err != nil {
		return fmt.Errorf("nft: %v", err)
	}
	return netlink.LinkDel(nc.link)
}

func (nc *netCtlr) IFName() string {
	return nc.link.Attrs().Name
}

func (nc *netCtlr) guestAddr() net.IPNet {
	ip := make(net.IP, len(nc.hostAddr.IP))
	copy(ip, nc.hostAddr.IP)
	ip[len(ip)-1]++
	return net.IPNet{
		IP:   ip,
		Mask: nc.hostAddr.Mask,
	}
}

func (nc *netCtlr) GuestAddr() string {
	g := nc.guestAddr()
	return (&g).String()
}

func (nc *netCtlr) computeFirewall() (firewall, error) {
	denySet, allowSet, err := computeIPLists()
	if err != nil {
		return firewall{}, fmt.Errorf("computing denylist: %v", err)
	}

	return firewall{
		allowUDP:  *allowUDP,
		allowTCP:  *allowTCP,
		allowICMP: *allowICMP,
		deny:      *denySet,
		allow:     *allowSet,
	}, nil
}

func (nc *netCtlr) BringUp() error {
	if err := ipv4EnableForwarding(true); err != nil {
		return err
	}

	fw, err := nc.computeFirewall()
	if err != nil {
		return fmt.Errorf("computing firewall: %v", err)
	}
	if err := nc.nft.makeFirewall(*id, nc.link, fw); err != nil {
		return err
	}

	if err := nc.nft.makeNAT(*id, nc.link); err != nil {
		return fmt.Errorf("nat setup: %v", err)
	}
	if err := netlink.LinkSetUp(nc.link); err != nil {
		return err
	}
	nc.up = true
	return nil
}

func newNet(wd, hostAddrSpec string) (*netCtlr, error) {
	hostAddr, err := netlink.ParseIPNet(hostAddrSpec)
	if err != nil {
		return nil, err
	}

	link := &netlink.Tuntap{
		LinkAttrs: netlink.LinkAttrs{Name: "tap-" + *id},
		Mode:      netlink.TUNTAP_MODE_TAP,
	}
	if err := netlink.LinkAdd(link); err != nil {
		return nil, err
	}
	ift, err := netlink.LinkByName(link.LinkAttrs.Name)
	if err != nil {
		return nil, err
	}
	if err := netlink.AddrAdd(ift, &netlink.Addr{IPNet: hostAddr}); err != nil {
		return nil, err
	}

	return &netCtlr{
		link:     ift,
		hostAddr: *hostAddr,
		nft: &rulesNFTables{
			nft: &nftables.Conn{},
		},
	}, nil
}

// ipv4EnableForwarding enables or disables forwarding of IPv4 packets.
func ipv4EnableForwarding(state bool) error {
	outData := "0"
	if state {
		outData = "1"
	}
	return ioutil.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte(outData), 0644)
}
