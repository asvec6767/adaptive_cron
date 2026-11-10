# Этап 1: Сборка
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые инструменты для сборки
RUN apk add --no-cache git make

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download && go mod tidy

# Копируем исходный код
COPY . .

# Собираем приложение
RUN make build

# Этап 2: Финальный образ
FROM alpine:latest

# Устанавливаем CA-сертификаты для HTTPS и время
RUN apk add --no-cache ca-certificates tzdata

# Создаем пользователя без прав
RUN adduser -D -u 1000 appuser

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем собранный бинарник из этапа сборки
COPY --from=builder /app/bin/adaptive-cron /app/adaptive-cron

# Копируем конфиги
COPY --from=builder /app/configs /app/configs

# Создаем директории для данных и логов
RUN mkdir -p /app/data /app/logs && chown -R appuser:appuser /app

# Переключаемся на пользователя без прав
USER appuser

# Открываем порты для HTTP сервисов
EXPOSE 8080 8081 8082 8083 8084

# Точка входа
ENTRYPOINT ["/app/adaptive-cron"]

# Команда по умолчанию (переопределяется через docker-compose)
CMD ["--profile", "standard", "--config", "configs/standard.yaml"]