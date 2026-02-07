# Testing Documentation

## Overview

This document describes the testing strategy and implementation for the PakaiWA Platform.

## Test Coverage

Current test coverage: **91.3%**

### Package Coverage Breakdown

| Package | Coverage | Status |
|---------|----------|--------|
| `errors` | 100.0% | ✅ Excellent |
| `db/postgres` | 86.7% | ✅ Good |

## Test Structure

### Unit Tests

#### `errors/panic_test.go`

Tests for error handling utilities:

- **TestMust_Success**: Verifies `Must` returns value when no error
- **TestMust_Panic**: Verifies `Must` panics on error
- **TestMust_WithString**: Tests `Must` with string type
- **TestMust_WithStruct**: Tests `Must` with struct type
- **TestCheck_NoError**: Verifies `Check` doesn't panic without error
- **TestCheck_WithError**: Verifies `Check` panics on error
- **TestPanicIfError_NoError**: Verifies `PanicIfError` doesn't panic without error
- **TestPanicIfError_WithError**: Verifies `PanicIfError` panics on error

#### `db/postgres/config_test.go`

Tests for database configuration:

- **TestConfig_Creation**: Verifies all config fields are set correctly
- **TestConfig_ZeroValues**: Verifies default zero values

#### `db/postgres/pgsql_test.go`

Tests for database connection:

- **TestNewDatabase_InvalidDSN**: Verifies panic on invalid DSN
- **TestNewDatabase_ValidConfig**: Integration test with real PostgreSQL (skipped if `TEST_DATABASE_URL` not set)

### Integration Tests

Integration tests require a PostgreSQL database. Set the `TEST_DATABASE_URL` environment variable:

```bash
export TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
```

## Running Tests

### Local Development

```bash
# Run all tests
make test

# Run with coverage report
make test-coverage

# Run with race detector
make test-race

# View coverage in browser
make test-coverage
open coverage.html
```

### CI/CD

Tests run automatically on:
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

The CI pipeline includes:
1. **Test Matrix**: Go versions 1.23, 1.24, 1.25
2. **PostgreSQL Service**: PostgreSQL 16 for integration tests
3. **Race Detection**: All tests run with `-race` flag
4. **Coverage Reporting**: Results uploaded to Codecov

## Test Best Practices

### 1. Table-Driven Tests

For testing multiple scenarios, use table-driven tests:

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "output1"},
        {"case2", "input2", "output2"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### 2. Testing Panics

Use `defer` and `recover` to test panic behavior:

```go
func TestPanic(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Error("Expected panic, but didn't panic")
        }
    }()
    
    FunctionThatPanics()
}
```

### 3. Integration Test Skipping

Skip integration tests when dependencies aren't available:

```go
func TestIntegration(t *testing.T) {
    if os.Getenv("TEST_DATABASE_URL") == "" {
        t.Skip("Skipping: TEST_DATABASE_URL not set")
    }
    
    // Integration test code
}
```

### 4. Test Cleanup

Always clean up resources:

```go
func TestWithResource(t *testing.T) {
    resource := setupResource()
    defer resource.Close()
    
    // Test code
}
```

## Coverage Goals

- **Minimum**: 80% overall coverage
- **Target**: 90% overall coverage
- **Critical packages**: 100% coverage for error handling and core utilities

## Adding New Tests

When adding new functionality:

1. Write tests first (TDD approach)
2. Ensure all public functions have tests
3. Test both success and error cases
4. Add integration tests for database operations
5. Run `make test-coverage` to verify coverage
6. Aim for 100% coverage on new code

## Continuous Improvement

### Current Focus Areas

1. ✅ Error handling utilities - 100% coverage
2. ✅ Database configuration - 100% coverage
3. 🔄 Database connection pooling - 86.7% coverage (improve to 100%)

### Future Enhancements

- [ ] Add benchmark tests for performance-critical code
- [ ] Add fuzzing tests for input validation
- [ ] Add end-to-end tests for complete workflows
- [ ] Set up mutation testing to verify test quality

## Troubleshooting

### Tests Failing Locally

1. **Check Go version**: Ensure you're using Go 1.23+
   ```bash
   go version
   ```

2. **Update dependencies**:
   ```bash
   make deps
   ```

3. **Clear test cache**:
   ```bash
   make clean
   ```

### Integration Tests Failing

1. **Verify PostgreSQL is running**:
   ```bash
   psql $TEST_DATABASE_URL -c "SELECT 1"
   ```

2. **Check database permissions**:
   - Ensure user has CREATE/DROP privileges
   - Verify database exists

3. **Check connection string format**:
   ```
   postgres://user:password@host:port/database?sslmode=disable
   ```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Test Coverage](https://go.dev/blog/cover)
- [pgx Testing Guide](https://github.com/jackc/pgx/wiki/Testing)
