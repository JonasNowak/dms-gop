{
  description = "Dms_GOP (dms-gop): Apple Silicon macOS system monitoring CLI, TUI and REST API";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      lib = nixpkgs.lib;

      # Apple Silicon only: dms-gop targets macOS on aarch64-darwin.
      supportedSystems = [ "aarch64-darwin" ];

      # Match the Go toolchain to the `go` directive in go.mod (currently 1.27).
      goModVersion =
        let
          content = builtins.readFile ./go.mod;
          lines = builtins.filter builtins.isString (builtins.split "\n" content);
          goLines = builtins.filter (l: builtins.match "go [0-9]+\\..*" l != null) lines;
          matched =
            if goLines != [ ] then builtins.match "go ([0-9]+)\\.([0-9]+).*" (builtins.head goLines) else null;
        in
        if matched != null then
          {
            major = builtins.elemAt matched 0;
            minor = builtins.elemAt matched 1;
          }
        else
          {
            major = "1";
            minor = "27";
          };
      goForPkgs = pkgs: pkgs.${"go_${goModVersion.major}_${goModVersion.minor}"};

      pkgsFor =
        system:
        import nixpkgs {
          inherit system;
          overlays = [ self.overlays.default ];
        };

      forAllSystems = f: lib.genAttrs supportedSystems (system: f (pkgsFor system));

      mkPackage =
        pkgs:
        (pkgs.buildGoModule.override { go = goForPkgs pkgs; }) (finalAttrs: {
          pname = "dms-gop";
          version = "2.1.0";
          src = ./.;

          # Updated after go mod vendor / nix build
          vendorHash = "sha256-ajtmBCQ4OdIpjOQeiL+hj3e6PL2Np4NLnMAcA1AsW/w=";

          subPackages = [ "cmd/dms-gop" ];
          env.CGO_ENABLED = 0;

          ldflags = [
            "-s"
            "-w"
            "-X main.Version=${finalAttrs.version}"
            "-X main.buildTime=1970-01-01_00:00:00"
            "-X main.Commit=${finalAttrs.version}"
          ];

          nativeBuildInputs = with pkgs; [
            installShellFiles
          ];

          postInstall = ''
            installShellCompletion --cmd dms-gop \
              --bash <($out/bin/dms-gop completion bash) \
              --fish <($out/bin/dms-gop completion fish) \
              --zsh <($out/bin/dms-gop completion zsh)
            # Convenience alias: dgop -> dms-gop
            ln -s dms-gop $out/bin/dgop
          '';

          meta = {
            description = "Dms_GOP (dms-gop): Apple Silicon macOS system monitoring CLI, TUI and REST API";
            homepage = "https://github.com/JonasNowak/dms-gop";
            mainProgram = "dms-gop";
            license = lib.licenses.mit;
            platforms = [ "aarch64-darwin" ];
          };
        });
    in
    {
      packages = forAllSystems (pkgs: {
        dms-gop = mkPackage pkgs;
        default = self.packages.${pkgs.stdenv.hostPlatform.system}.dms-gop;
      });

      overlays.default = final: _prev: {
        dms-gop = mkPackage final;
      };

      darwinModules.default =
        { lib, pkgs, ... }:
        {
          imports = [ ./nix/darwin-module.nix ];
          programs.dms-gop.package = lib.mkDefault self.packages.${pkgs.stdenv.hostPlatform.system}.dms-gop;
        };

      homeManagerModules.default =
        { lib, pkgs, ... }:
        {
          imports = [ ./nix/home-manager-module.nix ];
          programs.dms-gop.package = lib.mkDefault self.packages.${pkgs.stdenv.hostPlatform.system}.dms-gop;
        };

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [
            pkgs.go_1_27
            pkgs.golangci-lint
            pkgs.go-mockery
          ];
        };
      });

      checks = forAllSystems (pkgs: {
        # Syntax-parse every .nix file shipped by this flake.
        nix-syntax = pkgs.runCommand "dms-gop-nix-syntax" { } ''
          for f in ${./flake.nix} ${./nix/darwin-module.nix} ${./nix/home-manager-module.nix}; do
            ${pkgs.nix}/bin/nix-instantiate --parse "$f" > /dev/null
          done
          touch "$out"
        '';
        # Evaluate both modules (options + config) using minimal stub
        # declarations for the nix-darwin / Home Manager namespaces.
        module-options =
          let
            evalHM = extra: lib.evalModules {
              modules = [
                ./nix/home-manager-module.nix
                ({ lib, ... }: {
                  options = {
                    home.packages = lib.mkOption { type = lib.types.listOf lib.types.package; default = []; };
                  };
                })
                ({ ... }: {
                  programs.dms-gop = {
                    enable = true;
                    package = self.packages.${pkgs.stdenv.hostPlatform.system}.dms-gop;
                  } // extra;
                })
              ];
              specialArgs = { inherit pkgs; };
            };
            hmEmpty = evalHM { };
            darwinEval = lib.evalModules {
              modules = [
                ./nix/darwin-module.nix
                ({ lib, ... }: {
                  options = {
                    system.primaryUser = lib.mkOption { type = lib.types.nullOr lib.types.str; default = "tester"; };
                    environment.systemPackages = lib.mkOption { type = lib.types.listOf lib.types.package; default = []; };
                    launchd.user.agents = lib.mkOption { type = lib.types.attrsOf lib.types.anything; default = {}; };
                  };
                  config.programs.dms-gop = {
                    enable = true;
                    server.enable = true;
                    package = self.packages.${pkgs.stdenv.hostPlatform.system}.dms-gop;
                  };
                })
              ];
              specialArgs = { inherit pkgs; };
            };
          in
          pkgs.runCommand "dms-gop-module-options" { } ''
            test -n "${builtins.toJSON (builtins.attrNames hmEmpty.options.programs.dms-gop)}"
            test -n "${builtins.toJSON hmEmpty.config.home.packages}"
            test -n "${darwinEval.config.launchd.user.agents.dms-gop.serviceConfig.EnvironmentVariables.API_PORT}"
            test -n "${darwinEval.config.launchd.user.agents.dms-gop.serviceConfig.EnvironmentVariables.PATH}"
            touch "$out"
          '';
      });
    };
}
