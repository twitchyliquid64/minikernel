// This connects to a 9p filesystem over virtio-vsock.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/u-root/u-root/pkg/mount"
	"golang.org/x/sys/unix"
)

var (
	cid = flag.Int("cid", 2, "Host CID to connect to")
	port = flag.Int("port", 1234, "vsock port to connect to")
)

func main() {
	flag.Parse()
	if flag.Arg(0) == "" {
		fmt.Fprintln(os.Stderr, "Error: Mountpoint must be provided")
		os.Exit(1)
	}

	fd, err := nixfs_connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open connection: %v\n", err)
		os.Exit(1)
	}

	mnt, err := mount.Mount("9p", flag.Arg(0), "9p", fmt.Sprintf("trans=fd,rfdno=%d,wfdno=%d,cache=loose,ro", fd, fd), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Mount failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Mount: %+v\n", mnt)
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