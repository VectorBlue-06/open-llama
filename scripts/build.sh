#!/usr/bin/env bash
set -euo pipefail

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS="-s -w -X main.version=${VERSION}"

echo "Building OpenLlama ${VERSION}..."

case "${1:-$(go env GOOS)}" in
    linux)
        GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/openllama ./cmd/openllama
        ;;
    darwin)
        GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o bin/openllama ./cmd/openllama
        ;;
    windows)
        GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/openllama.exe ./cmd/openllama
        ;;
    all)
        GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/openllama-linux-amd64 ./cmd/openllama
        GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o bin/openllama-darwin-arm64 ./cmd/openllama
        GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o bin/openllama-windows-amd64.exe ./cmd/openllama
        ;;
    *)
        echo "Usage: $0 {linux|darwin|windows|all}"
        exit 1
        ;;
esac

echo "Build complete."
