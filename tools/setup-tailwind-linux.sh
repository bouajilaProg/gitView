#!/usr/bin/env sh
set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
TOOLS_DIR="$ROOT_DIR/tools"
BIN_PATH="$TOOLS_DIR/tailwindcss"
VERSION="v3.4.17"
URL="https://github.com/tailwindlabs/tailwindcss/releases/download/$VERSION/tailwindcss-linux-x64"

mkdir -p "$TOOLS_DIR"
curl -fL "$URL" -o "$BIN_PATH"
chmod +x "$BIN_PATH"

echo "Installed Tailwind CLI to $BIN_PATH"
