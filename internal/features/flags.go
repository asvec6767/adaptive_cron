package features

// Все возможные feature flags
const (
	// Database features
	FlagDBPostgres    = "db.postgres"
	FlagDBSQLite      = "db.sqlite"
	FlagDBMySQL       = "db.mysql"
	FlagDBReplication = "db.replication"

	// Queue features
	FlagQueueMemory     = "queue.memory"
	FlagQueueRedis      = "queue.redis"
	FlagQueueNATS       = "queue.nats"
	FlagQueuePersistent = "queue.persistent"

	// Logging features
	FlagLogFile       = "log.file"
	FlagLogJSON       = "log.json"
	FlagLogClickHouse = "log.clickhouse"
	FlagLogELK        = "log.elk"
	FlagLogStructured = "log.structured"

	// Transport features
	FlagTransportREST = "transport.rest"
	FlagTransportGRPC = "transport.grpc"
	FlagTransportNATS = "transport.nats"

	// Metrics features
	FlagMetricsPrometheus = "metrics.prometheus"
	FlagMetricsInMemory   = "metrics.inmemory"
	FlagMetricsStatsD     = "metrics.statsd"

	// Service features
	FlagServiceDiscovery     = "service.discovery"
	FlagDistributedScheduler = "scheduler.distributed"
	FlagWorkerPool           = "worker.pool"
	FlagTaskDependencies     = "task.dependencies"
	FlagTaskHistory          = "task.history"

	// Monitoring features
	FlagTracing      = "monitoring.tracing"
	FlagHealthChecks = "monitoring.health"
	FlagProfiling    = "monitoring.profiling"

	// Performance features
	FlagBatching    = "performance.batching"
	FlagCaching     = "performance.caching"
	FlagCompression = "performance.compression"
)

// Default features for each profile
var ProfileFeatures = map[string][]string{
	"lite": {
		FlagDBSQLite,
		FlagQueueMemory,
		FlagLogFile,
		FlagTransportREST,
		FlagMetricsInMemory,
		FlagHealthChecks,
	},
	"standard": {
		FlagDBPostgres,
		FlagQueueRedis,
		FlagLogFile,
		FlagLogJSON,
		FlagTransportREST,
		FlagTransportGRPC,
		FlagMetricsPrometheus,
		FlagWorkerPool,
		FlagTaskHistory,
		FlagHealthChecks,
		FlagCaching,
		FlagProfiling,
	},
	"enterprise": {
		FlagDBPostgres,
		FlagDBReplication,
		FlagQueueNATS,
		FlagQueuePersistent,
		FlagLogClickHouse,
		FlagLogStructured,
		FlagTransportGRPC,
		FlagTransportNATS,
		FlagMetricsPrometheus,
		FlagServiceDiscovery,
		FlagDistributedScheduler,
		FlagWorkerPool,
		FlagTaskDependencies,
		FlagTaskHistory,
		FlagTracing,
		FlagHealthChecks,
		FlagBatching,
		FlagCaching,
		FlagCompression,
		FlagProfiling,
	},
	"cloud": {
		FlagDBPostgres,
		FlagDBReplication,
		FlagQueueNATS,
		FlagQueuePersistent,
		FlagLogClickHouse,
		FlagLogStructured,
		FlagTransportGRPC,
		FlagTransportNATS,
		FlagMetricsPrometheus,
		FlagServiceDiscovery,
		FlagDistributedScheduler,
		FlagWorkerPool,
		FlagTaskDependencies,
		FlagTaskHistory,
		FlagTracing,
		FlagHealthChecks,
		FlagBatching,
		FlagCaching,
		FlagCompression,
		FlagProfiling,
	},
}
