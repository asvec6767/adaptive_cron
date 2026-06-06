package models

import (
    "time"
    "fmt"
)

type SystemMetrics struct {
    Timestamp    time.Time `json:"timestamp"`
    CPUPercent   float64   `json:"cpu_percent"`
    LoadAvg1     float64   `json:"load_avg1"`
    LoadAvg5     float64   `json:"load_avg5"`
    LoadAvg15    float64   `json:"load_avg15"`
    MemoryUsed   uint64    `json:"memory_used"`
    MemoryTotal  uint64    `json:"memory_total"`
    MemoryPercent float64  `json:"memory_percent"`
    SwapUsed     uint64    `json:"swap_used"`
    SwapTotal    uint64    `json:"swap_total"`
    SwapPercent  float64   `json:"swap_percent"`
}

type MetricsHistory struct {
    Metrics []SystemMetrics `json:"metrics"`
}

func (m *SystemMetrics) String() string {
    return fmt.Sprintf("CPU: %.2f%%, Memory: %.2f%%, Load: %.2f", 
        m.CPUPercent, m.MemoryPercent, m.LoadAvg1)
}

func (m *SystemMetrics) GetLoadScore() float64 {
    // Комбинированный показатель нагрузки (0-1)
    cpuScore := m.CPUPercent / 100.0
    memScore := m.MemoryPercent / 100.0
    loadScore := m.LoadAvg1 / float64(8) // Предполагаем 8 ядер
    
    // Взвешенная сумма
    return (cpuScore*0.4 + memScore*0.3 + loadScore*0.3)
}