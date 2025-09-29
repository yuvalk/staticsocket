# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=staticsocket
BINARY_UNIX=$(BINARY_NAME)_unix

# Build info
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_DATE = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS = -ldflags "-w -s -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)"

.PHONY: all build clean test coverage deps lint help docker

all: clean deps test build ## Build everything

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v

build-linux: ## Build for Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) -v

build-all: ## Build for all platforms
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -rf dist/

test: ## Run tests
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-short: ## Run short tests
	$(GOTEST) -short -v ./...

coverage: test ## Generate coverage report
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

deps-update: ## Update dependencies
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

lint: ## Run linters
	golangci-lint run

lint-fix: ## Fix linting issues
	golangci-lint run --fix

format: ## Format code
	$(GOCMD) fmt ./...
	goimports -w .

vet: ## Run go vet
	$(GOCMD) vet ./...

security: ## Run security checks
	gosec ./...
	govulncheck ./...

docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):latest .

docker-run: docker-build ## Run Docker container
	docker run --rm -v $(PWD)/testdata:/app/testdata $(BINARY_NAME):latest -path /app/testdata/samples

install: build ## Install binary to GOPATH/bin
	$(GOCMD) install $(LDFLAGS)

uninstall: ## Remove binary from GOPATH/bin
	rm -f $(GOPATH)/bin/$(BINARY_NAME)

run: build ## Build and run with sample data
	./$(BINARY_NAME) -path testdata/samples -format json

demo: build ## Run demo with all output formats
	@echo "=== JSON Output ==="
	./$(BINARY_NAME) -path testdata/samples -format json
	@echo ""
	@echo "=== YAML Output ==="
	./$(BINARY_NAME) -path testdata/samples -format yaml
	@echo ""
	@echo "=== CSV Output ==="
	./$(BINARY_NAME) -path testdata/samples -format csv

benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

profile: build ## Run with profiling
	./$(BINARY_NAME) -path testdata/samples -cpuprofile cpu.prof -memprofile mem.prof

check: deps lint vet test ## Run all checks

ci: check build ## Run CI pipeline locally

help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)