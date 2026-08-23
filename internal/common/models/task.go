package models

import (
	"time"
)

type Task struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name           string    `json:"name" gorm:"uniqueIndex;not null"`
	Command        string    `json:"command" gorm:"not null"`
	Schedule       string    `json:"schedule" gorm:"not null"`
	IsCritical     bool      `json:"is_critical" gorm:"default:false"`
	MaxLoadScore   float64   `json:"max_load_score" gorm:"default:0.7"`
	MaxCPUPercent  float64   `json:"max_cpu_percent" gorm:"default:80"`
	MaxMemoryMB    int64     `json:"max_memory_mb" gorm:"default:1024"`
	TimeoutSeconds int       `json:"timeout_seconds" gorm:"default:300"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateTaskRequest struct {
	Name           string  `json:"name" validate:"required"`
	Command        string  `json:"command" validate:"required"`
	Schedule       string  `json:"schedule" validate:"required"`
	IsCritical     bool    `json:"is_critical"`
	MaxLoadScore   float64 `json:"max_load_score"`
	MaxCPUPercent  float64 `json:"max_cpu_percent"`
	MaxMemoryMB    int64   `json:"max_memory_mb"`
	TimeoutSeconds int     `json:"timeout_seconds"`
}

type UpdateTaskRequest struct {
	Name           *string  `json:"name"`
	Command        *string  `json:"command"`
	Schedule       *string  `json:"schedule"`
	IsCritical     *bool    `json:"is_critical"`
	MaxLoadScore   *float64 `json:"max_load_score"`
	MaxCPUPercent  *float64 `json:"max_cpu_percent"`
	MaxMemoryMB    *int64   `json:"max_memory_mb"`
	TimeoutSeconds *int     `json:"timeout_seconds"`
	IsActive       *bool    `json:"is_active"`
}

type TaskFilter struct {
	Active   bool   `json:"active,omitempty"`
	Critical bool   `json:"critical,omitempty"`
	Name     string `json:"name,omitempty"`
}

type DelayedTask struct {
	TaskID     string    `json:"task_id"`
	DelayUntil time.Time `json:"delay_until"`
	Attempt    int       `json:"attempt"`
}
