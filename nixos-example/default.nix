let
  nixos = import <nixpkgs/nixos> { configuration = import ./config.nix; };
  minikernel = import ../default.nix { nixos = nixos; };
in

  # You should be able to VNC to 198.51.100.2:8080 once the VM comes up.
  # Press O+Enter for a shell.
  minikernel.make-nixos-environment {
    name = "nixos-example";
    inherit nixos;
  }