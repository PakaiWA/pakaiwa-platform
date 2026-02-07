# Quick Start Guide

## Setup

```bash
# Clone the repository
git clone https://github.com/PakaiWA/pakaiwa-platform.git
cd pakaiwa-platform

# Install dependencies
make deps

# Install pre-commit hooks (optional)
pip install pre-commit
pre-commit install
```

## Common Commands

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Format code
make fmt

# Run all CI checks locally
make ci

# Clean build artifacts
make clean
```

## Running Tests

### Unit Tests Only

```bash
make test
```

### With Coverage Report

```bash
make test-coverage
# Opens coverage.html in your browser
```

### With Race Detection

```bash
make test-race
```

### Integration Tests

```bash
# Start PostgreSQL (using Docker)
docker run -d \
  --name postgres-test \
  -e POSTGRES_USER=testuser \
  -e POSTGRES_PASSWORD=testpass \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  postgres:16-alpine

# Set environment variable
export TEST_DATABASE_URL="postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"

# Run tests
make test

# Stop PostgreSQL
docker stop postgres-test
docker rm postgres-test
```

## Development Workflow

1. **Create a feature branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make changes and write tests**
   - Write tests first (TDD)
   - Ensure tests pass: `make test`
   - Check coverage: `make test-coverage`

3. **Format and lint**
   ```bash
   make fmt
   make lint
   ```

4. **Run all CI checks**
   ```bash
   make ci
   ```

5. **Commit and push**
   ```bash
   git add .
   git commit -m "Add my feature"
   git push origin feature/my-feature
   ```

6. **Create Pull Request**
   - GitHub Actions will automatically run tests
   - Ensure all checks pass

## Troubleshooting

### Tests Failing

```bash
# Clear test cache
make clean

# Update dependencies
make deps

# Try again
make test
```

### Linter Errors

```bash
# Auto-fix formatting issues
make fmt

# Check specific issues
make lint
```

### Import Issues

```bash
# Tidy go.mod
make tidy
```

## Project Structure

```
pakaiwa-platform/
├── db/postgres/       # Database utilities
├── errors/            # Error handling
├── .github/workflows/ # CI/CD configuration
├── Makefile          # Development commands
├── README.md         # Project overview
└── TESTING.md        # Testing documentation
```

## Getting Help

- Check [README.md](README.md) for detailed documentation
- Check [TESTING.md](TESTING.md) for testing guide
- Run `make help` for available commands
- Open an issue on GitHub

## CI/CD Status

The project uses GitHub Actions for CI/CD:

- ✅ Tests run on Go 1.23, 1.24, 1.25
- ✅ PostgreSQL integration tests
- ✅ Code coverage reporting
- ✅ Linting with golangci-lint
- ✅ Build verification

Check the [Actions tab](https://github.com/PakaiWA/pakaiwa-platform/actions) for status.
