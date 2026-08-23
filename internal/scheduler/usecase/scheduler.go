package usecase

import (
	"adaptive-cron/internal/scheduler/domain"
	"adaptive-cron/internal/scheduler/repository"
	"adaptive-cron/pkg/cronparser"
	"context"
	"time"
)

type WorkerClient interface {
	Execute(ctx context.Context, taskID, command string, timeout int) (string, error)
}

type SchedulerUseCase struct {
	taskRepo      repository.TaskRepository
	delayedQueue  repository.DelayedQueueRepository
	checker       *AdaptiveChecker
	workerClient  WorkerClient
	parser        *cronparser.Parser
	checkInterval time.Duration
}

func NewSchedulerUseCase(
	taskRepo repository.TaskRepository,
	delayedQueue repository.DelayedQueueRepository,
	checker *AdaptiveChecker,
	workerClient WorkerClient,
	checkInterval time.Duration,
) *SchedulerUseCase {
	return &SchedulerUseCase{
		taskRepo:      taskRepo,
		delayedQueue:  delayedQueue,
		checker:       checker,
		workerClient:  workerClient,
		parser:        cronparser.NewParser(),
		checkInterval: checkInterval,
	}
}

func (s *SchedulerUseCase) ScheduleTasks(ctx context.Context, now time.Time) error {
	tasks, err := s.taskRepo.GetActiveTasks(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if s.shouldRunAt(now, task.Schedule) {
			s.handleTask(ctx, task)
		}
	}

	return nil
}

func (s *SchedulerUseCase) shouldRunAt(now time.Time, schedule string) bool {
	next, err := s.parser.Next(now.Add(-time.Minute), schedule)
	if err != nil {
		return false
	}
	return next.Before(now) || next.Equal(now)
}

func (s *SchedulerUseCase) handleTask(ctx context.Context, task *domain.Task) {
	canRun, delay, _ := s.checker.CanRun(ctx, task)

	if canRun {
		jobID, err := s.workerClient.Execute(ctx, task.ID, task.Command, task.TimeoutSeconds)
		if err == nil {
			_ = jobID
		}
		return
	}

	delayedTask := &domain.DelayedTask{
		TaskID:     task.ID,
		DelayUntil: time.Now().Add(delay),
		Attempt:    1,
	}
	_ = s.delayedQueue.Push(ctx, delayedTask)
}

func (s *SchedulerUseCase) ProcessDelayedTasks(ctx context.Context, now time.Time) error {
	tasks, err := s.delayedQueue.PopDue(ctx, now)
	if err != nil {
		return err
	}

	for _, delayed := range tasks {
		task, err := s.taskRepo.GetTask(ctx, delayed.TaskID)
		if err != nil {
			continue
		}

		if !task.IsActive {
			continue
		}

		canRun, delay, _ := s.checker.CanRun(ctx, task)
		if canRun {
			_, _ = s.workerClient.Execute(ctx, task.ID, task.Command, task.TimeoutSeconds)
			continue
		}

		delayed.DelayUntil = time.Now().Add(delay)
		delayed.Attempt++
		_ = s.delayedQueue.Push(ctx, delayed)
	}

	return nil
}

func (s *SchedulerUseCase) TriggerTask(ctx context.Context, taskID string, force bool) error {
	task, err := s.taskRepo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if !task.IsActive {
		return domain.ErrTaskNotActive
	}

	if !force {
		canRun, delay, _ := s.checker.CanRun(ctx, task)
		if !canRun {
			delayedTask := &domain.DelayedTask{
				TaskID:     task.ID,
				DelayUntil: time.Now().Add(delay),
				Attempt:    1,
			}
			return s.delayedQueue.Push(ctx, delayedTask)
		}
	}

	_, err = s.workerClient.Execute(ctx, task.ID, task.Command, task.TimeoutSeconds)
	return err
}

func (s *SchedulerUseCase) GetDelayedQueue(ctx context.Context) ([]*domain.DelayedTask, error) {
	return s.delayedQueue.List(ctx)
}

func (s *SchedulerUseCase) RemoveFromQueue(ctx context.Context, taskID string) error {
	return s.delayedQueue.Remove(ctx, taskID)
}
