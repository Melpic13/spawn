package scheduler

import (
	"fmt"
	"sync"
	"time"
)

// Scheduler is an in-memory priority scheduler.
type Scheduler struct {
	mu      sync.Mutex
	queue   taskHeap
	policy  BackpressurePolicy
	metrics Metrics
}

// Metrics expose scheduler counters.
type Metrics struct {
	Queued    int64
	Dequeued  int64
	Rejected  int64
	QueueSize int
}

// New creates a scheduler.
func New(policy BackpressurePolicy) *Scheduler {
	return &Scheduler{policy: policy}
}

// Enqueue pushes a task into queue.
func (s *Scheduler) Enqueue(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if task == nil {
		return fmt.Errorf("enqueue task: nil task")
	}
	if task.ETA.IsZero() {
		task.ETA = time.Now().UTC()
	}
	if s.policy.MaxQueueDepth > 0 && len(s.queue) >= s.policy.MaxQueueDepth {
		s.metrics.Rejected++
		if s.policy.DropWhenFull {
			return nil
		}
		return fmt.Errorf("enqueue task: queue is full")
	}
	s.queue.pushTask(task)
	s.metrics.Queued++
	s.metrics.QueueSize = len(s.queue)
	return nil
}

// Dequeue pops highest-priority ready task.
func (s *Scheduler) Dequeue() (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.queue) == 0 {
		return nil, fmt.Errorf("dequeue task: queue empty")
	}
	t := s.queue.popTask()
	s.metrics.Dequeued++
	s.metrics.QueueSize = len(s.queue)
	return t, nil
}

// Metrics returns current scheduler metrics.
func (s *Scheduler) Metrics() Metrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.metrics
}
