# System Architecture Documentation

## Tổng quan hệ thống

Hệ thống được thiết kế theo kiến trúc **Microservices** với **API Gateway Pattern**, sử dụng **gRPC** cho internal communication và **GraphQL/REST** cho external APIs.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT LAYER                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Web App    │  │  Mobile App  │  │  Admin Panel │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└────────────┬────────────────┬────────────────┬──────────────────┘
             │                │                │
             │ HTTP/WS        │ HTTP/WS        │ HTTP/WS
             └────────────────┴────────────────┘
                              │
┌─────────────────────────────▼─────────────────────────────────┐
│                    BACKEND GATEWAY LAYER                       │
│  ┌────────────────────────────────────────────────────────┐   │
│  │              Backend Gateway (Port 8081)                │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │   │
│  │  │   GraphQL    │  │   REST API   │  │  Playground  │ │   │
│  │  │   /query     │  │   /api/v1    │  │      /       │ │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘ │   │
│  │                                                         │   │
│  │  ┌──────────────┐  ┌──────────────┐                   │   │
│  │  │ Redis Cache  │  │    Minio     │                   │   │
│  │  │ (Caching)    │  │ (File Store) │                   │   │
│  │  └──────────────┘  └──────────────┘                   │   │
│  └────────────────────────────────────────────────────────┘   │
└────────────┬─────────┬─────────┬──────────┬──────────┬────────┘
             │         │         │          │          │
             │ gRPC    │ gRPC    │ gRPC     │ gRPC     │ gRPC
             │         │         │          │          │
┌────────────▼─────────▼─────────▼──────────▼──────────▼────────┐
│                    MICROSERVICES LAYER                         │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐         │
│  │   User   │ │ Academic │ │  Thesis  │ │   File   │         │
│  │ Service  │ │ Service  │ │ Service  │ │ Service  │         │
│  │  :5001   │ │  :5002   │ │  :5003   │ │  :5004   │         │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘         │
│       │            │            │            │                │
│  ┌──────────┐ ┌──────────┐                                    │
│  │ Council  │ │   Role   │                                    │
│  │ Service  │ │ Service  │                                    │
│  │  :5005   │ │  :5006   │                                    │
│  └────┬─────┘ └────┬─────┘                                    │
│       │            │                                           │
│  ┌────▼────────────▼────────────────────────────────┐         │
│  │           Redis-Minio (BullMQ - Minio)           │         │
│  │              (Background Jobs)                   │         │
│  └──────────────────────────────────────────────────┘         │
└────────────┬─────────┬─────────┬──────────┬──────────────────┘
             │         │         │          │
             │         │         │          │
┌────────────▼─────────▼─────────▼──────────▼──────────────────┐
│                      DATABASE LAYER                           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │
│  │   User   │ │ Academic │ │  Thesis  │ │   File   │        │
│  │    DB    │ │    DB    │ │    DB    │ │    DB    │        │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │
│                                                                │
│  ┌──────────┐ ┌──────────┐                                   │
│  │ Council  │ │   Role   │                                   │
│  │    DB    │ │    DB    │                                   │
│  └──────────┘ └──────────┘                                   │
└────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Client Layer

**Mục đích**: User interface cho end users

**Components**:
- **Web Application**: Frontend web (React, Vue, Angular)
- **Mobile App**: iOS/Android apps
- **Admin Panel**: Dashboard quản trị

**Protocols**: HTTP/HTTPS, WebSocket (cho real-time)

---

### 2. Backend Gateway Layer

**Mục đích**: API Gateway, routing, authentication, caching

#### 2.1 Backend Gateway Service

**Port**: 8081
**Language**: Go
**Framework**: Gin + gqlgen

**Responsibilities**:
- Expose GraphQL API (`/query`)
- Expose REST API (`/api/v1`)
- Authentication & Authorization (JWT)
- Request routing to microservices
- Response aggregation
- Rate limiting
- API documentation (Playground)

**APIs**:

**GraphQL** (`/query`):
- Single endpoint cho tất cả queries/mutations
- Real-time với subscriptions (WebSocket)
- Authentication required

**REST API** (`/api/v1`):
- `/auth/google/login` - OAuth login
- `/auth/google/callback` - OAuth callback
- `/users/*` - User management
- `/files/*` - File operations

**GraphQL Playground** (`/`):
- Interactive API documentation
- Query builder & testing

#### 2.2 Redis Cache

**Purpose**: Application-level caching

**Use cases**:
- Cache GraphQL query results
- Cache user sessions
- Cache frequently accessed data
- Temporary data storage (OTP, tokens)

**Configuration**:
```env
REDIS_CACHE_HOST=localhost
REDIS_CACHE_PORT=6379
REDIS_CACHE_DB=0
```

#### 2.3 Minio

**Purpose**: Object storage for files

**Use cases**:
- Upload user avatars
- Store document files (PDF, DOCX)
- Store thesis files
- Store academic documents
- Image storage

**Buckets**:
- `avatars` - User profile pictures
- `documents` - General documents
- `thesis` - Thesis files
- `academic` - Academic materials

**Configuration**:
```env
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
```

---

### 3. Microservices Layer

Mỗi service là một ứng dụng độc lập, có database riêng (Database per Service pattern).

#### 3.1 User Service

**Port**: 5001
**Database**: MySQL/PostgreSQL

**Responsibilities**:
- User authentication & authorization
- User profile management
- Role-based access control
- User preferences

**gRPC Methods**:
```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc CreateUser(CreateUserRequest) returns (User);
  rpc UpdateUser(UpdateUserRequest) returns (User);
  rpc DeleteUser(DeleteUserRequest) returns (Empty);
  rpc AuthenticateUser(AuthRequest) returns (AuthResponse);
}
```

#### 3.2 Academic Service

**Port**: 5002
**Database**: MySQL/PostgreSQL

**Responsibilities**:
- Academic year management
- Semester management
- Course management
- Department management

**gRPC Methods**:
```protobuf
service AcademicService {
  rpc GetAcademicYear(GetAcademicYearRequest) returns (AcademicYear);
  rpc CreateCourse(CreateCourseRequest) returns (Course);
  rpc GetDepartments(GetDepartmentsRequest) returns (DepartmentList);
}
```

#### 3.3 Thesis Service

**Port**: 5003
**Database**: MySQL/PostgreSQL

**Responsibilities**:
- Thesis registration
- Thesis submission
- Thesis review workflow
- Thesis status tracking

**gRPC Methods**:
```protobuf
service ThesisService {
  rpc CreateThesis(CreateThesisRequest) returns (Thesis);
  rpc UpdateThesisStatus(UpdateStatusRequest) returns (Thesis);
  rpc GetThesis(GetThesisRequest) returns (Thesis);
  rpc SubmitThesis(SubmitThesisRequest) returns (Thesis);
}
```

#### 3.4 File Service

**Port**: 5004
**Database**: MySQL/PostgreSQL (metadata only)

**Responsibilities**:
- File metadata management
- File upload coordination với Minio
- File access control
- File versioning

**gRPC Methods**:
```protobuf
service FileService {
  rpc UploadFile(stream FileChunk) returns (FileMetadata);
  rpc GetFile(GetFileRequest) returns (FileMetadata);
  rpc DeleteFile(DeleteFileRequest) returns (Empty);
  rpc GetFileURL(GetFileURLRequest) returns (FileURL);
}
```

**Data Flow**:
```
Client → Backend Gateway → File Service → Minio
                                ↓
                           Database (metadata)
```

#### 3.5 Council Service

**Port**: 5005
**Database**: MySQL/PostgreSQL

**Responsibilities**:
- Council member management
- Council meeting scheduling
- Thesis defense scheduling
- Review assignment

#### 3.6 Role Service

**Port**: 5006
**Database**: MySQL/PostgreSQL

**Responsibilities**:
- Role definition
- Permission management
- Role assignment
- Access control rules

#### 3.7 Redis BullMQ

**Purpose**: Background job processing

**Use cases**:
- Send emails (thesis submission notification)
- Generate reports (PDF generation)
- Data synchronization
- Scheduled tasks (deadline reminders)
- Batch processing

**Job Types**:
- `email:send` - Send email notifications
- `thesis:deadline-reminder` - Thesis deadline reminders
- `report:generate` - Generate PDF reports
- `data:sync` - Sync data between services

**Configuration**:
```env
REDIS_QUEUE_HOST=localhost
REDIS_QUEUE_PORT=6380
REDIS_QUEUE_DB=0
```

**Example Job Flow**:
```
Thesis Service → Add job to BullMQ → Worker picks up job
                                           ↓
                                    Send email notification
```

---

### 4. Database Layer

**Strategy**: Database per Service (Microservices best practice)

**Technology**: MySQL hoặc PostgreSQL

**Databases**:
- `user_db` - User service data
- `academic_db` - Academic service data
- `thesis_db` - Thesis service data
- `file_db` - File metadata
- `council_db` - Council service data
- `role_db` - Role service data

**Benefits**:
- Service independence
- Schema evolution freedom
- Technology heterogeneity
- Fault isolation

**Challenges**:
- Distributed transactions (use Saga pattern)
- Data consistency (eventual consistency)
- Cross-service queries (API composition)

---

## Communication Patterns

### 1. Client ↔ Backend Gateway

**Protocol**: HTTP/HTTPS, WebSocket
**Format**: JSON (REST), GraphQL queries

**REST Example**:
```bash
POST /api/v1/auth/google/login
Content-Type: application/json

{
  "redirect_uri": "http://localhost:3000/callback"
}
```

**GraphQL Example**:
```graphql
query {
  getUser(id: "123") {
    id
    name
    email
  }
}
```

### 2. Backend Gateway ↔ Microservices

**Protocol**: gRPC (HTTP/2)
**Format**: Protocol Buffers

**Example**:
```go
// Backend Gateway
userClient := userService.GetUser(ctx, &pb.GetUserRequest{
    UserId: "123",
})
```

**Benefits**:
- High performance (binary format)
- Strong typing
- Code generation
- Bidirectional streaming

### 3. Microservices ↔ Database

**Protocol**: Native database drivers
**ORM**: GORM (Go), TypeORM (Node.js)

### 4. Service ↔ Redis

**Caching (Backend Gateway)**:
```go
// Set cache
cache.Set("user:123", userData, 10*time.Minute)

// Get cache
data, err := cache.Get("user:123")
```

**Job Queue (Services)**:
```go
// Add job
queue.Add("email:send", {
    to: "user@example.com",
    subject: "Thesis Submitted",
    template: "thesis_notification"
})

// Worker
queue.Process("email:send", async (job) => {
    await sendEmail(job.data)
})
```

### 5. Service ↔ Minio

**Upload Flow**:
```go
// 1. Client uploads to Backend Gateway
POST /api/v1/files/upload

// 2. Backend Gateway calls File Service
fileService.UploadFile(stream)

// 3. File Service uploads to Minio
minio.PutObject(bucket, objectName, reader)

// 4. File Service saves metadata to DB
db.Save(fileMetadata)

// 5. Return file URL to client
```

---

## Technology Stack

### Backend Gateway
- **Language**: Go 1.24
- **Framework**: Gin (REST), gqlgen (GraphQL)
- **Authentication**: JWT (golang-jwt)
- **Cache**: Redis
- **Storage**: Minio
- **Communication**: gRPC clients

### Microservices
- **Language**: Go (hoặc Node.js tùy service)
- **Framework**: gRPC
- **Database**: MySQL/PostgreSQL
- **ORM**: GORM (Go), TypeORM (Node.js)
- **Job Queue**: BullMQ (Node.js)

### Infrastructure
- **Redis**: 2 instances
  - Cache (port 6379)
  - Queue (port 6380)
- **Minio**: Object storage (port 9000)
- **Database**: MySQL/PostgreSQL
- **Container**: Docker, Docker Compose

---

## Data Flow Examples

### Example 1: User Login với Google OAuth

```
1. Client → Backend Gateway
   POST /api/v1/auth/google/login

2. Backend Gateway → Client
   Return OAuth URL

3. Client → Google OAuth
   User authorizes

4. Google → Client
   Redirect với code

5. Client → Backend Gateway
   POST /api/v1/auth/google/callback { code }

6. Backend Gateway → Google
   Exchange code for token

7. Backend Gateway → User Service (gRPC)
   CreateUser() hoặc GetUser()

8. User Service → Database
   Save/Get user data

9. Backend Gateway → Client
   Return JWT token

10. Backend Gateway → Redis Cache
    Cache user session
```

### Example 2: Upload Thesis File

```
1. Client → Backend Gateway
   POST /api/v1/files/upload
   multipart/form-data

2. Backend Gateway → File Service (gRPC)
   UploadFile(stream)

3. File Service → Minio
   PutObject(bucket: "thesis", file)

4. File Service → Database
   Save file metadata (name, size, url, owner)

5. File Service → Backend Gateway
   Return file metadata

6. Backend Gateway → Thesis Service (gRPC)
   UpdateThesis(fileId)

7. Thesis Service → Redis BullMQ
   Add job: "email:send" - notify advisor

8. BullMQ Worker
   Pick up job & send email

9. Backend Gateway → Client
   Return success response
```

### Example 3: Get Thesis với Advisor Info (Aggregation)

```
1. Client → Backend Gateway
   GraphQL query {
     thesis(id: "123") {
       title
       file { url }
       advisor { name, email }
     }
   }

2. Backend Gateway → Redis Cache
   Check cache for "thesis:123"

3. If cache miss:

   a. Backend Gateway → Thesis Service (gRPC)
      GetThesis(id: "123")
      Response: { title, fileId, advisorId }

   b. Backend Gateway → File Service (gRPC)
      GetFile(id: fileId)
      Response: { url }

   c. Backend Gateway → User Service (gRPC)
      GetUser(id: advisorId)
      Response: { name, email }

   d. Backend Gateway
      Aggregate responses

   e. Backend Gateway → Redis Cache
      Cache aggregated data

4. Backend Gateway → Client
   Return complete data
```

---

## Security Considerations

### 1. Authentication
- JWT tokens (short-lived access token + refresh token)
- OAuth 2.0 (Google, Facebook)
- Session management với Redis

### 2. Authorization
- Role-Based Access Control (RBAC)
- Permission checks ở Gateway level
- Service-level authorization

### 3. Data Security
- TLS/SSL cho external communication
- mTLS cho gRPC internal communication
- Database encryption at rest
- Secrets management (env variables, vault)

### 4. File Security
- Minio bucket policies
- Pre-signed URLs (temporary access)
- File type validation
- Size limits

---

## Scalability & Performance

### Horizontal Scaling
- Backend Gateway: Multiple instances behind load balancer
- Microservices: Independent scaling per service
- Database: Read replicas

### Caching Strategy
- Redis cache ở gateway layer (hot data)
- Query result caching
- Session caching

### Job Processing
- BullMQ workers scale independently
- Job prioritization
- Retry mechanisms

---

## Monitoring & Logging

### Metrics
- Request rate, latency
- Service health checks
- Database connection pools
- Cache hit/miss ratio

### Logging
- Structured logging (JSON)
- Centralized log aggregation
- Request tracing (correlation IDs)

### Tools
- Prometheus (metrics)
- Grafana (visualization)
- ELK Stack (logging)
- Jaeger (distributed tracing)

---

## Deployment Architecture

```yaml
# docker-compose.yml structure
services:
  # Gateway
  backend-gateway:
    ports: ["8081:8081"]
    depends_on: [redis-cache, minio]

  # Storage
  redis-cache:
    ports: ["6379:6379"]

  redis-queue:
    ports: ["6380:6379"]

  minio:
    ports: ["9000:9000", "9001:9001"]

  # Microservices
  user-service:
    ports: ["5001:5001"]
    depends_on: [user-db, redis-queue]

  academic-service:
    ports: ["5002:5002"]

  thesis-service:
    ports: ["5003:5003"]

  file-service:
    ports: ["5004:5004"]

  council-service:
    ports: ["5005:5005"]

  role-service:
    ports: ["5006:5006"]

  # Databases
  user-db:
    image: mysql:8

  academic-db:
    image: mysql:8

  # ... other DBs
```

---

## Environment Variables

Xem file `.env.example` để có danh sách đầy đủ các biến môi trường cần thiết.

**Key Variables**:
```env
# Backend Gateway
PORT=8081
GIN_MODE=release

# Services
SERVICE_USER=user-service:5001
SERVICE_ACADEMIC=academic-service:5002
SERVICE_THESIS=thesis-service:5003
SERVICE_FILE=file-service:5004
SERVICE_COUNCIL=council-service:5005
SERVICE_ROLE=role-service:5006

# Redis
REDIS_CACHE_URL=redis://redis-cache:6379/0
REDIS_QUEUE_URL=redis://redis-queue:6379/0

# Minio
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# Database
DB_HOST=user-db
DB_PORT=3306
DB_USER=root
DB_PASS=password
```

---

## Future Enhancements

