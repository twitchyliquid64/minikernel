{ lib, buildGoModule }:

  buildGoModule rec {
   name = "mk-minikernel";
   version = "unstable-20220206";

   src = ./mk-minikernel;

   vendorSha256 = "sha256:051yrm6nfir6p57psfw2k25l4brzq874908sflhx70l0frb0m55x"; # lib.fakeSha256;

   buildInputs = [];

   meta = with lib; {
     description = "Launcher / setup binary for an instance of minikernel.";
     license = licenses.bsd3;
     platforms = platforms.unix;
   };
 }