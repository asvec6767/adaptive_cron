package models

import (
	"time"
)

// ClickHouseLog структура для хранения логов в ClickHouse
type ClickHouseLog struct {
	Timestamp  time.Time `ch:"timestamp"`
	Level      string    `ch:"level"`
	Service    string    `ch:"service"`
	TaskID     string    `ch:"task_id"`
	TaskName   string    `ch:"task_name"`
	JobID      string    `ch:"job_id"`
	Message    string    `ch:"message"`
	Error      string    `ch:"error"`
	DurationMs int64     `ch:"duration_ms"`
	ExitCode   int       `ch:"exit_code"`
	CPUPercent float64   `ch:"cpu_percent"`
	MemoryMB   int64     `ch:"memory_mb"`
	LoadScore  float64   `ch:"load_score"`
	IPAddress  string    `ch:"ip_address"`
	UserAgent  string    `ch:"user_agent"`
	RequestID  string    `ch:"request_id"`
	Tags       []string  `ch:"tags"`
	Metadata   string    `ch:"metadata"` // JSON поле
}

// ClickHouseLogBatch для массовой вставки
type ClickHouseLogBatch struct {
	Logs []ClickHouseLog
}

// ClickHouseQuery для запросов к ClickHouse
type ClickHouseQuery struct {
	StartTime time.Time
	EndTime   time.Time
	Level     string
	Service   string
	TaskID    string
	TaskName  string
	Limit     int
	OrderBy   string
}
