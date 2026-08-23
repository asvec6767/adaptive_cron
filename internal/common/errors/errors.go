package errors

import "errors"

var (
	// Common errors
	ErrNotFound       = errors.New("not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrInternalError  = errors.New("internal error")
	ErrNotImplemented = errors.New("not implemented")

	// Task errors
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task with this name already exists")
	ErrEmptyName         = errors.New("task name cannot be empty")
	ErrEmptyCommand      = errors.New("task command cannot be empty")
	ErrEmptySchedule     = errors.New("task schedule cannot be empty")
	ErrInvalidLoadScore  = errors.New("load score must be between 0 and 1")
	ErrInvalidCPUPercent = errors.New("cpu percent must be between 0 and 100")
	ErrInvalidMemoryMB   = errors.New("memory mb must be greater than 0")
	ErrTaskNotActive     = errors.New("task is not active")

	// Job errors
	ErrJobNotFound       = errors.New("job not found")
	ErrJobAlreadyRunning = errors.New("job already running")
	ErrExecutionTimeout  = errors.New("execution timeout")
	ErrCommandFailed     = errors.New("command execution failed")

	// Metric errors
	ErrMetricsUnavailable  = errors.New("metrics unavailable")
	ErrCollectorNotRunning = errors.New("metric collector not running")

	// Queue errors
	ErrQueueFull   = errors.New("queue is full")
	ErrQueueEmpty  = errors.New("queue is empty")
	ErrQueueClosed = errors.New("queue is closed")

	// Service errors
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrServiceTimeout     = errors.New("service timeout")
)

type AppError struct {
	Err  error
	Msg  string
	Code int
}

func (e *AppError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.Err.Error()
}

func NewAppError(err error, msg string, code int) *AppError {
	return &AppError{
		Err:  err,
		Msg:  msg,
		Code: code,
	}
}
