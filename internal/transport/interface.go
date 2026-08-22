package transport

import (
	"context"
	"time"
)

// ServiceBinding определяет как компоненты общаются друг с другом
type ServiceBinding interface {
	// Соединение
	Connect(ctx context.Context) error
	Close() error

	// Получение клиентов
	GetScheduler() SchedulerClient
	GetWorker() WorkerClient
	GetTaskManager() TaskManagerClient
	GetMetricAgent() MetricAgentClient

	// Регистрация сервисов (для серверной стороны)
	RegisterScheduler(service SchedulerService) error
	RegisterWorker(service WorkerService) error
	RegisterTaskManager(service TaskManagerService) error
	RegisterMetricAgent(service MetricAgentService) error

	// Статус
	Health(ctx context.Context) error
	GetType() BindingType
}

type BindingType string

const (
	InMemoryBinding    BindingType = "inmemory"
	LocalBinding       BindingType = "local"       // localhost REST/gRPC
	DistributedBinding BindingType = "distributed" // NATS/gRPC across network
)

// Клиенты для взаимодействия с сервисами
type SchedulerClient interface {
	ScheduleTask(ctx context.Context, task *Task) (*ScheduleResponse, error)
	GetNextTasks(ctx context.Context) ([]*Task, error)
	GetDelayedQueue(ctx context.Context) ([]*DelayedTask, error)
	TriggerTask(ctx context.Context, taskID string, force bool) error
}

type WorkerClient interface {
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
	GetStatus(ctx context.Context, jobID string) (*JobStatus, error)
	GetStats(ctx context.Context) (*WorkerStats, error)
}

type TaskManagerClient interface {
	CreateTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, id string) (*Task, error)
	ListTasks(ctx context.Context, filter *TaskFilter) ([]*Task, error)
	UpdateTask(ctx context.Context, task *Task) error
	DeleteTask(ctx context.Context, id string) error
	ToggleTask(ctx context.Context, id string) error
}

type MetricAgentClient interface {
	GetCurrentMetrics(ctx context.Context) (*SystemMetrics, error)
	GetMetricsHistory(ctx context.Context, limit int) ([]*SystemMetrics, error)
	GetLoadScore(ctx context.Context) (float64, error)
}

// Сервисы (для регистрации)
type SchedulerService interface {
	ScheduleTask(ctx context.Context, task *Task) (*ScheduleResponse, error)
	GetNextTasks(ctx context.Context) ([]*Task, error)
	GetDelayedQueue(ctx context.Context) ([]*DelayedTask, error)
	TriggerTask(ctx context.Context, taskID string, force bool) error
}

type WorkerService interface {
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
	GetStatus(ctx context.Context, jobID string) (*JobStatus, error)
	GetStats(ctx context.Context) (*WorkerStats, error)
}

type TaskManagerService interface {
	CreateTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, id string) (*Task, error)
	ListTasks(ctx context.Context, filter *TaskFilter) ([]*Task, error)
	UpdateTask(ctx context.Context, task *Task) error
	DeleteTask(ctx context.Context, id string) error
	ToggleTask(ctx context.Context, id string) error
}

type MetricAgentService interface {
	GetCurrentMetrics(ctx context.Context) (*SystemMetrics, error)
	GetMetricsHistory(ctx context.Context, limit int) ([]*SystemMetrics, error)
	GetLoadScore(ctx context.Context) (float64, error)
}

// Модели данных для транспорта
type Task struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Command        string    `json:"command"`
	Schedule       string    `json:"schedule"`
	IsCritical     bool      `json:"is_critical"`
	MaxLoadScore   float64   `json:"max_load_score"`
	MaxCPUPercent  float64   `json:"max_cpu_percent"`
	MaxMemoryMB    int64     `json:"max_memory_mb"`
	TimeoutSeconds int       `json:"timeout_seconds"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SystemMetrics struct {
	Timestamp     time.Time `json:"timestamp"`
	CPUPercent    float64   `json:"cpu_percent"`
	LoadAvg1      float64   `json:"load_avg1"`
	LoadAvg5      float64   `json:"load_avg5"`
	LoadAvg15     float64   `json:"load_avg15"`
	MemoryUsed    uint64    `json:"memory_used"`
	MemoryTotal   uint64    `json:"memory_total"`
	MemoryPercent float64   `json:"memory_percent"`
	SwapUsed      uint64    `json:"swap_used"`
	SwapTotal     uint64    `json:"swap_total"`
	SwapPercent   float64   `json:"swap_percent"`
}

type DelayedTask struct {
	TaskID     string    `json:"task_id"`
	DelayUntil time.Time `json:"delay_until"`
	Attempt    int       `json:"attempt"`
}

type ExecuteRequest struct {
	TaskID  string `json:"task_id"`
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

type ExecuteResponse struct {
	JobID string `json:"job_id"`
}

type JobStatus struct {
	JobID     string    `json:"job_id"`
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	ExitCode  int       `json:"exit_code"`
	Output    string    `json:"output"`
	Error     string    `json:"error"`
}

type WorkerStats struct {
	ActiveJobs    int `json:"active_jobs"`
	CompletedJobs int `json:"completed_jobs"`
	FailedJobs    int `json:"failed_jobs"`
}

type ScheduleResponse struct {
	Scheduled bool   `json:"scheduled"`
	JobID     string `json:"job_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Delay     string `json:"delay,omitempty"`
}

type TaskFilter struct {
	Active   bool   `json:"active,omitempty"`
	Critical bool   `json:"critical,omitempty"`
	Name     string `json:"name,omitempty"`
}
