```
nix-build --show-trace --cores 24 -A kernel
firecracker --no-api --config-file firecracker_config.json
```


nix-build -A kernel
nix-instantiate -A kernel

nix-shell -E 'with import ./default.nix; kernel.overrideAttrs (o: {nativeBuildInputs=o.nativeBuildInputs ++ [ pkgs.pkg-config pkgs.ncurses ];})'
nix repl default.nix