{
  description = "CLI for checking if there are updates available.";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";

    flake-utils.url = "github:numtide/flake-utils";

    flake-compat.url = "github:edolstra/flake-compat";
    flake-compat.flake = false;

    gomod2nix.url = "github:nix-community/gomod2nix";
    gomod2nix.inputs.nixpkgs.follows = "nixpkgs";
    gomod2nix.inputs.utils.follows = "flake-utils";
  };
  outputs = inputs @ {flake-utils, ...}:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        overlays = [
          inputs.gomod2nix.overlays.default
        ];
        pkgs = import inputs.nixpkgs {
          inherit system overlays;
        };
      in rec {
        packages = rec {
          nixcfg = pkgs.buildGoApplication {
            pname = "nixcfg";
            version = "0.0.1";
            src = ./.;
            go = pkgs.go_1_22;
            modules = ./gomod2nix.toml;
          };
          default = nixcfg;
        };

        apps = rec {
          nixcfg = {
            type = "app";
            program = "${packages.nixcfg}/bin/nixcfg";
          };
          default = nixcfg;
        };

        devShells = {
          default = pkgs.mkShell {
            packages = with pkgs; [
              ## golang
              delve
              go-outline
              go
              golangci-lint
              golangci-lint-langserver
              gopkgs
              gopls
              gotools
              ## nix
              gomod2nix
            ];

            shellHook = ''
              zsh
            '';

            # Need to disable fortify hardening because GCC is not built with -oO,
            # which means that if CGO_ENABLED=1 (which it is by default) then the golang
            # debugger fails.
            # see https://github.com/NixOS/nixpkgs/pull/12895/files
            hardeningDisable = ["fortify"];
          };
        };
      }
    );
}
