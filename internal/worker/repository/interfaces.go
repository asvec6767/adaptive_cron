package repository

import (
	"adaptive-cron/internal/worker/domain"
	"context"
)

type JobRepository interface {
	Save(ctx context.Context, job *domain.Job) error
	Get(ctx context.Context, id string) (*domain.Job, error)
	List(ctx context.Context) ([]*domain.Job, error)
	GetStats(ctx context.Context) (*domain.WorkerStats, error)
}
