# Test Automation - Quick Start Guide

Hướng dẫn nhanh để bắt đầu sử dụng Test Automation cho dự án LVTN.

## 🚀 Setup Nhanh

### 1. Cài đặt Dependencies

```bash
# Install test tools
make test-deps

# Install git hooks (optional)
make install-hooks
```

### 2. Chạy Tests

#### Linux/Mac:
```bash
# Quick test - tất cả services
make test

# hoặc
./scripts/run-all-tests.sh quick

# Full test với race detection
make test-all

# Test với coverage
make test-coverage

# Test một service cụ thể
make test-service SERVICE=academic
```

#### Windows PowerShell:
```powershell
# Quick test - tất cả services
.\scripts\run-all-tests.ps1 quick

# Full test với race detection
.\scripts\run-all-tests.ps1 full

# Test với coverage
.\scripts\run-all-tests.ps1 coverage

# Test một service cụ thể
cd src\service\academic
go test .\tests\unit\... -v
```

### 3. Generate Coverage Reports

#### Linux/Mac:
```bash
# Coverage cho tất cả services
make coverage-report

# hoặc
./scripts/coverage-report.sh all

# Coverage cho một service
make coverage-service SERVICE=academic

# hoặc
./scripts/coverage-report.sh academic
```

#### Windows PowerShell:
```powershell
# Coverage cho tất cả services
.\scripts\coverage-report.ps1 all

# Coverage cho một service
.\scripts\coverage-report.ps1 academic
```

Kết quả sẽ được tạo trong thư mục `coverage/`:
- `<service>_coverage.html` - HTML report (mở trong browser)
- `<service>_coverage.out` - Raw coverage data
- `<service>_coverage.txt` - Function-level coverage
- `summary_<timestamp>.txt` - Tổng hợp coverage

## 📊 Test Modes

### Quick Mode (Mặc định)
```bash
./scripts/run-all-tests.sh quick
```
- Chạy tất cả 218 tests
- Timeout: 30s
- Không có race detection
- Nhanh nhất

### Full Mode
```bash
./scripts/run-all-tests.sh full
```
- Chạy tất cả tests với race detection
- Timeout: 1m
- Phát hiện race conditions
- Chậm hơn nhưng an toàn hơn

### Coverage Mode
```bash
./scripts/run-all-tests.sh coverage
```
- Generate coverage reports
- Hiển thị coverage % cho mỗi service
- Lưu coverage files

### Watch Mode
```bash
./scripts/run-all-tests.sh watch
```
- Chạy tests liên tục
- Auto-refresh mỗi 5 giây
- Ctrl+C để thoát
- Tốt cho development

## 🔍 Makefile Commands

```bash
# Testing
make test              # Quick test
make test-all          # Full test suite
make test-coverage     # Coverage reports
make test-service SERVICE=academic  # Test một service
make test-race         # Race detection
make test-watch        # Watch mode

# Coverage
make coverage-report           # All services
make coverage-service SERVICE=academic  # Một service

# Maintenance
make lint              # Run linter
make test-clean        # Clean artifacts
make test-deps         # Install dependencies
make install-hooks     # Setup git hooks

# Help
make help              # Xem tất cả commands
```

## 🎯 Test Structure

```
src/service/
├── academic/tests/
│   ├── fixtures/      # Test data
│   │   ├── faculty.go
│   │   ├── major.go
│   │   └── semester.go
│   └── unit/          # Unit tests
│       ├── faculty_test.go    (14 tests)
│       ├── major_test.go      (14 tests)
│       └── semester_test.go   (14 tests)
│
├── council/tests/
│   ├── fixtures/council.go
│   └── unit/council_test.go   (15 tests)
│
├── role/tests/
│   ├── fixtures/role.go
│   └── unit/role_test.go      (15 tests)
│
├── thesis/tests/
│   ├── fixtures/               (7 entities)
│   └── unit/                   (100 tests total)
│       ├── topic_test.go       (15 tests)
│       ├── enrollment_test.go  (15 tests)
│       ├── midterm_test.go     (14 tests)
│       ├── final_test.go       (14 tests)
│       ├── grade_review_test.go (14 tests)
│       ├── topic_council_test.go (14 tests)
│       └── topic_council_supervisor_test.go (14 tests)
│
├── user/tests/
│   ├── fixtures/
│   │   ├── student.go
│   │   └── teacher.go
│   └── unit/
│       ├── student_test.go     (14 tests)
│       └── teacher_test.go     (14 tests)
│
└── file/tests/
    ├── fixtures/file.go
    └── unit/file_test.go       (18 tests)
```

**Total: 218 tests across 6 services**

## 🎨 Output Examples

### Successful Test Run:
```
🧪 LVTN Microservices Test Suite
=================================
Mode: quick
Total Expected Tests: 218

Testing academic service...
✅ academic: PASSED

Testing council service...
✅ council: PASSED

...

=================================
Test Summary
=================================
✅ All services passed!
   218 tests executed successfully
```

### Coverage Report:
```
📊 Generating Coverage Reports
==================================
Processing academic service...
✅ academic: 85.4% coverage
Processing council service...
✅ council: 88.2% coverage
...

📋 Coverage Summary:
academic       : 85.4%
council        : 88.2%
file           : 90.1%
role           : 87.5%
thesis         : 89.3%
user           : 86.7%
```

## 🤖 CI/CD Integration

### GitHub Actions
Tests tự động chạy khi:
- Push lên `main` hoặc `develop`
- Tạo Pull Request
- Manual trigger

Xem: `.github/workflows/tests.yml`

### Pre-commit Hooks
Tự động chạy khi commit:
- Code formatting check (gofmt)
- Go vet
- Quick tests cho changed services

Install: `make install-hooks`

## 📈 Coverage Goals

- **Unit Tests**: Minimum 80% coverage
- **Handler Functions**: 100% coverage  
- **Business Logic**: 90%+ coverage
- **Error Handling**: 100% coverage

## 🐛 Troubleshooting

### Tests fail locally
```bash
# Clear test cache
go clean -testcache

# Update dependencies
go mod tidy
go mod download
```

### Coverage không generate
```bash
# Kiểm tra Go tools
go tool cover -h

# Clean và retry
make test-clean
make test-coverage
```

### Permission denied (Linux/Mac)
```bash
# Make scripts executable
chmod +x scripts/*.sh
chmod +x .githooks/pre-commit
```

## 📚 Documentation

Xem hướng dẫn chi tiết tại: [TEST_AUTOMATION_GUIDE.md](./TEST_AUTOMATION_GUIDE.md)

## ✅ Quick Checklist

- [ ] Đã cài đặt test dependencies (`make test-deps`)
- [ ] Đã chạy thử tests (`make test`)
- [ ] Đã xem coverage reports (`make coverage-report`)
- [ ] Đã setup git hooks nếu cần (`make install-hooks`)
- [ ] Đã đọc TEST_AUTOMATION_GUIDE.md

---

**Happy Testing! 🎉**
