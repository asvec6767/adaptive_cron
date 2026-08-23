package usecase

import (
	"adaptive-cron/internal/metricagent/domain"
	"adaptive-cron/internal/metricagent/repository"
	"context"
)

type MetricAgentUseCase struct {
	repo      repository.MetricRepository
	collector *MetricCollector
}

func NewMetricAgentUseCase(repo repository.MetricRepository, collector *MetricCollector) *MetricAgentUseCase {
	return &MetricAgentUseCase{
		repo:      repo,
		collector: collector,
	}
}

func (u *MetricAgentUseCase) GetCurrentMetrics(ctx context.Context) (*domain.SystemMetrics, error) {
	return u.collector.GetCurrent(ctx)
}

func (u *MetricAgentUseCase) GetMetricsHistory(ctx context.Context, limit int) ([]*domain.SystemMetrics, error) {
	return u.collector.GetHistory(ctx, limit)
}

func (u *MetricAgentUseCase) GetLoadScore(ctx context.Context) (float64, error) {
	return u.collector.GetLoadScore(ctx)
}

func (u *MetricAgentUseCase) Start(ctx context.Context) error {
	return u.collector.Start(ctx)
}

func (u *MetricAgentUseCase) Stop() error {
	return u.collector.Stop()
}
