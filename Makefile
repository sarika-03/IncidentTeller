# IncidentTeller Makefile
# Provides convenient commands for development, testing, and deployment

.PHONY: help build run test clean docker docker-run deps lint format vet migrate

# Default target
help: ## Show this help message
	@echo 'IncidentTeller - AI-Powered SRE Incident Analysis'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
deps: ## Install Go dependencies
	@echo 'Installing dependencies...'
	go mod download
	go mod tidy

build: deps ## Build the application
	@echo 'Building IncidentTeller...'
	go build -o bin/incident-teller ./cmd/incident-teller

run: ## Run the application (development mode)
	@echo 'Running IncidentTeller in development mode...'
	go run ./cmd/incident-teller

run-memory: ## Run with in-memory database for testing
	@echo 'Running with in-memory database...'
	DB_TYPE=memory OBSERVABILITY_LOG_LEVEL=debug go run ./cmd/incident-teller

run-sqlite: ## Run with SQLite database
	@echo 'Running with SQLite database...'
	DB_TYPE=sqlite DB_SQLITE_PATH=./dev.db go run ./cmd/incident-teller

# Testing targets
test: ## Run all tests
	@echo 'Running tests...'
	go test -v ./...

test-cover: ## Run tests with coverage
	@echo 'Running tests with coverage...'
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo 'Coverage report generated: coverage.html'

test-integration: ## Run integration tests
	@echo 'Running integration tests...'
	go test -v -tags=integration ./...

test-race: ## Run tests with race detector
	@echo 'Running tests with race detector...'
	go test -v -race ./...

# Quality targets
lint: ## Run linter
	@echo 'Running linter...'
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo 'golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2'; \
	fi

format: ## Format Go code
	@echo 'Formatting code...'
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo 'Running go vet...'
	go vet ./...

check: lint vet test ## Run all quality checks

# Database targets
migrate-up: ## Run database migrations up
	@echo 'Running database migrations...'
	DB_TYPE=sqlite DB_SQLITE_PATH=./incident_teller.db go run ./cmd/incident-teller -migrate

migrate-down: ## Run database migrations down
	@echo 'Rolling back database migrations...'
	DB_TYPE=sqlite DB_SQLITE_PATH=./incident_teller.db go run ./cmd/incident-teller -migrate-down

# Docker targets
docker: ## Build Docker image
	@echo 'Building Docker image...'
	docker build -t incident-teller:latest .

docker-run: docker ## Run Docker container
	@echo 'Running Docker container...'
	docker run -p 8080:8080 -p 9090:9090 \
		-v $(PWD)/config.yaml:/app/config.yaml \
		incident-teller:latest

docker-dev: ## Run Docker container with development config
	@echo 'Running Docker container in development mode...'
	docker run -p 8080:8080 -p 9090:9090 \
		-e DB_TYPE=memory \
		-e AI_ENABLED=true \
		-e OBSERVABILITY_LOG_LEVEL=debug \
		incident-teller:latest

docker-compose-up: ## Start services with Docker Compose
	@echo 'Starting services with Docker Compose...'
	docker-compose -f docker-compose.yml up -d

docker-compose-down: ## Stop Docker Compose services
	@echo 'Stopping Docker Compose services...'
	docker-compose -f docker-compose.yml down

# Deployment targets
build-all: ## Build for multiple platforms
	@echo 'Building for multiple platforms...'
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/incident-teller-linux-amd64 ./cmd/incident-teller
	GOOS=darwin GOARCH=amd64 go build -o bin/incident-teller-darwin-amd64 ./cmd/incident-teller
	GOOS=windows GOARCH=amd64 go build -o bin/incident-teller-windows-amd64.exe ./cmd/incident-teller
	GOOS=linux GOARCH=arm64 go build -o bin/incident-teller-linux-arm64 ./cmd/incident-teller
	@echo 'Built binaries:'
	@ls -la bin/

package: build-all ## Create distribution packages
	@echo 'Creating distribution packages...'
	mkdir -p dist
	cd bin && \
	for binary in incident-teller-*; do \
		$$binary --version > ../dist/$${binary%.exe}.txt 2>/dev/null || true; \
		tar -czf ../dist/$$binary.tar.gz $$binary ../config.yaml ../README.md; \
	done
	@echo 'Packages created in dist/'

# Production targets
prod-build: ## Build optimized production binary
	@echo 'Building production binary...'
	CGO_ENABLED=0 GOOS=linux go build \
		-ldflags='-w -s -extldflags "-static"' \
		-a -installsuffix cgo \
		-o bin/incident-teller-linux ./cmd/incident-teller

prod-docker: ## Production Docker build
	@echo 'Building production Docker image...'
	docker build -f Dockerfile.prod -t incident-teller:prod .

prod-run: prod-docker ## Run production Docker container
	@echo 'Running production Docker container...'
	docker run -d \
		--name incident-teller \
		-p 8080:8080 \
		-p 9090:9090 \
		-v $(PWD)/config.yaml:/app/config.yaml \
		--restart unless-stopped \
		incident-teller:prod

# Cleanup targets
clean: ## Clean build artifacts
	@echo 'Cleaning up...'
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html
	rm -f *.db
	go clean -cache

clean-docker: ## Clean Docker resources
	@echo 'Cleaning Docker resources...'
	docker system prune -f
	docker volume prune -f

# Security targets
security-scan: ## Run security scan
	@echo 'Running security scan...'
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo 'gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest'; \
	fi

deps-check: ## Check for outdated dependencies
	@echo 'Checking for outdated dependencies...'
	go list -u -m all

# Monitoring targets
dev-deps: ## Install development dependencies
	@echo 'Installing development dependencies...'
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/air-verse/air@latest

dev-watch: ## Run with hot reload
	@echo 'Running with hot reload...'
	@air

# CI/CD targets
ci: lint vet test security-scan ## Run all CI checks
	@echo 'All CI checks passed!'

# Quick start
quick-start: ## Quick start for local development
	@echo 'Setting up quick start environment...'
	cp config.yaml.template config.yaml.local
	@echo 'Created config.yaml.local - customize as needed'
	@echo ''
	@echo 'To start development:'
	@echo '  make run-memory          # Run with in-memory database'
	@echo '  make run-sqlite         # Run with SQLite database'
	@echo '  make docker-dev          # Run with Docker'
	@echo ''
	@echo 'Config file: config.yaml.local'

# Version targets
version: ## Show version information
	@echo 'IncidentTeller v1.0.0'
	@echo 'Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)'
	@echo 'Build date: $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")'

# Health check
health: ## Check application health
	@echo 'Checking application health...'
	curl -f http://localhost:8080/health || echo 'Application not running or unhealthy'

# Documentation targets
docs: ## Generate documentation
	@echo 'Generating documentation...'
	@if command -v godoc >/dev/null 2>&1; then \
		echo 'Starting godoc server on :6060...'; \
		godoc -http=:6060; \
	else \
		echo 'godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest'; \
	fi

docs-serve: docs ## Alias for docs target