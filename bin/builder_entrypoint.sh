#!/usr/bin/env bash

set -eo pipefail

function main() {
  local OUTPUT_DIR
  OUTPUT_DIR='dist/fips'

  local SHORT_COMMIT_HASH
  SHORT_COMMIT_HASH="$(git rev-parse --short HEAD)"

  go mod tidy

  rm -rf "$OUTPUT_DIR"
  mkdir -p "$OUTPUT_DIR/conjur-cli-go_linux_amd64_v1"

  CGO_ENABLED=1 \
  GOOS=linux \
  GOARCH=amd64 \
  GOEXPERIMENT=boringcrypto \
  go build \
    -ldflags "-w -X github.com/cyberark/conjur-cli-go/pkg/version.Tag=$SHORT_COMMIT_HASH -X main.version=${VERSION}" \
    -o "$OUTPUT_DIR/conjur-cli-go_linux_amd64_v1/conjur" \
    ./cmd/conjur/main.go
}

main "$@"
