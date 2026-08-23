package domain

import (
	"time"
)

type Job struct {
	ID        string
	TaskID    string
	Command   string
	Timeout   time.Duration
	StartTime time.Time
	EndTime   time.Time
	ExitCode  int
	Output    string
	Error     string
	Status    JobStatus
}

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusTimeout   JobStatus = "timeout"
)

type ExecuteRequest struct {
	TaskID  string
	Command string
	Timeout int
}

type ExecuteResponse struct {
	JobID string
}

type JobStatusResponse struct {
	JobID     string
	TaskID    string
	Status    JobStatus
	StartTime time.Time
	EndTime   time.Time
	ExitCode  int
	Output    string
	Error     string
}

type WorkerStats struct {
	ActiveJobs    int
	CompletedJobs int
	FailedJobs    int
	TotalJobs     int
}
