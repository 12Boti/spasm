self:
{ lib, config, pkgs, ... }:
let
  cfg = config.services.spasm;
in
{
  options.services.spasm = {
    enable = lib.mkEnableOption "spasm";
    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.system}.default;
    };
    configFile = lib.mkOption {
      type = lib.types.path;
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.spasm = {
      wantedBy = [ "multi-user.target" ];
      serviceConfig = {
        ExecStart = lib.getExe cfg.package;
        EnvironmentFile = cfg.configFile;

        # security
        DynamicUser = true;
        PrivateDevices = true;
        PrivateIPC = true;
        PrivateUsers = true;
        ProtectClock = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectKernelLogs = true;
        ProtectControlGroups = true;
        ProtectProc = "invisible";
        RestrictAddressFamilies = "AF_INET AF_INET6";
        ProtectHome = true;
        RestrictNamespaces = true;
        RestrictRealtime = true;
        MemoryDenyWriteExecute = true;
        LockPersonality = true;
        CapabilityBoundingSet = "";
        ReadOnlyPaths = "/";
        SystemCallArchitectures = "native";
        ProtectHostname = true;
        ProcSubset = "pid";
        SystemCallFilter = "@basic-io @file-system @signal @process @io-event @network-io @ipc setrlimit madvise uname";
      };
    };
  };
}
