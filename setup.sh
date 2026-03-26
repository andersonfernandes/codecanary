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

TAG=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"v//' | sed 's/".*//')
if [ "$CANARY" = true ]; then
  echo "Downloading CodeCanary Setup (canary)..."
else
  echo "Downloading CodeCanary Setup v${TAG}..."
fi
ARCHIVE="codecanary-setup_${TAG}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/v${TAG}/${ARCHIVE}"
CHECKSUMS_URL="https://github.com/$REPO/releases/download/v${TAG}/checksums.txt"

curl -fsSL -o "$TMPDIR/$ARCHIVE" "$URL"
curl -fsSL -o "$TMPDIR/checksums.txt" "$CHECKSUMS_URL"

# Verify checksum (sha256sum on Linux, shasum on macOS).
# NOTE: This verification must happen in shell because the Go binary is the
# thing being verified — we can't run it before confirming its integrity.
if command -v sha256sum >/dev/null 2>&1; then
  (cd "$TMPDIR" && grep -F "$ARCHIVE" checksums.txt | sha256sum -c --quiet -)
elif command -v shasum >/dev/null 2>&1; then
  (cd "$TMPDIR" && grep -F "$ARCHIVE" checksums.txt | shasum -a 256 -c --quiet -)
else
  echo "Warning: no sha256 tool found, skipping checksum verification" >&2
fi

tar -xz -C "$TMPDIR" -f "$TMPDIR/$ARCHIVE" "$BINARY"
chmod +x "$TMPDIR/$BINARY"
"$TMPDIR/$BINARY" "$@" < /dev/tty
