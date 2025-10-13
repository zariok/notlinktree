.PHONY: build-ui build-embed build-go build

# Default target
all: build

# Detect container runtime (Docker or Podman)
CONTAINER_RUNTIME := $(shell command -v docker >/dev/null 2>&1 && echo docker || (command -v podman >/dev/null 2>&1 && echo podman) || echo docker)
COMPOSE_CMD := $(shell if [ "$(CONTAINER_RUNTIME)" = "podman" ]; then echo "podman-compose"; else echo "docker-compose"; fi)

# Clean build artifacts
clean:
	rm -rf dist/* embed/ui/*

# Create embed directory
embed-dir:
	mkdir -p embed/ui

# Build UI application
build-ui: embed-dir
	@echo "Building Next.js application..."
	cd ui && rm -rf .next out && npm install && npm run build && cp -r out/* ../embed/ui/

# Get latest tag or default
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")
# Get short hash or default
GIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
# Build date
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
# Version string
VERSION := v$(GIT_TAG)-$(GIT_HASH)
# LDFLAGS
LDFLAGS := -X main.version=$(VERSION) -X 'main.buildDate=$(BUILD_DATE)'

# Build for Linux x86_64
build-linux: build-ui
	@echo "Building for Linux x86_64..."
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/notlinktree-linux-amd64

# Build for macOS ARM64
build-darwin: build-ui
	@echo "Building for macOS ARM64..."
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/notlinktree-darwin-arm64

# Build for all platforms
build: clean embed-dir
	$(MAKE) build-linux
	$(MAKE) build-darwin
	@echo "Build complete! Binaries are in the dist directory:"
	ls -l dist/

# Development target
dev-ui:
	cd ui && npm start

deps:
	go mod download
	cd ui && npm install

build-embed: clean-embed build-ui

clean-embed:
	rm -rf embed/ui/*

build-go:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/notlinktree-linux-amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/notlinktree-darwin-arm64

# Container runtime info
runtime-info:
	@echo "Detected container runtime: $(CONTAINER_RUNTIME)"
	@echo "Using compose command: $(COMPOSE_CMD)"
	@if [ "$(CONTAINER_RUNTIME)" = "podman" ]; then \
		echo "Note: Using Podman instead of Docker"; \
	fi

# Test targets
.PHONY: test test-go test-ui test-coverage test-all docker-test docker-lint docker-clean runtime-info

# Run all Go tests
test-go:
	@echo "Running Go unit tests..."
	go test -v -race ./...

# Run Go tests with coverage
test-coverage:
	@echo "Running Go tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=count ./...
	go tool cover -func=coverage.out

# Run UI tests (linting and build)
test-ui:
	@echo "Running UI tests..."
	cd ui && npm ci
	cd ui && npm run lint
	cd ui && npm run build

# Run all tests
test-all: test-go test-ui
	@echo "All tests completed successfully"

# Quick test for development (Go tests only)
test: test-go
	@echo "Quick test completed"

# Docker/Podman test targets
.PHONY: docker-test docker-lint docker-clean runtime-info

# Run container tests
docker-test: docker-lint build-ui
	@echo "Building main notlinktree image..."
	$(CONTAINER_RUNTIME) build --build-arg NLT_JWT_SECRET=testingonly -t notlinktree:test .
	@echo "Verifying image was built..."
	$(CONTAINER_RUNTIME) images notlinktree:test
	@echo "Running container tests using $(CONTAINER_RUNTIME)..."
	cd test/docker && $(COMPOSE_CMD) -f docker-compose.test.yml up --build --abort-on-container-exit; \
	test_exit_code=$$?; \
	cd ../.. && $(MAKE) docker-clean; \
	exit $$test_exit_code
	@echo "Container tests complete"

# Lint Dockerfile
docker-lint:
	@echo "Linting Dockerfile using $(CONTAINER_RUNTIME)..."
	@if command -v hadolint >/dev/null 2>&1; then \
		hadolint Dockerfile || true; \
	elif [ "$(CONTAINER_RUNTIME)" = "podman" ]; then \
		podman run --rm -i hadolint/hadolint:latest < Dockerfile || true; \
	else \
		docker run --rm -i hadolint/hadolint:latest < Dockerfile || true; \
	fi

# Clean container test artifacts
docker-clean:
	@echo "Cleaning up container test artifacts..."
	cd test/docker && $(COMPOSE_CMD) -f docker-compose.test.yml down -v --remove-orphans
	@echo "Removing test containers..."
	$(CONTAINER_RUNTIME) container prune -f
	@echo "Removing test images..."
	$(CONTAINER_RUNTIME) image rm -f localhost/notlinktree:test localhost/notlinktree:test-runner localhost/notlinktree:test-runner-local localhost/notlinktree:test-runner-integration 2>/dev/null || true
	@echo "Cleaning up unused resources..."
	$(CONTAINER_RUNTIME) system prune -f --filter "label=test=true"
	@echo "Docker cleanup complete"

build: build-embed build-go 