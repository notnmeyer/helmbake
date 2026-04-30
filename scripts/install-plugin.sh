#!/usr/bin/env bash
set -euo pipefail

PLUGIN_DIR="${HELM_PLUGIN_DIR:-$(cd "$(dirname "$0")/.." && pwd)}"

VERSION="$(awk '/^version:/ { gsub(/["'"'"']/, "", $2); print $2 }' "$PLUGIN_DIR/plugin.yaml")"
if [ -z "$VERSION" ]; then
  echo "could not read version from $PLUGIN_DIR/plugin.yaml" >&2
  exit 1
fi

case "$(uname -s)" in
  Darwin) OS="Darwin" ;;
  Linux)  OS="Linux"  ;;
  *) echo "unsupported os: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64)   ARCH="x86_64" ;;
  arm64|aarch64)  ARCH="arm64"  ;;
  i386|i686)      ARCH="i386"   ;;
  *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
esac

ARCHIVE="helmbake_${OS}_${ARCH}.tar.gz"
URL="https://github.com/notnmeyer/helmbake/releases/download/v${VERSION}/${ARCHIVE}"

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "downloading helmbake v${VERSION} (${OS}/${ARCH})"
if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "$TMP/$ARCHIVE"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$URL" -O "$TMP/$ARCHIVE"
else
  echo "neither curl nor wget is available" >&2
  exit 1
fi

tar -xzf "$TMP/$ARCHIVE" -C "$TMP"
mkdir -p "$PLUGIN_DIR/bin"
mv "$TMP/helmbake" "$PLUGIN_DIR/bin/helmbake"
chmod +x "$PLUGIN_DIR/bin/helmbake"

echo "helmbake v${VERSION} installed. try: helm bake --help"
