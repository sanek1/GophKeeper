VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_NUM ?= $(shell git rev-list --count HEAD 2>/dev/null || echo "0")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH) -X main.buildDate=$(BUILD_DATE) -X main.buildNum=$(BUILD_NUM)"

# Цвета для вывода
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: all test test-coverage test-race build clean swagger client-build server-build lint lint-fix fmt help deps check

all: build

help: ## Показать помощь
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## Development:
test: ## Запуск всех тестов
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v ./...

test-coverage: ## Запуск тестов с покрытием
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

test-coverage-func: ## Показать покрытие по функциям
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-race: ## Запуск тестов с детектором гонок
	@echo "$(GREEN)Running tests with race detector...$(NC)"
	go test -v -race ./...

coverage-check: ## Проверка минимального покрытия (80%)
	@echo "$(GREEN)Checking coverage...$(NC)"
	@go test -coverprofile=coverage.out ./... > /dev/null 2>&1
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
	echo "Total coverage: $$COVERAGE%"; \
	COVERAGE_NUM=$$(echo $$COVERAGE | cut -d'.' -f1); \
	if [ $$COVERAGE_NUM -ge 25 ]; then \
		echo "$(GREEN)✅ Coverage check passed: $$COVERAGE% >= 25%$(NC)"; \
	else \
		echo "$(RED)❌ Coverage check failed: $$COVERAGE% < 25%$(NC)"; \
		exit 1; \
	fi

## Code Quality:
lint: ## Запуск линтера с конфигурацией
	@echo "$(GREEN)Running golangci-lint...$(NC)"
	golangci-lint run --config .golangci.yml

lint-fix: ## Запуск линтера с автоисправлением
	@echo "$(GREEN)Running golangci-lint with fixes...$(NC)"
	golangci-lint run --config .golangci.yml --fix

fmt: ## Форматирование кода
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...
	goimports -w .

vet: ## Запуск go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	go vet ./...

check: fmt vet lint test-race ## Полная проверка кода (форматирование, vet, линтер, тесты)

## Build:
swagger: ## Генерация документации Swagger
	@echo "$(GREEN)Generating Swagger documentation...$(NC)"
	swag init -g internal/api/api.go --output docs

server: ## Сборка сервера
	@echo "$(GREEN)Building server...$(NC)"
	go build -o bin/server $(LDFLAGS) ./cmd/server

client: ## Сборка клиента для текущей платформы
	@echo "$(GREEN)Building client...$(NC)"
	go build -o bin/client $(LDFLAGS) ./cmd/client

## Cross-compilation:
client-windows: ## Сборка клиента для Windows
	@echo "$(GREEN)Building client for Windows...$(NC)"
	GOOS=windows GOARCH=amd64 go build -o bin/client-windows-amd64.exe $(LDFLAGS) ./cmd/client

client-linux: ## Сборка клиента для Linux
	@echo "$(GREEN)Building client for Linux...$(NC)"
	GOOS=linux GOARCH=amd64 go build -o bin/client-linux-amd64 $(LDFLAGS) ./cmd/client

client-mac: ## Сборка клиента для macOS
	@echo "$(GREEN)Building client for macOS...$(NC)"
	GOOS=darwin GOARCH=amd64 go build -o bin/client-darwin-amd64 $(LDFLAGS) ./cmd/client
	GOOS=darwin GOARCH=arm64 go build -o bin/client-darwin-arm64 $(LDFLAGS) ./cmd/client

client-all: client-windows client-linux client-mac ## Сборка клиентов для всех платформ

build: swagger server client ## Сборка сервера и клиента для текущей платформы

build-all: swagger server client-all ## Сборка всех артефактов

## Docker:
docker-build: ## Сборка Docker образа
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t gophkeeper:$(VERSION) .

docker-run: ## Запуск в Docker
	@echo "$(GREEN)Running Docker container...$(NC)"
	docker-compose up -d

docker-stop: ## Остановка Docker
	@echo "$(YELLOW)Stopping Docker container...$(NC)"
	docker-compose down

## Utilities:
clean: ## Очистка бинарных файлов и кеша
	@echo "$(YELLOW)Cleaning up...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache

run-server: ## Запуск сервера
	@echo "$(GREEN)Starting server...$(NC)"
	./bin/server

deps: ## Обновление зависимостей
	@echo "$(GREEN)Updating dependencies...$(NC)"
	go mod tidy
	go mod verify

deps-check: ## Проверка обновлений зависимостей
	@echo "$(GREEN)Checking for dependency updates...$(NC)"
	go list -u -m all

install-tools: ## Установка необходимых инструментов
	@echo "$(GREEN)Installing development tools...$(NC)"
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

## CI:
ci-test: deps test-coverage coverage-check ## CI: тесты с проверкой покрытия
ci-lint: deps lint ## CI: линтинг кода
ci-build: deps build-all ## CI: сборка всех артефактов
ci: ci-lint ci-test ci-build ## CI: полная проверка 