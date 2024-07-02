package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"spawn.dev/pkg/capability"
	"spawn.dev/pkg/llm"
)

// Supervisor is an in-memory manager implementation.
type Supervisor struct {
	mu     sync.RWMutex
	agents map[string]*Agent
	watch  chan Event
}

// NewSupervisor creates a new supervisor.
func NewSupervisor() *Supervisor {
	return &Supervisor{
		agents: make(map[string]*Agent),
		watch:  make(chan Event, 128),
	}
}

// Create registers an agent from config.
func (s *Supervisor) Create(_ context.Context, config *AgentConfig) (*Agent, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}
	a := &Agent{
		ID:           uuid.NewString(),
		Name:         config.Metadata.Name,
		Namespace:    config.Metadata.Namespace,
		Config:       config,
		State:        StateInitializing,
		Capabilities: make(map[string]capability.Capability),
		Inbox:        make(chan Message, 32),
		Outbox:       make(chan Message, 32),
	}
	s.mu.Lock()
	s.agents[a.ID] = a
	s.mu.Unlock()
	s.emit("created", a.ID)
	return a, nil
}

// Start marks an agent running.
func (s *Supervisor) Start(_ context.Context, id string) error {
	a, err := s.Get(context.Background(), id)
	if err != nil {
		return err
	}
	a.State = StateRunning
	a.StartedAt = time.Now().UTC()
	s.emit("started", id)
	return nil
}

// Stop marks an agent terminated.
func (s *Supervisor) Stop(_ context.Context, id string) error {
	a, err := s.Get(context.Background(), id)
	if err != nil {
		return err
	}
	a.State = StateTerminated
	if a.Context != nil {
		a.Context.Cancel()
	}
	s.emit("stopped", id)
	return nil
}

// Restart stop/starts an agent.
func (s *Supervisor) Restart(ctx context.Context, id string) error {
	if err := s.Stop(ctx, id); err != nil {
		return err
	}
	return s.Start(ctx, id)
}

// Delete deletes an agent.
func (s *Supervisor) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.agents[id]; !ok {
		return fmt.Errorf("delete agent: %s not found", id)
	}
	delete(s.agents, id)
	s.emit("deleted", id)
	return nil
}

// Get returns one agent.
func (s *Supervisor) Get(_ context.Context, id string) (*Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.agents[id]
	if !ok {
		return nil, fmt.Errorf("get agent: %s not found", id)
	}
	return a, nil
}

// List returns agents with optional namespace filter.
func (s *Supervisor) List(_ context.Context, opts ListOptions) ([]*Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Agent, 0, len(s.agents))
	for _, a := range s.agents {
		if opts.Namespace != "" && a.Namespace != opts.Namespace {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}

// SendMessage sends message into agent inbox.
func (s *Supervisor) SendMessage(_ context.Context, id string, msg Message) error {
	a, err := s.Get(context.Background(), id)
	if err != nil {
		return err
	}
	select {
	case a.Inbox <- msg:
		return nil
	default:
		return fmt.Errorf("send message: inbox full")
	}
}

// Execute executes a task using current LLM provider.
func (s *Supervisor) Execute(ctx context.Context, id string, task Task) (*TaskResult, error) {
	a, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	result := &TaskResult{TaskID: task.ID}
	if a.LLM == nil {
		result.Error = "no llm provider configured"
		result.Duration = time.Since(start)
		return result, nil
	}
	resp, err := a.LLM.Chat(ctx, &llm.ChatRequest{
		Model:    a.Config.Spec.Model.Name,
		Messages: []llm.Message{{Role: "user", Content: task.Prompt}},
	})
	if err != nil {
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result, nil
	}
	result.Output = resp.Content
	result.Duration = time.Since(start)
	a.TasksRun++
	if resp.Usage != nil {
		a.TokensUsed += int64(resp.Usage.InputTokens + resp.Usage.OutputTokens)
	}
	return result, nil
}

// Logs streams synthetic logs for now.
func (s *Supervisor) Logs(_ context.Context, id string, _ LogOptions) (<-chan LogEntry, error) {
	_, err := s.Get(context.Background(), id)
	if err != nil {
		return nil, err
	}
	ch := make(chan LogEntry, 1)
	ch <- LogEntry{Time: time.Now().UTC(), Level: "info", Message: "log stream started"}
	close(ch)
	return ch, nil
}

// Metrics returns aggregated metrics.
func (s *Supervisor) Metrics(_ context.Context, id string) (*AgentMetrics, error) {
	a, err := s.Get(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return &AgentMetrics{TokensUsed: a.TokensUsed, CostUSD: a.CostUSD, TasksRun: a.TasksRun}, nil
}

// Watch streams lifecycle events.
func (s *Supervisor) Watch(ctx context.Context, _ WatchOptions) (<-chan Event, error) {
	out := make(chan Event, 128)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-s.watch:
				out <- ev
			}
		}
	}()
	return out, nil
}

func (s *Supervisor) emit(eventType, agentID string) {
	select {
	case s.watch <- Event{Type: eventType, AgentID: agentID, Timestamp: time.Now().UTC()}:
	default:
	}
}
