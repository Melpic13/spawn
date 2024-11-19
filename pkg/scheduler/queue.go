package scheduler

import (
	"container/heap"
	"time"
)

// Task defines a scheduled task.
type Task struct {
	ID       string
	AgentID  string
	Priority int
	Payload  map[string]interface{}
	ETA      time.Time
	index    int
}

type taskHeap []*Task

func (h taskHeap) Len() int { return len(h) }
func (h taskHeap) Less(i, j int) bool {
	if h[i].Priority == h[j].Priority {
		return h[i].ETA.Before(h[j].ETA)
	}
	return h[i].Priority > h[j].Priority
}
func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}
func (h *taskHeap) Push(x interface{}) {
	item := x.(*Task)
	item.index = len(*h)
	*h = append(*h, item)
}
func (h *taskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return item
}

func (h *taskHeap) pushTask(task *Task) { heap.Push(h, task) }
func (h *taskHeap) popTask() *Task      { return heap.Pop(h).(*Task) }
