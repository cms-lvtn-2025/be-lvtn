#!/bin/bash

# mTLS Certificate Generation Script for gRPC Microservices
# This script generates:
# 1. CA (Certificate Authority)
# 2. Server certificates for each microservice
# 3. Client certificates for GraphQL gateway

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VALIDITY_DAYS=3650  # 10 years for development, reduce for production

echo "========================================="
echo "mTLS Certificate Generation"
echo "========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ============================================
# 1. Generate CA (Certificate Authority)
# ============================================
echo -e "${GREEN}[1/4] Generating Certificate Authority (CA)...${NC}"

if [ -f "$SCRIPT_DIR/ca/ca.key" ]; then
    echo -e "${YELLOW}CA already exists. Skipping...${NC}"
else
    # Generate CA private key
    openssl genrsa -out "$SCRIPT_DIR/ca/ca.key" 4096

    # Generate CA certificate
    openssl req -new -x509 \
        -key "$SCRIPT_DIR/ca/ca.key" \
        -out "$SCRIPT_DIR/ca/ca.crt" \
        -days $VALIDITY_DAYS \
        -subj "/C=VN/ST=HoChiMinh/L=HoChiMinh/O=HeheheBE/OU=CA/CN=heheheh-ca" \
        -sha256

    echo -e "${GREEN}✓ CA certificate generated${NC}"
fi

# ============================================
# 2. Generate Server Certificates
# ============================================
echo -e "\n${GREEN}[2/4] Generating Server Certificates...${NC}"

SERVICES=("user" "council" "thesis" "academic" "role" "file")

for SERVICE in "${SERVICES[@]}"; do
    echo -e "${YELLOW}Generating certificate for ${SERVICE}-service...${NC}"

    SERVICE_DIR="$SCRIPT_DIR/services/$SERVICE"

    if [ -f "$SERVICE_DIR/${SERVICE}-server.key" ]; then
        echo -e "${YELLOW}  Certificate already exists. Skipping...${NC}"
        continue
    fi

    # Generate private key
    openssl genrsa -out "$SERVICE_DIR/${SERVICE}-server.key" 4096

    # Create config file for SAN (Subject Alternative Names)
    cat > "$SERVICE_DIR/${SERVICE}-server.cnf" <<EOF
[req]
default_bits = 4096
prompt = no
default_md = sha256
distinguished_name = dn
req_extensions = v3_req

[dn]
C = VN
ST = HoChiMinh
L = HoChiMinh
O = HeheheBE
OU = Services
CN = ${SERVICE}-service

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${SERVICE}-service
DNS.2 = localhost
DNS.3 = ${SERVICE}
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

    # Generate CSR (Certificate Signing Request)
    openssl req -new \
        -key "$SERVICE_DIR/${SERVICE}-server.key" \
        -out "$SERVICE_DIR/${SERVICE}-server.csr" \
        -config "$SERVICE_DIR/${SERVICE}-server.cnf"

    # Sign with CA
    openssl x509 -req \
        -in "$SERVICE_DIR/${SERVICE}-server.csr" \
        -CA "$SCRIPT_DIR/ca/ca.crt" \
        -CAkey "$SCRIPT_DIR/ca/ca.key" \
        -CAcreateserial \
        -out "$SERVICE_DIR/${SERVICE}-server.crt" \
        -days $VALIDITY_DAYS \
        -extensions v3_req \
        -extfile "$SERVICE_DIR/${SERVICE}-server.cnf" \
        -sha256

    # Cleanup CSR file
    rm "$SERVICE_DIR/${SERVICE}-server.csr"

    echo -e "${GREEN}  ✓ ${SERVICE}-service certificate generated${NC}"
done

# ============================================
# 3. Generate Client Certificates
# ============================================
echo -e "\n${GREEN}[3/4] Generating Client Certificates...${NC}"

CLIENT_DIR="$SCRIPT_DIR/clients"

if [ -f "$CLIENT_DIR/client.key" ]; then
    echo -e "${YELLOW}Client certificate already exists. Skipping...${NC}"
else
    # Generate client private key
    openssl genrsa -out "$CLIENT_DIR/client.key" 4096

    # Create config file
    cat > "$CLIENT_DIR/client.cnf" <<EOF
[req]
default_bits = 4096
prompt = no
default_md = sha256
distinguished_name = dn

[dn]
C = VN
ST = HoChiMinh
L = HoChiMinh
O = HeheheBE
OU = GraphQL Gateway
CN = graphql-gateway
EOF

    # Generate CSR
    openssl req -new \
        -key "$CLIENT_DIR/client.key" \
        -out "$CLIENT_DIR/client.csr" \
        -config "$CLIENT_DIR/client.cnf"

    # Sign with CA
    openssl x509 -req \
        -in "$CLIENT_DIR/client.csr" \
        -CA "$SCRIPT_DIR/ca/ca.crt" \
        -CAkey "$SCRIPT_DIR/ca/ca.key" \
        -CAcreateserial \
        -out "$CLIENT_DIR/client.crt" \
        -days $VALIDITY_DAYS \
        -sha256

    # Cleanup
    rm "$CLIENT_DIR/client.csr"

    echo -e "${GREEN}✓ Client certificate generated${NC}"
fi

# Copy CA cert to client directory for convenience
cp "$SCRIPT_DIR/ca/ca.crt" "$CLIENT_DIR/ca.crt"

# ============================================
# 4. Set Permissions
# ============================================
echo -e "\n${GREEN}[4/4] Setting Permissions...${NC}"

# Secure private keys (readable only by owner)
find "$SCRIPT_DIR" -name "*.key" -exec chmod 600 {} \;

# Public certs can be readable
find "$SCRIPT_DIR" -name "*.crt" -exec chmod 644 {} \;

echo -e "${GREEN}✓ Permissions set${NC}"

# ============================================
# Summary
# ============================================
echo -e "\n${GREEN}========================================="
echo "Certificate Generation Complete!"
echo "=========================================${NC}"
echo ""
echo "Generated certificates:"
echo "  CA Certificate:     $SCRIPT_DIR/ca/ca.crt"
echo "  CA Private Key:     $SCRIPT_DIR/ca/ca.key"
echo ""
echo "Service Certificates:"
for SERVICE in "${SERVICES[@]}"; do
    echo "  ${SERVICE}-service: $SCRIPT_DIR/services/$SERVICE/${SERVICE}-server.{crt,key}"
done
echo ""
echo "Client Certificate:"
echo "  GraphQL Gateway:    $SCRIPT_DIR/clients/client.{crt,key}"
echo "  CA Certificate:     $SCRIPT_DIR/clients/ca.crt"
echo ""
echo -e "${YELLOW}⚠️  IMPORTANT SECURITY NOTES:${NC}"
echo "  1. Keep *.key files secure and NEVER commit to git"
echo "  2. Add 'certs/' to .gitignore"
echo "  3. In production, use proper secrets management (Vault, AWS Secrets Manager)"
echo "  4. Rotate certificates before expiry"
echo "  5. These certificates are valid for $VALIDITY_DAYS days"
echo ""
echo -e "${GREEN}Next steps:${NC}"
echo "  1. Update service main.go files to use TLS"
echo "  2. Update client connections to use mTLS"
echo "  3. Test connections"
echo ""
