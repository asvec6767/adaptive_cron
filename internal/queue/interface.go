package queue

import "time"

type Queue interface {
	Push(task *Task) error
	Pop() (*Task, error)
	Peek() (*Task, error)
	Size() int
	Close() error
}

type DelayedQueue interface {
	Queue
	PushDelayed(task *Task, delay time.Duration) error
	GetDelayed() ([]*Task, error)
	RemoveDelayed(taskID string) error
}
