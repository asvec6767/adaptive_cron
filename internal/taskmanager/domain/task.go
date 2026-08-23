package domain

import (
	"adaptive-cron/internal/common/errors"
	"adaptive-cron/internal/common/models"
	"time"
)

type Task struct {
	ID             string
	Name           string
	Command        string
	Schedule       string
	IsCritical     bool
	MaxLoadScore   float64
	MaxCPUPercent  float64
	MaxMemoryMB    int64
	TimeoutSeconds int
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (t *Task) Validate() error {
	if t.Name == "" {
		return errors.ErrEmptyName
	}
	if t.Command == "" {
		return errors.ErrEmptyCommand
	}
	if t.Schedule == "" {
		return errors.ErrEmptySchedule
	}
	if t.MaxLoadScore < 0 || t.MaxLoadScore > 1 {
		return errors.ErrInvalidLoadScore
	}
	if t.MaxCPUPercent < 0 || t.MaxCPUPercent > 100 {
		return errors.ErrInvalidCPUPercent
	}
	if t.MaxMemoryMB < 0 {
		return errors.ErrInvalidMemoryMB
	}
	return nil
}

func (t *Task) ToModel() *models.Task {
	return &models.Task{
		ID:             t.ID,
		Name:           t.Name,
		Command:        t.Command,
		Schedule:       t.Schedule,
		IsCritical:     t.IsCritical,
		MaxLoadScore:   t.MaxLoadScore,
		MaxCPUPercent:  t.MaxCPUPercent,
		MaxMemoryMB:    t.MaxMemoryMB,
		TimeoutSeconds: t.TimeoutSeconds,
		IsActive:       t.IsActive,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

func TaskFromModel(m *models.Task) *Task {
	return &Task{
		ID:             m.ID,
		Name:           m.Name,
		Command:        m.Command,
		Schedule:       m.Schedule,
		IsCritical:     m.IsCritical,
		MaxLoadScore:   m.MaxLoadScore,
		MaxCPUPercent:  m.MaxCPUPercent,
		MaxMemoryMB:    m.MaxMemoryMB,
		TimeoutSeconds: m.TimeoutSeconds,
		IsActive:       m.IsActive,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

type CreateTaskRequest struct {
	Name           string  `json:"name"`
	Command        string  `json:"command"`
	Schedule       string  `json:"schedule"`
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
	Active   bool
	Critical bool
	Name     string
}
