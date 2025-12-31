.PHONY: help build test lint clean docker-build release

VERSION := $(shell cat VERSION)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building version $(VERSION) ($(GIT_COMMIT))"
	go build -ldflags "$(LDFLAGS)" -o bin/server ./cmd/server

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-integration: ## Run integration tests
	go test -v -tags=integration ./tests/integration/...

lint: ## Run linters
	golangci-lint run --timeout=5m

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out

docker-build: ## Build Docker image
	docker build -t anonymous-support-api:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		.

release: ## Create a new release
	@echo "Creating release $(VERSION)"
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)

.DEFAULT_GOAL := help
