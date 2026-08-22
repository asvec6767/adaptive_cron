package worker

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"adaptive-cron/internal/transport"

	"github.com/google/uuid"
)

type Service struct {
	mu            sync.RWMutex
	jobs          map[string]*Job
	maxConcurrent int
	semaphore     chan struct{}
}

type Job struct {
	ID        string
	TaskID    string
	Command   string
	Timeout   time.Duration
	StartTime time.Time
	EndTime   time.Time
	ExitCode  int
	Output    string
	Error     string
	Done      chan bool
}

func NewService(maxConcurrent int) *Service {
	return &Service{
		jobs:          make(map[string]*Job),
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
	}
}

func (s *Service) Execute(ctx context.Context, req *transport.ExecuteRequest) (*transport.ExecuteResponse, error) {
	jobID := uuid.New().String()
	timeout := time.Duration(req.Timeout) * time.Second
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	job := &Job{
		ID:      jobID,
		TaskID:  req.TaskID,
		Command: req.Command,
		Timeout: timeout,
		Done:    make(chan bool),
	}

	s.mu.Lock()
	s.jobs[jobID] = job
	s.mu.Unlock()

	go s.runJob(job)

	return &transport.ExecuteResponse{JobID: jobID}, nil
}

func (s *Service) runJob(job *Job) {
	// Acquire semaphore
	select {
	case s.semaphore <- struct{}{}:
		defer func() { <-s.semaphore }()
	case <-time.After(30 * time.Second):
		job.Error = "Failed to acquire execution slot"
		job.Done <- true
		return
	}

	job.StartTime = time.Now()

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
	defer cancel()

	// Выполняем команду
	cmd := exec.CommandContext(ctx, "sh", "-c", job.Command)
	output, err := cmd.CombinedOutput()

	job.EndTime = time.Now()
	job.Output = string(output)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			job.Error = "Execution timeout"
			job.ExitCode = -1
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			job.ExitCode = exitErr.ExitCode()
			job.Error = err.Error()
		} else {
			job.ExitCode = -1
			job.Error = err.Error()
		}
	} else {
		job.ExitCode = 0
	}

	job.Done <- true
}

func (s *Service) GetStatus(ctx context.Context, jobID string) (*transport.JobStatus, error) {
	s.mu.RLock()
	job, exists := s.jobs[jobID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("job not found")
	}

	return &transport.JobStatus{
		JobID:     job.ID,
		TaskID:    job.TaskID,
		Status:    s.getStatus(job),
		StartTime: job.StartTime,
		EndTime:   job.EndTime,
		ExitCode:  job.ExitCode,
		Output:    job.Output,
		Error:     job.Error,
	}, nil
}

func (s *Service) GetStats(ctx context.Context) (*transport.WorkerStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	active := 0
	completed := 0
	failed := 0

	for _, job := range s.jobs {
		if job.EndTime.IsZero() {
			active++
		} else if job.ExitCode == 0 {
			completed++
		} else {
			failed++
		}
	}

	return &transport.WorkerStats{
		ActiveJobs:    active,
		CompletedJobs: completed,
		FailedJobs:    failed,
	}, nil
}

func (s *Service) getStatus(job *Job) string {
	if job.EndTime.IsZero() {
		return "running"
	}
	if job.ExitCode == 0 {
		return "completed"
	}
	return "failed"
}
