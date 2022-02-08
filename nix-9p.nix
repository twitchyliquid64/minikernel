{ lib, fetchFromGitHub, rustPlatform }:

rustPlatform.buildRustPackage rec {
  pname = "nix-9p";
  version = "unstable-20220205";

  src = fetchFromGitHub {
    owner = "twitchyliquid64";
    repo = pname;
    rev = "89ef863d9c4ae847d730914c66786029b2b6c45f";
    sha256 = "04b6idfmi5rxn6b67x5n1k9560spvj627fz2h53v0fw57frypzyv";
  };

  cargoSha256 = "09bab6gd95dhy232lq01za4kp5bwbz9j4isxiyw73y7yfhcpriyv";

  meta = with lib; {
    description = "A 9p to unix socket fs server for a subset of your nix store";
    homepage = "https://github.com/twitchyliquid64/nix-9p";
    license = licenses.bsd3;
    maintainers = [ maintainers.twitchyliquid64 ];
  };
}
