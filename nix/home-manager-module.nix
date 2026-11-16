{ config, lib, pkgs, ... }:

let
  cfg = config.programs.dms-gop;

  package =
    if cfg.package != null then cfg.package else throw "programs.dms-gop.package must be set (the flake module sets it by default)";
in

{
  options.programs.dms-gop = {
    enable = lib.mkEnableOption "Dms_GOP (dms-gop)";

    package = lib.mkOption {
      type = lib.types.nullOr lib.types.package;
      default = null;
      description = "The dms-gop package to install into home.packages. Set by the flake's homeManagerModules.default.";
    };
  };

  config = lib.mkIf cfg.enable {
    home.packages = [ package ];
  };
}
