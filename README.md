## Test instructions

```
nix-build --show-trace --cores 24 -A demo-files
sudo ip tuntap add mode tap test0 || true
firecracker --no-api --config-file firecracker_config.json
```

## Useful commands

```
nix-build -A kernel
nix-instantiate -A kernel

nix repl default.nix
```