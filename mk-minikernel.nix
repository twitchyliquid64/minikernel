{ lib, buildGoModule }:

  buildGoModule rec {
   name = "mk-minikernel";
   version = "unstable-20220206";

   src = ./mk-minikernel;

   vendorSha256 = "sha256-Zd65e6jQ0T/LpibB/B+6Ltf571C6MLVkZtRVKydTHPg="; # lib.fakeSha256;

   buildInputs = [];

   meta = with lib; {
     description = "Launcher / setup binary for an instance of minikernel.";
     license = licenses.bsd3;
     platforms = platforms.unix;
   };
 }