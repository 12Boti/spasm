{ buildGoModule }:
buildGoModule {
  pname = "spasm";
  version = "1.0.0";
  src = ./src;
  vendorHash = "sha256-xPkbDMVxQFiS0PPRBv+H6lDjDe7F2EO+si4lCWiVFLo=";
  meta = {
    mainProgram = "spasm";
  };
}
