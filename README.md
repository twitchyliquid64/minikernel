# minikernel

Run nix programs in an isolated, mini-VM. Great as a containment primitive.

 * Minimal, hardened kernel
 * Firecracker is used to make a strong security boundary
 * Only a read-only subset of the nix store is exposed
 * Startup time ~2 seconds (4 secs on my 2011 thinkpad)

## Technicals

 * Readonly nix store exposed over vsockets using 9p
 * Firecracker + TAP interface provides networking
 * Host-side nftables filtering/masquerade rules constrain networking
 * Custom, u-root based initrd means very simple early userspace with minimal attack surface

## Example

There are two main interfaces for defining minikernels: `make-environment`, and `make-nixos-environment`.


#### make-environment

Use `make-environment` when you need some programs from nixpkgs, but otherwise dont care about having
a running nixos system with systemd and all that jazz.

```
sudo $(nix-build --show-trace --cores 24 -A demo-files)/bin/demo-files
```

This `demo-files` attribute is the result of this minikernel config:

```nix

    make-environment {
      name = "demo-files";
      startScript = ''#! ${pkgs.bash}/bin/bash
        echo 'hello from minikernel!'
        echo 'ifconfig output:'
        ${pkgs.inetutils}/bin/ifconfig
      '';
      extraPkgs = with pkgs;[ htop ];
    }
```

#### make-nixos-environment

Use `make-nixos-environment` when you want to boot nixos in your minikernel and configure your
minikernel as if it was a nixos system.
Due to the use of virtualization and a minimal kernel not all features will work.

See `nixos-example/` for an example of this.