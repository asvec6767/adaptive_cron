package models

import (
    "time"
    "encoding/json"
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

type JobLog struct {
    ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    TaskID         string    `json:"task_id" gorm:"index"`
    TaskName       string    `json:"task_name"`
    StartTime      time.Time `json:"start_time"`
    EndTime        time.Time `json:"end_time"`
    ExitCode       int       `json:"exit_code"`
    Output         string    `json:"output" gorm:"type:text"`
    Error          string    `json:"error" gorm:"type:text"`
    MetricsAtStart string    `json:"metrics_at_start" gorm:"type:jsonb"`
    CreatedAt      time.Time `json:"created_at"`
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

func (t *Task) Validate() error {
    if t.Name == "" {
        return ErrEmptyName
    }
    if t.Command == "" {
        return ErrEmptyCommand
    }
    if t.Schedule == "" {
        return ErrEmptySchedule
    }
    if t.MaxLoadScore < 0 || t.MaxLoadScore > 1 {
        return ErrInvalidLoadScore
    }
    if t.MaxCPUPercent < 0 || t.MaxCPUPercent > 100 {
        return ErrInvalidCPUPercent
    }
    if t.MaxMemoryMB < 0 {
        return ErrInvalidMemoryMB
    }
    return nil
}

func (j *JobLog) ToJSON() (string, error) {
    bytes, err := json.Marshal(j)
    if err != nil {
        return "", err
    }
    return string(bytes), nil
}