#!/bin/bash

# Test Coverage Script for RexiERP
# This script runs comprehensive tests with coverage reporting

set -e

echo "🧪 Running RexiERP Test Suite with Coverage..."
echo "============================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    print_status "Checking dependencies..."

    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi

    if ! command -v git &> /dev/null; then
        print_error "Git is not installed or not in PATH"
        exit 1
    fi

    print_status "Dependencies check passed"
}

# Run tests with coverage
run_tests() {
    print_status "Running unit tests with coverage..."

    # Create coverage directory
    mkdir -p coverage

    # Run tests for shared packages
    SHARED_PACKAGES=(
        "./internal/shared/config"
        "./internal/shared/auth"
        "./internal/shared/database"
        "./internal/shared/health"
    )

    FAILED_TESTS=()

    for package in "${SHARED_PACKAGES[@]}"; do
        print_status "Testing package: $package"

        if [ -d "$package" ]; then
            if go test -v -race -coverprofile="coverage/$(basename $package).out" -covermode=atomic "$package"; then
                print_status "✅ $package tests passed"
            else
                print_error "❌ $package tests failed"
                FAILED_TESTS+=("$package")
            fi
        else
            print_warning "Package directory $package does not exist, skipping..."
        fi
    done

    # Run tests for other packages if they exist
    if [ -d "./internal/authentication" ]; then
        print_status "Testing package: ./internal/authentication"
        if go test -v -race -coverprofile="coverage/authentication.out" -covermode=atomic ./internal/authentication/...; then
            print_status "✅ Authentication tests passed"
        else
            print_error "❌ Authentication tests failed"
            FAILED_TESTS+=("authentication")
        fi
    fi

    if [ -d "./internal/inventory" ]; then
        print_status "Testing package: ./internal/inventory"
        if go test -v -race -coverprofile="coverage/inventory.out" -covermode=atomic ./internal/inventory/...; then
            print_status "✅ Inventory tests passed"
        else
            print_error "❌ Inventory tests failed"
            FAILED_TESTS+=("inventory")
        fi
    fi

    # Return failure if any tests failed
    if [ ${#FAILED_TESTS[@]} -gt 0 ]; then
        print_error "The following test packages failed: ${FAILED_TESTS[*]}"
        return 1
    fi
}

# Generate coverage report
generate_coverage_report() {
    print_status "Generating combined coverage report..."

    # Combine all coverage files
    COVERAGE_FILES=$(find coverage -name "*.out" -type f)

    if [ -z "$COVERAGE_FILES" ]; then
        print_warning "No coverage files found to combine"
        return 0
    fi

    # Install gocovmerge if not present
    if ! command -v gocovmerge &> /dev/null; then
        print_status "Installing gocovmerge..."
        go install github.com/wadey/gocovmerge@latest
    fi

    # Merge coverage files
    gocovmerge $COVERAGE_FILES > coverage/combined.out

    # Generate HTML report
    go tool cover -html=coverage/combined.out -o coverage/coverage.html

    # Generate coverage summary
    echo ""
    print_status "Coverage Summary:"
    go tool cover -func=coverage/combined.out

    # Extract total coverage percentage
    TOTAL_COVERAGE=$(go tool cover -func=coverage/combined.out | grep "total:" | awk '{print $3}')

    echo ""
    print_status "Total Test Coverage: $TOTAL_COVERAGE"
    print_status "HTML coverage report generated: coverage/coverage.html"

    # Check if coverage meets threshold (recommended: 80%)
    COVERAGE_NUM=$(echo $TOTAL_COVERAGE | sed 's/%//')
    if (( $(echo "$COVERAGE_NUM >= 80" | bc -l) )); then
        print_status "✅ Coverage threshold (80%) met: $TOTAL_COVERAGE"
    else
        print_warning "⚠️  Coverage threshold (80%) not met: $TOTAL_COVERAGE"
    fi
}

# Run integration tests (if available)
run_integration_tests() {
    if [ -d "./tests/integration" ]; then
        print_status "Running integration tests..."

        if go test -v -race ./tests/integration/...; then
            print_status "✅ Integration tests passed"
        else
            print_error "❌ Integration tests failed"
            return 1
        fi
    else
        print_status "No integration tests found, skipping..."
    fi
}

# Run benchmarks
run_benchmarks() {
    print_status "Running benchmark tests..."

    BENCHMARK_PACKAGES=(
        "./internal/shared/config"
        "./internal/shared/auth"
        "./internal/shared/database"
        "./internal/shared/health"
    )

    for package in "${BENCHMARK_PACKAGES[@]}"; do
        if [ -d "$package" ]; then
            print_status "Benchmarking package: $package"
            go test -bench=. -benchmem "$package" || true
        fi
    done
}

# Check code quality with go vet
run_vet() {
    print_status "Running go vet..."

    if go vet ./...; then
        print_status "✅ go vet passed"
    else
        print_error "❌ go vet failed"
        return 1
    fi
}

# Check code formatting
run_fmt_check() {
    print_status "Checking code formatting..."

    UNFORMATTED=$(gofmt -s -l .)
    if [ -z "$UNFORMATTED" ]; then
        print_status "✅ Code is properly formatted"
    else
        print_error "❌ Code is not properly formatted:"
        echo "$UNFORMATTED"
        print_status "Run 'gofmt -s -w .' to fix formatting"
        return 1
    fi
}

# Main execution
main() {
    echo "Starting RexiERP Test Suite..."
    echo ""

    # Check dependencies
    check_dependencies

    # Run code quality checks
    run_fmt_check || echo "Warning: Code formatting check failed"
    run_vet || echo "Warning: go vet check failed"

    # Run tests
    if run_tests; then
        print_status "🎉 All tests passed!"

        # Generate coverage report
        generate_coverage_report

        # Run integration tests
        run_integration_tests || echo "Warning: Integration tests failed"

        # Run benchmarks
        run_benchmarks

        echo ""
        print_status "✅ Test suite completed successfully!"
        print_status "📊 View coverage report: file://$(pwd)/coverage/coverage.html"

        exit 0
    else
        print_error "❌ Test suite failed!"
        exit 1
    fi
}

# Handle script interruption
trap 'print_error "Test suite interrupted"; exit 1' INT TERM

# Run main function
main "$@"