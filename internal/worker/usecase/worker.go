package usecase

import (
	"adaptive-cron/internal/worker/domain"
	"adaptive-cron/internal/worker/repository"
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/google/uuid"
)

type WorkerUseCase struct {
	repo          repository.JobRepository
	maxConcurrent int
	semaphore     chan struct{}
	jobs          map[string]*domain.Job
	mu            sync.RWMutex
}

func NewWorkerUseCase(repo repository.JobRepository, maxConcurrent int) *WorkerUseCase {
	return &WorkerUseCase{
		repo:          repo,
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
		jobs:          make(map[string]*domain.Job),
	}
}

func (w *WorkerUseCase) Execute(ctx context.Context, req *domain.ExecuteRequest) (*domain.ExecuteResponse, error) {
	jobID := uuid.New().String()
	timeout := time.Duration(req.Timeout) * time.Second
	if timeout == 0 {
		timeout = 300 * time.Second
	}

	job := &domain.Job{
		ID:        jobID,
		TaskID:    req.TaskID,
		Command:   req.Command,
		Timeout:   timeout,
		Status:    domain.JobStatusPending,
		StartTime: time.Now(),
	}

	w.mu.Lock()
	w.jobs[jobID] = job
	w.mu.Unlock()

	go w.runJob(job)

	return &domain.ExecuteResponse{JobID: jobID}, nil
}

func (w *WorkerUseCase) runJob(job *domain.Job) {
	select {
	case w.semaphore <- struct{}{}:
		defer func() { <-w.semaphore }()
	case <-time.After(30 * time.Second):
		job.Status = domain.JobStatusFailed
		job.Error = "failed to acquire execution slot"
		job.EndTime = time.Now()
		return
	}

	job.Status = domain.JobStatusRunning
	job.StartTime = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), job.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", job.Command)
	output, err := cmd.CombinedOutput()

	job.EndTime = time.Now()
	job.Output = string(output)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			job.Status = domain.JobStatusTimeout
			job.Error = "execution timeout"
			job.ExitCode = -1
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			job.Status = domain.JobStatusFailed
			job.ExitCode = exitErr.ExitCode()
			job.Error = err.Error()
		} else {
			job.Status = domain.JobStatusFailed
			job.ExitCode = -1
			job.Error = err.Error()
		}
	} else {
		job.Status = domain.JobStatusCompleted
		job.ExitCode = 0
	}
}

func (w *WorkerUseCase) GetStatus(ctx context.Context, jobID string) (*domain.JobStatusResponse, error) {
	w.mu.RLock()
	job, exists := w.jobs[jobID]
	w.mu.RUnlock()

	if !exists {
		return nil, domain.ErrJobNotFound
	}

	return &domain.JobStatusResponse{
		JobID:     job.ID,
		TaskID:    job.TaskID,
		Status:    job.Status,
		StartTime: job.StartTime,
		EndTime:   job.EndTime,
		ExitCode:  job.ExitCode,
		Output:    job.Output,
		Error:     job.Error,
	}, nil
}

func (w *WorkerUseCase) GetStats(ctx context.Context) (*domain.WorkerStats, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	stats := &domain.WorkerStats{
		TotalJobs: len(w.jobs),
	}

	for _, job := range w.jobs {
		switch job.Status {
		case domain.JobStatusRunning, domain.JobStatusPending:
			stats.ActiveJobs++
		case domain.JobStatusCompleted:
			stats.CompletedJobs++
		case domain.JobStatusFailed, domain.JobStatusTimeout:
			stats.FailedJobs++
		}
	}

	return stats, nil
}

func (w *WorkerUseCase) ListJobs(ctx context.Context) ([]*domain.Job, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	jobs := make([]*domain.Job, 0, len(w.jobs))
	for _, job := range w.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}
