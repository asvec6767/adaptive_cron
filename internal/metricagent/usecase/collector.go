package usecase

import (
	"adaptive-cron/internal/metricagent/domain"
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

type MetricCollector struct {
	mu              sync.RWMutex
	currentMetrics  *domain.SystemMetrics
	history         []*domain.SystemMetrics
	maxHistorySize  int
	numCores        int
	collectInterval time.Duration
	running         bool
	stopCh          chan struct{}
}

func NewMetricCollector(collectInterval time.Duration, maxHistorySize int) *MetricCollector {
	return &MetricCollector{
		history:         make([]*domain.SystemMetrics, 0, maxHistorySize),
		maxHistorySize:  maxHistorySize,
		numCores:        runtime.NumCPU(),
		collectInterval: collectInterval,
		stopCh:          make(chan struct{}),
	}
}

func (c *MetricCollector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = true
	c.mu.Unlock()

	go c.collectLoop(ctx)
	return nil
}

func (c *MetricCollector) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.running {
		return nil
	}
	c.running = false
	close(c.stopCh)
	return nil
}

func (c *MetricCollector) collectLoop(ctx context.Context) {
	ticker := time.NewTicker(c.collectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.collectOnce()
		}
	}
}

func (c *MetricCollector) collectOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metrics := &domain.SystemMetrics{
		Timestamp: time.Now(),
	}

	// CPU
	cpuPercent, err := cpu.PercentWithContext(ctx, 0, false)
	if err == nil && len(cpuPercent) > 0 {
		metrics.CPUPercent = cpuPercent[0]
	}

	// Load Average
	loadAvg, err := load.AvgWithContext(ctx)
	if err == nil {
		metrics.LoadAvg1 = loadAvg.Load1
		metrics.LoadAvg5 = loadAvg.Load5
		metrics.LoadAvg15 = loadAvg.Load15
	}

	// Memory
	memInfo, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		metrics.MemoryUsed = memInfo.Used
		metrics.MemoryTotal = memInfo.Total
		metrics.MemoryPercent = memInfo.UsedPercent
	}

	// Swap
	swapInfo, err := mem.SwapMemoryWithContext(ctx)
	if err == nil {
		metrics.SwapUsed = swapInfo.Used
		metrics.SwapTotal = swapInfo.Total
		metrics.SwapPercent = swapInfo.UsedPercent
	}

	c.mu.Lock()
	c.currentMetrics = metrics
	c.history = append(c.history, metrics)
	if len(c.history) > c.maxHistorySize {
		c.history = c.history[1:]
	}
	c.mu.Unlock()
}

func (c *MetricCollector) GetCurrent(ctx context.Context) (*domain.SystemMetrics, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.currentMetrics == nil {
		return nil, domain.ErrMetricsUnavailable
	}
	return c.currentMetrics, nil
}

func (c *MetricCollector) GetHistory(ctx context.Context, limit int) ([]*domain.SystemMetrics, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.history) == 0 {
		return nil, domain.ErrHistoryEmpty
	}

	start := len(c.history) - limit
	if start < 0 {
		start = 0
	}

	return c.history[start:], nil
}

func (c *MetricCollector) GetLoadScore(ctx context.Context) (float64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.currentMetrics == nil {
		return 0, domain.ErrMetricsUnavailable
	}
	return c.currentMetrics.GetLoadScore(c.numCores), nil
}

func (c *MetricCollector) GetSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.history)
}
