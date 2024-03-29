{ stdenv, lib, fetchFromGitHub, buildGoPackage, fetchgit, fetchhg, fetchbzr, fetchsvn, pkgs, config }:
let
  u-root-source = fetchFromGitHub {
    owner = "u-root";
    repo = "u-root";
    rev = "4dad982f78a72202985296afdfbc47c274ccc944";
    sha256 = "15d75nyjl3hy77a0wa8plgs0mbhvpvz846rm7k01586s1ci2s512"; # use nix-prefetch-git
  };
  u-root-binary = buildGoPackage rec {
   name = "u-root";
   version = u-root-source.rev;
   goPackagePath = "github.com/u-root/u-root";

   buildPhase = ''
   runHook preBuild
   go build -o "$TMPDIR/u-root" -v github.com/u-root/u-root
   runHook postBuild
   '';

   installPhase = ''
   runHook preInstall

   mkdir -p $out
   cp -rv "$TMPDIR/u-root" $out

   runHook postInstall
   '';

   src = u-root-source;

   meta = with lib; {
     description = "A fully Go userland! u-root can create a root file system (initramfs) containing a busybox-like set of tools written in Go.";
     license = licenses.bsd3;
     homepage = https://u-root.tk;
     platforms = platforms.all;
   };
 };
in
  rec {
    source = u-root-source;
    binary = u-root-binary;

    mk-init = ./mk-init.go;
    bringup = pkgs.writeScriptBin "bringup.sh" ''#!/bbin/elvish
      /bbin/mkdir -p /nix/store
      exec /bbin/mkinit /nix/store
      '';

    cpio = stdenv.mkDerivation {
      name = "uroot-cpio";
      buildInputs = [ ];
      nativeBuildInputs = [ pkgs.go pkgs.coreutils binary ];
      src = source;

      # ${pkgs.coreutils}/bin/install -C -m 775 "${./usetup}"/* $TMPDIR/go/src/usetup
      installPhase = ''
        dir=$(pwd)
        mkdir -p $TMPDIR/go/src/{usetup,github.com/u-root}

        mv $dir $TMPDIR/go/src/github.com/u-root/u-root

        mkdir mkinit
        cat ${mk-init} >mkinit/mk-init.go

        export GOCACHE=$TMPDIR/go-cache
        export GOPATH="$TMPDIR/go"
        export GOSUMDB=off
        export GOPROXY=off
        export GO111MODULE=off
        export GOROOT=${pkgs.go}/share/go

        ${binary}/u-root -o "$out" \
          -base=/dev/null \
          -files "${bringup}/bin/bringup.sh:bringup" \
          -initcmd '/bringup' \
          core boot ./mkinit
      '';
    };
  }
