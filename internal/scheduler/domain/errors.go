package domain

import "errors"

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskNotActive      = errors.New("task is not active")
	ErrInvalidSchedule    = errors.New("invalid schedule format")
	ErrQueueFull          = errors.New("delayed queue is full")
	ErrMetricsUnavailable = errors.New("metrics are unavailable")
	ErrWorkerUnavailable  = errors.New("worker service is unavailable")
)
