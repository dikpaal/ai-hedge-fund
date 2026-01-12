.PHONY: help build test clean docker-build docker-compose-up docker-compose-down k8s-deploy k8s-clean

# Go settings
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
CLI_BINARY=hedge-fund
GATEWAY_BINARY=api-gateway
PORTFOLIO_BINARY=portfolio-service
RISK_BINARY=risk-service
MARKET_BINARY=market-data-service

# Build output directory
BUILD_DIR=build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

build-cli: ## Build CLI binary
	$(GOBUILD) -o $(BUILD_DIR)/$(CLI_BINARY) ./cmd/cli

build-gateway: ## Build API Gateway binary
	$(GOBUILD) -o $(BUILD_DIR)/$(GATEWAY_BINARY) ./cmd/gateway

build-portfolio: ## Build Portfolio Service binary
	$(GOBUILD) -o $(BUILD_DIR)/$(PORTFOLIO_BINARY) ./cmd/portfolio

build-risk: ## Build Risk Service binary
	$(GOBUILD) -o $(BUILD_DIR)/$(RISK_BINARY) ./cmd/risk

build-market: ## Build Market Data Service binary
	$(GOBUILD) -o $(BUILD_DIR)/$(MARKET_BINARY) ./cmd/market

build-all: build-cli build-gateway build-portfolio build-risk build-market ## Build all binaries

docker-build: ## Build all Docker images
	docker build -f deployments/docker/Dockerfile.gateway -t hedge-fund/api-gateway:latest .
	docker build -f deployments/docker/Dockerfile.portfolio -t hedge-fund/portfolio-service:latest .
	docker build -f deployments/docker/Dockerfile.risk -t hedge-fund/risk-service:latest .
	docker build -f deployments/docker/Dockerfile.market -t hedge-fund/market-data-service:latest .
	docker build -f deployments/docker/Dockerfile.ai-service -t hedge-fund/ai-service:latest ./ai-service
	docker build -f deployments/docker/Dockerfile.cli -t hedge-fund/cli:latest .

docker-compose-up: ## Start services with Docker Compose
	docker-compose -f deployments/docker/docker-compose.yml up -d

docker-compose-down: ## Stop services with Docker Compose
	docker-compose -f deployments/docker/docker-compose.yml down

k8s-deploy: ## Deploy to Kubernetes
	kubectl apply -f deployments/k8s/

k8s-clean: ## Clean Kubernetes resources
	kubectl delete -f deployments/k8s/

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	$(GOCMD) fmt ./...

dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	$(GOMOD) init hedge-fund
	$(GOMOD) tidy
	@echo "Installing pre-commit hooks..."
	@echo "Development environment ready!"

ai-deps: ## Install Python AI service dependencies
	cd ai-service && python -m pip install -r requirements.txt

proto-gen: ## Generate gRPC code from protobuf files
	protoc --go_out=. --go-grpc_out=. pkg/shared/proto/*.proto

monitoring-up: ## Start monitoring stack
	docker-compose -f monitoring/docker-compose.monitoring.yml up -d

monitoring-down: ## Stop monitoring stack
	docker-compose -f monitoring/docker-compose.monitoring.yml down