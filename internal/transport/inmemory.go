package transport

import (
	"context"
	"sync"
)

type InMemoryBinding struct {
	mu sync.RWMutex

	schedulerClient   *InMemorySchedulerClient
	workerClient      *InMemoryWorkerClient
	taskManagerClient *InMemoryTaskManagerClient
	metricAgentClient *InMemoryMetricAgentClient

	schedulerService   SchedulerService
	workerService      WorkerService
	taskManagerService TaskManagerService
	metricAgentService MetricAgentService
}

func NewInMemoryBinding() *InMemoryBinding {
	return &InMemoryBinding{
		schedulerClient:   &InMemorySchedulerClient{},
		workerClient:      &InMemoryWorkerClient{},
		taskManagerClient: &InMemoryTaskManagerClient{},
		metricAgentClient: &InMemoryMetricAgentClient{},
	}
}

func (b *InMemoryBinding) Connect(ctx context.Context) error {
	// In-memory не требует подключения
	return nil
}

func (b *InMemoryBinding) Close() error {
	return nil
}

func (b *InMemoryBinding) GetScheduler() SchedulerClient {
	return b.schedulerClient
}

func (b *InMemoryBinding) GetWorker() WorkerClient {
	return b.workerClient
}

func (b *InMemoryBinding) GetTaskManager() TaskManagerClient {
	return b.taskManagerClient
}

func (b *InMemoryBinding) GetMetricAgent() MetricAgentClient {
	return b.metricAgentClient
}

func (b *InMemoryBinding) RegisterScheduler(service SchedulerService) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.schedulerService = service
	b.schedulerClient.service = service
	return nil
}

func (b *InMemoryBinding) RegisterWorker(service WorkerService) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.workerService = service
	b.workerClient.service = service
	return nil
}

func (b *InMemoryBinding) RegisterTaskManager(service TaskManagerService) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.taskManagerService = service
	b.taskManagerClient.service = service
	return nil
}

func (b *InMemoryBinding) RegisterMetricAgent(service MetricAgentService) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.metricAgentService = service
	b.metricAgentClient.service = service
	return nil
}

func (b *InMemoryBinding) Health(ctx context.Context) error {
	return nil
}

func (b *InMemoryBinding) GetType() BindingType {
	return InMemoryBinding
}

// InMemory клиенты
type InMemorySchedulerClient struct {
	service SchedulerService
}

func (c *InMemorySchedulerClient) ScheduleTask(ctx context.Context, task *Task) (*ScheduleResponse, error) {
	return c.service.ScheduleTask(ctx, task)
}

func (c *InMemorySchedulerClient) GetNextTasks(ctx context.Context) ([]*Task, error) {
	return c.service.GetNextTasks(ctx)
}

func (c *InMemorySchedulerClient) GetDelayedQueue(ctx context.Context) ([]*DelayedTask, error) {
	return c.service.GetDelayedQueue(ctx)
}

func (c *InMemorySchedulerClient) TriggerTask(ctx context.Context, taskID string, force bool) error {
	return c.service.TriggerTask(ctx, taskID, force)
}

type InMemoryWorkerClient struct {
	service WorkerService
}

func (c *InMemoryWorkerClient) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	return c.service.Execute(ctx, req)
}

func (c *InMemoryWorkerClient) GetStatus(ctx context.Context, jobID string) (*JobStatus, error) {
	return c.service.GetStatus(ctx, jobID)
}

func (c *InMemoryWorkerClient) GetStats(ctx context.Context) (*WorkerStats, error) {
	return c.service.GetStats(ctx)
}

type InMemoryTaskManagerClient struct {
	service TaskManagerService
}

func (c *InMemoryTaskManagerClient) CreateTask(ctx context.Context, task *Task) error {
	return c.service.CreateTask(ctx, task)
}

func (c *InMemoryTaskManagerClient) GetTask(ctx context.Context, id string) (*Task, error) {
	return c.service.GetTask(ctx, id)
}

func (c *InMemoryTaskManagerClient) ListTasks(ctx context.Context, filter *TaskFilter) ([]*Task, error) {
	return c.service.ListTasks(ctx, filter)
}

func (c *InMemoryTaskManagerClient) UpdateTask(ctx context.Context, task *Task) error {
	return c.service.UpdateTask(ctx, task)
}

func (c *InMemoryTaskManagerClient) DeleteTask(ctx context.Context, id string) error {
	return c.service.DeleteTask(ctx, id)
}

func (c *InMemoryTaskManagerClient) ToggleTask(ctx context.Context, id string) error {
	return c.service.ToggleTask(ctx, id)
}

type InMemoryMetricAgentClient struct {
	service MetricAgentService
}

func (c *InMemoryMetricAgentClient) GetCurrentMetrics(ctx context.Context) (*SystemMetrics, error) {
	return c.service.GetCurrentMetrics(ctx)
}

func (c *InMemoryMetricAgentClient) GetMetricsHistory(ctx context.Context, limit int) ([]*SystemMetrics, error) {
	return c.service.GetMetricsHistory(ctx, limit)
}

func (c *InMemoryMetricAgentClient) GetLoadScore(ctx context.Context) (float64, error) {
	return c.service.GetLoadScore(ctx)
}
