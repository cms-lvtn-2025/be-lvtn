# Testing Guide

## Tổng quan

Dự án sử dụng Go testing framework với các loại test:
- **Unit Tests**: Test các function/method riêng lẻ
- **Integration Tests**: Test tích hợp giữa các components

## Cấu trúc Test

```
be-lvtn/
├── tests/
│   └── integration/          # Integration tests
│       ├── graphql_test.go   # GraphQL API tests
│       └── service_test.go   # gRPC service tests
│
└── src/service/
    ├── academic/tests/unit/  # Academic service unit tests
    │   ├── faculty_test.go
    │   ├── major_test.go
    │   └── semester_test.go
    ├── council/tests/unit/   # Council service unit tests
    ├── file/tests/unit/      # File service unit tests
    ├── role/tests/unit/      # Role service unit tests
    ├── thesis/tests/unit/    # Thesis service unit tests
    │   ├── topic_test.go
    │   ├── enrollment_test.go
    │   ├── midterm_test.go
    │   └── final_test.go
    └── user/tests/unit/      # User service unit tests
        ├── student_test.go
        └── teacher_test.go
```

## Chạy Tests

### Tất cả tests

```bash
go test ./... -v -count=1
```

### Unit tests theo service

```bash
# User service
go test ./src/service/user/tests/unit/... -v

# Academic service
go test ./src/service/academic/tests/unit/... -v

# Thesis service
go test ./src/service/thesis/tests/unit/... -v
```

### Integration tests

```bash
go test ./tests/integration/... -v
```

### Với race detector

```bash
go test ./... -race -v
```

### Với coverage

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out -covermode=atomic

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Coverage theo service

```bash
# User service coverage
go test ./src/service/user/tests/unit/... -coverprofile=coverage-user.out
go tool cover -func=coverage-user.out

# Academic service coverage
go test ./src/service/academic/tests/unit/... -coverprofile=coverage-academic.out
go tool cover -func=coverage-academic.out
```

### Chạy test cụ thể

```bash
# Chạy một test function
go test ./src/service/user/tests/unit/... -run TestCreateStudent_Success -v

# Chạy các test match pattern
go test ./src/service/user/tests/unit/... -run TestCreate -v
```

### Với timeout

```bash
go test ./... -timeout=5m -v
```

## CI/CD Integration

### GitHub Actions Workflows

#### 1. Tests Workflow (`.github/workflows/tests.yml`)

Tự động chạy khi:
- Push đến `main`, `develop`
- Pull request đến `main`, `develop`
- Manual trigger

**Jobs:**
- **Lint**: golangci-lint code quality check
- **Unit Tests**: Matrix strategy cho tất cả services
- **Integration Tests**: Với MySQL và Redis containers
- **Coverage Summary**: Tổng hợp coverage report

#### 2. CI Pipeline (`.github/workflows/ci.yml`)

Chạy cho Pull Requests và pushes đến `develop`, `test`:

**Jobs:**
- **Validate**: Dependencies, formatting, go vet
- **Test**: Gọi tests workflow
- **Coverage Check**: Đảm bảo coverage >= 70%
- **Security Scan**: Gosec security scanner
- **Build Check**: Verify build cho tất cả services

#### 3. DABE Workflow (`.github/workflows/dabe.yml`)

Chạy khi push đến `main`:
1. ✅ Run all tests (MUST PASS)
2. 🐳 Build & push Docker images

### Chạy Tests Locally Như CI

```bash
# Validate
go mod download
go mod verify
gofmt -l .
go vet ./...

# All tests
go test ./... -v -race -timeout=5m

# Coverage check
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

## Test Best Practices

### Viết Unit Tests

```go
func TestCreateStudent_Success(t *testing.T) {
    // Arrange
    mockDB, mock, _ := sqlmock.New()
    defer mockDB.Close()
    
    mock.ExpectExec("INSERT INTO students").
        WillReturnResult(sqlmock.NewResult(1, 1))
    
    // Act
    result, err := CreateStudent(ctx, req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

### Table-Driven Tests

```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid email", "test@example.com", false},
        {"invalid email", "invalid", true},
        {"empty email", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Mocking

Project sử dụng:
- `github.com/DATA-DOG/go-sqlmock` cho database mocking
- `github.com/stretchr/testify/mock` cho general mocking

```go
// Mock database expectations
mock.ExpectQuery("SELECT (.+) FROM users").
    WithArgs(userID).
    WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
        AddRow(1, "Test User"))
```

## Test Coverage

### Mục tiêu
- Minimum: **70%** tổng coverage
- Target: **80%+** cho business logic

### Xem Coverage Report

```bash
# Generate và open HTML report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Coverage by Package

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep -E "service/(user|academic|thesis)"
```

## Debugging Tests

### Verbose output

```bash
go test ./... -v
```

### Print statements

```go
t.Logf("Debug: value = %v", value)
```

### Skip tests

```go
func TestSomething(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping in short mode")
    }
    // test code
}
```

Chạy với `-short`:
```bash
go test ./... -short
```

### Clear test cache

```bash
go clean -testcache
go test ./... -count=1
```

## Integration Tests Setup

### Requirements

- MySQL 8.0
- Redis 7

### Environment Variables

```bash
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=root
DB_NAME=test_db
REDIS_HOST=localhost
REDIS_PORT=6379
```

### Với Docker

```bash
# Start services
docker-compose -f src/server/docker-compose.yml up -d

# Run integration tests
go test ./tests/integration/... -v

# Stop services
docker-compose -f src/server/docker-compose.yml down
```

## Troubleshooting

### Tests fail với "connection refused"

Check services đang chạy:
```bash
docker ps
netstat -an | grep 3306
netstat -an | grep 6379
```

### Tests timeout

Tăng timeout:
```bash
go test ./... -timeout=10m
```

### Coverage không đúng

Clear cache và chạy lại:
```bash
go clean -testcache
go test ./... -coverprofile=coverage.out -count=1
```

### Race conditions

Luôn chạy với race detector khi develop:
```bash
go test ./... -race
```

## Continuous Improvement

- [ ] Tăng coverage lên 80%+
- [ ] Thêm benchmark tests
- [ ] Thêm E2E tests
- [ ] Load/stress testing
- [ ] Mutation testing
- [ ] Property-based testing

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)
- [testify](https://github.com/stretchr/testify)
- [GitHub Actions](https://docs.github.com/en/actions)
