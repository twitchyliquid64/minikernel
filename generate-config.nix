{ 
	src, version, structuredExtraConfig

	, extraMakeFlags ? []
	, kernelArch ? stdenv.hostPlatform.linuxArch

	, lib, stdenv, buildPackages
	, go, gmp, libmpc, mpfr, bison, flex
} @ args:

stdenv.mkDerivation {
	pname = "linux-config";
	inherit src version kernelArch;

	generateConfig = ./generate-config.go;
	kernelBaseTarget = "allnoconfig";
	makeFlags = lib.optionals (stdenv.hostPlatform.linux-kernel ? makeFlags) stdenv.hostPlatform.linux-kernel.makeFlags
      ++ extraMakeFlags;

    kernelConfig = rec {
	  module = import <nixpkgs/nixos/modules/system/boot/kernel_config.nix>;
	  moduleStructuredConfig = (lib.evalModules {
	    modules = [
	      module
	      { settings = structuredExtraConfig; _file = "structuredExtraConfig"; }
	    ];
	  }).config;
	}.moduleStructuredConfig.intermediateNixConfig;
    passAsFile = [ "kernelConfig" ];

	depsBuildBuild = [ buildPackages.stdenv.cc ];
	nativeBuildInputs = [ go gmp libmpc mpfr bison flex ];

	prePatch = ''
	  # Patch kconfig to print "###" after every question so that
	  # generate-config.pl from the generic builder can answer them.
	  sed -e '/fflush(stdout);/i\printf("###");' -i scripts/kconfig/conf.c
	'';

	buildPhase = ''
	  export buildRoot="''${buildRoot:-build}"
	  export HOSTCC=$CC_FOR_BUILD
	  export HOSTCXX=$CXX_FOR_BUILD
	  export HOSTAR=$AR_FOR_BUILD
	  export HOSTLD=$LD_FOR_BUILD

	  # Get a basic config file for later refinement with $generateConfig.
	  make $makeFlags \
	      -C . O="$buildRoot" $kernelBaseTarget \
	      ARCH=$kernelArch \
	      HOSTCC=$HOSTCC HOSTCXX=$HOSTCXX HOSTAR=$HOSTAR HOSTLD=$HOSTLD \
	      CC=$CC OBJCOPY=$OBJCOPY OBJDUMP=$OBJDUMP READELF=$READELF \
	      $makeFlags

	  # Create the config file.
	  echo "generating kernel configuration..."
	  ln -s "$kernelConfigPath" "$buildRoot/kernel-config"

	  # Setup some crap so golang doesnt explode


	  DEBUG=1 ARCH=$kernelArch KERNEL_CONFIG="$buildRoot/kernel-config" \
	    BUILD_ROOT="$buildRoot" SRC=. MAKE_FLAGS="$makeFlags" \
	    HOME="$TMP/fake-home" go run $generateConfig
	'';

	installPhase = "mv $buildRoot/.config $out";

	enableParallelBuilding = true;
}