package domain

import (
	"adaptive-cron/internal/common/models"
	"time"
)

type SystemMetrics struct {
	Timestamp     time.Time `json:"timestamp"`
	CPUPercent    float64   `json:"cpu_percent"`
	LoadAvg1      float64   `json:"load_avg1"`
	LoadAvg5      float64   `json:"load_avg5"`
	LoadAvg15     float64   `json:"load_avg15"`
	MemoryUsed    uint64    `json:"memory_used"`
	MemoryTotal   uint64    `json:"memory_total"`
	MemoryPercent float64   `json:"memory_percent"`
	SwapUsed      uint64    `json:"swap_used"`
	SwapTotal     uint64    `json:"swap_total"`
	SwapPercent   float64   `json:"swap_percent"`
}

func (m *SystemMetrics) ToModel() *models.SystemMetrics {
	return &models.SystemMetrics{
		Timestamp:     m.Timestamp,
		CPUPercent:    m.CPUPercent,
		LoadAvg1:      m.LoadAvg1,
		LoadAvg5:      m.LoadAvg5,
		LoadAvg15:     m.LoadAvg15,
		MemoryUsed:    m.MemoryUsed,
		MemoryTotal:   m.MemoryTotal,
		MemoryPercent: m.MemoryPercent,
		SwapUsed:      m.SwapUsed,
		SwapTotal:     m.SwapTotal,
		SwapPercent:   m.SwapPercent,
	}
}

func MetricsFromModel(m *models.SystemMetrics) *SystemMetrics {
	return &SystemMetrics{
		Timestamp:     m.Timestamp,
		CPUPercent:    m.CPUPercent,
		LoadAvg1:      m.LoadAvg1,
		LoadAvg5:      m.LoadAvg5,
		LoadAvg15:     m.LoadAvg15,
		MemoryUsed:    m.MemoryUsed,
		MemoryTotal:   m.MemoryTotal,
		MemoryPercent: m.MemoryPercent,
		SwapUsed:      m.SwapUsed,
		SwapTotal:     m.SwapTotal,
		SwapPercent:   m.SwapPercent,
	}
}

func (m *SystemMetrics) GetLoadScore(numCores int) float64 {
	cpuScore := m.CPUPercent / 100.0
	memScore := m.MemoryPercent / 100.0
	loadScore := m.LoadAvg1 / float64(numCores)
	return (cpuScore * 0.4) + (memScore * 0.3) + (loadScore * 0.3)
}

type MetricsHistory struct {
	Metrics []*SystemMetrics `json:"metrics"`
	Total   int              `json:"total"`
}
