{ pkgs, ... }:
let
  inherit (pkgs) lib buildGoModule fetchFromGitHub;
in

buildGoModule {
  pname = "safe-rm";
  version = "0.1.0";

  src = ./..;

  vendorHash = lib.fakeHash;

  meta = {
    description = "A safe rm wrapper with trash support";
    license = lib.licenses.mit;
    mainProgram = "safe-rm";
  };
}
