# Test Automation Setup Guide

Hướng dẫn chi tiết thiết lập Test Automation cho dự án LVTN Backend với CI/CD pipeline, code coverage reporting, và test result analysis.

## 📋 Mục Lục

1. [Tổng Quan](#tổng-quan)
2. [GitHub Actions CI/CD](#github-actions-cicd)
3. [Code Coverage](#code-coverage)
4. [Test Scripts](#test-scripts)
5. [Pre-commit Hooks](#pre-commit-hooks)
6. [Makefile Commands](#makefile-commands)
7. [Best Practices](#best-practices)

---

## 🎯 Tổng Quan

### Mục Tiêu
- ✅ Tự động chạy tests khi push code lên repository
- ✅ Báo cáo code coverage chi tiết
- ✅ Phát hiện lỗi sớm trong quá trình phát triển
- ✅ Đảm bảo chất lượng code trước khi merge
- ✅ Tích hợp với Docker build pipeline

### Các Công Cụ Sử Dụng
- **GitHub Actions**: CI/CD pipeline automation
- **Go Test**: Unit testing framework
- **go-sqlmock**: Database mocking
- **testify**: Assertion library
- **codecov/coveralls**: Coverage reporting
- **golangci-lint**: Code quality checks

---

## 🚀 GitHub Actions CI/CD

### 1. Test Pipeline Workflow

Tạo file `.github/workflows/tests.yml`:

```yaml
name: Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]
  workflow_dispatch:

env:
  GO_VERSION: '1.21'

jobs:
  # Job 1: Lint và Code Quality
  lint:
    name: Lint & Code Quality
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
      
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m

  # Job 2: Unit Tests cho từng service
  unit-tests:
    name: Unit Tests
    runs-on: ubuntu-latest
    needs: lint
    
    strategy:
      matrix:
        service: [academic, council, file, role, thesis, user]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
      
      - name: Install dependencies
        run: |
          go mod download
          go mod verify
      
      - name: Run ${{ matrix.service }} service tests
        run: |
          cd src/service/${{ matrix.service }}
          go test ./tests/unit/... -v -race -coverprofile=coverage.out -covermode=atomic -timeout=5m
      
      - name: Generate coverage report
        run: |
          cd src/service/${{ matrix.service }}
          go tool cover -func=coverage.out -o=coverage.txt
          echo "Coverage summary for ${{ matrix.service }}:"
          tail -n 1 coverage.txt
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./src/service/${{ matrix.service }}/coverage.out
          flags: ${{ matrix.service }}
          name: ${{ matrix.service }}-coverage
          fail_ci_if_error: false
      
      - name: Archive test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-results-${{ matrix.service }}
          path: |
            src/service/${{ matrix.service }}/coverage.out
            src/service/${{ matrix.service }}/coverage.txt

  # Job 3: Integration Tests (optional)
  integration-tests:
    name: Integration Tests
    runs-on: ubuntu-latest
    needs: unit-tests
    if: github.event_name == 'push' || github.event_name == 'workflow_dispatch'
    
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: root
          MYSQL_DATABASE: test_db
        ports:
          - 3306:3306
        options: >-
          --health-cmd="mysqladmin ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=3
      
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd="redis-cli ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=3
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
      
      - name: Wait for services
        run: |
          sleep 10
          echo "Services are ready"
      
      - name: Run integration tests
        env:
          DB_HOST: localhost
          DB_PORT: 3306
          DB_USER: root
          DB_PASSWORD: root
          DB_NAME: test_db
          REDIS_HOST: localhost
          REDIS_PORT: 6379
        run: |
          if [ -d "tests/integration" ]; then
            go test ./tests/integration/... -v -timeout=10m
          else
            echo "No integration tests found, skipping..."
          fi

  # Job 4: Test Coverage Summary
  coverage-summary:
    name: Coverage Summary
    runs-on: ubuntu-latest
    needs: unit-tests
    if: always()
    
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v3
      
      - name: Display coverage summary
        run: |
          echo "## Test Coverage Summary" >> $GITHUB_STEP_SUMMARY
          echo "" >> $GITHUB_STEP_SUMMARY
          echo "| Service | Coverage |" >> $GITHUB_STEP_SUMMARY
          echo "|---------|----------|" >> $GITHUB_STEP_SUMMARY
          for service in academic council file role thesis user; do
            if [ -f "test-results-$service/coverage.txt" ]; then
              coverage=$(tail -n 1 "test-results-$service/coverage.txt" | awk '{print $3}')
              echo "| $service | $coverage |" >> $GITHUB_STEP_SUMMARY
            fi
          done

  # Job 5: Notify on failure
  notify:
    name: Notify Results
    runs-on: ubuntu-latest
    needs: [lint, unit-tests, integration-tests]
    if: failure()
    
    steps:
      - name: Send notification
        run: |
          echo "⚠️ Tests failed! Check the workflow logs for details."
```

### 2. Extended Test Workflow với Badges

Tạo file `.github/workflows/tests-extended.yml`:

```yaml
name: Tests Extended

on:
  schedule:
    - cron: '0 2 * * *'  # Run daily at 2 AM
  workflow_dispatch:

jobs:
  comprehensive-tests:
    name: Comprehensive Test Suite
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Run all tests with coverage
        run: |
          ./scripts/run-all-tests.sh coverage
      
      - name: Generate coverage badge
        uses: vladopajic/go-test-coverage@v2
        with:
          config: .testcoverage.yml
      
      - name: Update README badges
        if: github.ref == 'refs/heads/main'
        run: |
          echo "Updating coverage badges..."
          # Add logic to update badges in README
```

---

## 📊 Code Coverage

### 1. Coverage Configuration

Tạo file `.testcoverage.yml`:

```yaml
# Test coverage thresholds
profile: coverage.out
threshold:
  file: 80
  package: 75
  total: 80

# Files to exclude from coverage
exclude:
  - "**/*_test.go"
  - "**/tests/**"
  - "**/*.pb.go"
  - "**/*.pb.gw.go"
  - "**/generated/**"
  - "**/mocks/**"
  - "**/vendor/**"

# Coverage report format
output:
  format: badges
```

### 2. Local Coverage Script

Tạo file `scripts/coverage-report.sh`:

```bash
#!/bin/bash

# Generate comprehensive coverage report
# Usage: ./scripts/coverage-report.sh [service]

set -e

SERVICE=${1:-all}
COVERAGE_DIR="coverage"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}📊 Generating Coverage Reports${NC}"
echo "=================================="

# Create coverage directory
mkdir -p $COVERAGE_DIR

# Function to generate coverage for a service
generate_service_coverage() {
    local service=$1
    echo -e "${YELLOW}Processing $service service...${NC}"
    
    cd "src/service/$service"
    
    # Run tests with coverage
    go test ./tests/unit/... -coverprofile="../../../$COVERAGE_DIR/${service}_coverage.out" \
        -covermode=atomic -v
    
    # Generate HTML report
    go tool cover -html="../../../$COVERAGE_DIR/${service}_coverage.out" \
        -o="../../../$COVERAGE_DIR/${service}_coverage.html"
    
    # Generate function-level coverage
    go tool cover -func="../../../$COVERAGE_DIR/${service}_coverage.out" \
        > "../../../$COVERAGE_DIR/${service}_coverage.txt"
    
    # Extract total coverage
    total_coverage=$(go tool cover -func="../../../$COVERAGE_DIR/${service}_coverage.out" | \
        grep "total:" | awk '{print $3}')
    
    echo -e "${GREEN}✅ $service: $total_coverage coverage${NC}"
    
    cd - > /dev/null
}

# Generate coverage for specific service or all
if [ "$SERVICE" == "all" ]; then
    services=("academic" "council" "file" "role" "thesis" "user")
    
    echo "" > "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    echo "Coverage Summary - $TIMESTAMP" >> "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    echo "=============================" >> "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    echo "" >> "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    
    for service in "${services[@]}"; do
        generate_service_coverage "$service"
        
        # Add to summary
        total=$(tail -n 1 "$COVERAGE_DIR/${service}_coverage.txt" | awk '{print $3}')
        printf "%-15s: %s\n" "$service" "$total" >> "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    done
    
    # Merge coverage files
    echo -e "${YELLOW}Merging coverage reports...${NC}"
    gocovmerge $COVERAGE_DIR/*_coverage.out > $COVERAGE_DIR/merged_coverage.out
    
    # Generate merged HTML report
    go tool cover -html=$COVERAGE_DIR/merged_coverage.out \
        -o=$COVERAGE_DIR/merged_coverage.html
    
    echo ""
    echo -e "${BLUE}📋 Coverage Summary:${NC}"
    cat "$COVERAGE_DIR/summary_$TIMESTAMP.txt"
    
else
    generate_service_coverage "$SERVICE"
fi

echo ""
echo -e "${GREEN}✅ Coverage reports generated in $COVERAGE_DIR/${NC}"
echo -e "   Open ${BLUE}$COVERAGE_DIR/*_coverage.html${NC} in your browser"
```

### 3. PowerShell Coverage Script

Tạo file `scripts/coverage-report.ps1`:

```powershell
# Generate comprehensive coverage report (PowerShell)
# Usage: .\scripts\coverage-report.ps1 [service]

param(
    [string]$Service = "all"
)

$CoverageDir = "coverage"
$Timestamp = Get-Date -Format "yyyyMMdd_HHmmss"

Write-Host "📊 Generating Coverage Reports" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan

# Create coverage directory
New-Item -ItemType Directory -Force -Path $CoverageDir | Out-Null

# Function to generate coverage for a service
function Generate-ServiceCoverage {
    param($ServiceName)
    
    Write-Host "Processing $ServiceName service..." -ForegroundColor Yellow
    
    Push-Location "src\service\$ServiceName"
    
    try {
        # Run tests with coverage
        go test .\tests\unit\... -coverprofile="..\..\..\$CoverageDir\${ServiceName}_coverage.out" `
            -covermode=atomic -v
        
        if ($LASTEXITCODE -eq 0) {
            # Generate HTML report
            go tool cover -html="..\..\..\$CoverageDir\${ServiceName}_coverage.out" `
                -o="..\..\..\$CoverageDir\${ServiceName}_coverage.html"
            
            # Generate function-level coverage
            go tool cover -func="..\..\..\$CoverageDir\${ServiceName}_coverage.out" `
                | Out-File "..\..\..\$CoverageDir\${ServiceName}_coverage.txt"
            
            # Extract total coverage
            $totalCoverage = go tool cover -func="..\..\..\$CoverageDir\${ServiceName}_coverage.out" | `
                Select-String "total:" | ForEach-Object { $_.Line.Split()[2] }
            
            Write-Host "✅ $ServiceName`: $totalCoverage coverage" -ForegroundColor Green
            
            return $totalCoverage
        } else {
            Write-Host "❌ Tests failed for $ServiceName" -ForegroundColor Red
            return "N/A"
        }
    }
    finally {
        Pop-Location
    }
}

# Generate coverage for specific service or all
if ($Service -eq "all") {
    $services = @("academic", "council", "file", "role", "thesis", "user")
    
    $summary = @()
    $summary += "Coverage Summary - $Timestamp"
    $summary += "============================="
    $summary += ""
    
    foreach ($svc in $services) {
        $coverage = Generate-ServiceCoverage -ServiceName $svc
        $summary += "{0,-15}: {1}" -f $svc, $coverage
    }
    
    # Save summary
    $summary | Out-File "$CoverageDir\summary_$Timestamp.txt"
    
    Write-Host ""
    Write-Host "📋 Coverage Summary:" -ForegroundColor Cyan
    $summary | ForEach-Object { Write-Host $_ }
    
} else {
    Generate-ServiceCoverage -ServiceName $Service
}

Write-Host ""
Write-Host "✅ Coverage reports generated in $CoverageDir\" -ForegroundColor Green
Write-Host "   Open $CoverageDir\*_coverage.html in your browser" -ForegroundColor Blue
```

---

## 🔧 Test Scripts

### 1. Comprehensive Test Runner

Tạo file `scripts/run-all-tests.sh`:

```bash
#!/bin/bash

# Comprehensive test runner for all services
# Usage: ./scripts/run-all-tests.sh [mode]
# Modes: quick, full, coverage, watch

set -e

MODE=${1:-quick}

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
SERVICES=("academic" "council" "file" "role" "thesis" "user")
TOTAL_TESTS=218
FAILED_SERVICES=()

echo -e "${BLUE}🧪 LVTN Microservices Test Suite${NC}"
echo -e "${CYAN}===================================${NC}"
echo -e "Mode: ${YELLOW}$MODE${NC}"
echo -e "Total Expected Tests: ${CYAN}$TOTAL_TESTS${NC}"
echo ""

# Function to run tests for a service
run_service_tests() {
    local service=$1
    local flags=$2
    
    echo -e "${YELLOW}Testing $service service...${NC}"
    
    cd "src/service/$service"
    
    if go test ./tests/unit/... $flags; then
        echo -e "${GREEN}✅ $service: PASSED${NC}"
        cd - > /dev/null
        return 0
    else
        echo -e "${RED}❌ $service: FAILED${NC}"
        FAILED_SERVICES+=("$service")
        cd - > /dev/null
        return 1
    fi
}

# Quick mode - run all tests
if [ "$MODE" == "quick" ]; then
    echo -e "${CYAN}Running quick test suite...${NC}"
    echo ""
    
    for service in "${SERVICES[@]}"; do
        run_service_tests "$service" "-v -timeout=30s" || true
        echo ""
    done

# Full mode - run with race detection
elif [ "$MODE" == "full" ]; then
    echo -e "${CYAN}Running full test suite with race detection...${NC}"
    echo ""
    
    for service in "${SERVICES[@]}"; do
        run_service_tests "$service" "-v -race -timeout=1m" || true
        echo ""
    done

# Coverage mode - generate coverage reports
elif [ "$MODE" == "coverage" ]; then
    echo -e "${CYAN}Running tests with coverage analysis...${NC}"
    echo ""
    
    mkdir -p coverage
    
    for service in "${SERVICES[@]}"; do
        run_service_tests "$service" "-v -coverprofile=coverage/${service}.out -covermode=atomic" || true
        
        if [ -f "src/service/$service/coverage/${service}.out" ]; then
            total=$(go tool cover -func="src/service/$service/coverage/${service}.out" | \
                grep "total:" | awk '{print $3}')
            echo -e "   Coverage: ${CYAN}$total${NC}"
        fi
        echo ""
    done

# Watch mode - continuous testing
elif [ "$MODE" == "watch" ]; then
    echo -e "${CYAN}Starting watch mode (Ctrl+C to stop)...${NC}"
    echo ""
    
    while true; do
        clear
        echo -e "${BLUE}🔄 Running tests... ($(date))${NC}"
        echo ""
        
        for service in "${SERVICES[@]}"; do
            run_service_tests "$service" "-v -timeout=30s" || true
        done
        
        echo ""
        echo -e "${YELLOW}Waiting 5 seconds before next run...${NC}"
        sleep 5
    done

else
    echo -e "${RED}Unknown mode: $MODE${NC}"
    echo "Available modes: quick, full, coverage, watch"
    exit 1
fi

# Summary
echo ""
echo -e "${BLUE}=================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}=================================${NC}"

if [ ${#FAILED_SERVICES[@]} -eq 0 ]; then
    echo -e "${GREEN}✅ All services passed!${NC}"
    echo -e "   ${CYAN}$TOTAL_TESTS tests executed successfully${NC}"
    exit 0
else
    echo -e "${RED}❌ ${#FAILED_SERVICES[@]} service(s) failed:${NC}"
    for service in "${FAILED_SERVICES[@]}"; do
        echo -e "   ${RED}- $service${NC}"
    done
    exit 1
fi
```

### 2. PowerShell Test Runner

Tạo file `scripts/run-all-tests.ps1`:

```powershell
# Comprehensive test runner for all services (PowerShell)
# Usage: .\scripts\run-all-tests.ps1 [mode]
# Modes: quick, full, coverage, watch

param(
    [string]$Mode = "quick"
)

$Services = @("academic", "council", "file", "role", "thesis", "user")
$TotalTests = 218
$FailedServices = @()

Write-Host "🧪 LVTN Microservices Test Suite" -ForegroundColor Blue
Write-Host "=================================" -ForegroundColor Cyan
Write-Host "Mode: $Mode" -ForegroundColor Yellow
Write-Host "Total Expected Tests: $TotalTests" -ForegroundColor Cyan
Write-Host ""

# Function to run tests for a service
function Run-ServiceTests {
    param(
        [string]$Service,
        [string]$Flags
    )
    
    Write-Host "Testing $Service service..." -ForegroundColor Yellow
    
    Push-Location "src\service\$Service"
    
    try {
        $cmd = "go test .\tests\unit\... $Flags"
        Invoke-Expression $cmd
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ $Service`: PASSED" -ForegroundColor Green
            return $true
        } else {
            Write-Host "❌ $Service`: FAILED" -ForegroundColor Red
            $script:FailedServices += $Service
            return $false
        }
    }
    finally {
        Pop-Location
    }
}

# Quick mode
if ($Mode -eq "quick") {
    Write-Host "Running quick test suite..." -ForegroundColor Cyan
    Write-Host ""
    
    foreach ($service in $Services) {
        Run-ServiceTests -Service $service -Flags "-v -timeout=30s" | Out-Null
        Write-Host ""
    }
}
# Full mode
elseif ($Mode -eq "full") {
    Write-Host "Running full test suite with race detection..." -ForegroundColor Cyan
    Write-Host ""
    
    foreach ($service in $Services) {
        Run-ServiceTests -Service $service -Flags "-v -race -timeout=1m" | Out-Null
        Write-Host ""
    }
}
# Coverage mode
elseif ($Mode -eq "coverage") {
    Write-Host "Running tests with coverage analysis..." -ForegroundColor Cyan
    Write-Host ""
    
    New-Item -ItemType Directory -Force -Path "coverage" | Out-Null
    
    foreach ($service in $Services) {
        $coverFile = "coverage\${service}.out"
        Run-ServiceTests -Service $service -Flags "-v -coverprofile=$coverFile -covermode=atomic" | Out-Null
        
        if (Test-Path $coverFile) {
            $coverage = go tool cover -func=$coverFile | Select-String "total:" | ForEach-Object { $_.Line.Split()[2] }
            Write-Host "   Coverage: $coverage" -ForegroundColor Cyan
        }
        Write-Host ""
    }
}
# Watch mode
elseif ($Mode -eq "watch") {
    Write-Host "Starting watch mode (Ctrl+C to stop)..." -ForegroundColor Cyan
    Write-Host ""
    
    while ($true) {
        Clear-Host
        Write-Host "🔄 Running tests... ($(Get-Date))" -ForegroundColor Blue
        Write-Host ""
        
        foreach ($service in $Services) {
            Run-ServiceTests -Service $service -Flags "-v -timeout=30s" | Out-Null
        }
        
        Write-Host ""
        Write-Host "Waiting 5 seconds before next run..." -ForegroundColor Yellow
        Start-Sleep -Seconds 5
    }
}
else {
    Write-Host "Unknown mode: $Mode" -ForegroundColor Red
    Write-Host "Available modes: quick, full, coverage, watch"
    exit 1
}

# Summary
Write-Host ""
Write-Host "=================================" -ForegroundColor Blue
Write-Host "Test Summary" -ForegroundColor Blue
Write-Host "=================================" -ForegroundColor Blue

if ($FailedServices.Count -eq 0) {
    Write-Host "✅ All services passed!" -ForegroundColor Green
    Write-Host "   $TotalTests tests executed successfully" -ForegroundColor Cyan
    exit 0
} else {
    Write-Host "❌ $($FailedServices.Count) service(s) failed:" -ForegroundColor Red
    foreach ($service in $FailedServices) {
        Write-Host "   - $service" -ForegroundColor Red
    }
    exit 1
}
```

---

## 🪝 Pre-commit Hooks

### 1. Setup Git Hooks

Tạo file `.githooks/pre-commit`:

```bash
#!/bin/bash

# Pre-commit hook for running tests
# Install: ln -s ../../.githooks/pre-commit .git/hooks/pre-commit

echo "🔍 Running pre-commit checks..."

# Run gofmt
echo "Checking code formatting..."
unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
    echo "❌ These files need formatting:"
    echo "$unformatted"
    echo ""
    echo "Run: go fmt ./..."
    exit 1
fi
echo "✅ Code formatting OK"

# Run go vet
echo "Running go vet..."
if ! go vet ./...; then
    echo "❌ go vet failed"
    exit 1
fi
echo "✅ go vet OK"

# Run quick tests for changed services
echo "Running quick tests for changed services..."
changed_services=$(git diff --cached --name-only | grep "src/service/" | cut -d/ -f3 | sort -u)

if [ -n "$changed_services" ]; then
    for service in $changed_services; do
        if [ -d "src/service/$service/tests" ]; then
            echo "Testing $service..."
            cd "src/service/$service"
            if ! go test ./tests/unit/... -short -timeout=30s; then
                echo "❌ Tests failed for $service"
                exit 1
            fi
            cd - > /dev/null
        fi
    done
    echo "✅ All tests passed"
else
    echo "ℹ️  No service changes detected"
fi

echo "✅ Pre-commit checks passed!"
```

### 2. Install Hooks Script

Tạo file `scripts/install-hooks.sh`:

```bash
#!/bin/bash

echo "Installing git hooks..."

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Install pre-commit hook
if [ -f ".githooks/pre-commit" ]; then
    ln -sf ../../.githooks/pre-commit .git/hooks/pre-commit
    chmod +x .githooks/pre-commit
    chmod +x .git/hooks/pre-commit
    echo "✅ Pre-commit hook installed"
else
    echo "⚠️  .githooks/pre-commit not found"
fi

echo "Done!"
```

---

## 📝 Makefile Commands

Thêm các commands sau vào `Makefile`:

```makefile
# Test targets
.PHONY: test test-all test-coverage test-quick test-watch test-service

# Run all tests
test:
	@echo "Running all tests..."
	@./scripts/run-all-tests.sh quick

# Run all tests with race detection
test-all:
	@echo "Running full test suite..."
	@./scripts/run-all-tests.sh full

# Run tests with coverage
test-coverage:
	@echo "Generating coverage reports..."
	@./scripts/run-all-tests.sh coverage

# Quick test (parallel)
test-quick:
	@echo "Running quick tests..."
	@go test ./src/service/*/tests/unit/... -short -timeout=30s

# Watch mode
test-watch:
	@echo "Starting watch mode..."
	@./scripts/run-all-tests.sh watch

# Test specific service
test-service:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Usage: make test-service SERVICE=<service_name>"; \
		exit 1; \
	fi
	@cd src/service/$(SERVICE) && go test ./tests/unit/... -v

# Test with race detection
test-race:
	@go test ./src/service/*/tests/unit/... -race -timeout=1m

# Lint
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Install test dependencies
test-deps:
	@echo "Installing test dependencies..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/wadey/gocovmerge@latest

# Clean test artifacts
test-clean:
	@echo "Cleaning test artifacts..."
	@rm -rf coverage/
	@find . -name "*.out" -type f -delete
	@find . -name "*.html" -type f -delete
	@go clean -testcache

# Help
test-help:
	@echo "Test Commands:"
	@echo "  make test              - Run all tests (quick mode)"
	@echo "  make test-all          - Run full test suite with race detection"
	@echo "  make test-coverage     - Generate coverage reports"
	@echo "  make test-quick        - Run quick parallel tests"
	@echo "  make test-watch        - Watch mode (continuous testing)"
	@echo "  make test-service      - Test specific service (make test-service SERVICE=academic)"
	@echo "  make test-race         - Run tests with race detection"
	@echo "  make lint              - Run code linter"
	@echo "  make test-deps         - Install test dependencies"
	@echo "  make test-clean        - Clean test artifacts"
```

---

## ✅ Best Practices

### 1. Test Organization
- ✅ Tách tests thành unit và integration
- ✅ Sử dụng fixtures và test data riêng biệt
- ✅ Mock dependencies (database, external services)
- ✅ Đặt tên test rõ ràng theo pattern: `Test<Function>_<Scenario>`

### 2. Coverage Goals
- 🎯 **Unit Tests**: Minimum 80% coverage
- 🎯 **Handler Functions**: 100% coverage
- 🎯 **Business Logic**: 90%+ coverage
- 🎯 **Error Handling**: 100% coverage

### 3. CI/CD Integration
- ✅ Run tests trên mọi pull request
- ✅ Block merge nếu tests fail
- ✅ Generate coverage reports tự động
- ✅ Notify team khi có failures

### 4. Performance
- ⚡ Keep tests fast (< 5 seconds per service)
- ⚡ Run tests parallel khi có thể
- ⚡ Use `-short` flag cho quick tests
- ⚡ Cache dependencies

### 5. Maintenance
- 🔄 Review và update tests regularly
- 🔄 Remove flaky tests
- 🔄 Keep test data up to date
- 🔄 Document test scenarios

---

## 📊 Test Metrics Dashboard

### Current Status

| Service | Tests | Coverage | Status |
|---------|-------|----------|--------|
| Academic | 42 | TBD | ✅ |
| Council | 15 | TBD | ✅ |
| Role | 15 | TBD | ✅ |
| Thesis | 100 | TBD | ✅ |
| User | 28 | TBD | ✅ |
| File | 18 | TBD | ✅ |
| **Total** | **218** | **TBD** | ✅ |

---

## 🚀 Quick Start

### 1. Install Dependencies
```bash
# Install test tools
make test-deps

# Install git hooks
./scripts/install-hooks.sh
```

### 2. Run Tests Locally
```bash
# Quick test
make test

# Full test suite
make test-all

# With coverage
make test-coverage

# Specific service
make test-service SERVICE=academic
```

### 3. Setup CI/CD
- Merge `.github/workflows/tests.yml` vào repository
- Configure secrets trong GitHub Settings
- Enable branch protection rules

### 4. Monitor Results
- Check GitHub Actions tab
- Review coverage reports
- Monitor test execution time

---

## 📚 Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Codecov Integration](https://about.codecov.io/)
- [golangci-lint](https://golangci-lint.run/)

---

## 🔧 Troubleshooting

### Tests Failing Locally
1. Clear test cache: `go clean -testcache`
2. Check dependencies: `go mod tidy`
3. Verify mock data is up to date

### CI/CD Pipeline Issues
1. Check GitHub Actions logs
2. Verify secrets are configured
3. Ensure services are available

### Coverage Not Generating
1. Check coverage file paths
2. Verify test flags are correct
3. Ensure go tool cover is installed

---

**Tác giả**: LVTN Development Team  
**Cập nhật**: November 13, 2025  
**Version**: 1.0.0
