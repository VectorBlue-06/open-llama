#!/usr/bin/env bash
set -euo pipefail

LLAMA_CPP_VERSION="${LLAMA_CPP_VERSION:-b4567}"
BASE_URL="https://github.com/ggerganov/llama.cpp/releases/download/${LLAMA_CPP_VERSION}"

case "${1:-}" in
    linux-amd64)
        URL="${BASE_URL}/llama-server-linux-x86_64"
        ;;
    linux-amd64-cuda)
        URL="${BASE_URL}/llama-server-linux-x86_64-cuda"
        ;;
    darwin-arm64)
        URL="${BASE_URL}/llama-server-darwin-arm64"
        ;;
    windows-amd64)
        URL="${BASE_URL}/llama-server-windows-x86_64.exe"
        ;;
    *)
        echo "Usage: $0 {linux-amd64|linux-amd64-cuda|darwin-arm64|windows-amd64}"
        exit 1
        ;;
esac

mkdir -p assets/server
echo "Downloading llama-server for $1..."
curl -L -o "assets/server/llama-server" "$URL"
chmod +x "assets/server/llama-server"
echo "Downloaded llama-server for $1"
