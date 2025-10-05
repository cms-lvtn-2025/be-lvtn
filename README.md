# Thesis Management Backend

Hệ thống backend quản lý luận văn sử dụng gRPC microservices.

## Cấu trúc dự án

```
├── env/                      # Environment files cho từng service
│   ├── academic.env
│   ├── council.env
│   ├── file.env
│   ├── role.env
│   ├── thesis.env
│   └── user.env
├── docker/                   # Dockerfiles cho từng service
│   ├── academic.Dockerfile
│   ├── council.Dockerfile
│   ├── file.Dockerfile
│   ├── role.Dockerfile
│   ├── thesis.Dockerfile
│   └── user.Dockerfile
├── proto/                    # Protocol Buffer definitions
├── src/
│   ├── pkg/
│   │   ├── config/          # Config loader
│   │   └── database/        # Database connection
│   └── service/
│       ├── academic/        # Academic service (Semester, Faculty, Major)
│       ├── council/         # Council service
│       ├── file/           # File service
│       ├── role/           # Role service
│       ├── thesis/         # Thesis service
│       └── user/           # User service
└── docker-compose.yml       # Docker compose file

```

## Services

| Service  | Port  | Entities |
|----------|-------|----------|
| Academic | 50051 | Semester, Faculty, Major |
| Council  | 50052 | Council, Defence, GradeDefences, CouncilsSchedule |
| File     | 50053 | File |
| Role     | 50054 | RoleSystem |
| Thesis   | 50055 | Topic, Midterm, Final, Enrollment |
| User     | 50056 | Student, Teacher |

## Cài đặt

### Prerequisites

- Go 1.24+
- Protocol Buffer compiler (protoc)
- Docker & Docker Compose (optional)

### Makefile Commands

```bash
# Xem tất cả commands
make help

# Generate proto & services
make all                  # Generate tất cả
make proto-academic       # Generate service riêng lẻ

# Build services
make build                # Build tất cả
make build-academic       # Build service riêng lẻ

# Docker
make docker-build         # Build tất cả Docker images
make docker-build-academic # Build image riêng lẻ
make up                   # Start docker-compose
make down                 # Stop docker-compose
make logs                 # View logs

# Run locally
make run-academic         # Run service riêng lẻ

# Clean
make clean                # Clean tất cả
make clean-build          # Clean binaries
make clean-proto          # Clean proto files
make clean-gen            # Clean generated files
```

### Quick Start

**Với Docker Compose:**
```bash
make up          # Start tất cả services
make down        # Stop tất cả services
make logs        # View logs
```

**Chạy local:**
```bash
# 1. Cấu hình DB trong env/<service>.env
# 2. Build và chạy
make build-academic
cd src/service/academic && ./academic

# Hoặc run trực tiếp
make run-academic
```

## Development

### Thêm service mới

1. Tạo proto file trong `proto/<service_name>/<service_name>.proto`
2. Thêm vào Makefile:
```makefile
proto-<service_name>:
	$(PROTOC) proto/<service_name>/<service_name>.proto
	$(GEN) <service_name> <ServiceName> <port>
```
3. Chạy `make proto-<service_name>`

### Cấu trúc handler

Mỗi service có:
- `main.go` - Entry point, load config, connect DB
- `handler/handler.go` - Root handler với DB helpers
- `handler/<entity>.go` - Methods cho từng entity

Helper methods có sẵn:
```go
h.getDB()
h.execQuery(ctx, query, args...)
h.queryRow(ctx, query, args...)
h.query(ctx, query, args...)
```

## License

MIT
