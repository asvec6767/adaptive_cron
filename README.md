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
- CLI интерфейс (встроен в единый бинарник)
- Горизонтальное масштабирование (TODO)
- Машинное обучение для предсказания времени выполнения (TODO)

## Архитектура

Компоненты системы:

- API Gateway (8080) — единая точка входа, маршрутизация запросов к сервисам. Может отсутствовать в Lite профиле.
- Task Manager (8081) — CRUD операции с задачами, хранилище.
- Scheduler (8082) — планирование, адаптивный алгоритм, управление очередью.
- Metric Agent (8083) — сбор системных метрик (CPU, память, load average).
- Worker (8084) — выполнение команд.
- CLI — встроен в единый бинарник `adaptive-cron`, управление через флаги командной строки.

Коммуникация между сервисами через транспортный слой:
- In-Memory (для Lite профиля)
- REST (для Standard профиля)
- gRPC (TODO)
- NATS (TODO)

NATS используется только как транспорт между сервисами. Очередь задач всегда либо in-memory, либо Redis, в зависимости от профиля.

## Clean Code слои

Каждый сервис в internal/ имеет четыре слоя:

- Domain — сущности, ошибки, константы. Не имеет внешних зависимостей.
- Repository — интерфейсы доступа к данным.
- UseCase — бизнес-логика, координирует слои.
- Delivery — обработка запросов (HTTP).

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
- Транспорт: REST
- База данных: PostgreSQL (через GORM)
- Очередь: Redis
- Логи: Файл (JSON формат)
- Метрики: Prometheus

Enterprise:
- Транспорт: REST (gRPC и NATS в TODO)
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
- scheduler.local / scheduler.distributed (TODO)

## Адаптивный алгоритм

Load Score = (CPU% * 0.4) + (Memory% * 0.3) + (LoadAvg1 / cores * 0.3)

Алгоритм принятия решения:
1. Если задача критическая (is_critical=true) — запуск немедленно.
2. Проверить CPU > max_cpu_percent — отложить на 30 секунд.
3. Проверить Memory > max_memory_mb — отложить на 30 секунд.
4. Проверить Load Score > max_load_score — отложить на score * 60 секунд.
5. Если все проверки пройдены — запустить задачу через Worker.

Отложенные задачи помещаются в очередь с экспоненциальной задержкой. Базовый интервал 30 секунд, максимальный 1 час.

Алгоритм расширяем для добавления новых проверок (диск I/O, сеть, и т.д.). Поддержка машинного обучения для предсказания времени выполнения запланирована в TODO.

## Технологический стек

Базовый (все профили):
- Go 1.21+
- gorilla/mux (HTTP)
- viper (конфигурация через YAML)
- logrus (логирование)
- GORM (ORM, с возможностью замены на чистые SQL запросы)
- robfig/cron (парсер cron)
- testify (тестирование)
- google/uuid (генерация ID)

Опциональные компоненты по профилям:
- SQLite / PostgreSQL
- In-Memory / Redis (очередь)
- NATS (транспорт для Enterprise, TODO)
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
│   ├── apigateway/             # API Gateway
│   │   ├── domain/
│   │   ├── usecase/
│   │   └── delivery/
│   ├── transport/              # Транспортный слой
│   │   ├── interface.go
│   │   ├── inmemory.go
│   │   ├── rest.go
│   │   ├── grpc.go
│   │   └── nats.go
│   ├── config/                 # Конфигурация
│   │   ├── config.go
│   │   ├── loader.go
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
│   │   │   └── errors.go
│   │   └── utils/
│   │       └── timeutil.go
│   └── app/                    # Сборка приложения
│       ├── app.go
│       └── builder.go
├── pkg/
│   ├── logger/                 # Логирование
│   │   ├── logger.go
│   │   └── clickhouse.go
│   └── cronparser/             # Парсер cron
│       └── parser.go
├── configs/
│   ├── lite.yaml
│   ├── standard.yaml
│   └── enterprise.yaml
├── api/
│   ├── scheduler/openapi.yaml
│   ├── worker/openapi.yaml
│   ├── taskmanager/openapi.yaml
│   ├── metricagent/openapi.yaml
│   └── apigateway/openapi.yaml
├── docker-compose.yml
├── docker-compose.lite.yml
├── docker-compose.standard.yml
├── docker-compose.enterprise.yml
├── Makefile
└── go.mod
```

## Конфигурация

Переменные окружения с префиксом ADAPTIVE_CRON_:
- PROFILE — lite/standard/enterprise
- LOG_LEVEL — debug/info/warn/error
- LOG_FORMAT — text/json
- DATABASE_DSN — строка подключения (для SQLite)
- DATABASE_HOST, DATABASE_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME — для PostgreSQL
- QUEUE_TYPE — memory/redis
- QUEUE_HOST, QUEUE_PORT, QUEUE_DB — для Redis
- WORKER_MAX_CONCURRENT — число параллельных задач
- SERVER_HTTP_PORT — порт HTTP API
- SERVER_GRPC_PORT — порт gRPC (TODO)

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
  format: text
  file: ./logs/adaptive-cron.log
server:
  http_port: 8080
worker:
  max_concurrent: 3
scheduler:
  check_interval: 10s
  type: local
metrics:
  type: basic
features:
  - db.sqlite
  - queue.memory
  - log.file
  - transport.inmemory
  - metrics.basic
  - scheduler.local
```

## Запуск

Локально:
```
make build
./bin/adaptive-cron --profile lite
./bin/adaptive-cron --profile standard --config configs/standard.yaml
./bin/adaptive-cron --profile enterprise --config configs/enterprise.yaml
./bin/adaptive-cron --list-profiles
./bin/adaptive-cron --profile lite --enable metrics.prometheus
```

Docker Compose:
```
make docker-lite
make docker-standard
make docker-enterprise
```

## Команды Makefile

```
make build              - Сборка бинарника
make run-lite           - Запуск с профилем lite
make run-standard       - Запуск с профилем standard
make run-enterprise     - Запуск с профилем enterprise
make docker-lite        - Запуск lite в Docker
make docker-standard    - Запуск standard в Docker
make docker-enterprise  - Запуск enterprise в Docker
make docker-down        - Остановка всех контейнеров
make test               - Запуск всех тестов
make test-unit          - Unit тесты
make test-integration   - Интеграционные тесты
make lint               - Линтер
make clean              - Очистка
make install-cli        - Установка CLI
make deps               - Установка зависимостей
make help               - Справка
```

## API Endpoints

### API Gateway (8080)
```
GET  /health
GET  /api/v1/gateway/health
GET  /api/v1/gateway/services
GET  /api/v1/gateway/routes
/*   (прокси к сервисам)
```

### Task Manager (8081)
```
POST   /api/v1/tasks
GET    /api/v1/tasks
GET    /api/v1/tasks/{id}
PUT    /api/v1/tasks/{id}
DELETE /api/v1/tasks/{id}
POST   /api/v1/tasks/{id}/toggle
GET    /health
```

### Scheduler (8082)
```
GET  /api/v1/scheduler/next
POST /api/v1/scheduler/trigger/{id}?force=false
GET  /api/v1/scheduler/queue
DELETE /api/v1/scheduler/queue/{id}
GET  /health
```

### Metric Agent (8083)
```
GET /api/v1/metrics/current
GET /api/v1/metrics/history?limit=100
GET /api/v1/metrics/load
GET /api/v1/metrics/health
GET /health
```

### Worker (8084)
```
POST /api/v1/execute
GET  /api/v1/status/{id}
GET  /api/v1/worker/stats
GET  /api/v1/worker/jobs
GET  /health
```

## Порты сервисов

| Сервис | Порт |
|--------|------|
| API Gateway | 8080 |
| Task Manager | 8081 |
| Scheduler | 8082 |
| Metric Agent | 8083 |
| Worker | 8084 |
| gRPC | 9090 (TODO) |
| Prometheus | 9091 |
| PostgreSQL | 5432 |
| Redis | 6379 |
| ClickHouse | 8123/9000 |
| Grafana | 3000 |

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
- In-Memory и REST транспорт
- SQLite, PostgreSQL, Redis
- Базовое логирование

Версия 1.1 (в разработке):
- gRPC для внутренней коммуникации
- NATS транспорт
- ClickHouse логирование
- Prometheus + Grafana
- Web UI

Версия 2.0:
- Горизонтальное масштабирование (Leader Election для Scheduler)
- Машинное обучение для предсказания времени выполнения и нагрузки
- Kubernetes Helm
- Service Discovery (etcd/Consul)

## Лицензия

MIT
