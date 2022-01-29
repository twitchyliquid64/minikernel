nix-build -A kernel
nix-instantiate -A kernel

nix-shell -E 'with import ./default.nix; kernel.overrideAttrs (o: {nativeBuildInputs=o.nativeBuildInputs ++ [ pkgs.pkg-config pkgs.ncurses ];})'


