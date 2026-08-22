package scheduler

import (
	"context"
	"fmt"
	"time"

	"adaptive-cron/internal/transport"
)

type Service struct {
	// Зависимости через интерфейсы
	taskManager transport.TaskManagerClient
	worker      transport.WorkerClient
	metricAgent transport.MetricAgentClient

	// Внутренние компоненты
	repo    Repository
	queue   Queue
	checker *AdaptiveChecker
	ticker  *time.Ticker
	stopCh  chan struct{}
}

func NewService(
	taskManager transport.TaskManagerClient,
	worker transport.WorkerClient,
	metricAgent transport.MetricAgentClient,
	repo Repository,
	queue Queue,
) *Service {
	return &Service{
		taskManager: taskManager,
		worker:      worker,
		metricAgent: metricAgent,
		repo:        repo,
		queue:       queue,
		checker:     NewAdaptiveChecker(metricAgent),
		stopCh:      make(chan struct{}),
	}
}

func (s *Service) ScheduleTask(ctx context.Context, task *transport.Task) (*transport.ScheduleResponse, error) {
	// Проверка ресурсов
	canRun, delay, reason := s.checker.CanRun(ctx, task)

	if !canRun {
		// Откладываем
		delayedTask := &transport.DelayedTask{
			TaskID:     task.ID,
			DelayUntil: time.Now().Add(delay),
			Attempt:    1,
		}
		s.queue.PushDelayed(ctx, delayedTask)

		return &transport.ScheduleResponse{
			Scheduled: false,
			Reason:    reason,
			Delay:     delay.String(),
		}, nil
	}

	// Выполняем
	resp, err := s.worker.Execute(ctx, &transport.ExecuteRequest{
		TaskID:  task.ID,
		Command: task.Command,
		Timeout: task.TimeoutSeconds,
	})
	if err != nil {
		return nil, err
	}

	return &transport.ScheduleResponse{
		Scheduled: true,
		JobID:     resp.JobID,
	}, nil
}

func (s *Service) GetNextTasks(ctx context.Context) ([]*transport.Task, error) {
	return s.repo.GetTasksDue(ctx)
}

func (s *Service) GetDelayedQueue(ctx context.Context) ([]*transport.DelayedTask, error) {
	return s.queue.GetDelayed(ctx)
}

func (s *Service) TriggerTask(ctx context.Context, taskID string, force bool) error {
	task, err := s.taskManager.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if !force && !task.IsCritical {
		canRun, delay, _ := s.checker.CanRun(ctx, task)
		if !canRun {
			s.queue.PushDelayed(ctx, &transport.DelayedTask{
				TaskID:     task.ID,
				DelayUntil: time.Now().Add(delay),
				Attempt:    1,
			})
			return fmt.Errorf("task delayed due to resource constraints: %v", delay)
		}
	}

	_, err = s.worker.Execute(ctx, &transport.ExecuteRequest{
		TaskID:  task.ID,
		Command: task.Command,
		Timeout: task.TimeoutSeconds,
	})
	return err
}

func (s *Service) Start(ctx context.Context) error {
	s.ticker = time.NewTicker(10 * time.Second)

	go s.run(ctx)
	go s.processDelayedQueue(ctx)

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	close(s.stopCh)
	if s.ticker != nil {
		s.ticker.Stop()
	}
	return nil
}

func (s *Service) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-s.ticker.C:
			s.scheduleDueTasks(ctx)
		}
	}
}

func (s *Service) scheduleDueTasks(ctx context.Context) {
	tasks, err := s.repo.GetTasksDue(ctx)
	if err != nil {
		return
	}

	for _, task := range tasks {
		s.ScheduleTask(ctx, task)
	}
}

func (s *Service) processDelayedQueue(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.processDelayed(ctx)
		}
	}
}

func (s *Service) processDelayed(ctx context.Context) {
	delayedTasks, err := s.queue.GetDueTasks(ctx)
	if err != nil {
		return
	}

	for _, delayed := range delayedTasks {
		// Удаляем из очереди
		s.queue.Remove(ctx, delayed.TaskID)

		// Получаем задачу
		task, err := s.taskManager.GetTask(ctx, delayed.TaskID)
		if err != nil {
			continue
		}

		// Пытаемся выполнить
		s.ScheduleTask(ctx, task)
	}
}
