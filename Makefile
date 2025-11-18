.PHONY: build test

# Version information (for local builds)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Build flags
LDFLAGS := -X github.com/crenshaw-dev/miaka/cmd.version=$(VERSION) \
           -X github.com/crenshaw-dev/miaka/cmd.commit=$(COMMIT) \
           -X github.com/crenshaw-dev/miaka/cmd.date=$(DATE)

# Build the binary locally with version info
build:
	@mkdir -p dist
	go build -ldflags "$(LDFLAGS)" -o dist/miaka .

# Run tests
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
