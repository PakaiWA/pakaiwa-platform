# Makefile for PakaiWA Platform

.PHONY: test test-coverage test-race lint fmt vet clean help

# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=gofmt
GOLINT=golangci-lint

# Test parameters
TEST_FLAGS=-v -race -coverprofile=coverage.out -covermode=atomic
TEST_PACKAGES=./...

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	$(GOTEST) -v $(TEST_PACKAGES)

test-coverage: ## Run tests with coverage
	$(GOTEST) $(TEST_FLAGS) $(TEST_PACKAGES)
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detector
	$(GOTEST) -race $(TEST_PACKAGES)

test-short: ## Run short tests
	$(GOTEST) -short $(TEST_PACKAGES)

lint: ## Run linter
	$(GOLINT) run --timeout=5m

fmt: ## Format code
	$(GOFMT) -s -w .

vet: ## Run go vet
	$(GOVET) $(TEST_PACKAGES)

clean: ## Clean build artifacts and test cache
	$(GOCMD) clean -testcache
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	$(GOCMD) mod download
	$(GOCMD) mod verify

tidy: ## Tidy go.mod
	$(GOCMD) mod tidy

build: ## Build the project
	$(GOCMD) build -v $(TEST_PACKAGES)

ci: deps vet lint test-coverage ## Run all CI checks locally

.DEFAULT_GOAL := help
