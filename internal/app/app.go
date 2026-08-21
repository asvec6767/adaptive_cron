package app

import (
	"context"
	"fmt"
	"sync"

	"adaptive-cron/internal/config"
	"adaptive-cron/internal/db"
	"adaptive-cron/internal/features"
	"adaptive-cron/internal/log"
	"adaptive-cron/internal/queue"
)

type App struct {
	config   *config.Config
	db       db.Database
	queue    queue.Queue
	logger   log.Logger
	features map[string]*features.FeatureFlag

	scheduler *Scheduler
	worker    *Worker
	api       *API
	metrics   *MetricsCollector

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

type Option func(*App)

func WithDatabase(db db.Database) Option {
	return func(a *App) {
		a.db = db
	}
}

func WithQueue(q queue.Queue) Option {
	return func(a *App) {
		a.queue = q
	}
}

func WithLogger(l log.Logger) Option {
	return func(a *App) {
		a.logger = l
	}
}

func WithConfig(c *config.Config) Option {
	return func(a *App) {
		a.config = c
	}
}

func WithFeatures(f map[string]*features.FeatureFlag) Option {
	return func(a *App) {
		a.features = f
	}
}

func NewApp(options ...Option) *App {
	app := &App{}
	for _, opt := range options {
		opt(app)
	}
	app.ctx, app.cancel = context.WithCancel(context.Background())
	return app
}

func (a *App) Run() error {
	// Инициализация компонентов в зависимости от feature flags
	if err := a.initComponents(); err != nil {
		return err
	}

	// Запуск компонентов
	a.startComponents()

	// Ожидание завершения
	<-a.ctx.Done()
	a.wg.Wait()

	return nil
}

func (a *App) initComponents() error {
	// Инициализация БД
	if a.db != nil {
		if err := a.db.Connect(); err != nil {
			return fmt.Errorf("failed to connect DB: %w", err)
		}
		if err := a.db.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate DB: %w", err)
		}
	}

	// Инициализация планировщика
	a.scheduler = NewScheduler(SchedulerConfig{
		DB:            a.db,
		Queue:         a.queue,
		Logger:        a.logger,
		CheckInterval: a.config.Scheduler.CheckInterval,
		Distributed:   features.IsEnabled(features.FlagDistributedScheduler),
	})

	// Инициализация воркера
	if features.IsEnabled(features.FlagWorkerPool) {
		a.worker = NewWorker(WorkerConfig{
			Queue:         a.queue,
			Logger:        a.logger,
			MaxConcurrent: a.config.Worker.MaxConcurrent,
		})
	}

	// Инициализация API
	a.api = NewAPI(APIConfig{
		Port:       a.config.Server.HTTPPort,
		Logger:     a.logger,
		Scheduler:  a.scheduler,
		DB:         a.db,
		EnableGRPC: features.IsEnabled(features.FlagTransportGRPC),
	})

	// Инициализация метрик
	if features.IsEnabled(features.FlagMetricsPrometheus) {
		a.metrics = NewPrometheusMetrics(a.config.Metrics.Port)
	} else if features.IsEnabled(features.FlagMetricsInMemory) {
		a.metrics = NewInMemoryMetrics()
	}

	return nil
}

func (a *App) startComponents() {
	// Запуск планировщика
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.scheduler.Start(a.ctx); err != nil {
			a.logger.Error("Scheduler error", "error", err)
		}
	}()

	// Запуск воркера
	if a.worker != nil {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			if err := a.worker.Start(a.ctx); err != nil {
				a.logger.Error("Worker error", "error", err)
			}
		}()
	}

	// Запуск API
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.api.Start(a.ctx); err != nil {
			a.logger.Error("API error", "error", err)
		}
	}()

	// Запуск метрик
	if a.metrics != nil {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			a.metrics.Start(a.ctx)
		}()
	}
}
