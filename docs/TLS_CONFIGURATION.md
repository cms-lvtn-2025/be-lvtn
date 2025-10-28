# TLS Configuration Guide

## 🔒 Lỗi: TLSV1_ALERT_CERTIFICATE_REQUIRED

### Nguyên nhân:

Services của bạn đang sử dụng **mTLS (mutual TLS)**:
- Server yêu cầu client phải gửi certificate
- Client (từ API gateway, browser, hoặc service khác) không có certificate
- Kết nối bị từ chối

```
Error: SSL alert number 116
TLSV1_ALERT_CERTIFICATE_REQUIRED
```

### Code hiện tại:

```go
// src/pkg/tls/tls.go line 59
tlsConfig := &tls.Config{
    ClientAuth: tls.RequireAndVerifyClientCert,  // ← YÊU CẦU client cert
}
```

## ✅ Giải pháp

### **Solution 1: Disable mTLS (Recommended cho Docker/Production)**

Cho phép client kết nối mà không cần certificate.

#### Cập nhật service main.go files:

**Trước:**
```go
creds, err := tls.LoadServerTLSCredentials("academic")
```

**Sau:**
```go
creds, err := tls.LoadServerTLSFromEnv("academic")
```

#### Thêm environment variable:

```bash
# env/academic.env
TLS_MODE=optional  # Cho phép kết nối với hoặc không có client cert

# Hoặc
TLS_MODE=none     # Không yêu cầu client cert
TLS_MODE=strict   # Yêu cầu client cert (mTLS - như hiện tại)
```

#### Update docker-compose:

```yaml
services:
  academic:
    environment:
      - TLS_MODE=optional  # Hoặc none
```

### **Solution 2: Sử dụng Optional mTLS**

Client có certificate thì verify, không có thì vẫn cho kết nối.

#### Update service code:

```go
// Thay vì
creds, err := tls.LoadServerTLSCredentials("academic")

// Dùng
creds, err := tls.LoadServerTLSCredentialsOptional("academic", false)
```

### **Solution 3: Disable TLS hoàn toàn (Dev only)**

**WARNING: Chỉ dùng cho development!**

```go
// Comment out TLS
// creds, err := tls.LoadServerTLSCredentials("academic")
// grpcServer := grpc.NewServer(grpc.Creds(creds))

// Create server without TLS
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(logger.UnaryServerInterceptor()),
)
```

## 🔧 Cách sửa cho toàn bộ project:

### Bước 1: Update env files

```bash
# env/academic.env
TLS_MODE=optional

# env/council.env
TLS_MODE=optional

# env/file.env
TLS_MODE=optional

# env/role.env
TLS_MODE=optional

# env/thesis.env
TLS_MODE=optional

# env/user.env
TLS_MODE=optional
```

### Bước 2: Update service main.go files

Thay đổi tất cả 6 services:

**src/service/academic/main.go:**
```go
// Line 42-45
// BEFORE:
// creds, err := tls.LoadServerTLSCredentials("academic")

// AFTER:
creds, err := tls.LoadServerTLSFromEnv("academic")
if err != nil {
    log.Fatalf("Failed to load TLS credentials: %v", err)
}
```

Làm tương tự cho:
- `src/service/council/main.go`
- `src/service/file/main.go`
- `src/service/role/main.go`
- `src/service/thesis/main.go`
- `src/service/user/main.go`

### Bước 3: Update docker-compose

```yaml
# docker-compose.yml và docker-compose.prod.yml
services:
  academic:
    environment:
      - TLS_MODE=optional

  council:
    environment:
      - TLS_MODE=optional

  # ... các services khác
```

### Bước 4: Rebuild và restart

```bash
# Local
./docker-setup.sh build
./docker-setup.sh restart

# Production
docker-compose -f docker-compose.prod.yml build
docker-compose -f docker-compose.prod.yml up -d
```

## 🎯 TLS Modes Explained

### **TLS_MODE=strict** (mTLS - Highest Security)
- ✅ Client MUST provide valid certificate
- ✅ Server verifies client identity
- ❌ Cannot connect without client cert
- **Use case:** Internal microservices, high security

### **TLS_MODE=optional** (Recommended)
- ✅ Client CAN provide certificate (will be verified)
- ✅ Client can connect WITHOUT certificate
- ✅ Flexible for different clients
- **Use case:** Mixed environments, API gateway → services

### **TLS_MODE=none** (TLS without client auth)
- ✅ Encryption enabled (TLS)
- ❌ No client authentication
- **Use case:** Public APIs, external clients

### **No TLS** (Dev only)
- ❌ No encryption
- ❌ No authentication
- **Use case:** Local development only

## 🛡️ Security Recommendations

### **Production:**
```bash
# For service-to-service: Use mTLS
TLS_MODE=strict

# For API Gateway → Services: Use optional
TLS_MODE=optional
```

### **Staging:**
```bash
TLS_MODE=optional
```

### **Development:**
```bash
TLS_MODE=none
# Or disable TLS completely
```

## 🔍 Debugging TLS Issues

### Check certificates exist:

```bash
ls -la certs/services/academic/
# Should have: academic-server.crt, academic-server.key

ls -la certs/ca/
# Should have: ca.crt
```

### Test connection:

```bash
# With grpcurl
grpcurl -insecure localhost:50051 list

# Check TLS mode
docker-compose logs academic | grep TLS
```

### View certificate info:

```bash
openssl x509 -in certs/services/academic/academic-server.crt -text -noout
```

## 🚨 Common Errors

### Error 1: Certificate not found
```
failed to load server certificate: open certs/...
```
**Fix:** Regenerate certificates
```bash
cd certs && ./generate-certs.sh
```

### Error 2: Client certificate required
```
SSL alert number 116
TLSV1_ALERT_CERTIFICATE_REQUIRED
```
**Fix:** Set `TLS_MODE=optional` or `TLS_MODE=none`

### Error 3: Certificate has expired
```
certificate has expired or is not yet valid
```
**Fix:** Regenerate certificates with longer validity

### Error 4: Hostname mismatch
```
x509: certificate is valid for X, not Y
```
**Fix:** Update certificate with correct hostname/IP

## 📚 References

- [Go TLS Package](https://pkg.go.dev/crypto/tls)
- [gRPC Authentication](https://grpc.io/docs/guides/auth/)
- [mTLS Best Practices](https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/)

## 💡 Quick Fix Summary

**Fastest fix - Add to all env files:**
```bash
TLS_MODE=optional
```

**Then update all 6 service main.go files:**
```go
creds, err := tls.LoadServerTLSFromEnv("SERVICE_NAME")
```

**Rebuild and restart:**
```bash
docker-compose build
docker-compose up -d
```

Done! ✅
