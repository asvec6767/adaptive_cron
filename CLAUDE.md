# Adaptive Cron - Critical Documentation

## Scheduler Service (internal/scheduler)

**Port:** 8082

**Clean Code Layers:**
- domain/ - Task, DelayedTask, ScheduleResult entities, domain errors
- repository/ - TaskRepository, DelayedQueueRepository interfaces
- usecase/ - SchedulerUseCase, AdaptiveChecker
- delivery/rest/ - HTTP handlers

**Key Dependencies (interfaces):**
```
TaskRepository - GetActiveTasks, GetTask, UpdateTask
DelayedQueueRepository - Push, PopDue, Remove, List, Size
MetricProvider - GetCurrentMetrics, GetLoadScore
WorkerClient - Execute(taskID, command, timeout) returns jobID
```

**Core Logic:**
1. ScheduleTasks() - checks active tasks, runs if schedule matches
2. AdaptiveChecker.CanRun() - checks CPU, Memory, Load Score
3. ProcessDelayedTasks() - processes delayed queue with exponential backoff
4. TriggerTask() - manual trigger with force flag

**Delayed Queue Strategy:** Exponential backoff: baseDelay(30s) * 2^attempt, max 1h

**API Endpoints:**
```
GET  /api/v1/scheduler/next
POST /api/v1/scheduler/trigger/{id}?force=false
GET  /api/v1/scheduler/queue
DELETE /api/v1/scheduler/queue/{id}
GET  /health
```

## Worker Service (internal/worker)

**Port:** 8084

**Clean Code Layers:**
- domain/ - Job, JobStatus, ExecuteRequest, ExecuteResponse, JobStatusResponse, WorkerStats
- repository/ - JobRepository interface
- usecase/ - WorkerUseCase
- delivery/rest/ - HTTP handlers

**Key Structures:**
```
Job:
  ID, TaskID, Command, Timeout, StartTime, EndTime, ExitCode, Output, Error, Status

JobStatus: pending | running | completed | failed | timeout
```

**Core Logic:**
1. Execute() - creates job, runs async with semaphore (maxConcurrent)
2. runJob() - executes command via exec.CommandContext with timeout
3. GetStatus() - returns job status
4. GetStats() - returns aggregated worker statistics

**Resource Control:** Semaphore pattern limits concurrent executions

**API Endpoints:**
```
POST /api/v1/execute
GET  /api/v1/status/{id}
GET  /api/v1/worker/stats
GET  /api/v1/worker/jobs
GET  /health
```

## Task Manager Service (internal/taskmanager)

**Port:** 8081

**Clean Code Layers:**
- domain/ - Task, CreateTaskRequest, UpdateTaskRequest, TaskFilter, domain errors
- repository/ - TaskRepository interface
- usecase/ - TaskManagerUseCase
- delivery/rest/ - HTTP handlers

**Key Structures:**
```
Task:
  ID, Name, Command, Schedule, IsCritical, MaxLoadScore, MaxCPUPercent,
  MaxMemoryMB, TimeoutSeconds, IsActive, CreatedAt, UpdatedAt
```

**Default Values:**
```
MaxLoadScore = 0.7
MaxCPUPercent = 80
MaxMemoryMB = 1024
TimeoutSeconds = 300
```

**Core Logic:**
1. CreateTask() - validates, checks name uniqueness, sets defaults
2. UpdateTask() - partial updates using pointers
3. ToggleTask() - flips IsActive status
4. ListTasks() - filters by Active, Critical, Name

**API Endpoints:**
```
POST   /api/v1/tasks
GET    /api/v1/tasks?active=true&critical=true&name=foo
GET    /api/v1/tasks/{id}
PUT    /api/v1/tasks/{id}
DELETE /api/v1/tasks/{id}
POST   /api/v1/tasks/{id}/toggle
GET    /health
```

## Common Models (internal/common/models/)

**Task** - Core task entity
**SystemMetrics** - CPU, Memory, Load Average with GetLoadScore(cores) method
**JobLog** - Execution log entries
**DelayedTask** - Queue items with attempt counter

## Common Errors (internal/common/errors/)

Standard errors used across all services

## Cron Parser (pkg/cronparser/)

Parse cron expressions for scheduling

## Dependencies Required

```
github.com/gorilla/mux
github.com/google/uuid
github.com/spf13/viper
github.com/sirupsen/logrus
gorm.io/gorm
gorm.io/driver/postgres
gorm.io/driver/sqlite
```

## File Structure Summary

```
internal/
├── scheduler/
│   ├── domain/
│   │   ├── task.go
│   │   └── errors.go
│   ├── repository/
│   │   └── interfaces.go
│   ├── usecase/
│   │   ├── scheduler.go
│   │   └── adaptive_checker.go
│   └── delivery/
│       └── rest/
│           └── handler.go
├── worker/
│   ├── domain/
│   │   ├── job.go
│   │   └── errors.go
│   ├── repository/
│   │   └── interfaces.go
│   ├── usecase/
│   │   └── worker.go
│   └── delivery/
│       └── rest/
│           └── handler.go
├── taskmanager/
│   ├── domain/
│   │   ├── task.go
│   │   └── errors.go
│   ├── repository/
│   │   └── interfaces.go
│   ├── usecase/
│   │   └── taskmanager.go
│   └── delivery/
│       └── rest/
│           └── handler.go
└── common/
    ├── models/
    │   ├── task.go
    │   ├── metric.go
    │   └── log.go
    └── errors/
        └── errors.go
```

## OpenAPI Files Location

```
api/
├── scheduler/openapi.yaml
├── worker/openapi.yaml
└── taskmanager/openapi.yaml
```

## TODO Items Already Implemented

- Clean Architecture layers (Domain, Repository, UseCase, Delivery)
- Interfaces for dependency injection
- No direct dependencies on external packages except standard lib + gorm + uuid + mux
- All services independent and can run separately
- Common models shared across services

## TODO Items Marked for Future

- GORM repository implementations
- DelayedQueueRepository implementations (Redis/In-Memory)
- WorkerClient implementation (transport layer)
- MetricProvider implementation
- ClickHouse logging
- NATS transport
- gRPC handlers
- Web UI

## Metric Agent Service (internal/metricagent)

**Port:** 8083

**Clean Code Layers:**
- domain/ - SystemMetrics, MetricsHistory, domain errors
- repository/ - MetricRepository interface
- usecase/ - MetricCollector, MetricAgentUseCase
- delivery/rest/ - HTTP handlers

**Key Dependencies (interfaces):**
```
MetricRepository - Save, GetCurrent, GetHistory, GetLoadScore
```

**Core Logic:**
1. MetricCollector.Start() - starts background collection with configurable interval
2. collectOnce() - collects CPU, Load Average, Memory, Swap via gopsutil
3. GetCurrent() - returns latest metrics snapshot
4. GetHistory() - returns recent metrics with limit
5. GetLoadScore() - calculates weighted load score: (CPU*0.4 + Memory*0.3 + Load*0.3)

**Configuration:**
```
CollectInterval: 5s (configurable)
MaxHistorySize: 1000 entries
NumCores: runtime.NumCPU()
```

**API Endpoints:**
```
GET /api/v1/metrics/current
GET /api/v1/metrics/history?limit=100
GET /api/v1/metrics/load
GET /api/v1/metrics/health
GET /health
```

## API Gateway Service (internal/apigateway)

**Port:** 8080

**Clean Code Layers:**
- domain/ - ServiceConfig, GatewayConfig, domain errors
- usecase/ - GatewayUseCase (service registration, proxy routing, health checks)
- delivery/rest/ - HTTP handlers

**Core Logic:**
1. RegisterService() - registers service with URL and routes
2. Proxy() - routes requests to appropriate service based on path prefix
3. HealthCheck() - checks service health before proxying
4. ListServices() - returns all registered services

**Path Routing:**
```
/api/v1/tasks/*       -> taskmanager:8081
/api/v1/scheduler/*   -> scheduler:8082
/api/v1/metrics/*     -> metricagent:8083
/api/v1/execute/*     -> worker:8084
/api/v1/status/*      -> worker:8084
/api/v1/worker/*      -> worker:8084
```

**API Endpoints:**
```
GET /health
GET /api/v1/gateway/health
GET /api/v1/gateway/services
GET /api/v1/gateway/routes
/* (proxy to services)
```

## Transport Layer (internal/transport)

**Purpose:** Abstraction layer for service-to-service communication

**Supported Bindings:**
```
InMemoryBinding - for Lite profile (in-process communication)
RESTBinding - for Standard profile (HTTP REST)
GRPCBinding - placeholder for future gRPC implementation
NATSBinding - placeholder for future NATS implementation
```

**Client Interfaces:**
```
SchedulerClient - ScheduleTask, GetNextTasks, GetDelayedQueue, TriggerTask
WorkerClient - Execute, GetStatus, GetStats
TaskManagerClient - CreateTask, GetTask, ListTasks, UpdateTask, DeleteTask, ToggleTask
MetricAgentClient - GetCurrentMetrics, GetMetricsHistory, GetLoadScore
```

**Service Interfaces (for registration):**
```
SchedulerService, WorkerService, TaskManagerService, MetricAgentService
```

## Service Ports Summary

| Service | Port |
|---------|------|
| API Gateway | 8080 |
| Task Manager | 8081 |
| Scheduler | 8082 |
| Metric Agent | 8083 |
| Worker | 8084 |

## Transport Types by Profile

| Profile | Transport | Communication |
|---------|-----------|---------------|
| Lite | InMemory | Direct function calls (all in one process) |
| Standard | REST | HTTP calls between processes |
| Enterprise | gRPC/NATS | gRPC + NATS (distributed) |

## File Structure Summary

```
internal/
├── scheduler/           # Port 8082
├── worker/              # Port 8084
├── taskmanager/         # Port 8081
├── metricagent/         # Port 8083
│   ├── domain/
│   │   ├── metric.go
│   │   └── errors.go
│   ├── repository/
│   │   └── interfaces.go
│   ├── usecase/
│   │   ├── collector.go
│   │   └── metricagent.go
│   └── delivery/
│       └── rest/
│           └── handler.go
├── apigateway/          # Port 8080
│   ├── domain/
│   │   ├── config.go
│   │   └── errors.go
│   ├── usecase/
│   │   └── gateway.go
│   └── delivery/
│       └── rest/
│           └── handler.go
├── transport/
│   ├── interface.go
│   ├── inmemory.go
│   ├── rest.go
│   ├── grpc.go
│   └── nats.go
└── common/
    ├── models/
    ├── errors/
    └── utils/
```

## OpenAPI Files Location

```
api/
├── scheduler/openapi.yaml
├── worker/openapi.yaml
├── taskmanager/openapi.yaml
├── metricagent/openapi.yaml
└── apigateway/openapi.yaml
```

## TODO Items Already Implemented

- Metric Agent with gopsutil integration
- API Gateway with reverse proxy routing
- Transport layer with InMemory and REST bindings
- gRPC and NATS placeholders for future implementation
- All services have independent Clean Architecture layers
- OpenAPI specifications for all services

## TODO Items Marked for Future

- gRPC binding implementation (transport/grpc.go)
- NATS binding implementation (transport/nats.go)
- ClickHouse logging integration
- Prometheus metrics integration
- Web UI
- Distributed scheduler
- ML for task prediction
```

## Config Package (internal/config)

**Purpose:** Configuration management with profile support

**Key Structures:**
```
Config - Main configuration struct
DatabaseConfig - DB connection settings
QueueConfig - Queue settings (memory/redis)
LoggingConfig - Log settings with rotation
ServerConfig - HTTP/gRPC ports
WorkerConfig - Concurrency settings
SchedulerConfig - Check interval, type
MetricsConfig - Metrics type and port
```

**Functions:**
```
LoadConfig(profile, configFile) - Loads config with profile defaults
GetDSN() - Returns database connection string
GetCheckInterval() - Parses check interval duration
```

**Supported Profiles:**
```
lite - SQLite + In-Memory queue
standard - PostgreSQL + Redis
enterprise - PostgreSQL + Redis + ClickHouse + NATS
```

## Features Package (internal/features)

**Purpose:** Feature flag management

**Key Functions:**
```
Register(flag) - Register a feature flag
IsEnabled(name) - Check if feature is enabled
Enable(name) - Enable a feature
Disable(name) - Disable a feature
GetAll() - Get all registered features
InitFromConfig(features) - Initialize from config list
```

**Feature Categories:**
```
db.* - Database features (sqlite, postgres, replication)
queue.* - Queue features (memory, redis, persistent)
log.* - Logging features (file, json, clickhouse, structured)
transport.* - Transport features (inmemory, rest, grpc, nats)
metrics.* - Metrics features (basic, prometheus)
scheduler.* - Scheduler features (local, distributed)
service.* - Service features (discovery)
```

## Common Package (internal/common)

**Purpose:** Shared models, errors, utilities

**Models:**
```
models/task.go - Task, CreateTaskRequest, UpdateTaskRequest, TaskFilter, DelayedTask
models/metric.go - SystemMetrics with GetLoadScore() method
models/log.go - JobLog, LogEntry, ClickHouseLog
```

**Errors:**
```
errors/errors.go - Common errors: ErrNotFound, ErrInvalidInput, ErrTaskNotFound,
    ErrJobNotFound, ErrMetricsUnavailable, ErrQueueFull, etc.
```

**Utils:**
```
utils/timeutil.go - ShouldRunAt(), GetNextRunTime(), CalculateBackoff()
```

## App Package (internal/app)

**Purpose:** Application assembly and orchestration

**Key Structures:**
```
App - Main application struct with all services
SchedulerService, WorkerService, TaskManagerService, MetricAgentService - Service interfaces
TransportBinding - Transport layer interface
APIServer - HTTP/gRPC server interface
```

**Key Functions:**
```
NewApp(cfg, opts...) - Creates application with options
Run(ctx) - Starts all services and API server
Shutdown(ctx) - Graceful shutdown of all components
```

**Transport Types:**
```
InMemoryTransport - For Lite profile (in-process)
RESTTransport - For Standard profile (HTTP REST)
GRPCTransport - Placeholder for gRPC implementation
NATSTransport - Placeholder for NATS implementation
```

**Lifecycle:**
```
1. Build components based on features
2. Connect transport
3. Start services (scheduler, worker, metric agent)
4. Start API server
5. Wait for context cancellation
6. Graceful shutdown
```

## Service Registration (for InMemory binding)

```
transport.RegisterScheduler(schedulerService)
transport.RegisterWorker(workerService)
transport.RegisterTaskManager(taskManagerService)
transport.RegisterMetricAgent(metricAgentService)
```

## API Server (internal/app)

**Purpose:** Unified API server that uses transport clients

**Functionality:**
- Serves HTTP/gRPC endpoints
- Routes requests to appropriate services via transport layer
- Health checks
- Service discovery endpoints

## Dependencies

```
github.com/spf13/viper
github.com/robfig/cron/v3
```

## File Structure Summary

```
internal/
├── config/
│   ├── config.go
│   ├── loader.go
│   └── profiles.go
├── features/
│   ├── flags.go
│   └── registry.go
├── common/
│   ├── models/
│   │   ├── task.go
│   │   ├── metric.go
│   │   └── log.go
│   ├── errors/
│   │   └── errors.go
│   └── utils/
│       └── timeutil.go
└── app/
    ├── app.go
    └── builder.go
```

## TODO Items Already Implemented

- Config loading with profile support
- Feature flags registry with initialization from config
- Common models shared across all services
- Common errors for consistent error handling
- Time utilities for cron and backoff
- App assembly with transport selection
- Service interfaces for decoupling
- Transport placeholders for all types

## TODO Items Marked for Future

- Complete RESTTransport implementation with actual HTTP clients
- Complete GRPCTransport implementation
- Complete NATSTransport implementation
- Complete APIServer with actual endpoints
- Complete service implementations (scheduler, worker, taskmanager, metricagent)
- Database repository implementations
- Delayed queue implementations (Redis, In-Memory)
```


## Main Entry Point (cmd/adaptive-cron)

**Purpose:** Single binary entry point for all profiles

**Key Functions:**
```
main() - Parses flags, loads config, initializes logger, runs app
showAvailableProfiles() - Displays available profiles with features
```

**Command Line Flags:**
```
--profile       - Profile to use (lite, standard, enterprise)
--config        - Config file path
--list-profiles - Show available profiles
--enable        - Enable additional features (comma-separated)
--disable       - Disable features (comma-separated)
--version       - Show version
```

**Usage Examples:**
```
adaptive-cron --profile lite
adaptive-cron --profile standard --config configs/standard.yaml
adaptive-cron --profile lite --enable metrics.prometheus
adaptive-cron --list-profiles
```

## Logger Package (pkg/logger)

**Purpose:** Structured logging with file rotation support

**Key Functions:**
```
Init(config) - Initialize logger with level, format, file, rotation
Debug/Info/Warn/Error/Fatal - Logging functions
WithField/WithFields - Add structured fields
```

**Configuration:**
```
Level - debug, info, warn, error
Format - text, json
File - log file path
Rotation - MaxSize, MaxAge, MaxBackups
```

**ClickHouse Support:**
```
NewClickHouseLogger(config) - Create ClickHouse logger
Write(ctx, entry) - Write log entry to ClickHouse
Flush(ctx) - Flush batch
Close() - Close connection
```

## Cron Parser Package (pkg/cronparser)

**Purpose:** Parse and evaluate cron expressions

**Key Functions:**
```
Parse(schedule) - Parse cron expression
Next(now, schedule) - Get next run time
ShouldRunAt(now, schedule) - Check if should run at time
```

**Supported Format:**
```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6)
│ │ │ │ │
* * * * *
```

## Config Files (configs/)

| File | Profile | Description |
|------|---------|-------------|
| lite.yaml | Lite | SQLite + In-Memory, minimal dependencies |
| standard.yaml | Standard | PostgreSQL + Redis, full features |
| enterprise.yaml | Enterprise | Distributed with ClickHouse + NATS |

**Configuration Structure:**
```
profile - Profile name
database - DB settings (type, dsn, host, port, user, password, dbname)
queue - Queue settings (type, max_size, host, port, db, servers)
logging - Log settings (type, level, format, file, rotation, host, port, database, table, batch_size, flush_interval)
server - Server settings (http_port, grpc_port)
worker - Worker settings (max_concurrent, queue_size)
scheduler - Scheduler settings (check_interval, type)
metrics - Metrics settings (type, port)
features - Enabled feature flags
```

## Docker Compose Files

| File | Profile | Services |
|------|---------|----------|
| docker-compose.yml | Base | postgres, redis (optional) |
| docker-compose.lite.yml | Lite | app (single container) |
| docker-compose.standard.yml | Standard | postgres, redis, app |
| docker-compose.enterprise.yml | Enterprise | postgres, redis, clickhouse, grafana, app |

## Makefile Commands

```
make build         - Build the application
make run-lite      - Run with lite profile
make run-standard  - Run with standard profile
make run-enterprise- Run with enterprise profile
make docker-lite   - Run lite profile with Docker
make docker-standard - Run standard profile with Docker
make docker-enterprise - Run enterprise profile with Docker
make test          - Run all tests
make lint          - Run linter
make clean         - Clean build artifacts
make install-cli   - Install CLI to /usr/local/bin
make deps          - Download and tidy dependencies
make help          - Show available commands
```

## Service Ports Summary

| Service | Port | Profile |
|---------|------|---------|
| API Gateway | 8080 | All |
| Task Manager | 8081 | All |
| Scheduler | 8082 | All |
| Metric Agent | 8083 | All |
| Worker | 8084 | All |
| gRPC | 9090 | Standard, Enterprise |
| Prometheus | 9091 | Standard, Enterprise |
| PostgreSQL | 5432 | Standard, Enterprise |
| Redis | 6379 | Standard, Enterprise |
| ClickHouse | 8123/9000 | Enterprise |
| Grafana | 3000 | Enterprise |

## OpenAPI Files Location

```
api/
├── scheduler/openapi.yaml
├── worker/openapi.yaml
├── taskmanager/openapi.yaml
├── metricagent/openapi.yaml
└── apigateway/openapi.yaml
```

## TODO Items Already Implemented

- Single binary entry point with profile support
- Structured logging with file rotation
- Cron expression parser
- YAML configuration files for all profiles
- Docker Compose for all profiles
- Makefile with common tasks
- Feature flags integration
- Clean architecture layers

## TODO Items Marked for Future

- ClickHouse logger implementation
- gRPC transport implementation
- NATS transport implementation
- Web UI
- Prometheus metrics integration
- Kubernetes Helm charts
- Distributed scheduler with leader election
- ML for task prediction