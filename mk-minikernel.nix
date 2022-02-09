{ lib, buildGoModule }:

  buildGoModule rec {
   name = "mk-minikernel";
   version = "unstable-20220206";

   src = ./mk-minikernel;

   vendorSha256 = "0rmkyn3dvsh36897hvadyyyn8dpcr1wr46xzsz8nfvqyxx5nh35v"; # lib.fakeSha256;

   buildInputs = [];

   meta = with lib; {
     description = "Launcher / setup binary for an instance of minikernel.";
     license = licenses.bsd3;
     platforms = platforms.unix;
   };
 }