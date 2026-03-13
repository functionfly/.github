#!/usr/bin/env bash
# Install the fly CLI from GitHub releases.
# Usage: curl -fsSL https://raw.githubusercontent.com/functionfly/functionfly/main/scripts/install.sh | bash
# Or: VERSION=v1.0.0 BINDIR=~/.local/bin bash scripts/install.sh

set -e

REPO="${REPO:-functionfly/functionfly}"
VERSION="${VERSION:-latest}"
BINDIR="${BINDIR:-}"
SUDO="${SUDO:-}"

# Resolve latest to a real tag
if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" | grep -o '"tag_name": *"[^"]*"' | sed 's/"tag_name": *"\(.*\)"/\1/')
fi
# Strip 'v' prefix for asset names
VER_STRIP="${VERSION#v}"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$OS" in
  linux)   OS_ID="linux";;
  darwin)  OS_ID="macos";;
  *)       echo "Unsupported OS: $OS"; exit 1;;
esac
case "$ARCH" in
  x86_64|amd64) ARCH_ID="x86_64";;
  aarch64|arm64) ARCH_ID="arm64";;
  *)       echo "Unsupported arch: $ARCH"; exit 1;;
esac

# Install directory: BINDIR or ~/.local/bin if writable, else /usr/local/bin (with sudo)
if [ -z "$BINDIR" ]; then
  if [ -w "$HOME/.local/bin" ] 2>/dev/null || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
    BINDIR="$HOME/.local/bin"
  else
    BINDIR="/usr/local/bin"
    SUDO="sudo"
  fi
fi
BINDIR="${BINDIR/#\~/$HOME}"
mkdir -p "$BINDIR"

# Download URL (GoReleaser wrap_in_directory: archive contains fly_VER_OS_ARCH/fly)
ASSET="fly_${VER_STRIP}_${OS_ID}_${ARCH_ID}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

echo "Installing fly ${VERSION} (${OS_ID}/${ARCH_ID}) to ${BINDIR}"
TMPDIR=$(mktemp -d)
trap "rm -rf ${TMPDIR}" EXIT
curl -fsSL -o "${TMPDIR}/fly.tar.gz" "$URL"
tar -xzf "${TMPDIR}/fly.tar.gz" -C "${TMPDIR}"
# Binary may be in a subdir when wrap_in_directory is true
if [ -f "${TMPDIR}/fly/fly" ]; then
  $SUDO mv "${TMPDIR}/fly/fly" "${BINDIR}/fly"
else
  $SUDO mv "${TMPDIR}/fly" "${BINDIR}/fly"
fi
chmod +x "${BINDIR}/fly"

echo "Installed: ${BINDIR}/fly"
if ! command -v fly >/dev/null 2>&1; then
  echo "Add to PATH: export PATH=\"${BINDIR}:\$PATH\""
  echo "Or add the above to your shell profile (.bashrc, .zshrc, etc.)."
fi
