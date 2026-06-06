package errors

import "errors"

var (
    // Общие ошибки
    ErrNotFound      = errors.New("not found")
    ErrInvalidInput  = errors.New("invalid input")
    ErrInternalError = errors.New("internal error")
    
    // Ошибки задач
    ErrEmptyName       = errors.New("task name cannot be empty")
    ErrEmptyCommand    = errors.New("task command cannot be empty")
    ErrEmptySchedule   = errors.New("task schedule cannot be empty")
    ErrInvalidLoadScore = errors.New("load score must be between 0 and 1")
    ErrInvalidCPUPercent = errors.New("cpu percent must be between 0 and 100")
    ErrInvalidMemoryMB   = errors.New("memory mb must be greater than 0")
    ErrTaskNotFound      = errors.New("task not found")
    ErrTaskAlreadyExists = errors.New("task with this name already exists")
    
    // Ошибки метрик
    ErrMetricsUnavailable = errors.New("metrics are unavailable")
    ErrMetricsTimeout     = errors.New("metrics collection timeout")
    
    // Ошибки выполнения
    ErrExecutionTimeout = errors.New("execution timeout")
    ErrCommandFailed    = errors.New("command execution failed")
    
    // Ошибки сервисов
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