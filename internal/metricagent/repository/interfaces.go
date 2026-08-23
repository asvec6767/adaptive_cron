package repository

import (
	"adaptive-cron/internal/metricagent/domain"
	"context"
)

type MetricRepository interface {
	Save(ctx context.Context, metrics *domain.SystemMetrics) error
	GetCurrent(ctx context.Context) (*domain.SystemMetrics, error)
	GetHistory(ctx context.Context, limit int) ([]*domain.SystemMetrics, error)
	GetLoadScore(ctx context.Context) (float64, error)
	GetSize(ctx context.Context) int
}
