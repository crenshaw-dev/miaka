.PHONY: build test

# Build the binary
build:
	@mkdir -p dist
	go build -o dist/miaka .

# Run tests
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
