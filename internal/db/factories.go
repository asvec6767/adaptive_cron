package queue

import (
	"fmt"
	"sync"
)

type MemoryQueueFactory struct{}

func (f *MemoryQueueFactory) Create(config Config) (Queue, error) {
	return NewMemoryQueue(config.MaxSize), nil
}

type MemoryQueue struct {
	mu      sync.Mutex
	tasks   []*Task
	maxSize int
	cond    *sync.Cond
	closed  bool
}

func NewMemoryQueue(maxSize int) *MemoryQueue {
	q := &MemoryQueue{
		tasks:   make([]*Task, 0, maxSize),
		maxSize: maxSize,
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *MemoryQueue) Push(task *Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	if len(q.tasks) >= q.maxSize {
		return fmt.Errorf("queue is full")
	}

	q.tasks = append(q.tasks, task)
	q.cond.Signal()
	return nil
}

func (q *MemoryQueue) Pop() (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.tasks) == 0 && !q.closed {
		q.cond.Wait()
	}

	if q.closed && len(q.tasks) == 0 {
		return nil, fmt.Errorf("queue is closed")
	}

	task := q.tasks[0]
	q.tasks = q.tasks[1:]
	return task, nil
}

func (q *MemoryQueue) Peek() (*Task, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.tasks) == 0 {
		return nil, nil
	}
	return q.tasks[0], nil
}

func (q *MemoryQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tasks)
}

func (q *MemoryQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.cond.Broadcast()
	return nil
}

// RedisQueueFactory (будет использоваться когда включен feature queue.redis)
type RedisQueueFactory struct{}

func (f *RedisQueueFactory) Create(config Config) (Queue, error) {
	// Реализация с Redis
	return nil, nil
}

// NATSQueueFactory (для enterprise версии)
type NATSQueueFactory struct{}

func (f *NATSQueueFactory) Create(config Config) (Queue, error) {
	// Реализация с NATS
	return nil, nil
}
