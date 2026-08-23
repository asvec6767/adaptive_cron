package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"adaptive-cron/internal/config"
	"adaptive-cron/internal/features"
)

type App struct {
	config  *config.Config
	mu      sync.RWMutex
	running bool

	// Services
	scheduler   SchedulerService
	worker      WorkerService
	taskManager TaskManagerService
	metricAgent MetricAgentService

	// Components
	transport TransportBinding
	apiServer APIServer

	// Lifecycle
	cancelFuncs []context.CancelFunc
	wg          sync.WaitGroup
}

type Option func(*App) error

func NewApp(cfg *config.Config, opts ...Option) (*App, error) {
	app := &App{
		config: cfg,
	}

	// Initialize features from config
	features.InitFromConfig(cfg.Features)

	// Apply options
	for _, opt := range opts {
		if err := opt(app); err != nil {
			return nil, err
		}
	}

	// Build components
	if err := app.buildComponents(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) buildComponents() error {
	// Build transport based on features
	transport, err := a.buildTransport()
	if err != nil {
		return err
	}
	a.transport = transport

	// Build services
	if err := a.buildServices(); err != nil {
		return err
	}

	// Build API server
	a.apiServer = NewAPIServer(a.config, a.transport)

	return nil
}

func (a *App) buildTransport() (TransportBinding, error) {
	if features.IsEnabled("transport.inmemory") {
		return NewInMemoryTransport(), nil
	}
	if features.IsEnabled("transport.rest") {
		return NewRESTTransport(a.config.Server.HTTPPort), nil
	}
	if features.IsEnabled("transport.grpc") {
		return NewGRPCTransport(a.config.Server.GRPCPort), nil
	}
	if features.IsEnabled("transport.nats") {
		return NewNATSTransport(a.config.Queue.Servers), nil
	}
	return nil, fmt.Errorf("no transport feature enabled")
}

func (a *App) buildServices() error {
	// Build services with injected dependencies
	a.scheduler = NewSchedulerService(a.config, a.transport)
	a.worker = NewWorkerService(a.config, a.transport)
	a.taskManager = NewTaskManagerService(a.config, a.transport)
	a.metricAgent = NewMetricAgentService(a.config, a.transport)

	// Register services with transport (for in-memory binding)
	if features.IsEnabled("transport.inmemory") {
		if err := a.transport.RegisterScheduler(a.scheduler); err != nil {
			return err
		}
		if err := a.transport.RegisterWorker(a.worker); err != nil {
			return err
		}
		if err := a.transport.RegisterTaskManager(a.taskManager); err != nil {
			return err
		}
		if err := a.transport.RegisterMetricAgent(a.metricAgent); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) Run(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("app already running")
	}
	a.running = true
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.running = false
		a.mu.Unlock()
	}()

	// Connect transport
	if err := a.transport.Connect(ctx); err != nil {
		return err
	}
	defer a.transport.Close()

	// Start services
	if err := a.startServices(ctx); err != nil {
		return err
	}

	// Start API server
	if err := a.apiServer.Start(ctx); err != nil {
		return err
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return a.Shutdown(shutdownCtx)
}

func (a *App) startServices(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	a.cancelFuncs = append(a.cancelFuncs, cancel)

	// Start scheduler
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.scheduler.Start(ctx); err != nil {
			// Log error
		}
	}()

	// Start worker
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.worker.Start(ctx); err != nil {
			// Log error
		}
	}()

	// Start metric agent
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.metricAgent.Start(ctx); err != nil {
			// Log error
		}
	}()

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	// Cancel all running goroutines
	for _, cancel := range a.cancelFuncs {
		cancel()
	}

	// Wait for goroutines to finish
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *App) GetConfig() *config.Config {
	return a.config
}

func (a *App) GetFeatures() map[string]*features.FeatureFlag {
	return features.GetAll()
}
