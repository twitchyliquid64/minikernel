package main

import (
	"fmt"
	"os/exec"
	"path"
)

const fsSockName = "fs-vsock"

type fsDaemon struct {
	proc *exec.Cmd
}

func (d *fsDaemon) Close() error {
	return d.proc.Process.Kill()
}

func setupFS(wd string) (*fsDaemon, error) {
	d := fsDaemon{
		proc: exec.Command(*nix9pPath, path.Join(wd, fsSockName+"_1234"), "/nix/store"),
	}
	if *fsManifest != "" {
		d.proc.Args = append(d.proc.Args, *fsManifest)
	}

	if err := d.proc.Start(); err != nil {
		return nil, fmt.Errorf("start failed: %v", err)
	}

	return &d, nil
}
