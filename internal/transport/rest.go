package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RESTBinding struct {
	baseURL string
	client  *http.Client

	// Клиенты для сервисов
	schedulerClient   *RESTClient
	workerClient      *RESTClient
	taskManagerClient *RESTClient
	metricAgentClient *RESTClient
}

func NewRESTBinding(baseURL string, timeout time.Duration) *RESTBinding {
	return &RESTBinding{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (b *RESTBinding) Connect(ctx context.Context) error {
	// Проверяем доступность
	if err := b.Health(ctx); err != nil {
		return fmt.Errorf("failed to connect to REST services: %w", err)
	}
	return nil
}

func (b *RESTBinding) Close() error {
	b.client.CloseIdleConnections()
	return nil
}

func (b *RESTBinding) GetScheduler() SchedulerClient {
	if b.schedulerClient == nil {
		b.schedulerClient = NewRESTClient(b.baseURL, b.client)
	}
	return b.schedulerClient
}

func (b *RESTBinding) GetWorker() WorkerClient {
	if b.workerClient == nil {
		b.workerClient = NewRESTClient(b.baseURL, b.client)
	}
	return b.workerClient
}

func (b *RESTBinding) GetTaskManager() TaskManagerClient {
	if b.taskManagerClient == nil {
		b.taskManagerClient = NewRESTClient(b.baseURL, b.client)
	}
	return b.taskManagerClient
}

func (b *RESTBinding) GetMetricAgent() MetricAgentClient {
	if b.metricAgentClient == nil {
		b.metricAgentClient = NewRESTClient(b.baseURL, b.client)
	}
	return b.metricAgentClient
}

func (b *RESTBinding) RegisterScheduler(service SchedulerService) error {
	// REST не регистрирует сервисы, они должны быть запущены отдельно
	return fmt.Errorf("REST binding does not support service registration")
}

func (b *RESTBinding) RegisterWorker(service WorkerService) error {
	return fmt.Errorf("REST binding does not support service registration")
}

func (b *RESTBinding) RegisterTaskManager(service TaskManagerService) error {
	return fmt.Errorf("REST binding does not support service registration")
}

func (b *RESTBinding) RegisterMetricAgent(service MetricAgentService) error {
	return fmt.Errorf("REST binding does not support service registration")
}

func (b *RESTBinding) Health(ctx context.Context) error {
	resp, err := b.client.Get(b.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	return nil
}

func (b *RESTBinding) GetType() BindingType {
	return LocalBinding
}

// REST клиент
type RESTClient struct {
	baseURL string
	client  *http.Client
}

func NewRESTClient(baseURL string, client *http.Client) *RESTClient {
	return &RESTClient{
		baseURL: baseURL,
		client:  client,
	}
}

func (c *RESTClient) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %d, %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// Реализация клиентских методов для Scheduler
func (c *RESTClient) ScheduleTask(ctx context.Context, task *Task) (*ScheduleResponse, error) {
	var result ScheduleResponse
	err := c.doRequest(ctx, "POST", "/api/v1/scheduler/schedule", task, &result)
	return &result, err
}

func (c *RESTClient) GetNextTasks(ctx context.Context) ([]*Task, error) {
	var result []*Task
	err := c.doRequest(ctx, "GET", "/api/v1/scheduler/next", nil, &result)
	return result, err
}

func (c *RESTClient) GetDelayedQueue(ctx context.Context) ([]*DelayedTask, error) {
	var result []*DelayedTask
	err := c.doRequest(ctx, "GET", "/api/v1/scheduler/queue", nil, &result)
	return result, err
}

func (c *RESTClient) TriggerTask(ctx context.Context, taskID string, force bool) error {
	path := "/api/v1/scheduler/trigger/" + taskID
	if force {
		path += "?force=true"
	}
	return c.doRequest(ctx, "POST", path, nil, nil)
}

// Реализация клиентских методов для Worker
func (c *RESTClient) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
	var result ExecuteResponse
	err := c.doRequest(ctx, "POST", "/api/v1/execute", req, &result)
	return &result, err
}

func (c *RESTClient) GetStatus(ctx context.Context, jobID string) (*JobStatus, error) {
	var result JobStatus
	err := c.doRequest(ctx, "GET", "/api/v1/status/"+jobID, nil, &result)
	return &result, err
}

func (c *RESTClient) GetStats(ctx context.Context) (*WorkerStats, error) {
	var result WorkerStats
	err := c.doRequest(ctx, "GET", "/api/v1/worker/stats", nil, &result)
	return &result, err
}

// Реализация клиентских методов для TaskManager
func (c *RESTClient) CreateTask(ctx context.Context, task *Task) error {
	return c.doRequest(ctx, "POST", "/api/v1/tasks", task, nil)
}

func (c *RESTClient) GetTask(ctx context.Context, id string) (*Task, error) {
	var result Task
	err := c.doRequest(ctx, "GET", "/api/v1/tasks/"+id, nil, &result)
	return &result, err
}

func (c *RESTClient) ListTasks(ctx context.Context, filter *TaskFilter) ([]*Task, error) {
	var result []*Task
	path := "/api/v1/tasks"
	if filter != nil {
		// Добавляем параметры
	}
	err := c.doRequest(ctx, "GET", path, nil, &result)
	return result, err
}

func (c *RESTClient) UpdateTask(ctx context.Context, task *Task) error {
	return c.doRequest(ctx, "PUT", "/api/v1/tasks/"+task.ID, task, nil)
}

func (c *RESTClient) DeleteTask(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", "/api/v1/tasks/"+id, nil, nil)
}

func (c *RESTClient) ToggleTask(ctx context.Context, id string) error {
	return c.doRequest(ctx, "POST", "/api/v1/tasks/"+id+"/toggle", nil, nil)
}

// Реализация клиентских методов для MetricAgent
func (c *RESTClient) GetCurrentMetrics(ctx context.Context) (*SystemMetrics, error) {
	var result SystemMetrics
	err := c.doRequest(ctx, "GET", "/api/v1/metrics/current", nil, &result)
	return &result, err
}

func (c *RESTClient) GetMetricsHistory(ctx context.Context, limit int) ([]*SystemMetrics, error) {
	var result []*SystemMetrics
	err := c.doRequest(ctx, "GET", "/api/v1/metrics/history?limit="+string(rune(limit)), nil, &result)
	return result, err
}

func (c *RESTClient) GetLoadScore(ctx context.Context) (float64, error) {
	var result struct {
		Score float64 `json:"load_score"`
	}
	err := c.doRequest(ctx, "GET", "/api/v1/metrics/load", nil, &result)
	return result.Score, err
}
