{
  outputs = { self, nixpkgs }: {
    packages = nixpkgs.lib.genAttrs nixpkgs.lib.platforms.all (system:
      let pkgs = nixpkgs.legacyPackages.${system}; in
      { default = pkgs.callPackage ./. { }; }
    );

    nixosModules.default = import ./module.nix self;

    checks.x86_64-linux.nixos = nixpkgs.lib.nixos.runTest {
      name = "spasm";
      imports = [ ./test.nix ];
      hostPkgs = nixpkgs.legacyPackages.x86_64-linux;
      defaults.imports = [ self.nixosModules.default ];
    };
  };
}
