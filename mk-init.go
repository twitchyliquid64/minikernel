// mk-init does guest-side filesystem & networking bringup.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/u-root/u-root/pkg/libinit"
	"github.com/u-root/u-root/pkg/mount"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func init() {
	runtime.LockOSThread()
}

var (
	cid  = flag.Int("cid", 2, "Host CID to connect to")
	port = flag.Int("port", 1234, "vsock port to connect to")
)

func main() {
	flag.Parse()
	if flag.Arg(0) == "" {
		fmt.Fprintln(os.Stderr, "Error: Mountpoint must be provided")
		os.Exit(1)
	}

	// Brought across from the default init program in u-root.
	libinit.SetEnv()
	libinit.CreateRootfs()
	libinit.NetInit()

	// Mount the nix store
	fd, err := nixfs_connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open connection: %v\n", err)
		os.Exit(2)
	}
	if _, err = mount.Mount("9p", flag.Arg(0), "9p", fmt.Sprintf("version=9p2000.L,trans=fd,rfdno=%d,wfdno=%d,cache=loose,ro", fd, fd), 0); err != nil {
		fmt.Fprintf(os.Stderr, "Mount failed: %v\n", err)
		os.Exit(3)
	}

	// Parse out instructions for networking and next stage
	var (
		netAddr      string
		defaultRoute string
		bringup      string
	)
	cmdline, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed reading commandline: %v\n", err)
		os.Exit(4)
	}
	spl := strings.Split(strings.TrimSpace(string(cmdline)), " ")
	for _, opt := range spl {
		idx := strings.Index(opt, "=")
		if idx > 0 {
			switch opt[:idx] {
			case "mk-init.IP":
				netAddr = opt[idx+1:]
			case "mk-init.defaultRoute":
				defaultRoute = opt[idx+1:]
			case "mk-init.bringup":
				bringup = opt[idx+1:]
			}
		}
	}

	// Perform actions
	if netAddr != "" {
		if err := setNetAddr(netAddr); err != nil {
			fmt.Fprintf(os.Stderr, "Failed setting network: %v\n", err)
			os.Exit(5)
		}
	}
	if defaultRoute != "" {
		if err := setDefaultRoute(defaultRoute); err != nil {
			fmt.Fprintf(os.Stderr, "Failed setting default route: %v\n", err)
			os.Exit(6)
		}
	}
	if bringup != "" {
		unix.Exec(bringup, []string{bringup}, os.Environ())
	}
}

func setDefaultRoute(gwAddr string) error {
	iface, err := netlink.LinkByName("eth0")
	if err != nil {
		return err
	}

	gw := net.ParseIP(gwAddr)
	_, defNet, _ := net.ParseCIDR("0.0.0.0/0")
	err = netlink.RouteAdd(&netlink.Route{
		LinkIndex: iface.Attrs().Index,
		Scope:     netlink.SCOPE_UNIVERSE,
		Dst:       defNet,
		Gw:        gw,
	})
	return err
}

func setNetAddr(addr string) error {
	iface, err := netlink.LinkByName("eth0")
	if err != nil {
		return err
	}

	a, err := netlink.ParseAddr(addr)
	if err != nil {
		return fmt.Errorf("parsing address: %v", err)
	}
	if err := netlink.AddrAdd(iface, a); err != nil {
		return fmt.Errorf("setting addr: %v", err)
	}
	return netlink.LinkSetUp(iface)
}

func nixfs_connect() (int, error) {
	socket, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM, 0)
	if err != nil {
		return 0, fmt.Errorf("socket: %v", err)
	}

	sockaddr := &unix.SockaddrVM{
		CID:  uint32(*cid),
		Port: uint32(*port),
	}
	if err := unix.Connect(socket, sockaddr); err != nil {
		return 0, fmt.Errorf("connect: %v", err)
	}

	return socket, nil
}
