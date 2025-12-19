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

              # Frontend
              pnpm
              nodejs

              # Other
              gnumake
              kubernetes-helm
              infracost
              protobuf
              protoc-gen-go
              protoc-gen-go-grpc
            ];

            shellHook = ''
            '';
          };
        });
    };
}