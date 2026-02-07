# PakaiWA Platform

[![CI Tests](https://github.com/PakaiWA/pakaiwa-platform/actions/workflows/ci.yml/badge.svg)](https://github.com/PakaiWA/pakaiwa-platform/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/PakaiWA/pakaiwa-platform)](https://goreportcard.com/report/github.com/PakaiWA/pakaiwa-platform)
[![codecov](https://codecov.io/gh/PakaiWA/pakaiwa-platform/branch/main/graph/badge.svg)](https://codecov.io/gh/PakaiWA/pakaiwa-platform)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

A robust Go platform for PakaiWA with PostgreSQL database support and comprehensive error handling utilities.

## Features

- 🗄️ **PostgreSQL Integration**: Production-ready database connection pooling with pgx/v5
- 🛡️ **Error Handling**: Utility functions for panic-based error handling
- ✅ **Comprehensive Testing**: 91.3% code coverage with unit and integration tests
- 🔄 **CI/CD**: Automated testing with GitHub Actions
- 📊 **Code Quality**: Linting with golangci-lint and pre-commit hooks

## Requirements

- Go 1.23 or higher (tested with 1.23, 1.24, 1.25)
- PostgreSQL 12+ (for integration tests)

## Installation

```bash
git clone https://github.com/PakaiWA/pakaiwa-platform.git
cd pakaiwa-platform
go mod download
```

## Usage

### Database Connection

```go
package main

import (
    "context"
    "time"
    
    "github.com/PakaiWA/pakaiwa-platform/db/postgres"
    "github.com/sirupsen/logrus"
)

func main() {
    log := logrus.New()
    
    cfg := postgres.Config{
        DSN:               "postgres://user:pass@localhost:5432/dbname",
        MinConns:          2,
        MaxConns:          10,
        MaxConnIdleTime:   30 * time.Minute,
        HealthCheckPeriod: 1 * time.Minute,
        ConnectTimeout:    5 * time.Second,
    }
    
    ctx := context.Background()
    pool := postgres.NewDatabase(ctx, log, cfg)
    defer pool.Close()
    
    // Use pool for database operations
}
```

### Error Handling Utilities

```go
package main

import "github.com/PakaiWA/pakaiwa-platform/errors"

func main() {
    // Must - returns value or panics on error
    value := errors.Must(someFunction())
    
    // Check - panics if error is not nil
    errors.Check(someOperation())
    
    // PanicIfError - prints and panics if error is not nil
    errors.PanicIfError(anotherOperation())
}
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detector
make test-race

# Run short tests only
make test-short
```

### Code Quality

```bash
# Run linter
make lint

# Format code
make fmt

# Run go vet
make vet

# Run all CI checks locally
make ci
```

### Integration Tests

To run integration tests with a real PostgreSQL database:

```bash
# Set the test database URL
export TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"

# Run tests
go test -v ./...
```

## Project Structure

```
.
├── db/
│   └── postgres/          # PostgreSQL database utilities
│       ├── config.go      # Database configuration
│       ├── pgsql.go       # Connection pool management
│       ├── config_test.go # Config tests
│       └── pgsql_test.go  # Database tests
├── errors/                # Error handling utilities
│   ├── panic.go          # Panic-based error helpers
│   └── panic_test.go     # Error handling tests
├── .github/
│   └── workflows/
│       └── ci.yml        # GitHub Actions CI workflow
├── .golangci.yml         # Linter configuration
├── .pre-commit-config.yaml # Pre-commit hooks
├── Makefile              # Development tasks
├── go.mod                # Go module definition
└── README.md             # This file
```

## Testing

The project maintains high test coverage:

- **errors package**: 100% coverage
- **db/postgres package**: 86.7% coverage
- **Overall**: 91.3% coverage

Tests include:
- Unit tests for all public functions
- Integration tests with PostgreSQL (optional)
- Race condition detection
- Coverage reporting

## CI/CD

The project uses GitHub Actions for continuous integration:

- ✅ **Multi-version testing**: Tests run on Go 1.23, 1.24, and 1.25
- ✅ **PostgreSQL service**: Integration tests with PostgreSQL 16
- ✅ **Code coverage**: Automatic coverage reporting to Codecov
- ✅ **Linting**: golangci-lint with comprehensive rules
- ✅ **Build verification**: Ensures code compiles successfully

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Pre-commit Hooks

This project uses pre-commit hooks. Install them with:

```bash
pip install pre-commit
pre-commit install
```

## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

See [LICENSE](LICENSE) for more details.

## Author

**KAnggara** - [GitHub](https://github.com/PakaiWA)

## Acknowledgments

- Built with [pgx](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit
- Logging with [logrus](https://github.com/sirupsen/logrus)
- Linting with [golangci-lint](https://golangci-lint.run/)
