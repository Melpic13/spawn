package mesh

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryMeshSend(t *testing.T) {
	t.Parallel()
	m := NewInMemoryMesh()
	received := make(chan struct{}, 1)
	_, err := m.Subscribe(context.Background(), "topic-a", func(_ context.Context, _ *Message) error {
		select {
		case received <- struct{}{}:
		default:
		}
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if err := m.Send(context.Background(), &Message{Topic: "topic-a", Type: MessageTypeEvent}); err != nil {
		t.Fatalf("send: %v", err)
	}
	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatalf("expected handler to receive message")
	}
}
