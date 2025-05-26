VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_NUM ?= $(shell git rev-list --count HEAD 2>/dev/null || echo "0")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH) -X main.buildDate=$(BUILD_DATE) -X main.buildNum=$(BUILD_NUM)"

.PHONY: all test build clean swagger client-build server-build 

all: build

# Запуск тестов
test:
	go test -v ./...

# Генерация документации Swagger
swagger:
	swag init -g internal/api/api.go --output docs

# Сборка сервера
server:
	go build -o bin/server $(LDFLAGS) ./cmd/server

# Сборка клиента для текущей платформы
client:
	go build -o bin/client $(LDFLAGS) ./cmd/client

# Кросс-компиляция для разных платформ
client-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/client-windows-amd64.exe $(LDFLAGS) ./cmd/client

client-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/client-linux-amd64 $(LDFLAGS) ./cmd/client

client-mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/client-darwin-amd64 $(LDFLAGS) ./cmd/client
	GOOS=darwin GOARCH=arm64 go build -o bin/client-darwin-arm64 $(LDFLAGS) ./cmd/client

# Сборка всех клиентов
client-all: client-windows client-linux client-mac

# Сборка сервера и клиента для текущей платформы
build: swagger server client

# Очистка бинарных файлов
clean:
	rm -rf bin/

# Запуск сервера
run-server:
	./bin/server

# Запуск линтера
lint:
	golangci-lint run ./...

# Обновление зависимостей
deps:
	go mod tidy
	go mod verify 