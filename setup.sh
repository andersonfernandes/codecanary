#!/bin/sh
set -eu

REPO="alansikora/codecanary"
BINARY="codecanary-setup"
CANARY=false

for arg in "$@"; do
  case "$arg" in
    --canary) CANARY=true ;;
  esac
done

cleanup() { rm -rf "$TMPDIR"; }
trap cleanup EXIT INT TERM

TMPDIR=$(mktemp -d)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
esac

if [ "$CANARY" = true ]; then
  TAG=$(curl -fsSL "https://api.github.com/repos/$REPO/releases" | grep '"tag_name"' | head -1 | sed 's/.*"v//' | sed 's/".*//')
  echo "Downloading CodeCanary Setup v${TAG} (canary)..."
else
  TAG=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"v//' | sed 's/".*//')
  echo "Downloading CodeCanary Setup v${TAG}..."
fi
URL="https://github.com/$REPO/releases/download/v${TAG}/codecanary-setup_${TAG}_${OS}_${ARCH}.tar.gz"

curl -fsSL "$URL" | tar -xz -C "$TMPDIR" "$BINARY"
chmod +x "$TMPDIR/$BINARY"
"$TMPDIR/$BINARY" "$@" < /dev/tty
