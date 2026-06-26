{
  description = "A safe rm wrapper with trash support";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        packages.default = import ./nix/package.nix { inherit pkgs; };
        devShells.default = import ./nix/shell.nix { inherit pkgs; };
      }
    ) // {
      nixosModules.default = import ./nix/module.nix;
      homeManagerModules.default = import ./nix/module.nix;
    };
}
