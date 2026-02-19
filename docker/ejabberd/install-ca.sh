#!/bin/bash
# Installs the local CA certificate into Ubuntu's trust store
# so that Dino (and other apps) trust aiox.local certificates.
# Run with: sudo bash docker/ejabberd/install-ca.sh

set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
CA_CERT="$DIR/certs/ca.crt"

if [ ! -f "$CA_CERT" ]; then
  echo "ERROR: CA cert not found. Run gen-cert.sh first."
  exit 1
fi

if [ "$(id -u)" -ne 0 ]; then
  echo "ERROR: This script must be run as root (sudo)."
  exit 1
fi

echo "==> Installing AIOX Local CA into system trust store..."
cp "$CA_CERT" /usr/local/share/ca-certificates/aiox-local-ca.crt
update-ca-certificates

echo ""
echo "CA installed. Restart Dino and try connecting again."
