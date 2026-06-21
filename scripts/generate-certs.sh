#!/bin/bash
# Certificate generation utility for proxis-c2 mTLS
# Generates CA, server, and agent certificates for secure communication

set -euo pipefail

# Configuration
CERT_DIR="${1:-/etc/proxis/certs}"
KEY_DIR="${2:-/etc/proxis/keys}"
DAYS_VALID=365
KEY_SIZE=2048

# Create directories
mkdir -p "${CERT_DIR}"
mkdir -p "${KEY_DIR}"

# Generate CA private key
echo "Generating CA private key..."
openssl genrsa -out "${KEY_DIR}/ca.key" ${KEY_SIZE}

# Generate CA certificate
echo "Generating CA certificate..."
openssl req -new -x509 -days ${DAYS_VALID} -key "${KEY_DIR}/ca.key" -out "${CERT_DIR}/ca.crt" -subj "/C=US/ST=State/L=City/O=Proxis-C2/OU=Security/CN=Proxis-C2 CA"

# Generate server private key
echo "Generating server private key..."
openssl genrsa -out "${KEY_DIR}/server.key" ${KEY_SIZE}

# Generate server CSR
echo "Generating server CSR..."
openssl req -new -key "${KEY_DIR}/server.key" -out "${CERT_DIR}/server.csr" -subj "/C=US/ST=State/L=City/O=Proxis-C2/OU=Server/CN=proxis-c2-server"

# Sign server certificate with CA
echo "Signing server certificate..."
openssl x509 -req -days ${DAYS_VALID} -in "${CERT_DIR}/server.csr" -CA "${CERT_DIR}/ca.crt" -CAkey "${KEY_DIR}/ca.key" -CAcreateserial -out "${CERT_DIR}/server.crt" -extfile <(printf "subjectAltName=DNS:proxis-c2-server,DNS:localhost,IP:127.0.0.1")

# Generate agent private key
echo "Generating agent private key..."
openssl genrsa -out "${KEY_DIR}/agent.key" ${KEY_SIZE}

# Generate agent CSR
echo "Generating agent CSR..."
openssl req -new -key "${KEY_DIR}/agent.key" -out "${CERT_DIR}/agent.csr" -subj "/C=US/ST=State/L=City/O=Proxis-C2/OU=Agent/CN=proxis-c2-agent"

# Sign agent certificate with CA
echo "Signing agent certificate..."
openssl x509 -req -days ${DAYS_VALID} -in "${CERT_DIR}/agent.csr" -CA "${CERT_DIR}/ca.crt" -CAkey "${KEY_DIR}/ca.key" -CAcreateserial -out "${CERT_DIR}/agent.crt" -extfile <(printf "extendedKeyUsage=clientAuth")

# Generate master encryption key
echo "Generating master encryption key..."
openssl rand -hex 32 > "${KEY_DIR}/master.key"

# Generate HMAC secret
echo "Generating HMAC secret..."
openssl rand -hex 32 > "${KEY_DIR}/hmac.secret"

# Set appropriate permissions
chmod 600 "${KEY_DIR}/ca.key"
chmod 600 "${KEY_DIR}/server.key"
chmod 600 "${KEY_DIR}/agent.key"
chmod 600 "${KEY_DIR}/master.key"
chmod 600 "${KEY_DIR}/hmac.secret"

# Cleanup CSR files
rm -f "${CERT_DIR}/server.csr" "${CERT_DIR}/agent.csr"

echo "Certificate generation complete."
echo "CA certificate: ${CERT_DIR}/ca.crt"
echo "Server certificate: ${CERT_DIR}/server.crt"
echo "Agent certificate: ${CERT_DIR}/agent.crt"