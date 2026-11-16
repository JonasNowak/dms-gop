{ config, lib, pkgs, ... }:

let
  cfg = config.programs.dms-gop;

  primaryUser = config.system.primaryUser or "";
  homeDir = "/Users/${primaryUser}";

  package =
    if cfg.package != null then cfg.package else throw "programs.dms-gop.package must be set (the flake module sets it by default)";

  agentPath = lib.concatStringsSep ":" (
    lib.optionals (primaryUser != "") [
      "${homeDir}/.nix-profile/bin"
      "/etc/profiles/per-user/${primaryUser}/bin"
    ]
    ++ [
      "/run/current-system/sw/bin"
      "/nix/var/nix/profiles/default/bin"
      "/opt/homebrew/bin"
      "/usr/local/bin"
      "/usr/bin"
      "/bin"
      "/usr/sbin"
      "/sbin"
    ]
  );
in

{
  options.programs.dms-gop = {
    enable = lib.mkEnableOption "Dms_GOP (dms-gop)";

    package = lib.mkOption {
      type = lib.types.nullOr lib.types.package;
      default = null;
      description = "The dms-gop package to install. Set by the flake's darwinModules.default; override to pin a different build.";
    };

    server = {
      enable = lib.mkEnableOption "the dms-gop REST API LaunchAgent";

      port = lib.mkOption {
        type = lib.types.str;
        default = ":63484";
        description = "Listen address for `dms-gop server`, passed to the agent via the API_PORT environment variable.";
      };
    };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ package ];

    # Opt-in LaunchAgent running the REST API server (`dms-gop server`).
    launchd.user.agents.dms-gop = lib.mkIf cfg.server.enable {
      serviceConfig = {
        ProgramArguments = [
          "${package}/bin/dms-gop"
          "server"
        ];
        EnvironmentVariables = {
          API_PORT = cfg.server.port;
          PATH = agentPath;
        };
        RunAtLoad = true;
        KeepAlive = true;
        ProcessType = "Background";
        StandardOutPath = "${homeDir}/Library/Logs/dms-gop.log";
        StandardErrorPath = "${homeDir}/Library/Logs/dms-gop.err.log";
      };
    };
  };
}
