package app

import (
	"context"
	"fmt"
	"time"

	"adaptive-cron/internal/config"
	"adaptive-cron/internal/metrics"
	"adaptive-cron/internal/scheduler"
	"adaptive-cron/internal/taskmanager"
	"adaptive-cron/internal/transport"
	"adaptive-cron/internal/worker"
)

type App struct {
	binding     transport.ServiceBinding
	scheduler   *scheduler.Service
	worker      *worker.Service
	taskManager *taskmanager.Service
	metricAgent *metrics.Agent

	config *config.Config
}

func NewApp(cfg *config.Config) (*App, error) {
	app := &App{config: cfg}

	// Выбор биндинга на основе профиля
	if err := app.initBinding(); err != nil {
		return nil, err
	}

	// Создание сервисов
	if err := app.initServices(); err != nil {
		return nil, err
	}

	// Регистрация сервисов (только для in-memory)
	if app.binding.GetType() == transport.InMemoryBinding {
		if err := app.registerServices(); err != nil {
			return nil, err
		}
	}

	return app, nil
}

func (a *App) initBinding() error {
	switch a.config.Profile {
	case "lite":
		// In-memory биндинг
		a.binding = transport.NewInMemoryBinding()
		return nil

	case "standard":
		// REST биндинг (local)
		a.binding = transport.NewRESTBinding("http://localhost:8080", 30*time.Second)
		return nil

	case "enterprise":
		// REST + NATS (распределенный)
		// Для enterprise создаем комбинированный биндинг
		a.binding = transport.NewNATSBinding(a.config.Queue.Servers)
		return nil

	default:
		return fmt.Errorf("unknown profile: %s", a.config.Profile)
	}
}

func (a *App) initServices() error {
	// Создаем сервисы с клиентами от биндинга
	a.scheduler = scheduler.NewService(
		a.binding.GetTaskManager(),
		a.binding.GetWorker(),
		a.binding.GetMetricAgent(),
		// ... другие зависимости
	)

	a.worker = worker.NewService(a.config.Worker.MaxConcurrent)
	a.taskManager = taskmanager.NewService( /* зависимости */ )
	a.metricAgent = metrics.NewAgent( /* зависимости */ )

	return nil
}

func (a *App) registerServices() error {
	// Регистрируем сервисы в биндинге (только для in-memory)
	if err := a.binding.RegisterScheduler(a.scheduler); err != nil {
		return err
	}
	if err := a.binding.RegisterWorker(a.worker); err != nil {
		return err
	}
	if err := a.binding.RegisterTaskManager(a.taskManager); err != nil {
		return err
	}
	if err := a.binding.RegisterMetricAgent(a.metricAgent); err != nil {
		return err
	}
	return nil
}

func (a *App) Run(ctx context.Context) error {
	// Подключение биндинга
	if err := a.binding.Connect(ctx); err != nil {
		return err
	}
	defer a.binding.Close()

	// Запуск сервисов
	if err := a.scheduler.Start(ctx); err != nil {
		return err
	}
	defer a.scheduler.Stop(ctx)

	// ... аналогично для других сервисов

	// Ожидание завершения
	<-ctx.Done()
	return nil
}
