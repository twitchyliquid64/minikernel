{ lib, buildGoModule }:

  buildGoModule rec {
   name = "mk-minikernel";
   version = "unstable-20220206";

   src = ./mk-minikernel;

   vendorSha256 = "1gckkc251x0d158m9a0ylllz6ckypvl2q8s4615s1py9sn834p8r"; # lib.fakeSha256;

   buildInputs = [];

   meta = with lib; {
     description = "Launcher / setup binary for an instance of minikernel.";
     license = licenses.bsd3;
     platforms = platforms.unix;
   };
 }