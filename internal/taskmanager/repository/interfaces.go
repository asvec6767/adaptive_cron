package repository

import (
	"adaptive-cron/internal/taskmanager/domain"
	"context"
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	Get(ctx context.Context, id string) (*domain.Task, error)
	GetByName(ctx context.Context, name string) (*domain.Task, error)
	List(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id string) error
	Toggle(ctx context.Context, id string) error
}
