let
  totpKey = "G3HECQ37UWKHEIKH56OYRXCTXPYK3EXHNAYLDXP6667T2QLDZ3BSVJH52DVJW5NKEDLTCVOLSOAD6S763MWGL2WV7QUULZRKCHKHWKA=";
in
{
  nodes.machine = { pkgs, ... }: {
    services.spasm = {
      enable = true;
      configFile = pkgs.writeText "spasm-config" ''
        # 'mypassword'
        SPASM_PASS_HASH='$2b$05$2gDWO9d4.B96asSaSHqiuu6w4fpR5jaVqVfv5TfANKjYAw/ihkcZ6'
        SPASM_TOTP_KEY='${totpKey}'
        SPASM_ADDRESS='localhost:1234'
      '';
    };
    environment.systemPackages = [ pkgs.oath-toolkit ];
  };
  testScript = ''
    machine.wait_for_unit("default.target")
    pin = machine.succeed("oathtool --totp=SHA512 -b -s 10s ${totpKey}").strip()
    assert "303 See Other"  in machine.succeed("2>&1 curl -v localhost:1234/check")
    assert "200 OK"         in machine.succeed("2>&1 curl -v localhost:1234/login")
    assert "403 Forbidden"  in machine.succeed("2>&1 curl -v localhost:1234/login -d 'pass=wrong&totp=123456'")
    assert "303 See Other"  in machine.succeed("2>&1 curl -v localhost:1234/login -d 'pass=mypassword&totp='" + pin)
  '';
}
