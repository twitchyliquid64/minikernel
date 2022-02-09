package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"inet.af/netaddr"
)

type arrayStringFlag []string

func (i *arrayStringFlag) String() string {
	return strings.Join([]string(*i), ",")
}

func (i *arrayStringFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	nix9pPath       = flag.String("nix9p-path", "result/nix-9p", "Path to the nix-9p binary.")
	firecrackerPath = flag.String("firecracker-path", "firecracker", "Path to the firecracker binary.")
	initrdPath      = flag.String("initrd-path", "result/initrd", "Path to the initrd.")
	kernelPath      = flag.String("kernel-path", "result/vmlinux", "Path to the kernel image.")

	id         = flag.String("id", "default", "The unique identifier of this instance.")
	addr       = flag.String("net", "198.51.100.1/30", "The network to use.")
	fsManifest = flag.String("fs-manifest", "", "The whitelisted set of nix store paths available in the minikernel.")
	onBringup  = flag.String("on-bringup", "", "Store path to execute once the microVM comes up.")

	numCores = flag.Int("cores", 2, "Number of cores the microVM should have.")
	numMem   = flag.Int("mem_mb", 512, "Amount of memory the microVM should have in megabytes.")

	allowUDP    = flag.Bool("allow_udp", false, "Whether to permit UDP traffic.")
	allowTCP    = flag.Bool("allow_tcp", false, "Whether to permit TCP traffic.")
	allowICMP   = flag.Bool("allow_icmp", false, "Whether to permit ICMP traffic.")
	denySubnets arrayStringFlag
	denyRanges  arrayStringFlag
)

func computeDenylist(b *netaddr.IPSetBuilder) (*netaddr.IPSet, error) {
	for _, s := range denySubnets {
		p, err := netaddr.ParseIPPrefix(s)
		if err != nil {
			return nil, fmt.Errorf("parsing %q: %v", s, err)
		}
		b.AddPrefix(p)
	}
	for _, r := range denyRanges {
		spl := strings.Split(r, "-")
		if len(spl) < 2 {
			return nil, fmt.Errorf("parsing %q: %s", r, "expecting '-' separated ip range")
		}
		f, err := netaddr.ParseIP(spl[0])
		if err != nil {
			return nil, fmt.Errorf("parsing %q from-address: %v", r, err)
		}
		t, err := netaddr.ParseIP(spl[1])
		if err != nil {
			return nil, fmt.Errorf("parsing %q to-address: %v", r, err)
		}
		b.AddRange(netaddr.IPRangeFrom(f, t))
	}
	return b.IPSet()
}

func main() {
	flag.Var(&denySubnets, "ip4-deny-subnet", "IP networks which should not be externally reachable.")
	flag.Var(&denyRanges, "ip4-deny-range", "IP address ranges which should not be externally reachable.")
	flag.Parse()

	mk, err := newMinikernel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := mk.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Shutdown failed: %v\n", err)
		}
	}()

	if err := mk.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start: %v\n", err)
		return
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

type mk struct {
	started bool
	wd      string

	nc *netCtlr
	fs *fsDaemon

	vmm *exec.Cmd
}

func (mk *mk) Close() error {
	if mk.started {
		if err := mk.vmm.Process.Kill(); err != nil {
			return err
		}
	}
	if err := mk.fs.Close(); err != nil {
		return err
	}
	if err := mk.nc.Close(); err != nil {
		return err
	}
	return os.RemoveAll(mk.wd)
}

func (mk *mk) Start() error {
	if err := mk.nc.BringUp(); err != nil {
		return fmt.Errorf("net bringup: %v", err)
	}
	if err := mk.vmm.Start(); err != nil {
		return fmt.Errorf("vmm launch: %v", err)
	}
	mk.started = true
	return nil
}

func newMinikernel() (*mk, error) {
	wd, err := ioutil.TempDir("", "minikernel-"+*id)
	if err != nil {
		return nil, err
	}

	nc, err := newNet(wd, *addr)
	if err != nil {
		os.RemoveAll(wd)
		return nil, fmt.Errorf("net: %v", err)
	}

	fs, err := setupFS(wd)
	if err != nil {
		nc.Close()
		os.RemoveAll(wd)
		return nil, fmt.Errorf("fs: %v", err)
	}

	if err := writeFirecrackerConfig(wd, nc); err != nil {
		nc.Close()
		fs.Close()
		os.RemoveAll(wd)
		return nil, fmt.Errorf("vmm conf: %v", err)
	}

	vmm := exec.Command(*firecrackerPath, "--no-api", "--config-file", path.Join(wd, "firecracker_config.json"))
	vmm.Stdin = os.Stdin
	vmm.Stdout = os.Stdout
	vmm.Stderr = os.Stderr

	return &mk{
		wd:  wd,
		fs:  fs,
		nc:  nc,
		vmm: vmm,
	}, nil
}

func writeFirecrackerConfig(wd string, nc *netCtlr) error {
	cf, err := os.OpenFile(path.Join(wd, "firecracker_config.json"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("conf: %v", err)
	}

	if err := json.NewEncoder(cf).Encode(map[string]interface{}{
		"boot-source": map[string]interface{}{
			"kernel_image_path": *kernelPath,
			"boot_args": fmt.Sprintf("console=ttyS0 reboot=k panic=1 i8042.noaux mk-init.IP=%s mk-init.defaultRoute=%s mk-init.bringup=%s",
				nc.GuestAddr(), nc.hostAddr.IP.String(), *onBringup),
			"initrd_path": *initrdPath,
		},
		"drives": []string{},
		"machine-config": map[string]interface{}{
			"vcpu_count":        *numCores,
			"mem_size_mib":      *numMem,
			"track_dirty_pages": false,
		},
		"network-interfaces": []map[string]interface{}{
			{
				"iface_id":        "1",
				"host_dev_name":   nc.IFName(),
				"guest_mac":       "06:00:c0:a8:00:02",
				"rx_rate_limiter": nil,
				"tx_rate_limiter": nil,
			},
		},
		"vsock": map[string]interface{}{
			"vsock_id":  "fs",
			"uds_path":  path.Join(wd, fsSockName),
			"guest_cid": 3,
		},
	}); err != nil {
		return fmt.Errorf("conf write: %v", err)
	}

	return cf.Close()
}
