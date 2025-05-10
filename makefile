APP_NAME=auth-service
BINARY=$(APP_NAME)
DB_URL_DOCKER=postgres://postgres:05042007PULlup!@postgres:5432/postgres?sslmode=disable

all: generate build

# Генерация gRPC кода
generate:
	protoc \
		--proto_path=./api/proto/calculator \
		--go_out=./pkg/api/v1 \
		--go-grpc_out=./pkg/api/v1 \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		./api/proto/calculator/calculator.proto

# Миграции в Docker-окружении
migrations-docker-up:
	docker compose up -d postgres

	docker compose run --rm orchestrator migrate -path /app/migrations -database '$(DB_URL_DOCKER)' up

migrations-docker-down:
	docker compose run --rm orchestrator migrate -path /app/migrations -database '$(DB_URL_DOCKER)' down

# Сборка приложения
build:
	go build -o $(APP_NAME) cmd/main.go
	docker-compose build

# Запуск приложения
run: build migrations-up
	docker-compose up auth-service

# Тестирование
test-integration:
	docker-compose run --rm auth-service go test -v -tags=integration ./tests/integration/auth...

test:
	docker-compose run --rm auth-service go test -v -short ./...

# Очистка
clean:
	docker-compose down -v
	rm -f $(BINARY)

.PHONY: all generate migrations-up migrations-down build run test-integration test clean