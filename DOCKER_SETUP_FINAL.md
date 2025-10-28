# ✅ Docker Setup - Đã kiểm tra và sửa

## Tóm tắt các thay đổi

### 🔧 Đã sửa:

#### 1. **docker-compose.yml** (tất cả 6 services)

**Trước (SAI):**
```yaml
build:
  context: ../..              # ❌ Sai context
  dockerfile: Dockerfile       # ❌ Không có path đầy đủ
image: thaily/academic:latest  # ❌ Thiếu prefix lvtn-
env_file:
  - .env                       # ❌ Tên file không đúng
volumes:
  - ../../logs:/app/logs       # ❌ Path sai
```

**Sau (ĐÚNG):**
```yaml
build:
  context: ../../..                        # ✅ Context từ root project
  dockerfile: src/service/academic/Dockerfile  # ✅ Path đầy đủ
image: thaily/lvtn-academic:latest         # ✅ Prefix lvtn-
env_file:
  - academic.env                           # ✅ Tên file đúng
volumes:
  - .:/app/service:ro                      # ✅ Mount current dir
  - ../../../logs:/app/logs                # ✅ Path đúng
```

#### 2. **Dockerfile** (tất cả 6 services)

**Trước (SAI):**
```dockerfile
# Build từ current directory
RUN go build -o academic-service .  # ❌ SAI

# Mkdir sai
RUN mkdir -p /app/certs /app/logs   # ❌ Không cần /app/certs
```

**Sau (ĐÚNG):**
```dockerfile
# Build từ path đúng
RUN go build -o academic-service ./src/service/academic  # ✅

# Mkdir đúng
RUN mkdir -p /app/service /app/logs  # ✅ /app/service cho certs
```

#### 3. **tls.go**

**Đã cập nhật:**
```go
// Local fallback - certs in src/service/{service_name}/
basePath = filepath.Join(
    os.Getenv("HOME"),
    "code", "heheheh_be",
    "src", "service",      // ✅ Thêm "src"
    serviceName
)
```

## Cấu trúc cuối cùng

```
heheheh_be/
├── go.mod, go.sum           # Root level
├── proto/                   # Protobuf definitions
├── logs/                    # Logs directory
├── .github/workflows/
│   └── dabe.yml            # GitHub Actions
└── src/service/
    ├── academic/
    │   ├── academic.env           ✅
    │   ├── academic-server.crt    ✅
    │   ├── academic-server.key    ✅
    │   ├── Dockerfile             ✅ Fixed
    │   ├── docker-compose.yml     ✅ Fixed
    │   ├── main.go
    │   └── handler/
    ├── council/, file/, role/, thesis/, user/  (tương tự)
    └── pkg/
        ├── database/
        ├── logger/
        └── tls/
            └── tls.go       ✅ Fixed
```

## Giải thích chi tiết

### Context và Dockerfile path

```yaml
build:
  context: ../../..                      # Từ src/service/academic lên root
  dockerfile: src/service/academic/Dockerfile  # Từ root xuống Dockerfile
```

**Tại sao:**
- `context: ../../..` → `/home/thaily/code/heheheh_be/`
- Cần context ở root để access `go.mod`, `go.sum`, và toàn bộ source
- `dockerfile` path tính từ context root

### Build command trong Dockerfile

```dockerfile
# Context = /home/thaily/code/heheheh_be/
WORKDIR /app
COPY go.mod go.sum ./
COPY . .  # Copy toàn bộ project

# Build từ specific service directory
RUN go build -o academic-service ./src/service/academic
```

**Tại sao:**
- Copy toàn bộ project (`.`) vào `/app/`
- Build từ path `./src/service/academic` (relative to WORKDIR `/app/`)
- Binary output: `/app/academic-service`

### Volume mounts

```yaml
volumes:
  - .:/app/service:ro                    # Current dir → /app/service
  - ../../../logs:/app/logs              # Root logs → /app/logs
```

**Cách hoạt động:**
- Khi ở `src/service/academic/`:
  - `.` = `/home/.../src/service/academic/` → mount vào `/app/service/`
  - `../../../logs` = `/home/.../logs/` → mount vào `/app/logs/`

### Certificate paths

Code tự động tìm cert theo thứ tự:

1. **Env var** (cao nhất): `SERVICE_CERT_PATH=/custom/path`
2. **Docker**: `/app/service/` (từ volume mount)
3. **Local**: `$HOME/code/heheheh_be/src/service/academic/`

## Test Setup

### Test build một service:

```bash
cd src/service/academic
docker-compose build
```

**Sẽ:**
1. Context = `../../../` (root project)
2. Dockerfile = `src/service/academic/Dockerfile`
3. Copy go.mod, source code
4. Build: `go build -o academic-service ./src/service/academic`
5. Create image: `thaily/lvtn-academic:latest`

### Test run:

```bash
cd src/service/academic
docker-compose up -d
```

**Container sẽ:**
1. Mount `./` → `/app/service/` (có certs)
2. Mount `../../../logs/` → `/app/logs/`
3. Load env từ `academic.env`
4. Chạy binary `./academic-service`
5. TLS tìm cert tại `/app/service/*.crt`

## Kiểm tra

### 1. Verify Dockerfile paths:

```bash
cd src/service/academic
grep "go build" Dockerfile
# Output: RUN ... go build -o academic-service ./src/service/academic
```

### 2. Verify docker-compose:

```bash
grep "dockerfile:" docker-compose.yml
# Output: dockerfile: src/service/academic/Dockerfile

grep "image:" docker-compose.yml
# Output: image: thaily/lvtn-academic:latest
```

### 3. Verify tls.go:

```bash
grep "src.*service" ../pkg/tls/tls.go
# Output: basePath = filepath.Join(..., "src", "service", serviceName)
```

## GitHub Actions

File `.github/workflows/dabe.yml` đã đúng:

```yaml
strategy:
  matrix:
    service: [academic, council, file, role, thesis, user]

images: ${{ env.DOCKER_USERNAME }}/lvtn-${{ matrix.service }}
# Sẽ tạo: thaily/lvtn-academic, thaily/lvtn-council, ...
```

## Quick Commands

```bash
# Build tất cả
for service in academic council file role thesis user; do
  cd src/service/$service
  docker-compose build
  cd ../../..
done

# Run một service
cd src/service/academic
docker-compose up -d
docker-compose logs -f

# Stop
docker-compose down

# Push code và trigger GitHub Actions
git add .
git commit -m "Fixed Docker setup"
git push origin main
```

## Troubleshooting

### Build failed: "cannot find package"

**Nguyên nhân:** Context không đúng
**Fix:** Đảm bảo `context: ../../..` (3 levels up)

### Build failed: "no Go files"

**Nguyên nhân:** Build path sai
**Fix:** `go build -o service ./src/service/academic` (không phải `.`)

### Cert not found khi chạy

**Nguyên nhân:** Volume mount sai
**Fix:** `- .:/app/service:ro` (mount current directory)

### Image name sai trên Docker Hub

**Nguyên nhân:** Thiếu prefix `lvtn-`
**Fix:** `image: thaily/lvtn-academic:latest`

## Summary

✅ Tất cả 6 services đã được sửa:
- ✅ docker-compose.yml: Context, dockerfile path, image name, env_file, volumes
- ✅ Dockerfile: Build path, mkdir paths
- ✅ tls.go: Local cert path với `src/service/`
- ✅ Consistency: Tất cả services đều giống nhau

Sẵn sàng deploy! 🚀
