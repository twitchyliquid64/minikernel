{ nixos ? import <nixpkgs/nixos> { }
, overrideKconfig ? null
}:
  let
    pkgs = nixos.pkgs;
    config = nixos.config;

    baseKernel = pkgs.linuxKernel.packageAliases.linux_latest.kernel;
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

    kbuildFn = pkgs.callPackage <nixpkgs/pkgs/os-specific/linux/kernel/manual-config.nix> { inherit (pkgs) buildPackages;  };
    kernel = ((kbuildFn) {
      inherit (pkgs) lib stdenv;
      inherit (baseKernel) src version modDirVersion;
      configfile = if overrideKconfig != null then overrideKconfig else kconfig;
    }).overrideAttrs (oldAttrs: {
      postInstall = ''cp vmlinux arch/x86/boot/bzImage $out/'';
    });

    initrd = u-root.cpio;
    nix-9p = pkgs.callPackage ./nix-9p.nix { };

    demo-files = pkgs.linkFarm "demo-files" [
      {name = "vmlinux"; path = "${kernel}/vmlinux"; }
      {name = "bzImage"; path = "${kernel}/bzImage"; }
      {name = "initrd"; path = initrd; }
      {name = "nix-9p"; path = "${nix-9p}/bin/nix-9p"; }
    ];
  }
