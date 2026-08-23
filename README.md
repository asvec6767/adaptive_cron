# Adaptive Cron

Адаптивный планировщик задач с микросервисной и Clean Code архитектурой.

## Описание

Adaptive Cron — система планирования задач, аналогичная cron, с адаптивным управлением на основе системных метрик (CPU, память, load average). Задачи откладываются при высокой нагрузке для предотвращения перегрузки сервера.

Ключевые особенности:
- Адаптивное планирование на основе метрик
- Критические задачи (выполняются всегда)
- Микросервисная архитектура с Clean Code слоями
- Feature Flags для гибкой настройки
- Три профиля: Lite, Standard, Enterprise
- CLI интерфейс (отдельный сервис)
- Горизонтальное масштабирование (TODO)
- Машинное обучение для предсказания времени выполнения (TODO)

## Архитектура

Компоненты системы:

- API Gateway (8080) — единая точка входа, маршрутизация запросов к сервисам. Может отсутствовать в Lite профиле.
- Task Manager (8081) — CRUD операции с задачами, хранилище.
- Scheduler (8082) — планирование, адаптивный алгоритм, управление очередью.
- Metric Agent (8083) — сбор системных метрик (CPU, память, load average).
- Worker (8084) — выполнение команд.
- CLI (taskctl) — управление из командной строки, отдельный сервис.

Коммуникация между сервисами через транспортный слой:
- In-Memory (для Lite профиля)
- REST/gRPC (для Standard профиля)
- gRPC/NATS (для Enterprise профиля)

NATS используется только как транспорт между сервисами. Очередь задач всегда либо in-memory, либо Redis, в зависимости от профиля.

## Clean Code слои

Каждый сервис в internal/ имеет четыре слоя:

- Domain — сущности, ошибки, константы. Не имеет внешних зависимостей.
- Repository — интерфейсы доступа к данным.
- UseCase — бизнес-логика, координирует слои.
- Delivery — обработка запросов (HTTP/gRPC).

Зависимости направлены внутрь: Domain <- Repository <- UseCase <- Delivery.

Общие модели (Task, Metric, Log) находятся в internal/common/models.

## Профили и Feature Flags

Три профиля конфигурации:

Lite:
- Транспорт: In-Memory
- База данных: SQLite (через GORM)
- Очередь: In-Memory
- Логи: Файл (через pkg/logger)
- Метрики: Встроенные

Standard:
- Транспорт: REST + gRPC
- База данных: PostgreSQL (через GORM)
- Очередь: Redis
- Логи: Файл (JSON формат)
- Метрики: Prometheus

Enterprise:
- Транспорт: gRPC + NATS
- База данных: PostgreSQL (через GORM)
- Очередь: Redis
- Логи: ClickHouse + Grafana
- Метрики: Prometheus

Feature Flags загружаются из конфига при старте, но также имеют значения по умолчанию для каждого профиля. Примеры флагов:
- db.sqlite / db.postgres
- queue.memory / queue.redis
- log.file / log.clickhouse
- transport.inmemory / transport.rest / transport.grpc / transport.nats
- metrics.basic / metrics.prometheus
- scheduler.distributed (TODO)

## Адаптивный алгоритм

Load Score = (CPU% * 0.4) + (Memory% * 0.3) + (LoadAvg1 / cores * 0.3)

Алгоритм принятия решения:
1. Если задача критическая (is_critical=true) — запуск немедленно.
2. Проверить CPU > max_cpu_percent — отложить на 30 секунд.
3. Проверить Memory > max_memory_mb — отложить на 30 секунд.
4. Проверить Load Score > max_load_score — отложить на score * 60 секунд.
5. Если все проверки пройдены — запустить задачу через Worker.

Отложенные задачи помещаются в очередь с экспоненциальной задержкой.

Алгоритм расширяем для добавления новых проверок (диск I/O, сеть, и т.д.). Поддержка машинного обучения для предсказания времени выполнения запланирована в TODO.

## Технологический стек

Базовый (все профили):
- Go 1.21+
- gorilla/mux (HTTP)
- cobra/viper (CLI и конфиг)
- logrus (логирование)
- GORM (ORM, с возможностью замены на чистые SQL запросы)
- testify (тестирование)

Опциональные компоненты по профилям:
- SQLite / PostgreSQL
- In-Memory / Redis (очередь)
- NATS (транспорт для Enterprise)
- ClickHouse + Grafana (логи для Enterprise)
- Prometheus (метрики для Standard и Enterprise)

## Структура проекта

```
adaptive-cron/
├── cmd/
│   └── adaptive-cron/          # Единый бинарник
│       └── main.go
├── internal/
│   ├── scheduler/              # Планировщик
│   │   ├── domain/
│   │   ├── repository/
│   │   ├── usecase/
│   │   └── delivery/
│   ├── worker/                 # Исполнитель
│   │   ├── domain/
│   │   ├── repository/
│   │   ├── usecase/
│   │   └── delivery/
│   ├── taskmanager/            # CRUD задач
│   │   ├── domain/
│   │   ├── repository/
│   │   ├── usecase/
│   │   └── delivery/
│   ├── metricagent/            # Сбор метрик
│   │   ├── domain/
│   │   ├── repository/
│   │   ├── usecase/
│   │   └── delivery/
│   ├── apigateway/             # API Gateway (опционально)
│   │   └── delivery/
│   ├── transport/              # Транспортный слой
│   │   ├── interface.go
│   │   ├── inmemory.go
│   │   ├── rest.go
│   │   ├── grpc.go
│   │   └── nats.go
│   ├── config/                 # Конфигурация
│   │   ├── config.go
│   │   └── profiles.go
│   ├── features/               # Feature flags
│   │   ├── flags.go
│   │   └── registry.go
│   ├── common/                 # Общие компоненты
│   │   ├── models/
│   │   │   ├── task.go
│   │   │   ├── metric.go
│   │   │   └── log.go
│   │   ├── errors/
│   │   └── utils/
│   └── app/                    # Сборка приложения
│       ├── app.go
│       └── builder.go
├── pkg/
│   ├── logger/                 # Логирование (файл, ClickHouse)
│   └── cronparser/             # Парсер cron
├── configs/
│   ├── lite.yaml
│   ├── standard.yaml
│   └── enterprise.yaml
├── docker-compose*.yml
├── Makefile
└── go.mod
```

## Конфигурация

Переменные окружения с префиксом ADAPTIVE_CRON_:
- PROFILE — lite/standard/enterprise
- LOG_LEVEL — debug/info/warn/error
- LOG_FORMAT — text/json
- DATABASE_DSN — строка подключения
- QUEUE_TYPE — memory/redis
- WORKER_MAX_CONCURRENT — число параллельных задач

Файлы конфигурации в формате YAML. Пример lite.yaml:

```yaml
profile: lite
database:
  type: sqlite
  dsn: ./data/adaptive-cron.db
queue:
  type: memory
  max_size: 100
logging:
  type: file
  level: info
  file: ./logs/adaptive-cron.log
server:
  http_port: 8080
worker:
  max_concurrent: 3
```

## Запуск

Локально:
```
make build
./bin/adaptive-cron --profile lite
./bin/adaptive-cron --profile standard --config configs/standard.yaml
```

Docker Compose:
```
make docker-lite
make docker-standard
make docker-enterprise
```

## Тестирование

```
make test-unit
make test-integration
make test-e2e
make coverage
```

## TODO и планы

Версия 1.0 (текущая):
- Архитектура, Clean Code, профили
- CLI интерфейс
- Docker Compose

Версия 1.1 (в разработке):
- gRPC для внутренней коммуникации
- Prometheus + Grafana
- Web UI

Версия 2.0:
- Горизонтальное масштабирование (Leader Election для Scheduler)
- Машинное обучение для предсказания времени выполнения и нагрузки
- ClickHouse аналитика
- Kubernetes Helm

## Лицензия

MIT
