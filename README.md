## Test instructions

```
sudo $(nix-build --show-trace --cores 24 -A demo-files)/mk-minikernel
```

## Useful commands

```
nix-build -A kernel
nix-instantiate -A kernel

nix repl default.nix
```