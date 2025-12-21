.PHONY: build test lint clean install run help

# Build variables
BINARY_NAME=orchestr
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/vkrishna03/orchestr/internal/cli.Version=$(VERSION) -X github.com/vkrishna03/orchestr/internal/cli.Commit=$(COMMIT) -X github.com/vkrishna03/orchestr/internal/cli.BuildDate=$(BUILD_DATE)"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/orchestr

## build-prod: Build optimized production binary
build-prod:
	CGO_ENABLED=0 go build $(LDFLAGS) -ldflags "-s -w" -o $(BINARY_NAME) ./cmd/orchestr

## install: Install binary to $GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/orchestr

## run: Run the CLI (use ARGS="test ." to pass arguments)
run:
	go run $(LDFLAGS) ./cmd/orchestr $(ARGS)

## test: Run tests
test:
	go test -race ./...

## test-cover: Run tests with coverage
test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linter (using built-in go vet)
lint:
	go vet ./...

## fmt: Format code
fmt:
	go fmt ./...

## tidy: Tidy go modules
tidy:
	go mod tidy

## clean: Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## snapshot: Build snapshot release (for testing)
snapshot:
	goreleaser release --snapshot --clean

## release-dry: Dry run release
release-dry:
	goreleaser release --skip=publish --clean
