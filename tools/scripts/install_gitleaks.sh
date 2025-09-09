#!/usr/bin/env bash
set -euo pipefail

echo "[INFO] Installing gitleaks via curl|sh fallback..."

URL_DEFAULT="https://raw.githubusercontent.com/gitleaks/gitleaks/master/install.sh"
URL="${GITLEAKS_INSTALL_URL:-$URL_DEFAULT}"

INSTALL_DIRECTORY="${INSTALL_DIRECTORY:-./bin}"
mkdir -p "$INSTALL_DIRECTORY"

curl -sSfL "$URL" | sh -s -- -b "$INSTALL_DIRECTORY"

if ! command -v gitleaks >/dev/null 2>&1; then
  echo "[INFO] gitleaks installed to $INSTALL_DIRECTORY"
  echo "[HINT] add to PATH for future shells: export PATH=\"$(pwd)/bin:\$PATH\""
fi

echo "[OK] gitleaks install done."
