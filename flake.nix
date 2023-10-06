{
  outputs = { self, nixpkgs }: {
    packages = nixpkgs.lib.genAttrs nixpkgs.lib.platforms.all (system:
      let pkgs = nixpkgs.legacyPackages.${system}; in
      { default = pkgs.callPackage ./nix { }; }
    );

    nixosModules.default = import ./nix/module.nix self;

    checks.x86_64-linux.nixos = nixpkgs.lib.nixos.runTest {
      name = "spasm";
      imports = [ ./nix/test.nix ];
      hostPkgs = nixpkgs.legacyPackages.x86_64-linux;
      defaults.imports = [ self.nixosModules.default ];
    };
  };
}
