#!/bin/bash
# Generates a local CA + server certificate for aiox.local (XMPP/TLS)
# Run once: bash docker/ejabberd/gen-cert.sh
# Then install CA: sudo bash docker/ejabberd/install-ca.sh

set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
CERT_DIR="$DIR/certs"
mkdir -p "$CERT_DIR"

echo "==> Generating local CA..."
openssl genrsa -out "$CERT_DIR/ca.key" 4096

openssl req -x509 -new -nodes \
  -key "$CERT_DIR/ca.key" \
  -sha256 -days 3650 \
  -out "$CERT_DIR/ca.crt" \
  -subj "/C=BR/O=AIOX Local CA/CN=AIOX Local CA"

echo "==> Generating server key..."
openssl genrsa -out "$CERT_DIR/server.key" 2048

echo "==> Generating CSR..."
openssl req -new \
  -key "$CERT_DIR/server.key" \
  -out "$CERT_DIR/server.csr" \
  -subj "/C=BR/O=AIOX/CN=aiox.local"

echo "==> Signing server cert with local CA (SAN for XMPP)..."
cat > "$CERT_DIR/server.ext" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = aiox.local
DNS.2 = *.aiox.local
DNS.3 = agents.aiox.local
EOF

openssl x509 -req \
  -in "$CERT_DIR/server.csr" \
  -CA "$CERT_DIR/ca.crt" \
  -CAkey "$CERT_DIR/ca.key" \
  -CAcreateserial \
  -out "$CERT_DIR/server.crt" \
  -days 3650 \
  -sha256 \
  -extfile "$CERT_DIR/server.ext"

echo "==> Creating PEM bundle (cert + key) for ejabberd..."
cat "$CERT_DIR/server.crt" "$CERT_DIR/server.key" > "$CERT_DIR/server.pem"

echo ""
echo "Done! Files generated in $CERT_DIR:"
ls -la "$CERT_DIR"
echo ""
echo "Next step — install the CA in your OS:"
echo "  sudo bash docker/ejabberd/install-ca.sh"
