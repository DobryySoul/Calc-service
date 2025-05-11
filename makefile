APP_NAME=calc-service
BINARY=$(APP_NAME)
DB_URL_DOCKER=postgres://postgres:password@postgres:5432/postgres?sslmode=disable

all: test-orchestrator test-agent build run migrations-docker-up

# Миграции в Docker-окружении
migrations-docker-up:
	docker compose up -d postgres

	docker compose run --rm orchestrator migrate -path /app/migrations -database '$(DB_URL_DOCKER)' up

migrations-docker-down:
	docker compose run --rm orchestrator migrate -path /app/migrations -database '$(DB_URL_DOCKER)' down

# Сборка приложения
build:
	docker-compose build

# Запуск приложения
run: build migrations-docker-up
	docker-compose up -d --build

# Запуск тестов
test-agent:
	cd ./agent && go test -v -short ./...

test-orchestrator:
	cd ./orchestrator && go test -v -short ./...

# Очистка
clean:
	docker-compose down -v
	rm -f $(BINARY)

.PHONY: all migrations-up migrations-down build run test-agent test-orchestrator clean