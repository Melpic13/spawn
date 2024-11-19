package mesh

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InMemoryMesh is a local mesh implementation.
type InMemoryMesh struct {
	mu       sync.RWMutex
	agents   map[string]*AgentInfo
	channels map[string]Channel
}

// NewInMemoryMesh creates an in-memory mesh.
func NewInMemoryMesh() *InMemoryMesh {
	return &InMemoryMesh{
		agents:   map[string]*AgentInfo{},
		channels: map[string]Channel{},
	}
}

func (m *InMemoryMesh) Register(_ context.Context, agent *AgentInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[agent.ID] = agent
	return nil
}

func (m *InMemoryMesh) Deregister(_ context.Context, agentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.agents, agentID)
	return nil
}

func (m *InMemoryMesh) Discover(_ context.Context, query *DiscoveryQuery) ([]*AgentInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*AgentInfo, 0, len(m.agents))
	for _, a := range m.agents {
		if query != nil {
			if query.Namespace != "" && a.Namespace != query.Namespace {
				continue
			}
			if query.Healthy != nil && a.Healthy != *query.Healthy {
				continue
			}
		}
		out = append(out, a)
	}
	return out, nil
}

func (m *InMemoryMesh) Send(ctx context.Context, msg *Message) error {
	if msg == nil {
		return fmt.Errorf("send mesh message: nil message")
	}
	ch, err := m.GetChannel(ctx, msg.Topic)
	if err != nil {
		return err
	}
	if msg.ID == "" {
		msg.ID = uuid.NewString()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}
	return ch.Send(ctx, msg)
}

func (m *InMemoryMesh) Request(ctx context.Context, msg *Message, timeout time.Duration) (*Message, error) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if err := m.Send(ctx, msg); err != nil {
		return nil, err
	}
	return &Message{ID: uuid.NewString(), Type: MessageTypeReply, CorrelationID: msg.ID, Timestamp: time.Now().UTC()}, nil
}

func (m *InMemoryMesh) Subscribe(ctx context.Context, topic string, handler MessageHandler) (Subscription, error) {
	ch, err := m.GetChannel(ctx, topic)
	if err != nil {
		return nil, err
	}
	return ch.Subscribe(handler)
}

func (m *InMemoryMesh) CreateChannel(_ context.Context, config *ChannelConfig) (Channel, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if config == nil || config.Name == "" {
		return nil, fmt.Errorf("create channel: name is required")
	}
	ch := newInMemoryChannel(config)
	m.channels[config.Name] = ch
	return ch, nil
}

func (m *InMemoryMesh) GetChannel(ctx context.Context, name string) (Channel, error) {
	m.mu.RLock()
	ch, ok := m.channels[name]
	m.mu.RUnlock()
	if ok {
		return ch, nil
	}
	return m.CreateChannel(ctx, &ChannelConfig{Name: name, Type: ChannelPubSub})
}

func (m *InMemoryMesh) Topology(_ context.Context) (*TopologyGraph, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	graph := &TopologyGraph{}
	for _, agent := range m.agents {
		graph.Agents = append(graph.Agents, &AgentNode{ID: agent.ID, Name: agent.Name, Namespace: agent.Namespace})
	}
	for name, ch := range m.channels {
		graph.Channels = append(graph.Channels, &ChannelEdge{Name: name, Type: ch.Type()})
	}
	return graph, nil
}
