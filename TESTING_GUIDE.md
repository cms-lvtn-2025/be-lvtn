# Testing Guide for LVTN Project

## 🎯 **Overview**
Comprehensive testing strategy for the LVTN microservices project, including unit tests, integration tests, and automated testing workflows.

## 📋 **Testing Strategy**

### Testing Pyramid
```
        🔺 End-to-End Tests
       🔺🔺 Integration Tests  
    🔺🔺🔺🔺 Unit Tests
```

### Test Types
- **Unit Tests**: Individual component testing
- **Integration Tests**: Service-to-service communication
- **GraphQL Tests**: API endpoint testing
- **Performance Tests**: Load and benchmark testing
- **Security Tests**: Vulnerability scanning

## 🧪 **Test Framework Stack**

### Core Libraries
- **Testing**: `github.com/stretchr/testify`
- **Mocking**: `github.com/DATA-DOG/go-sqlmock`
- **HTTP Testing**: `net/http/httptest`
- **gRPC Testing**: `google.golang.org/grpc/test/bufconn`

### Additional Tools
- **Coverage**: Built-in `go test -cover`
- **Benchmarking**: Built-in `go test -bench`
- **Race Detection**: `go test -race`
- **Linting**: `golangci-lint`
- **Security**: `gosec`

## 📁 **Test Organization**

```
tests/
├── unit/                 # Unit tests
│   ├── handlers/         # Handler tests  
│   ├── services/         # Business logic tests
│   └── utils/            # Utility tests
├── integration/          # Integration tests
│   ├── graphql_test.go   # GraphQL API tests
│   ├── service_test.go   # gRPC service tests
│   └── database_test.go  # Database tests
├── fixtures/             # Test data
├── mocks/                # Generated mocks
├── .env.test             # Test configuration
└── README.md             # Test documentation
```

## 🚀 **Running Tests**

### Quick Commands
```bash
# Run all tests
./run-tests.sh all

# Run unit tests only
./run-tests.sh unit

# Run integration tests
./run-tests.sh integration

# Run with coverage
go test -cover ./...

# Run specific package
go test ./src/service/academic/handler/

# Run specific test
go test -run TestHandler_GetAllAcademic ./src/service/academic/handler/
```

### Test Script Options
```bash
./run-tests.sh {unit|integration|bench|lint|security|all|ci}

# Examples:
./run-tests.sh unit        # Unit tests + coverage
./run-tests.sh integration # Integration tests
./run-tests.sh bench       # Benchmarks
./run-tests.sh lint        # Code linting
./run-tests.sh all         # Everything
./run-tests.sh ci          # CI/CD optimized
```

## 📊 **Code Coverage**

### Coverage Goals
- **Overall Project**: ≥80%
- **Critical Handlers**: ≥90%
- **Business Logic**: ≥85%
- **Utilities**: ≥75%

### Coverage Reports
```bash
# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# View coverage by function
go tool cover -func=coverage.out

# Coverage per package
go test -cover ./src/service/academic/...
```

## 🧩 **Unit Testing**

### Handler Testing Pattern
```go
func TestHandler_GetAllAcademic(t *testing.T) {
    // Setup mock database
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    handler := NewHandler(db)

    t.Run("Success case", func(t *testing.T) {
        // Setup mock expectations
        rows := sqlmock.NewRows([]string{"id", "name"}).
            AddRow("1", "Test Academic")
        mock.ExpectQuery("SELECT (.+) FROM academic").
            WillReturnRows(rows)

        // Execute
        req := &academic.GetAllAcademicRequest{}
        resp, err := handler.GetAllAcademic(context.Background(), req)

        // Assert
        assert.NoError(t, err)
        assert.Equal(t, common.Status_SUCCESS, resp.Status)
        assert.Len(t, resp.Academics, 1)

        // Verify mock expectations
        assert.NoError(t, mock.ExpectationsWereMet())
    })
}
```

### Test Structure
- **Arrange**: Setup test data and mocks
- **Act**: Execute the function under test
- **Assert**: Verify results and side effects

### Best Practices
1. **One assertion per test case**
2. **Descriptive test names**
3. **Use table-driven tests for multiple scenarios**
4. **Mock external dependencies**
5. **Test error cases**

## 🔗 **Integration Testing**

### Service Integration Tests
```go
func (suite *ServiceIntegrationTestSuite) TestAcademicServiceCRUD() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Test Create
    createReq := &academic.CreateAcademicRequest{
        UserId:        "test-user",
        AcademicTitle: "Professor",
        // ... other fields
    }
    
    createResp, err := suite.academicClient.CreateAcademic(ctx, createReq)
    require.NoError(suite.T(), err)
    
    // Test Get
    getReq := &academic.GetAcademicByIdRequest{Id: createResp.Id}
    getResp, err := suite.academicClient.GetAcademicById(ctx, getReq)
    require.NoError(suite.T(), err)
    
    // ... continue with Update and Delete tests
}
```

### GraphQL Integration Tests
```go
func (suite *GraphQLIntegrationTestSuite) TestAcademicWorkflow() {
    query := `
        mutation CreateAcademic($input: CreateAcademicInput!) {
            createAcademic(input: $input) {
                id
                academicTitle
            }
        }
    `
    
    variables := map[string]interface{}{
        "input": map[string]interface{}{
            "userId":        "test-user",
            "academicTitle": "Professor",
        },
    }
    
    resp, err := suite.executeGraphQLQuery(query, variables)
    require.NoError(suite.T(), err)
    assert.Empty(suite.T(), resp.Errors)
}
```

## 🏃 **Performance Testing**

### Benchmark Tests
```go
func BenchmarkAcademicServiceGetAll(b *testing.B) {
    // Setup
    client := setupTestClient()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := client.GetAllAcademic(context.Background(), &academic.GetAllAcademicRequest{})
            if err != nil {
                b.Error(err)
            }
        }
    })
}
```

### Load Testing
```bash
# Use Apache Bench for simple load testing
ab -n 1000 -c 10 http://localhost:8080/health

# Use hey for more advanced testing
hey -n 1000 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -d '{"query":"{ users { id name } }"}' \
  http://localhost:8080/query
```

## 🛡️ **Security Testing**

### Static Analysis
```bash
# Install gosec
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Run security scan
gosec ./...

# Check for known vulnerabilities
go list -json -m all | nancy sleuth
```

### Common Security Tests
- SQL injection prevention
- XSS protection
- Authentication bypass
- Authorization checks
- Input validation
- Rate limiting

## 🤖 **Test Automation**

### GitHub Actions Workflow
```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: 1.21
      
      - name: Run tests
        run: ./run-tests.sh ci
        
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: coverage.out
```

### Pre-commit Hooks
```bash
# Install pre-commit
pip install pre-commit

# Setup hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

## 📝 **Test Data Management**

### Fixtures and Factories
```go
// Test data factory
func CreateTestAcademic() *model.Academic {
    return &model.Academic{
        ID:            uuid.New().String(),
        UserID:        "test-user-id",
        AcademicTitle: "Test Professor",
        Degree:        stringPtr("PhD"),
        FieldOfStudy:  stringPtr("Computer Science"),
        Institution:   stringPtr("Test University"),
        YearObtained:  intPtr(2020),
        IsVerified:    false,
    }
}

// Database seeding for integration tests
func (suite *IntegrationTestSuite) SetupTest() {
    suite.seedTestData()
}

func (suite *IntegrationTestSuite) TearDownTest() {
    suite.cleanupTestData()
}
```

### Database Testing
```go
func TestWithDatabase(t *testing.T) {
    // Use testcontainers for real database testing
    ctx := context.Background()
    
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "mysql:8.0",
            ExposedPorts: []string{"3306/tcp"},
            Env: map[string]string{
                "MYSQL_ROOT_PASSWORD": "test",
                "MYSQL_DATABASE":      "test_db",
            },
        },
        Started: true,
    })
    require.NoError(t, err)
    defer container.Terminate(ctx)
    
    // Get connection details and run tests
    // ...
}
```

## 🎯 **Test Best Practices**

### DO ✅
- Write tests before or alongside code (TDD)
- Test both happy and error paths
- Use meaningful test names
- Keep tests simple and focused
- Mock external dependencies
- Use test helpers to reduce duplication
- Run tests frequently during development
- Maintain test data properly

### DON'T ❌
- Test implementation details
- Write overly complex test setup
- Ignore failing tests
- Skip error case testing
- Use production data in tests
- Make tests dependent on each other
- Hardcode test values without meaning

## 🔧 **Testing Tools and Utilities**

### Test Helpers
```go
// Common test utilities
package testutil

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func AssertNoError(t *testing.T, err error) {
    require.NoError(t, err)
}

func AssertEqual(t *testing.T, expected, actual interface{}) {
    require.Equal(t, expected, actual)
}

func CreateTestDB() (*sql.DB, sqlmock.Sqlmock, error) {
    return sqlmock.New()
}
```

### Mock Generation
```bash
# Generate mocks using mockgen
go install github.com/golang/mock/mockgen@latest

# Generate service mocks
mockgen -source=proto/academic/academic_grpc.pb.go \
        -destination=tests/mocks/academic_mock.go

# Generate interface mocks
mockgen -destination=tests/mocks/database_mock.go \
        database/sql/driver Driver
```

## 📈 **Continuous Improvement**

### Metrics to Track
- Test coverage percentage
- Test execution time
- Test failure rate
- Code quality scores
- Security vulnerabilities

### Regular Reviews
- Review test coverage reports monthly
- Update test data and fixtures quarterly
- Refactor slow or flaky tests
- Add tests for new features
- Remove obsolete tests

## 🆘 **Troubleshooting**

### Common Issues
1. **Tests timeout**: Increase timeout or optimize test setup
2. **Flaky tests**: Remove race conditions and external dependencies
3. **Slow tests**: Use mocks instead of real services
4. **Coverage gaps**: Add missing test cases
5. **Mock issues**: Verify mock expectations and setup

### Debug Tips
```bash
# Run with verbose output
go test -v ./...

# Run single test with debug
go test -v -run TestSpecificFunction ./package

# Show test coverage gaps
go test -cover -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%"
```

## 📚 **Resources**

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Framework](https://github.com/stretchr/testify)
- [Go Testing Best Practices](https://golang.org/doc/code.html#Testing)
- [GraphQL Testing Guide](https://graphql.org/learn/testing/)
- [gRPC Testing](https://grpc.io/docs/guides/testing/)

## 🎉 **Conclusion**

This comprehensive testing setup ensures:
- **High code quality** through extensive test coverage
- **Reliable deployments** via automated testing
- **Fast feedback loops** for development
- **Maintainable codebase** with proper test structure
- **Confident refactoring** with safety nets

Remember: Good tests are an investment in code quality and developer productivity! 🚀