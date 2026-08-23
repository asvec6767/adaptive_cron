package models

import (
	"time"
)

type JobLog struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TaskID         string    `json:"task_id" gorm:"index"`
	TaskName       string    `json:"task_name"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	ExitCode       int       `json:"exit_code"`
	Output         string    `json:"output" gorm:"type:text"`
	Error          string    `json:"error" gorm:"type:text"`
	MetricsAtStart string    `json:"metrics_at_start" gorm:"type:jsonb"`
	CreatedAt      time.Time `json:"created_at"`
}

type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Service    string                 `json:"service"`
	TaskID     string                 `json:"task_id,omitempty"`
	TaskName   string                 `json:"task_name,omitempty"`
	JobID      string                 `json:"job_id,omitempty"`
	Message    string                 `json:"message"`
	Error      string                 `json:"error,omitempty"`
	DurationMs int64                  `json:"duration_ms,omitempty"`
	ExitCode   int                    `json:"exit_code,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
}

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
	Tags       []string  `ch:"tags"`
	Metadata   string    `ch:"metadata"`
}
