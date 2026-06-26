{ config, lib, pkgs, ... }:

with lib;

let
  cfg = config.programs.safe-rm;
  tomlValue = v:
    if isString v then ''"${v}"''
    else if isBool v then (if v then "true" else "false")
    else if isList v then "[${concatMapStringsSep ", " tomlValue v}]"
    else "null";
  configFile = pkgs.writeText "safe-rm-config" ''
    ${optionalString (cfg.trashDir != null) "trash_dir = ${tomlValue cfg.trashDir}"}
    ${optionalString (cfg.bypassList != []) "bypass_list = ${tomlValue cfg.bypassList}"}
    ${optionalString (cfg.dangerList != []) "danger_list = ${tomlValue cfg.dangerList}"}
  '';
in {
  options.programs.safe-rm = {
    enable = mkEnableOption "safe-rm";

    trashDir = mkOption {
      type = types.nullOr types.str;
      default = null;
      description = "Custom trash directory";
    };

    bypassList = mkOption {
      type = types.listOf types.str;
      default = [];
      description = "Patterns that bypass trash (permanent delete)";
    };

    dangerList = mkOption {
      type = types.listOf types.str;
      default = [];
      description = "Patterns that require TUI confirmation";
    };
  };

  config = mkIf cfg.enable {
    home.packages = [ pkgs.safe-rm ];
    home.file.".config/safe-rm/config.toml".source = configFile;
  };
}
