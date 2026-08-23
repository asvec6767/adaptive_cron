package models

import (
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

func (m *SystemMetrics) GetLoadScore(numCores int) float64 {
	cpuScore := m.CPUPercent / 100.0
	memScore := m.MemoryPercent / 100.0
	loadScore := m.LoadAvg1 / float64(numCores)
	return (cpuScore * 0.4) + (memScore * 0.3) + (loadScore * 0.3)
}
