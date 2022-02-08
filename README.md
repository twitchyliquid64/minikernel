## Test instructions

```
sudo $(nix-build --show-trace --cores 24 -A demo-files)/launcher
```

This `demo-files` attribute is the result of this minikernel config:

```nix

    make-environment {
      name = "demo-files";
      startScript = ''#! ${pkgs.bash}/bin/bash
        echo 'hello from minikernel!'
        echo 'ifconfig output:'
        ${pkgs.inetutils}/bin/ifconfig
      '';
      extraPkgs = with pkgs;[ htop ];
    };
```


## Useful commands

```
nix-build -A kernel
nix-instantiate -A kernel

nix repl default.nix
```