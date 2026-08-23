package domain

import "errors"

var (
	ErrJobNotFound       = errors.New("job not found")
	ErrJobAlreadyRunning = errors.New("job already running")
	ErrWorkerBusy        = errors.New("worker is busy")
	ErrCommandTimeout    = errors.New("command execution timeout")
	ErrCommandFailed     = errors.New("command execution failed")
)
