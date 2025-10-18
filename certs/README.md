# mTLS Certificates for gRPC Services

This directory contains TLS certificates for securing gRPC communication between microservices.

## Directory Structure

```
certs/
├── ca/                           # Certificate Authority
│   ├── ca.crt                    # CA public certificate
│   └── ca.key                    # CA private key (⚠️ KEEP SECURE)
├── services/                     # Service server certificates
│   ├── user/
│   │   ├── user-server.crt       # User service certificate
│   │   └── user-server.key       # User service private key (⚠️ KEEP SECURE)
│   ├── council/
│   ├── thesis/
│   ├── academic/
│   ├── role/
│   └── file/
├── clients/                      # Client certificates (for GraphQL gateway)
│   ├── client.crt                # Client certificate
│   ├── client.key                # Client private key (⚠️ KEEP SECURE)
│   └── ca.crt                    # CA cert copy (for client validation)
└── generate-certs.sh             # Certificate generation script
```

## Quick Start

### 1. Generate Certificates

```bash
cd /home/thaily/code/heheheh_be/certs
chmod +x generate-certs.sh
./generate-certs.sh
```

This will generate:
- 1 CA certificate
- 6 server certificates (one for each microservice)
- 1 client certificate (for GraphQL gateway)

### 2. Verify Certificates

```bash
# View CA certificate
openssl x509 -in ca/ca.crt -text -noout

# View a service certificate
openssl x509 -in services/user/user-server.crt -text -noout

# Verify certificate chain
openssl verify -CAfile ca/ca.crt services/user/user-server.crt
```

### 3. Test Connection

```bash
# Start a test server
openssl s_server -accept 8443 \
  -cert services/user/user-server.crt \
  -key services/user/user-server.key \
  -CAfile ca/ca.crt -verify 1

# Connect with client certificate
openssl s_client -connect localhost:8443 \
  -cert clients/client.crt \
  -key clients/client.key \
  -CAfile clients/ca.crt
```

## Certificate Details

### CA Certificate
- **Validity**: 10 years
- **Key Size**: 4096 bits
- **Algorithm**: RSA + SHA256
- **Common Name**: heheheh-ca

### Service Certificates
- **Validity**: 10 years (adjust for production)
- **Key Size**: 4096 bits
- **Subject Alternative Names (SAN)**:
  - DNS: `{service}-service`, `localhost`, `{service}`
  - IP: `127.0.0.1`, `::1`
- **Common Name**: `{service}-service`

### Client Certificate
- **Validity**: 10 years
- **Key Size**: 4096 bits
- **Common Name**: graphql-gateway

## Security Best Practices

### ⚠️ CRITICAL
1. **NEVER** commit `*.key` files to git
2. **NEVER** share private keys
3. **ALWAYS** use proper file permissions (600 for .key files)

### Production Recommendations
1. **Reduce certificate validity** to 90 days or less
2. **Implement automatic rotation** using tools like cert-manager (Kubernetes) or Vault
3. **Use HSM or Key Management Service** for CA private key storage
4. **Enable certificate revocation** (CRL or OCSP)
5. **Monitor certificate expiry** with alerting
6. **Use different CA** for different environments (dev/staging/prod)

### File Permissions

```bash
# Private keys - readable only by owner
chmod 600 *.key

# Certificates - readable by all
chmod 644 *.crt

# Directory
chmod 700 ca/
chmod 755 services/
chmod 755 clients/
```

## Troubleshooting

### Certificate Verification Failed

```bash
# Check certificate details
openssl x509 -in services/user/user-server.crt -text -noout | grep -A 1 "Subject Alternative Name"

# Verify certificate chain
openssl verify -CAfile ca/ca.crt services/user/user-server.crt
```

### Common Errors

1. **"certificate has expired"**
   - Regenerate certificates: `./generate-certs.sh`

2. **"certificate verify failed"**
   - Ensure CA certificate is correct
   - Check file paths in code

3. **"x509: cannot validate certificate for 127.0.0.1"**
   - Check SAN includes the IP address
   - Regenerate with correct SAN

## Rotation Schedule

| Component | Validity | Rotation Frequency |
|-----------|----------|-------------------|
| CA | 10 years | Never (or every 5 years) |
| Service Certs | 10 years (dev) / 90 days (prod) | Before expiry |
| Client Certs | 10 years (dev) / 90 days (prod) | Before expiry |

## References

- [gRPC Authentication](https://grpc.io/docs/guides/auth/)
- [Go gRPC Security](https://github.com/grpc/grpc-go/tree/master/examples/features/encryption)
- [mTLS Best Practices](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/)
