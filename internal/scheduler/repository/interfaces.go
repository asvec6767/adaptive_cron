package repository

import (
	"adaptive-cron/internal/scheduler/domain"
	"context"
	"time"
)

type TaskRepository interface {
	GetActiveTasks(ctx context.Context) ([]*domain.Task, error)
	GetTask(ctx context.Context, id string) (*domain.Task, error)
	UpdateTask(ctx context.Context, task *domain.Task) error
}

type DelayedQueueRepository interface {
	Push(ctx context.Context, task *domain.DelayedTask) error
	PopDue(ctx context.Context, now time.Time) ([]*domain.DelayedTask, error)
	Remove(ctx context.Context, taskID string) error
	List(ctx context.Context) ([]*domain.DelayedTask, error)
	Size(ctx context.Context) (int, error)
}
