package observability

import (
	"sync"
	"time"
)

// Event is an observable event record.
type Event struct {
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventStream is an in-memory pub/sub event stream.
type EventStream struct {
	mu   sync.RWMutex
	subs map[int]chan Event
	next int
}

// NewEventStream creates stream.
func NewEventStream() *EventStream {
	return &EventStream{subs: map[int]chan Event{}}
}

// Publish publishes an event to subscribers.
func (s *EventStream) Publish(ev Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

// Subscribe subscribes and returns channel + cancel fn.
func (s *EventStream) Subscribe() (<-chan Event, func()) {
	s.mu.Lock()
	id := s.next
	s.next++
	ch := make(chan Event, 64)
	s.subs[id] = ch
	s.mu.Unlock()

	cancel := func() {
		s.mu.Lock()
		if c, ok := s.subs[id]; ok {
			delete(s.subs, id)
			close(c)
		}
		s.mu.Unlock()
	}
	return ch, cancel
}
