{ nixos ? import <nixpkgs/nixos> { }
, overrideKconfig ? null
}:
  let
    pkgs = nixos.pkgs;
    config = nixos.config;

    kbuildFn = pkgs.callPackage <nixpkgs/pkgs/os-specific/linux/kernel/manual-config.nix> { inherit (pkgs) buildPackages;  };
    baseKernel = pkgs.linuxKernel.packages.linux_6_2.kernel;
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
        ${pkgs.bash}/bin/bash
      '';
      extraPkgs = with pkgs;[ htop ];
    };

    # Creates a minikernel environment.
    # startScript is run by the guest once it comes up, and additionally
    # extraPkgs are available within the guest.
    make-environment = {
      name,
      startScript ? "",
      extraPkgs ? [],

      cores ? 2,
      mem_mb ? 512,

      network ? "198.51.100.1/30",
      deny_subnets ? ["10.0.0.0/8" "172.16.0.0/16" "192.168.1.1/24"],
      deny_ranges ? [],
      allow_udp ? true, allow_tcp ? true, allow_icmp ? true,

      allow_ips ? [],
      allow_subnets ? [],
      allow_ranges ? [],

      unsafe_firecracker_overrides ? null,
    }:
    let
      bringup = pkgs.writeScriptBin (name+"-bringup") startScript;
      manifest = pkgs.writeReferencesToFile (pkgs.buildEnv {
        name = (name+"-manifest");
        paths = [ bringup ] ++ extraPkgs;
      });

      _fc_overrides = if unsafe_firecracker_overrides != null then
        "--unsafe_firecracker_overrides " + (
          pkgs.lib.escapeShellArg (builtins.toJSON unsafe_firecracker_overrides)
        ) else "";

      _allow_ips = builtins.concatStringsSep " "
        (pkgs.lib.forEach allow_ips (x: "--ip4-allow-addr ${x}"));
      _allow_subnets = builtins.concatStringsSep " "
        (pkgs.lib.forEach allow_subnets (x: "--ip4-allow-subnet ${x}"));
      _allow_ranges = builtins.concatStringsSep " "
        (pkgs.lib.forEach allow_ranges (x: "--ip4-allow-range ${x}"));
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
             ${_deny_subnets} ${_deny_ranges} ${_proto_flags} ${_allow_ips} ${_allow_subnets} ${_allow_ranges} \
             ${_fc_overrides}
      '';

# Creates a minikernel using the provided nixos closure.
make-nixos-environment = {
      nixos,
      name,

      cores ? 2,
      mem_mb ? 512,

      network ? "198.51.100.1/30",
      deny_subnets ? ["10.0.0.0/8" "172.16.0.0/16" "192.168.1.1/24"],
      deny_ranges ? [],
      allow_udp ? true, allow_tcp ? true, allow_icmp ? true,

      allow_ips ? [],
      allow_subnets ? [],
      allow_ranges ? [],

      unsafe_firecracker_overrides ? null,
  }:
  make-environment {
      inherit name cores mem_mb;
      inherit network deny_subnets deny_ranges allow_udp allow_tcp allow_icmp;
      inherit allow_ips allow_subnets allow_ranges;
      inherit unsafe_firecracker_overrides;

      startScript = ''#! ${nixos.pkgs.bash}/bin/bash
      set -e
        
      ${pkgs.coreutils}/bin/rm /init

      # Make the root filesystem look mostly like
      # the toplevel environment.
      for p in ${nixos.config.system.build.toplevel}/*; do
        base=$(${nixos.pkgs.coreutils}/bin/basename $p)
        dest=$(${nixos.pkgs.coreutils}/bin/readlink $p || true)

        if [[ $base == "etc" || $base == "bin" ]]; then
          continue
        fi

        if [[ $dest == "" ]]; then
          ${nixos.pkgs.coreutils}/bin/ln -sv $p "/$base"
        else
          ${nixos.pkgs.coreutils}/bin/ln -sv $dest "/$base"
        fi
      done

      exec ${nixos.config.system.build.toplevel}/init
      '';
      extraPkgs = [ nixos.config.system.build.toplevel ];
    };
  }
