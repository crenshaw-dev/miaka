.PHONY: build test lint release

# Version information (for local builds, not used for releases)
BUILD_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Linter version
GOLANGCI_LINT_VERSION ?= latest

# Build flags
LDFLAGS := -X github.com/crenshaw-dev/miaka/cmd.version=$(BUILD_VERSION) \
           -X github.com/crenshaw-dev/miaka/cmd.commit=$(COMMIT) \
           -X github.com/crenshaw-dev/miaka/cmd.date=$(DATE)

# Build the binary locally with version info
build:
	@mkdir -p dist
	go build -ldflags "$(LDFLAGS)" -o dist/miaka .

# Run tests
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run linter
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION))
	golangci-lint run ./...

# Trigger a release by creating and pushing a tag
# Usage: make release VERSION=v0.1.0
release:
ifndef VERSION
	@echo "Error: VERSION must be set (e.g., make release VERSION=v0.1.0)"
	@exit 1
endif
	@if ! echo "$(VERSION)" | grep -q "^v[0-9]"; then \
		echo "Error: VERSION must start with 'v' followed by a number (e.g., v0.1.0)"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "✓ Tag $(VERSION) pushed. Check GitHub Actions for build progress."
