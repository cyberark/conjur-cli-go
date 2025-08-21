#!/usr/bin/env bash

set -eo pipefail

function main() {
  local OUTPUT_DIR
  OUTPUT_DIR='dist/fips'

  git config --global --add safe.directory "$(pwd)"
  local SHORT_COMMIT_HASH
  SHORT_COMMIT_HASH="$(git rev-parse --short HEAD)"

  go mod tidy

  rm -rf "$OUTPUT_DIR"
  mkdir -p "$OUTPUT_DIR/conjur-cli_linux_amd64_v1"

  CGO_ENABLED=1 \
  GOOS=linux \
  GOARCH=amd64 \
  GOEXPERIMENT=systemcrypto \
  go build \
    -ldflags "-w \
      -X github.com/cyberark/conjur-cli-go/pkg/version.Tag=$SHORT_COMMIT_HASH \
      -X github.com/cyberark/conjur-cli-go/pkg/version.Version=$VERSION" \
    -o "$OUTPUT_DIR/conjur-cli_linux_amd64_v1/conjur" \
    ./cmd/conjur/main.go
  
  # Ensure the binary is compiled with FIPS enabled
  if ! go tool nm "$OUTPUT_DIR/conjur-cli_linux_amd64_v1/conjur" | grep 'openssl_FIPS_mode' >/dev/null; then
    echo "FIPS mode not enabled in the binary"
    exit 1
  fi
}

main "$@"
