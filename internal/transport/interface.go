package transport

import (
	"adaptive-cron/internal/common/models"
	"context"
	"time"
)

// ServiceBinding defines how services communicate
type ServiceBinding interface {
	Connect(ctx context.Context) error
	Close() error

	// Client interfaces
	GetScheduler() SchedulerClient
	GetWorker() WorkerClient
	GetTaskManager() TaskManagerClient
	GetMetricAgent() MetricAgentClient

	// Service registration (for in-memory binding)
	RegisterScheduler(service SchedulerService) error
	RegisterWorker(service WorkerService) error
	RegisterTaskManager(service TaskManagerService) error
	RegisterMetricAgent(service MetricAgentService) error

	Health(ctx context.Context) error
	GetType() BindingType
}

type BindingType string

const (
	InMemoryBinding BindingType = "inmemory"
	RESTBinding     BindingType = "rest"
	GRPCBinding     BindingType = "grpc"
	NATSBinding     BindingType = "nats"
)

// Scheduler client interface
type SchedulerClient interface {
	ScheduleTask(ctx context.Context, task *models.Task) (*ScheduleResponse, error)
	GetNextTasks(ctx context.Context) ([]*models.Task, error)
	GetDelayedQueue(ctx context.Context) ([]*models.DelayedTask, error)
	TriggerTask(ctx context.Context, taskID string, force bool) error
}

// Worker client interface
type WorkerClient interface {
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
	GetStatus(ctx context.Context, jobID string) (*JobStatus, error)
	GetStats(ctx context.Context) (*WorkerStats, error)
}

// TaskManager client interface
type TaskManagerClient interface {
	CreateTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, id string) (*models.Task, error)
	ListTasks(ctx context.Context, filter *TaskFilter) ([]*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) error
	DeleteTask(ctx context.Context, id string) error
	ToggleTask(ctx context.Context, id string) error
}

// MetricAgent client interface
type MetricAgentClient interface {
	GetCurrentMetrics(ctx context.Context) (*models.SystemMetrics, error)
	GetMetricsHistory(ctx context.Context, limit int) ([]*models.SystemMetrics, error)
	GetLoadScore(ctx context.Context) (float64, error)
}

// Service interfaces (for registration)
type SchedulerService interface {
	ScheduleTask(ctx context.Context, task *models.Task) (*ScheduleResponse, error)
	GetNextTasks(ctx context.Context) ([]*models.Task, error)
	GetDelayedQueue(ctx context.Context) ([]*models.DelayedTask, error)
	TriggerTask(ctx context.Context, taskID string, force bool) error
}

type WorkerService interface {
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
	GetStatus(ctx context.Context, jobID string) (*JobStatus, error)
	GetStats(ctx context.Context) (*WorkerStats, error)
}

type TaskManagerService interface {
	CreateTask(ctx context.Context, task *models.Task) error
	GetTask(ctx context.Context, id string) (*models.Task, error)
	ListTasks(ctx context.Context, filter *TaskFilter) ([]*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) error
	DeleteTask(ctx context.Context, id string) error
	ToggleTask(ctx context.Context, id string) error
}

type MetricAgentService interface {
	GetCurrentMetrics(ctx context.Context) (*models.SystemMetrics, error)
	GetMetricsHistory(ctx context.Context, limit int) ([]*models.SystemMetrics, error)
	GetLoadScore(ctx context.Context) (float64, error)
}

// Transport models
type ScheduleResponse struct {
	Scheduled bool   `json:"scheduled"`
	JobID     string `json:"job_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Delay     string `json:"delay,omitempty"`
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
	TotalJobs     int `json:"total_jobs"`
}

type TaskFilter struct {
	Active   bool   `json:"active,omitempty"`
	Critical bool   `json:"critical,omitempty"`
	Name     string `json:"name,omitempty"`
}
