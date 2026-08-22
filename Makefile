.PHONY: build lite standard enterprise clean test

# Сборка единого бинарника
build:
	go build -o bin/adaptive-cron ./cmd/adaptive-cron

# Запуск разных профилей
lite: build
	./bin/adaptive-cron --profile lite

standard: build
	./bin/adaptive-cron --profile standard --config configs/standard.yaml

enterprise: build
	./bin/adaptive-cron --profile enterprise --config configs/enterprise.yaml

# Docker запуск
docker-lite:
	docker-compose --profile lite up

docker-standard:
	docker-compose --profile standard up

docker-enterprise:
	docker-compose --profile enterprise up

# Остановка всех
docker-down:
	docker-compose down

# Тесты
test:
	go test -v -race ./...

# Очистка
clean:
	rm -rf bin/
	docker-compose down -v

# Установка зависимостей
deps:
	go mod download
	go mod tidy