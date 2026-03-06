#!/usr/bin/env bash
set -euo pipefail

REPO="spritz-finance/spritz-cli"
VERSION="${SPRITZ_VERSION:-latest}"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
CURL_ARGS=(--proto '=https' --tlsv1.2 --fail --silent --show-error --location)

case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

GITHUB="https://github.com/${REPO}/releases"
if [ "$VERSION" = "latest" ]; then
  RELEASE_URL=$(curl "${CURL_ARGS[@]}" -o /dev/null -w '%{url_effective}' "$GITHUB/latest")
  RELEASE_TAG=${RELEASE_URL##*/}
else
  RELEASE_TAG="$VERSION"
fi
URL="$GITHUB/download/$RELEASE_TAG"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

ARCHIVE="spritz_${OS}_${ARCH}.tar.gz"

if ! command -v cosign >/dev/null 2>&1; then
  echo "Error: cosign is required to verify Spritz release signatures." >&2
  echo "Install cosign from https://docs.sigstore.dev/cosign/system_config/installation/ and retry." >&2
  exit 1
fi

echo "Downloading spritz ${RELEASE_TAG} metadata for ${OS}/${ARCH}..."
curl "${CURL_ARGS[@]}" "$URL/checksums.txt" -o "$TMP/checksums.txt"
curl "${CURL_ARGS[@]}" "$URL/checksums.txt.sigstore.json" -o "$TMP/checksums.txt.sigstore.json"

cosign verify-blob \
  --bundle "$TMP/checksums.txt.sigstore.json" \
  --certificate-identity "https://github.com/${REPO}/.github/workflows/release.yml@refs/tags/${RELEASE_TAG}" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  "$TMP/checksums.txt" >/dev/null

# Verify checksum
EXPECTED=$(awk -v name="$ARCHIVE" '$2 == name {print $1}' "$TMP/checksums.txt")
if [ -z "$EXPECTED" ]; then
  echo "Error: no checksum found for $ARCHIVE in checksums.txt" >&2
  exit 1
fi

echo "Signed checksums verified. Downloading spritz ${RELEASE_TAG} for ${OS}/${ARCH}..."
curl "${CURL_ARGS[@]}" "$URL/$ARCHIVE" -o "$TMP/$ARCHIVE"

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL=$(sha256sum "$TMP/$ARCHIVE" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL=$(shasum -a 256 "$TMP/$ARCHIVE" | awk '{print $1}')
else
  echo "Error: sha256sum or shasum is required to verify the downloaded archive." >&2
  exit 1
fi

if [ "$ACTUAL" != "$EXPECTED" ]; then
  echo "Error: checksum mismatch!" >&2
  echo "  Expected: $EXPECTED" >&2
  echo "  Got:      $ACTUAL" >&2
  echo "  The download may be corrupted or tampered with." >&2
  exit 1
fi

echo "Checksum verified."

tar -xzf "$TMP/$ARCHIVE" -C "$TMP"

BINARY_PATH=""
if [ -f "$TMP/spritz" ]; then
  BINARY_PATH="$TMP/spritz"
elif [ -f "$TMP/spritz-cli" ]; then
  BINARY_PATH="$TMP/spritz-cli"
else
  echo "Error: release archive did not contain a spritz binary." >&2
  exit 1
fi

INSTALL_DIR="${SPRITZ_INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$INSTALL_DIR"
install -m755 "$BINARY_PATH" "$INSTALL_DIR/spritz"

# Add to PATH if not already there
SHELL_RC=""
case "${SHELL:-}" in
  */zsh)  SHELL_RC="$HOME/.zshrc" ;;
  */bash) SHELL_RC="$HOME/.bashrc" ;;
esac

if [ -n "$SHELL_RC" ] && ! grep -q "$INSTALL_DIR" "$SHELL_RC" 2>/dev/null; then
  echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"
  echo "Added $INSTALL_DIR to PATH in $SHELL_RC"
fi

INSTALLED_VERSION=$("$INSTALL_DIR/spritz" version 2>/dev/null || echo "unknown")
echo "✓ ${INSTALLED_VERSION} installed to $INSTALL_DIR/spritz"
echo "  Run 'spritz login' to authenticate."
