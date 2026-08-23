package domain

import "errors"

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrTaskAlreadyExists = errors.New("task with this name already exists")
	ErrInvalidTaskData   = errors.New("invalid task data")
)
