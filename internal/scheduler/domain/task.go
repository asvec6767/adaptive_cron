package domain

import (
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

type DelayedTask struct {
	TaskID     string
	DelayUntil time.Time
	Attempt    int
}

type ScheduleResult struct {
	Scheduled bool
	JobID     string
	Reason    string
	Delay     time.Duration
}
