	let
  conf_src = import ./config.nix;

  nixos = import <nixpkgs/nixos> { configuration = conf_src; };
  pkgs = nixos.pkgs;
  config = nixos.config;

  baseKernel = pkgs.linuxKernel.packageAliases.linux_latest.kernel;
  manualConfig = pkgs.linuxKernel.manualConfig;

in {
  # Maybe we want to simply override the kernel defconfig (ie: "tinyconfig")
  # Then use structuredExtraConfig for the rest?

  kernel = manualConfig {
      inherit (pkgs) stdenv lib;
      inherit (baseKernel) src version;

      configfile = pkgs.linuxKernel.linuxConfig {
        makeTarget = "tinyconfig";
        src = baseKernel.src;
      };
      allowImportFromDerivation = true;
  };
  pkgs = pkgs; # needed for nix-shell

  toplevel = config.system.build.toplevel;

  squashfs = pkgs.callPackage <nixpkgs/nixos/lib/make-squashfs.nix> {
    storeContents = [ config.system.build.toplevel ];
  };

  #initrd = pkgs.callPackage ./initrd { config = config; };
}
