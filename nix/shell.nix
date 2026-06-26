{ pkgs }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    gopls
    gotools
    golangci-lint
  ];

  shellHook = ''
    echo "safe-rm dev shell"
    if [ -z "$GOPATH" ]; then
      export GOPATH="$HOME/go"
    fi
  '';
}
