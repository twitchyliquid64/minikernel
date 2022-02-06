{ lib, fetchFromGitHub, rustPlatform }:

rustPlatform.buildRustPackage rec {
  pname = "nix-9p";
  version = "unstable-20220205";

  src = fetchFromGitHub {
    owner = "twitchyliquid64";
    repo = pname;
    rev = "1b50ab575d18c554971abb32b533dec12e6db861";
    sha256 = "09jaq7mr4p7ailg1rq3ngnlwymr6q6n04pfj2hnmixsx15jnpklc";
  };

  cargoSha256 = "183rg7dzc13mwk8v0gfmxfn4hbk45fmnsv1jygm3lyndkgyf9r5g";

  meta = with lib; {
    description = "A 9p to unix socket fs server for a subset of your nix store";
    homepage = "https://github.com/twitchyliquid64/nix-9p";
    license = licenses.bsd3;
    maintainers = [ maintainers.twitchyliquid64 ];
  };
}