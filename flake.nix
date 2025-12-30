{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              # Backend
              go
              golangci-lint
              evans
              protobuf
              protoc-gen-go
              protoc-gen-go-grpc
              firebase-tools

              # Frontend
              pnpm
              nodejs

              # Other
              gnumake
              kubernetes-helm
              infracost
            ];

            shellHook = ''
            '';
          };
        });
    };
}