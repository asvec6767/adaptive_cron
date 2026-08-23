package usecase

import (
	"adaptive-cron/internal/common/models"
	"adaptive-cron/internal/scheduler/domain"
	"context"
	"time"
)

type MetricProvider interface {
	GetCurrentMetrics(ctx context.Context) (*models.SystemMetrics, error)
	GetLoadScore(ctx context.Context) (float64, error)
}

type AdaptiveChecker struct {
	metricProvider MetricProvider
	numCores       int
}

func NewAdaptiveChecker(metricProvider MetricProvider, numCores int) *AdaptiveChecker {
	return &AdaptiveChecker{
		metricProvider: metricProvider,
		numCores:       numCores,
	}
}

func (c *AdaptiveChecker) CanRun(ctx context.Context, task *domain.Task) (bool, time.Duration, string) {
	if task.IsCritical {
		return true, 0, "critical_task"
	}

	metrics, err := c.metricProvider.GetCurrentMetrics(ctx)
	if err != nil {
		return true, 0, "metrics_unavailable"
	}

	if metrics.CPUPercent > task.MaxCPUPercent {
		return false, 30 * time.Second, "cpu_high"
	}

	memoryUsedMB := int64(metrics.MemoryUsed / 1024 / 1024)
	if memoryUsedMB > task.MaxMemoryMB {
		return false, 30 * time.Second, "memory_high"
	}

	loadScore := metrics.GetLoadScore(c.numCores)
	if loadScore > task.MaxLoadScore {
		delay := time.Duration(loadScore*60) * time.Second
		return false, delay, "load_score_high"
	}

	return true, 0, "resources_ok"
}

func (c *AdaptiveChecker) CalculateBackoff(attempt int) time.Duration {
	baseDelay := 30 * time.Second
	maxDelay := 1 * time.Hour

	delay := baseDelay * time.Duration(1<<uint(attempt))
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}
