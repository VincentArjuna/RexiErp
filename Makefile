# RexiERP Makefile
# RexiERP: Indonesian ERP System for MSMEs

.PHONY: help build test lint fmt clean up down logs install-tools docker-build docker-clean pre-commit-install ci

# Default target
.DEFAULT_GOAL := help

# Variables
DOCKER_COMPOSE_FILE := deployments/docker-compose/docker-compose.yml
GO_FILES := $(shell find . -name "*.go" -type f)
SERVICES := authentication-service inventory-service accounting-service hr-service crm-service notification-service integration-service

# Colors
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
PURPLE := \033[0;35m
CYAN := \033[0;36m
WHITE := \033[0;37m
RESET := \033[0m

help: ## Show this help message
	@echo "$(CYAN)RexiERP - Indonesian ERP System for MSMEs$(RESET)"
	@echo "$(CYAN)==============================================$(RESET)"
	@echo ""
	@echo "$(YELLOW)Available commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development Commands
install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(RESET)"
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	@echo "Installing gosec..."
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "Installing swag..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Installing mockgen..."
	@go install github.com/golang/mock/mockgen@latest
	@echo "Installing pre-commit..."
	@pip3 install pre-commit || pip install pre-commit
	@echo "$(GREEN)Development tools installed successfully!$(RESET)"

pre-commit-install: ## Install pre-commit hooks
	@echo "$(BLUE)Installing pre-commit hooks...$(RESET)"
	@pre-commit install
	@pre-commit install --hook-type commit-msg
	@echo "$(GREEN)Pre-commit hooks installed successfully!$(RESET)"

# Go Commands
fmt: ## Format Go code
	@echo "$(BLUE)Formatting Go code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)Code formatted successfully!$(RESET)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	@go vet ./...
	@echo "$(GREEN)Vet completed successfully!$(RESET)"

lint: ## Run golangci-lint
	@echo "$(BLUE)Running golangci-lint...$(RESET)"
	@golangci-lint run ./...
	@echo "$(GREEN)Linting completed successfully!$(RESET)"

security: ## Run security scan with gosec
	@echo "$(BLUE)Running security scan...$(RESET)"
	@gosec ./...
	@echo "$(GREEN)Security scan completed!$(RESET)"

test: ## Run all tests
	@echo "$(BLUE)Running tests...$(RESET)"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests completed successfully!$(RESET)"
	@echo "$(CYAN)Coverage report: coverage.out$(RESET)"

test-coverage: ## Run tests with coverage and generate HTML report
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(RESET)"

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(RESET)"
	@go test -v -tags=integration ./tests/integration/...
	@echo "$(GREEN)Integration tests completed!$(RESET)"

build: ## Build all services
	@echo "$(BLUE)Building all services...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Building $$service...$(RESET)"; \
		go build -o bin/$$service ./cmd/$$service/; \
		if [ $$? -eq 0 ]; then \
			echo "$(GREEN)✓ $$service built successfully$(RESET)"; \
		else \
			echo "$(RED)✗ $$service build failed$(RESET)"; \
			exit 1; \
		fi; \
	done
	@echo "$(GREEN)All services built successfully!$(RESET)"

build-service: ## Build a specific service (usage: make build-service SERVICE=authentication-service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: SERVICE parameter is required$(RESET)"; \
		echo "$(YELLOW)Usage: make build-service SERVICE=authentication-service$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Building $(SERVICE)...$(RESET)"
	@go build -o bin/$(SERVICE) ./cmd/$(SERVICE)/
	@echo "$(GREEN)$(SERVICE) built successfully!$(RESET)"

run-service: ## Run a specific service (usage: make run-service SERVICE=authentication-service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: SERVICE parameter is required$(RESET)"; \
		echo "$(YELLOW)Usage: make run-service SERVICE=authentication-service$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Running $(SERVICE)...$(RESET)"
	@go run ./cmd/$(SERVICE)/main.go

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(RESET)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -f *.log
	@go clean -cache -modcache -testcache
	@echo "$(GREEN)Clean completed successfully!$(RESET)"

# Docker Commands
docker-build: ## Build Docker images for all services
	@echo "$(BLUE)Building Docker images...$(RESET)"
	@for service in $(SERVICES); do \
		echo "$(YELLOW)Building Docker image for $$service...$(RESET)"; \
		docker build -f cmd/$$service/Dockerfile -t rexi-$$service:latest .; \
		if [ $$? -eq 0 ]; then \
			echo "$(GREEN)✓ $$service Docker image built successfully$(RESET)"; \
		else \
			echo "$(RED)✗ $$service Docker image build failed$(RESET)"; \
			exit 1; \
		fi; \
	done
	@echo "$(GREEN)All Docker images built successfully!$(RESET)"

docker-clean: ## Clean Docker resources
	@echo "$(BLUE)Cleaning Docker resources...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans
	@docker system prune -f
	@echo "$(GREEN)Docker cleanup completed!$(RESET)"

# Docker Compose Commands
up: ## Start all services with Docker Compose
	@echo "$(BLUE)Starting all services...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "$(GREEN)Services started successfully!$(RESET)"
	@echo "$(CYAN)Services available at:$(RESET)"
	@echo "$(YELLOW)- API Gateway: http://localhost:8080$(RESET)"
	@echo "$(YELLOW)- Grafana: http://localhost:3000 (admin/admin)$(RESET)"
	@echo "$(YELLOW)- Prometheus: http://localhost:9090$(RESET)"
	@echo "$(YELLOW)- RabbitMQ Management: http://localhost:15672 (guest/guest)$(RESET)"

down: ## Stop all services with Docker Compose
	@echo "$(BLUE)Stopping all services...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_FILE) down
	@echo "$(GREEN)Services stopped successfully!$(RESET)"

logs: ## Show logs from all services
	@docker compose -f $(DOCKER_COMPOSE_FILE) logs -f

logs-service: ## Show logs from a specific service (usage: make logs-service SERVICE=postgres)
	@if [ -z "$(SERVICE)" ]; then \
		echo "$(RED)Error: SERVICE parameter is required$(RESET)"; \
		echo "$(YELLOW)Usage: make logs-service SERVICE=postgres$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Showing logs for $(SERVICE)...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_FILE) logs -f $(SERVICE)

restart: ## Restart all services
	@echo "$(BLUE)Restarting all services...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_FILE) restart
	@echo "$(GREEN)Services restarted successfully!$(RESET)"

status: ## Show status of all services
	@echo "$(BLUE)Service status:$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_FILE) ps

# Database Commands
migrate-up: ## Run database migrations
	@echo "$(BLUE)Running database migrations...$(RESET)"
	@# Placeholder for migration command
	@echo "$(YELLOW)TODO: Add migration command$(RESET)"

migrate-down: ## Rollback database migrations
	@echo "$(BLUE)Rolling back database migrations...$(RESET)"
	@# Placeholder for migration command
	@echo "$(YELLOW)TODO: Add migration rollback command$(RESET)"

db-seed: ## Seed database with test data
	@echo "$(BLUE)Seeding database with test data...$(RESET)"
	@# Placeholder for seeding command
	@echo "$(YELLOW)TODO: Add database seeding command$(RESET)"

# Development Workflow Commands
dev-setup: install-tools pre-commit-install ## Complete development environment setup
	@echo "$(GREEN)Development environment setup completed!$(RESET)"
	@echo "$(CYAN)Next steps:$(RESET)"
	@echo "$(YELLOW)1. Copy .env.example to .env and configure your environment$(RESET)"
	@echo "$(YELLOW)2. Run 'make up' to start all services$(RESET)"
	@echo "$(YELLOW)3. Run 'make test' to run tests$(RESET)"

dev-check: fmt vet lint security test ## Run all development checks
	@echo "$(GREEN)All development checks passed!$(RESET)"

# CI/CD Commands
ci: ## Run CI pipeline locally
	@echo "$(BLUE)Running CI pipeline...$(RESET)"
	@make fmt
	@make vet
	@make lint
	@make security
	@make test
	@echo "$(GREEN)CI pipeline completed successfully!$(RESET)"

# Documentation Commands
docs-api: ## Generate API documentation
	@echo "$(BLUE)Generating API documentation...$(RESET)"
	@# Placeholder for API documentation generation
	@echo "$(YELLOW)TODO: Add API documentation generation$(RESET)"

docs-lint: ## Lint documentation files
	@echo "$(BLUE)Linting documentation files...$(RESET)"
	@yamllint docs/ deployments/
	@echo "$(GREEN)Documentation linting completed!$(RESET)"

# Utility Commands
version: ## Show version information
	@echo "$(CYAN)RexiERP Information:$(RESET)"
	@echo "$(YELLOW)Go Version:$(RESET) $(shell go version)"
	@echo "$(YELLOW)Docker Version:$(RESET) $(shell docker --version)"
	@echo "$(YELLOW)Docker Compose Version:$(RESET) $(shell docker compose version)"

check-deps: ## Check for outdated dependencies
	@echo "$(BLUE)Checking for outdated dependencies...$(RESET)"
	@go list -u -m all
	@echo "$(GREEN)Dependency check completed!$(RESET)"

update-deps: ## Update Go dependencies
	@echo "$(BLUE)Updating Go dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)Dependencies updated successfully!$(RESET)"

# Health Check Commands
health: ## Check health of all services
	@echo "$(BLUE)Checking service health...$(RESET)"
	@curl -s http://localhost:8080/health || echo "$(RED)API Gateway is down$(RESET)"
	@curl -s http://localhost:3000/api/health || echo "$(RED)Grafana is down$(RESET)"
	@curl -s http://localhost:9090/-/healthy || echo "$(RED)Prometheus is down$(RESET)"
	@echo "$(GREEN)Health check completed!$(RESET)"

# Quick Development Commands
quick-build: fmt build ## Format and build all services
	@echo "$(GREEN)Quick build completed!$(RESET)"

quick-test: fmt test ## Format and run tests
	@echo "$(GREEN)Quick test completed!$(RESET)"

# Backup and Restore Commands
backup-db: ## Backup database
	@echo "$(BLUE)Creating database backup...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_FILE) exec -T postgres pg_dump -U rexi rexi_erp > backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)Database backup created!$(RESET)"

# Environment Commands
env-check: ## Check if .env file exists
	@if [ ! -f .env ]; then \
		echo "$(RED).env file not found!$(RESET)"; \
		echo "$(YELLOW)Please copy .env.example to .env and configure it.$(RESET)"; \
		exit 1; \
	else \
		echo "$(GREEN).env file found!$(RESET)"; \
	fi

# Project Statistics Commands
stats: ## Show project statistics
	@echo "$(CYAN)Project Statistics:$(RESET)"
	@echo "$(YELLOW)Go files:$(RESET) $(shell find . -name "*.go" -type f | wc -l)"
	@echo "$(YELLOW)Lines of Go code:$(RESET) $(shell find . -name "*.go" -type f -exec wc -l {} + | tail -1 | awk '{print $$1}')"
	@echo "$(YELLOW)Test files:$(RESET) $(shell find . -name "*_test.go" -type f | wc -l)"
	@echo "$(YELLOW)Docker files:$(RESET) $(shell find . -name "Dockerfile" -type f | wc -l)"
	@echo "$(YELLOW)YAML files:$(RESET) $(shell find . -name "*.yml" -o -name "*.yaml" -type f | wc -l)"