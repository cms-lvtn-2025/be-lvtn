# Academic Service Testing Structure

This directory contains the organized testing framework for the Academic Service, with a clean separation between test code and production code.

## Directory Structure

```
src/service/academic/
├── handler/                      # Production code
│   ├── faculty.go
│   ├── major.go
│   ├── semester.go
│   └── handler.go
├── tests/                        # Test code (separated from production)
│   ├── unit/                     # Unit tests
│   │   ├── faculty_test.go       # Faculty operation tests
│   │   ├── major_test.go         # Major operation tests
│   │   └── semester_test.go      # Semester operation tests
│   ├── mocks/                    # Mock objects and test helpers
│   │   ├── db_mock.go           # Database mock interface
│   │   ├── faculty_mock.go      # Faculty mock implementation
│   │   ├── major_mock.go        # Major mock implementation
│   │   └── semester_mock.go     # Semester mock implementation
│   └── fixtures/                 # Test data and fixtures
│       ├── faculty_fixtures.go  # Faculty test data
│       ├── major_fixtures.go    # Major test data
│       └── semester_fixtures.go # Semester test data
```

## Benefits of This Structure

### 🎯 **Clean Separation**
- Production code in `handler/` directory
- Test code completely separate in `tests/` directory
- No test files mixed with production code
- Clear organizational boundaries

### 📦 **Modular Organization**
- **Unit Tests**: Focused on individual function testing
- **Mocks**: Reusable mock objects for consistent testing
- **Fixtures**: Centralized test data management
- **Scalable**: Easy to add integration tests, e2e tests, etc.

### 🔧 **Maintainability**
- Easy to locate and modify specific tests
- Shared mock objects reduce code duplication
- Consistent test data across all tests
- Clear dependency structure

## Running Tests

### Using Test Runner Script
```powershell
# Run unit tests only
.\academic-test.ps1 unit

# Run with coverage analysis
.\academic-test.ps1 coverage

# Run all tests
.\academic-test.ps1
```

### Direct Go Commands
```bash
# Unit tests only
go test ./src/service/academic/tests/unit/ -v

# With coverage
go test ./src/service/academic/tests/unit/ -cover -coverprofile=coverage.out

# Coverage report
go tool cover -func=coverage.out
```

## Test Coverage

Current test suite covers:

### Faculty Operations
- ✅ Create Faculty (validation, success, error cases)
- ✅ List Faculties (success, database errors)
- ✅ Get Faculty (found, not found, invalid ID)
- ✅ Update Faculty (success, not found)
- ✅ Delete Faculty (success, not found, invalid ID)

### Major Operations
- ✅ Create Major (validation, success, error cases)
- ✅ List Majors (success, database errors)
- ✅ Get Major (found, not found, invalid ID)
- ✅ Update Major (success, not found)
- ✅ Delete Major (success, not found, invalid ID)

### Semester Operations
- ✅ Create Semester (validation, success, error cases)
- ✅ List Semesters (success, database errors)
- ✅ Get Semester (found, not found, invalid ID)
- ✅ Update Semester (success, not found)
- ✅ Delete Semester (success, not found, invalid ID)

## Mock Objects

### Database Mock (`db_mock.go`)
Provides database interface abstraction for testing:
```go
type MockHandler struct {
    db *sql.DB
}

func (h *MockHandler) ExecQuery(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
func (h *MockHandler) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
func (h *MockHandler) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
```

### Entity Mocks
Each entity (Faculty, Major, Semester) has dedicated mock implementations:
- Simplified SQL queries for testing
- Consistent error handling
- gRPC status code validation
- Database transaction mocking

## Test Fixtures

Centralized test data provides:
- **Valid entities** for success scenarios
- **Invalid entities** for validation testing
- **Request objects** for different test cases
- **Common data** used across multiple tests

Example usage:
```go
import "thaily/src/service/academic/tests/fixtures"

// Use predefined test data
req := fixtures.CreateFacultyRequests.Valid
invalidReq := fixtures.CreateFacultyRequests.NoTitle
```

## Best Practices

### 1. Test Isolation
- Each test case is independent
- Database mocks reset between tests
- No shared state between test functions

### 2. Consistent Error Handling
- gRPC status codes validated
- Error messages checked
- Database errors properly mocked

### 3. Comprehensive Coverage
- Success paths tested
- Error conditions covered
- Edge cases included
- Input validation verified

### 4. Maintainable Code
- Descriptive test names
- Clear test structure
- Reusable mock objects
- Centralized test data

## Adding New Tests

### 1. Unit Tests
Add new test files to `tests/unit/`:
```go
package unit

import (
    "testing"
    "thaily/src/service/academic/tests/mocks"
    "thaily/src/service/academic/tests/fixtures"
)

func TestNewFeature(t *testing.T) {
    // Use mocks and fixtures
    mock := mocks.NewFacultyMock(db)
    testData := fixtures.FacultyTestData.ValidFaculty
    
    // Write test cases
}
```

### 2. Mock Objects
Add new mocks to `tests/mocks/`:
```go
package mocks

type NewFeatureMock struct {
    *MockHandler
}

func NewNewFeatureMock(db *sql.DB) *NewFeatureMock {
    return &NewFeatureMock{
        MockHandler: NewMockHandler(db),
    }
}
```

### 3. Test Fixtures
Add test data to `tests/fixtures/`:
```go
package fixtures

var NewFeatureTestData = struct {
    ValidEntity   *pb.NewEntity
    InvalidEntity *pb.NewEntity
}{
    ValidEntity: &pb.NewEntity{
        Id:   "test-1",
        Name: "Test Entity",
    },
    InvalidEntity: &pb.NewEntity{
        Id:   "",
        Name: "",
    },
}
```

This organized testing structure provides a solid foundation for maintaining high code quality while keeping tests separate from production code.