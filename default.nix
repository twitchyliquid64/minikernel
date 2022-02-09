{ nixos ? import <nixpkgs/nixos> { }
, overrideKconfig ? null
}:
  let
    pkgs = nixos.pkgs;
    config = nixos.config;

    kbuildFn = pkgs.callPackage <nixpkgs/pkgs/os-specific/linux/kernel/manual-config.nix> { inherit (pkgs) buildPackages;  };
    baseKernel = pkgs.linuxKernel.packages.linux_5_16.kernel;
    manualConfig = pkgs.linuxKernel.manualConfig;

    u-root = pkgs.callPackage ./u-root.nix { };

  in rec {
    kconfig = ./kconfig;
    # nix-shell -A shell default.nix
    ## cd /tmp
    ## unpackPhase
    ## cd linux-*
    ## make oldconfig
    shell = kernel.overrideAttrs (o: {nativeBuildInputs=o.nativeBuildInputs ++ [ pkgs.pkg-config pkgs.ncurses ];});

    
    kernel = ((kbuildFn) {
      inherit (pkgs) lib stdenv;
      inherit (baseKernel) src version modDirVersion;
      configfile = if overrideKconfig != null then overrideKconfig else kconfig;
    }).overrideAttrs (oldAttrs: {
      postInstall = ''cp vmlinux arch/x86/boot/bzImage $out/'';
    });

    mk-minikernel = pkgs.callPackage ./mk-minikernel.nix { };
    initrd = u-root.cpio;
    nix-9p = pkgs.callPackage ./nix-9p.nix { };


    # Example
    demo-files = make-environment {
      name = "demo-files";
      startScript = ''#! ${pkgs.bash}/bin/bash
        echo 'hello from minikernel!'
        echo 'ifconfig output:'
        ${pkgs.inetutils}/bin/ifconfig
      '';
      extraPkgs = with pkgs;[ htop ];
    };

    # Creates a minikernel environment.
    # startScript is run by the guest once it comes up, and additionally
    # extraPkgs are available within the guest.
    make-environment = {
      name ? "",
      startScript ? "",
      extraPkgs ? [],

      cores ? 2,
      mem_mb ? 512,

      network ? "198.51.100.1/30",
      deny_subnets ? ["10.0.0.0/8" "172.16.0.0/16" "192.168.1.1/24"],
      deny_ranges ? [],
      allow_udp ? true, allow_tcp ? true, allow_icmp ? false,
    }:
    let
      bringup = pkgs.writeScriptBin (name+"-bringup") startScript;
      manifest = pkgs.writeReferencesToFile (pkgs.buildEnv {
        name = (name+"-manifest");
        paths = [ bringup ] ++ extraPkgs;
      });

      _deny_subnets = builtins.concatStringsSep " "
        (pkgs.lib.forEach deny_subnets (x: "--ip4-deny-subnet ${x}"));
      _deny_ranges = builtins.concatStringsSep " "
        (pkgs.lib.forEach deny_ranges (x: "--ip4-deny-range ${x}"));
      _proto_flags = (pkgs.lib.optionalString allow_udp "--allow_udp ") +
        (pkgs.lib.optionalString allow_tcp "--allow_tcp ") +
        (pkgs.lib.optionalString allow_icmp "--allow_icmp ");

    in
      pkgs.writeShellScriptBin name ''
        exec ${mk-minikernel}/bin/mk-minikernel \
             --nix9p-path "${nix-9p}/bin/nix-9p" \
             --kernel-path "${kernel}/vmlinux" \
             --initrd-path "${initrd}" \
             --firecracker-path "${pkgs.firecracker}/bin/firecracker" \
             --fs-manifest "${manifest}" \
             --on-bringup "${bringup}/bin/${name+"-bringup"}" \
             --id "${name}" --net "${network}" --cores "${builtins.toString cores}" --mem_mb "${builtins.toString mem_mb}" \
             ${_deny_subnets} ${_deny_ranges} ${_proto_flags}
      '';
  }
