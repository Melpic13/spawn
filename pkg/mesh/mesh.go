package mesh

import (
	"context"
	"time"
)

// Mesh coordinates multi-agent communication.
type Mesh interface {
	Register(ctx context.Context, agent *AgentInfo) error
	Deregister(ctx context.Context, agentID string) error
	Discover(ctx context.Context, query *DiscoveryQuery) ([]*AgentInfo, error)
	Send(ctx context.Context, msg *Message) error
	Request(ctx context.Context, msg *Message, timeout time.Duration) (*Message, error)
	Subscribe(ctx context.Context, topic string, handler MessageHandler) (Subscription, error)
	CreateChannel(ctx context.Context, config *ChannelConfig) (Channel, error)
	GetChannel(ctx context.Context, name string) (Channel, error)
	Topology(ctx context.Context) (*TopologyGraph, error)
}

// Channel represents a communication channel between agents.
type Channel interface {
	Name() string
	Type() ChannelType
	Send(ctx context.Context, msg *Message) error
	Receive(ctx context.Context) (*Message, error)
	Subscribe(handler MessageHandler) (Subscription, error)
	Close() error
}

// ChannelType identifies channel behavior.
type ChannelType string

const (
	ChannelPubSub       ChannelType = "pubsub"
	ChannelRequestReply ChannelType = "request-reply"
	ChannelStream       ChannelType = "stream"
	ChannelBroadcast    ChannelType = "broadcast"
)

// MessageType identifies message semantics.
type MessageType string

const (
	MessageTypeEvent   MessageType = "event"
	MessageTypeRequest MessageType = "request"
	MessageTypeReply   MessageType = "reply"
)

// Message represents an inter-agent message.
type Message struct {
	ID            string            `json:"id"`
	From          string            `json:"from"`
	To            string            `json:"to,omitempty"`
	Topic         string            `json:"topic,omitempty"`
	Type          MessageType       `json:"type"`
	Payload       interface{}       `json:"payload"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	ReplyTo       string            `json:"replyTo,omitempty"`
	CorrelationID string            `json:"correlationId,omitempty"`
}

// TopologyGraph represents the mesh topology.
type TopologyGraph struct {
	Agents      []*AgentNode   `json:"agents"`
	Channels    []*ChannelEdge `json:"channels"`
	Connections []*Connection  `json:"connections"`
}
