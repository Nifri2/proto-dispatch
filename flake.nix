
# This flake works buy simply using go get <package> to fetch dependencies.
# It uses vendor/ for dependencies, so you can run `go mod vendor` to populate it.
# Or run `task mod` to automatically fetch dependencies and populate the vendor directory.

# To build the Go project, run:
#   task build

# To run the Go project, run:
#   nix run

{
  description = "Basic Go Project with Nix Flake";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs?ref=nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      # Merge all packages under one attribute set to expose gopls
      packages.${system} = {
        default = pkgs.buildGoModule {
          pname = "put-name-here";  # Set the name of your package
          version = "0.0.1";

          src = pkgs.lib.cleanSource ./.;

          env.CGO_ENABLED = 1;  # Disable CGO for Static Compilation

          ldflags = [
            "-s" "-w" "-extldflags '-static'"
          ];  # Strip Binary and Disable Debug Information, static linking

          vendorHash = null;  # Null if you don't have a vendor directory

          buildInputs = [
            pkgs.musl
            pkgs.go
            pkgs.gopls
            pkgs.gotools
            pkgs.go-tools
            pkgs.golangci-lint
            pkgs.tinygo
            pkgs.picotool
          ];
        };

        # Expose gopls directly for scripts to use
        gopls = pkgs.gopls;
      };

      devShells.${system}.default = pkgs.mkShell {
        nativeBuildInputs = [
          pkgs.go
          pkgs.go-task
          pkgs.gopls
          pkgs.golangci-lint
          pkgs.gotools
          pkgs.go-tools
          pkgs.musl
          pkgs.tinygo
          pkgs.picotool
        ];

        shellHook = ''
          if [ "$SHELL" = "$(which fish)" ]; then
            source .dev-fish-setup.fish
          fi
        ''; # Optional: Fish shell setup script
      };
    };
}
