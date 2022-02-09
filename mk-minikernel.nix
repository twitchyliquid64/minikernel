{ lib, buildGoModule }:

  buildGoModule rec {
   name = "mk-minikernel";
   version = "unstable-20220206";

   src = ./mk-minikernel;

   vendorSha256 = "0w9iyvp5sc0x6rbl1vmahfdivpnjd5cal3x74kxx3sz9j00lk0xa"; # lib.fakeSha256;

   buildInputs = [];

   meta = with lib; {
     description = "Launcher / setup binary for an instance of minikernel.";
     license = licenses.bsd3;
     platforms = platforms.unix;
   };
 }