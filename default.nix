{ nixos ? import <nixpkgs/nixos> { }
, overrideKconfig ? null
}:
  let
    pkgs = nixos.pkgs;
    config = nixos.config;

    baseKernel = pkgs.linuxKernel.packageAliases.linux_latest.kernel;
    manualConfig = pkgs.linuxKernel.manualConfig;

  in rec {
    kconfig = pkgs.callPackage ./generate-config.nix {
      inherit (baseKernel) src version;

      structuredExtraConfig = import ./kernel_config.nix { inherit pkgs; };
    };

    kbuildFn = pkgs.callPackage <nixpkgs/pkgs/os-specific/linux/kernel/manual-config.nix> { inherit (pkgs) buildPackages;  };
    kernel = ((kbuildFn) {
      inherit (pkgs) lib stdenv;
      inherit (baseKernel) src version modDirVersion;
      configfile = if overrideKconfig != null then overrideKconfig else kconfig;
    }).overrideAttrs (oldAttrs: {
      postInstall = ''cp vmlinux $out/'';
    });


    toplevel = config.system.build.toplevel;
    squashfs = pkgs.callPackage <nixpkgs/nixos/lib/make-squashfs.nix> {
      storeContents = [ config.system.build.toplevel ];
    };

    #initrd = pkgs.callPackage ./initrd { config = config; };
  }
