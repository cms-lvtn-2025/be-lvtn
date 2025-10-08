# Thesis Management System - Backend

Hệ thống backend quản lý luận văn sử dụng **Microservices Architecture** với GraphQL/REST API Gateway, gRPC, Redis, và Minio.

---

## 📚 Documentation

### 📖 Architecture & Design
- **[SYSTEM_ARCHITECTURE.md](./SYSTEM_ARCHITECTURE.md)** - Kiến trúc tổng quan hệ thống (diagrams, components, data flow)
- **[CLEAN_ARCHITECTURE.md](./CLEAN_ARCHITECTURE.md)** - Clean code implementation guide
- **[README_ARCHITECTURE.md](./README_ARCHITECTURE.md)** - Chi tiết cấu trúc code & best practices

### 🔌 Integration Guides
- **[MINIO_INTEGRATION.md](./MINIO_INTEGRATION.md)** - Object storage integration
- **[REDIS_INTEGRATION.md](./REDIS_INTEGRATION.md)** - Caching & BullMQ job queue

### ⚙️ Setup & Configuration
- **[.env.example](./.env.example)** - Environment variables template
- **[docker-compose.example.yml](./docker-compose.example.yml)** - Infrastructure setup

---

## 🏗️ High-Level Architecture

```
┌──────────────────────────────────────────────────┐
│         Client Layer (Web/Mobile/Admin)          │
└───────────────────┬──────────────────────────────┘
                    │ HTTP/WebSocket
┌───────────────────▼──────────────────────────────┐
│          Backend Gateway (Port 8081)             │
│  ┌──────────────┐  ┌──────────────┐             │
│  │   GraphQL    │  │   REST API   │             │
│  │    /query    │  │   /api/v1    │             │
│  └──────────────┘  └──────────────┘             │
│  ┌──────────────┐  ┌──────────────┐             │
│  │ Redis Cache  │  │    Minio     │             │
│  └──────────────┘  └──────────────┘             │
└───────────────────┬──────────────────────────────┘
                    │ gRPC
┌───────────────────▼──────────────────────────────┐
│              Microservices Layer                 │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐            │
│  │  User   │ │Academic │ │ Thesis  │            │
│  │  :5001  │ │  :5002  │ │  :5003  │            │
│  └─────────┘ └─────────┘ └─────────┘            │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐            │
│  │  File   │ │ Council │ │  Role   │            │
│  │  :5004  │ │  :5005  │ │  :5006  │            │
│  └─────────┘ └─────────┘ └─────────┘            │
│  ┌──────────────────────────────────┐            │
│  │   Redis Queue (BullMQ Jobs)      │            │
│  └──────────────────────────────────┘            │
└───────────────────┬──────────────────────────────┘
                    │
┌───────────────────▼──────────────────────────────┐
│          Database Layer (MySQL)                  │
│  Per-service databases (user_db, academic_db...) │
└──────────────────────────────────────────────────┘
```

---

## 🎯 Services Overview

### Backend Gateway (Go - Port 8081)

**Mục đích**: API Gateway cho external clients

**APIs**:
- **GraphQL**: `/query` - Single endpoint
- **REST**: `/api/v1/*` - RESTful APIs
- **Playground**: `/` - API documentation

**Features**:
- Authentication (JWT, OAuth)
- Request routing
- Response aggregation
- Caching (Redis)
- File upload (Minio)

### Microservices (gRPC)

| Service | Port | Entities | Responsibilities |
|---------|------|----------|------------------|
| **User** | 5001 | Student, Teacher | Authentication, user management |
| **Academic** | 5002 | Semester, Faculty, Major | Academic data management |
| **Thesis** | 5003 | Topic, Midterm, Final, Enrollment | Thesis workflow |
| **File** | 5004 | File | File metadata, Minio integration |
| **Council** | 5005 | Council, Defence, GradeDefences | Defense management |
| **Role** | 5006 | RoleSystem | RBAC, permissions |

---

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Node.js (for BullMQ services)
- Protocol Buffer compiler (protoc)

### 1. Setup Environment

```bash
# Clone repository
git clone <repository-url>
cd heheheh_be

# Copy và configure env
cp .env.example .server.env
# Edit .server.env với credentials của bạn
```

### 2. Start Infrastructure

```bash
# Copy docker compose
cp docker-compose.example.yml docker-compose.yml

# Start Redis, Minio, MySQL
docker-compose up -d redis-cache redis-queue minio minio-setup

# Hoặc start tất cả
docker-compose up -d
```

### 3. Run Backend Gateway

```bash
cd src/server
go run server.go
```

Server running at:
- 🎮 GraphQL Playground: http://localhost:8081/
- 🔌 GraphQL Endpoint: http://localhost:8081/query
- 🌐 REST API: http://localhost:8081/api/v1

### 4. Access Services

**Minio Console**: http://localhost:9001
- User: `minioadmin`
- Pass: `minioadmin`

**Redis Cache**: `localhost:6379`
**Redis Queue**: `localhost:6380`

---

## 📁 Project Structure

```
heheheh_be/
├── src/
│   ├── server/
│   │   ├── server.go          # ⭐ Entry point (35 dòng)
│   │   └── client/            # gRPC clients
│   ├── config/
│   │   └── config.go          # Configuration management
│   ├── router/
│   │   └── router.go          # Routes setup (GraphQL + REST)
│   ├── api/                   # REST API handlers
│   │   ├── handler.go         # DI handler
│   │   ├── middleware.go      # Auth middleware
│   │   ├── auth.go            # Google OAuth
│   │   ├── user.go            # User endpoints
│   │   └── file.go            # File endpoints
│   ├── pkg/                   # Shared packages
│   │   ├── container/         # Dependency injection
│   │   ├── response/          # Unified API response
│   │   ├── cache/             # Redis cache client
│   │   └── storage/           # Minio client
│   ├── graph/                 # GraphQL
│   │   ├── schema/
│   │   ├── resolver/
│   │   ├── controller/
│   │   └── generated/
│   └── service/               # Microservices
│       ├── user/              # User service (gRPC)
│       ├── academic/          # Academic service
│       ├── thesis/            # Thesis service
│       ├── file/              # File service
│       ├── council/           # Council service
│       └── role/              # Role service
├── proto/                     # Protocol Buffers definitions
├── env/                       # Service env files
├── docker/                    # Dockerfiles
├── docs/                      # Documentation
├── .server.env               # Gateway env
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## 🔌 API Examples

### GraphQL

```graphql
# Query
query {
  getUser(id: "123") {
    id
    name
    email
    theses {
      id
      title
      status
    }
  }
}

# Mutation
mutation {
  createThesis(input: {
    title: "AI in Healthcare"
    description: "Research on AI applications"
  }) {
    id
    title
    createdAt
  }
}
```

### REST API

```bash
# Google OAuth login
POST /api/v1/auth/google/login

# Get current user (authenticated)
GET /api/v1/users/me
Authorization: Bearer <token>

# Update profile
PUT /api/v1/users/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "John Doe",
  "avatar": "https://..."
}

# Upload file
POST /api/v1/files/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <binary>
bucket: thesis
```

---

## 🛠️ Development

### Makefile Commands

```bash
# View all commands
make help

# Generate Proto & Services
make all                    # Generate tất cả
make proto-user             # Generate proto cho user service
make proto-thesis           # Generate proto cho thesis service

# Build Services
make build                  # Build tất cả services
make build-user             # Build user service
make build-gateway          # Build backend gateway

# Run Services
make run-user               # Run user service
make run-gateway            # Run backend gateway

# Docker
make docker-build           # Build tất cả images
make docker-build-user      # Build user service image
make up                     # Start docker-compose
make down                   # Stop docker-compose
make logs                   # View logs

# Clean
make clean                  # Clean everything
make clean-build            # Clean binaries
make clean-proto            # Clean proto generated files
```

### Generate Code

```bash
# GraphQL
cd src/graph
go run github.com/99designs/gqlgen generate

# gRPC Proto
make proto-<service>

# Or manually
protoc --go_out=. --go-grpc_out=. proto/<service>/<service>.proto
```

### Run Tests

```bash
# All tests
go test ./...

# Specific package
go test ./src/api/...

# With coverage
go test -cover ./...

# Integration tests
go test -tags=integration ./...
```

---

## 🐳 Docker Deployment

### Full Stack

```bash
# Start everything
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down

# Rebuild
docker-compose up -d --build
```

### Selective Services

```bash
# Infrastructure only
docker-compose up -d redis-cache redis-queue minio

# Backend gateway only
docker-compose up -d backend-gateway

# Specific microservice
docker-compose up -d user-service thesis-service
```

---

## ⚙️ Configuration

### Environment Variables

Tất cả config trong `.server.env`:

```env
# Server
PORT=8081
GIN_MODE=release

# Services (gRPC endpoints)
SERVICE_USER=localhost:5001
SERVICE_ACADEMIC=localhost:5002
SERVICE_THESIS=localhost:5003
SERVICE_FILE=localhost:5004
SERVICE_COUNCIL=localhost:5005
SERVICE_ROLE=localhost:5006

# Redis
REDIS_CACHE_HOST=localhost
REDIS_CACHE_PORT=6379
REDIS_QUEUE_HOST=localhost
REDIS_QUEUE_PORT=6380

# Minio
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
```

Xem [.env.example](./.env.example) cho chi tiết đầy đủ.

---

## 🎨 Design Patterns

1. **Microservices** - Independent, scalable services
2. **API Gateway** - Single entry point
3. **Database per Service** - Data isolation
4. **Dependency Injection** - Container pattern (Go)
5. **Repository Pattern** - Data access layer
6. **CQRS** - GraphQL for queries, REST for commands
7. **Event-Driven** - BullMQ background jobs
8. **Caching** - Redis multi-level caching

---

## 🔐 Security

- **Authentication**: JWT + OAuth 2.0 (Google)
- **Authorization**: RBAC with Role Service
- **Transport**: TLS/SSL (external), mTLS (gRPC)
- **Data**: Encryption at rest, secure file storage
- **Secrets**: Environment variables, no hardcoded

---

## 📈 Performance

### Caching Strategy
- Application cache (Redis) - 6379
- Tag-based invalidation
- Query result caching

### Database Optimization
- Connection pooling
- Database per service
- Indexes on frequently queried fields

### File Storage
- Direct upload to Minio
- Pre-signed URLs
- Bucket policies

---

## 📊 Monitoring & Health Checks

```bash
# Backend Gateway health
curl http://localhost:8081/health

# Minio health
curl http://localhost:9000/minio/health/live

# Redis cache
redis-cli -p 6379 ping

# Redis queue
redis-cli -p 6380 ping

# View logs
docker-compose logs -f backend-gateway
docker-compose logs -f user-service
```

---

## 🗺️ Roadmap

### ✅ Phase 1 (Current)
- [x] Backend Gateway (GraphQL + REST)
- [x] Clean architecture implementation
- [x] Microservices structure
- [x] Basic authentication
- [ ] Minio integration (in progress)
- [ ] Redis caching (in progress)

### 🎯 Phase 2
- [ ] Complete all microservices
- [ ] BullMQ job processing
- [ ] Email notifications
- [ ] Advanced RBAC
- [ ] File upload/download

### 🚀 Phase 3
- [ ] Monitoring (Prometheus + Grafana)
- [ ] Distributed tracing (Jaeger)
- [ ] Performance optimization
- [ ] Rate limiting
- [ ] API versioning

### 🌟 Phase 4
- [ ] Service mesh (Istio)
- [ ] Event sourcing
- [ ] Real-time features
- [ ] CDN integration

---

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Commit changes
4. Push to branch
5. Open Pull Request

---

## 📝 License

MIT

---

## 📞 Support

- 📖 Documentation: See docs folder
- 🐛 Issues: GitHub Issues
- 💬 Discussions: GitHub Discussions

---

## 🙏 Acknowledgments

- Built with Go, GraphQL (gqlgen), gRPC
- Infrastructure: Redis, Minio, MySQL
- Patterns: Microservices, Clean Architecture, DI

---

**Made with ❤️ for Thesis Management**
