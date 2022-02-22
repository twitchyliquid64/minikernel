let
  nixos = import <nixpkgs/nixos> { configuration = import ./config.nix; };
  minikernel = import ../default.nix { nixos = nixos; };
in

  minikernel.make-nixos-environment {
    name = "nixos-example";
    inherit nixos;
  }