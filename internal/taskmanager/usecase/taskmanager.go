package usecase

import (
	"adaptive-cron/internal/taskmanager/domain"
	"adaptive-cron/internal/taskmanager/repository"
	"context"
	"time"

	"github.com/google/uuid"
)

type TaskManagerUseCase struct {
	repo repository.TaskRepository
}

func NewTaskManagerUseCase(repo repository.TaskRepository) *TaskManagerUseCase {
	return &TaskManagerUseCase{
		repo: repo,
	}
}

func (u *TaskManagerUseCase) CreateTask(ctx context.Context, req *domain.CreateTaskRequest) (*domain.Task, error) {
	task := &domain.Task{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Command:        req.Command,
		Schedule:       req.Schedule,
		IsCritical:     req.IsCritical,
		MaxLoadScore:   req.MaxLoadScore,
		MaxCPUPercent:  req.MaxCPUPercent,
		MaxMemoryMB:    req.MaxMemoryMB,
		TimeoutSeconds: req.TimeoutSeconds,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if task.MaxLoadScore == 0 {
		task.MaxLoadScore = 0.7
	}
	if task.MaxCPUPercent == 0 {
		task.MaxCPUPercent = 80
	}
	if task.MaxMemoryMB == 0 {
		task.MaxMemoryMB = 1024
	}
	if task.TimeoutSeconds == 0 {
		task.TimeoutSeconds = 300
	}

	if err := task.Validate(); err != nil {
		return nil, err
	}

	existing, _ := u.repo.GetByName(ctx, task.Name)
	if existing != nil {
		return nil, domain.ErrTaskAlreadyExists
	}

	if err := u.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (u *TaskManagerUseCase) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	return u.repo.Get(ctx, id)
}

func (u *TaskManagerUseCase) ListTasks(ctx context.Context, filter *domain.TaskFilter) ([]*domain.Task, error) {
	return u.repo.List(ctx, filter)
}

func (u *TaskManagerUseCase) UpdateTask(ctx context.Context, id string, req *domain.UpdateTaskRequest) (*domain.Task, error) {
	task, err := u.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		task.Name = *req.Name
	}
	if req.Command != nil {
		task.Command = *req.Command
	}
	if req.Schedule != nil {
		task.Schedule = *req.Schedule
	}
	if req.IsCritical != nil {
		task.IsCritical = *req.IsCritical
	}
	if req.MaxLoadScore != nil {
		task.MaxLoadScore = *req.MaxLoadScore
	}
	if req.MaxCPUPercent != nil {
		task.MaxCPUPercent = *req.MaxCPUPercent
	}
	if req.MaxMemoryMB != nil {
		task.MaxMemoryMB = *req.MaxMemoryMB
	}
	if req.TimeoutSeconds != nil {
		task.TimeoutSeconds = *req.TimeoutSeconds
	}
	if req.IsActive != nil {
		task.IsActive = *req.IsActive
	}

	task.UpdatedAt = time.Now()

	if err := task.Validate(); err != nil {
		return nil, err
	}

	if err := u.repo.Update(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}

func (u *TaskManagerUseCase) DeleteTask(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *TaskManagerUseCase) ToggleTask(ctx context.Context, id string) (*domain.Task, error) {
	if err := u.repo.Toggle(ctx, id); err != nil {
		return nil, err
	}
	return u.repo.Get(ctx, id)
}
