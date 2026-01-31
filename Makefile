.PHONY: build build-all clean test install

# Binary name
BINARY=powerctl

# Build directory
DIST=dist

# Version info
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Build flags
LDFLAGS=-ldflags "-s -w \
	-X github.com/kristofferrisa/powerctl-cli/internal/commands.Version=$(VERSION) \
	-X github.com/kristofferrisa/powerctl-cli/internal/commands.GitCommit=$(GIT_COMMIT) \
	-X github.com/kristofferrisa/powerctl-cli/internal/commands.BuildDate=$(BUILD_DATE)"

# Default target
build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/powerctl

# Install to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/powerctl

# Run tests
test:
	go test -v ./...

# Build for all platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-amd64 ./cmd/powerctl
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-arm64 ./cmd/powerctl

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-amd64 ./cmd/powerctl
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-arm64 ./cmd/powerctl

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-windows-amd64.exe ./cmd/powerctl

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -rf $(DIST)

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy
