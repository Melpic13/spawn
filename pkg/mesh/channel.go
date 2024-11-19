package mesh

import (
	"context"
	"fmt"
	"sync"
)

// MessageHandler handles received messages.
type MessageHandler func(context.Context, *Message) error

// Subscription is a cancelable subscription.
type Subscription interface {
	Unsubscribe() error
}

type subscription struct {
	onClose func()
}

func (s *subscription) Unsubscribe() error {
	if s.onClose != nil {
		s.onClose()
	}
	return nil
}

// ChannelConfig configures a mesh channel.
type ChannelConfig struct {
	Name string
	Type ChannelType
}

type inMemoryChannel struct {
	name     string
	channelT ChannelType
	ch       chan *Message
	mu       sync.RWMutex
	handlers map[int]MessageHandler
	nextID   int
}

func newInMemoryChannel(cfg *ChannelConfig) *inMemoryChannel {
	c := &inMemoryChannel{
		name:     cfg.Name,
		channelT: cfg.Type,
		ch:       make(chan *Message, 256),
		handlers: map[int]MessageHandler{},
	}
	go c.fanout()
	return c
}

func (c *inMemoryChannel) Name() string      { return c.name }
func (c *inMemoryChannel) Type() ChannelType { return c.channelT }

func (c *inMemoryChannel) Send(_ context.Context, msg *Message) error {
	if msg == nil {
		return fmt.Errorf("send channel message: nil message")
	}
	c.ch <- msg
	return nil
}

func (c *inMemoryChannel) Receive(_ context.Context) (*Message, error) {
	msg := <-c.ch
	return msg, nil
}

func (c *inMemoryChannel) Subscribe(handler MessageHandler) (Subscription, error) {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	c.handlers[id] = handler
	c.mu.Unlock()
	return &subscription{onClose: func() {
		c.mu.Lock()
		delete(c.handlers, id)
		c.mu.Unlock()
	}}, nil
}

func (c *inMemoryChannel) Close() error {
	close(c.ch)
	return nil
}

func (c *inMemoryChannel) fanout() {
	for msg := range c.ch {
		c.mu.RLock()
		for _, h := range c.handlers {
			_ = h(context.Background(), msg)
		}
		c.mu.RUnlock()
	}
}
