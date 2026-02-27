.PHONY: build test clean lint run package build-all

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

# Build for current platform
build:
	go build $(LDFLAGS) -o bin/openllama ./cmd/openllama

# Run the application
run: build
	./bin/openllama

# Build for all platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/openllama-linux-amd64 ./cmd/openllama

build-darwin:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/openllama-darwin-arm64 ./cmd/openllama

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/openllama-windows-amd64.exe ./cmd/openllama

# Run tests
test:
	go test ./... -v -race -count=1

# Run tests (short)
test-short:
	go test ./... -short -count=1

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -rf bin/ dist/

# Package for distribution
package: build-all
	mkdir -p dist
	tar -czf dist/openllama-$(VERSION)-linux-amd64.tar.gz -C bin openllama-linux-amd64
	tar -czf dist/openllama-$(VERSION)-darwin-arm64.tar.gz -C bin openllama-darwin-arm64
	zip -j dist/openllama-$(VERSION)-windows-amd64.zip bin/openllama-windows-amd64.exe
