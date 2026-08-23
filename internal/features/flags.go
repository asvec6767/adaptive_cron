package features

const (
	// Database features
	FlagDBSQLite      = "db.sqlite"
	FlagDBPostgres    = "db.postgres"
	FlagDBReplication = "db.replication"

	// Queue features
	FlagQueueMemory     = "queue.memory"
	FlagQueueRedis      = "queue.redis"
	FlagQueuePersistent = "queue.persistent"

	// Logging features
	FlagLogFile       = "log.file"
	FlagLogJSON       = "log.json"
	FlagLogClickHouse = "log.clickhouse"
	FlagLogStructured = "log.structured"

	// Transport features
	FlagTransportInMemory = "transport.inmemory"
	FlagTransportREST     = "transport.rest"
	FlagTransportGRPC     = "transport.grpc"
	FlagTransportNATS     = "transport.nats"

	// Metrics features
	FlagMetricsBasic      = "metrics.basic"
	FlagMetricsPrometheus = "metrics.prometheus"

	// Scheduler features
	FlagSchedulerLocal       = "scheduler.local"
	FlagSchedulerDistributed = "scheduler.distributed"

	// Service features
	FlagServiceDiscovery = "service.discovery"
)

var AllFlags = []string{
	FlagDBSQLite,
	FlagDBPostgres,
	FlagDBReplication,
	FlagQueueMemory,
	FlagQueueRedis,
	FlagQueuePersistent,
	FlagLogFile,
	FlagLogJSON,
	FlagLogClickHouse,
	FlagLogStructured,
	FlagTransportInMemory,
	FlagTransportREST,
	FlagTransportGRPC,
	FlagTransportNATS,
	FlagMetricsBasic,
	FlagMetricsPrometheus,
	FlagSchedulerLocal,
	FlagSchedulerDistributed,
	FlagServiceDiscovery,
}

var FlagDescriptions = map[string]string{
	FlagDBSQLite:             "Use SQLite as database",
	FlagDBPostgres:           "Use PostgreSQL as database",
	FlagDBReplication:        "Enable database replication",
	FlagQueueMemory:          "Use in-memory queue",
	FlagQueueRedis:           "Use Redis as queue",
	FlagQueuePersistent:      "Enable persistent queue",
	FlagLogFile:              "Log to file",
	FlagLogJSON:              "Use JSON log format",
	FlagLogClickHouse:        "Log to ClickHouse",
	FlagLogStructured:        "Use structured logging",
	FlagTransportInMemory:    "Use in-memory transport",
	FlagTransportREST:        "Use REST transport",
	FlagTransportGRPC:        "Use gRPC transport",
	FlagTransportNATS:        "Use NATS transport",
	FlagMetricsBasic:         "Use basic metrics",
	FlagMetricsPrometheus:    "Use Prometheus metrics",
	FlagSchedulerLocal:       "Use local scheduler",
	FlagSchedulerDistributed: "Use distributed scheduler",
	FlagServiceDiscovery:     "Enable service discovery",
}
