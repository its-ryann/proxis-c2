.PHONY: build-server build-agent-windows build-agent-linux build-agent-darwin test clean lint docker-build generate-certs help

# Default target
help:
	@echo "Available targets:"
	@echo "  build-server        - Build server binary for current platform"
	@echo "  build-agent-windows - Build agent binary for Windows amd64"
	@echo "  build-agent-linux   - Build agent binary for Linux amd64"
	@echo "  build-agent-darwin  - Build agent binary for macOS amd64"
	@echo "  test                - Run tests with race detection and coverage"
	@echo "  lint                - Run golangci-lint"
	@echo "  docker-build        - Build Docker image"
	@echo "  generate-certs      - Generate mTLS certificates"
	@echo "  clean               - Remove build artifacts"

# Build server binary for current platform
build-server:
	go build -ldflags "-s -w" -o bin/proxis-server ./cmd/server/

# Build agent binary for Windows
build-agent-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -tags "windows" -o bin/proxis-agent-windows-amd64.exe ./cmd/agent/

# Build agent binary for Linux
build-agent-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -tags "linux" -o bin/proxis-agent-linux-amd64 ./cmd/agent/

# Build agent binary for macOS
build-agent-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -tags "darwin" -o bin/proxis-agent-darwin-amd64 ./cmd/agent/

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run linter
lint:
	golangci-lint run ./...

# Build Docker image
docker-build:
	docker build -t proxis-c2:latest -f deployments/docker/Dockerfile .

# Generate certificates
generate-certs:
	chmod +x scripts/generate-certs.sh
	./scripts/generate-certs.sh

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out