package bus

import (
	"context"
	"testing"
	"time"
)

func TestMessageBus_InboundRoundTrip(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	msg := InboundMessage{
		UserID:  1,
		Channel: "web",
		Content: "hello",
	}

	go func() {
		mb.PublishInbound(msg)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, ok := mb.ConsumeInbound(ctx)
	if !ok {
		t.Fatal("ConsumeInbound returned false; expected message")
	}
	if got.Content != "hello" {
		t.Errorf("Content = %q; want \"hello\"", got.Content)
	}
	if got.UserID != 1 {
		t.Errorf("UserID = %d; want 1", got.UserID)
	}
}

func TestMessageBus_OutboundRoundTrip(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	msg := OutboundMessage{
		UserID:  2,
		Channel: "slack",
		Content: "world",
	}

	go func() {
		mb.PublishOutbound(msg)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, ok := mb.SubscribeOutbound(ctx)
	if !ok {
		t.Fatal("SubscribeOutbound returned false; expected message")
	}
	if got.Content != "world" {
		t.Errorf("Content = %q; want \"world\"", got.Content)
	}
}

func TestMessageBus_RegisterHandler(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	handler := func(msg InboundMessage) error { return nil }
	mb.RegisterHandler("slack", handler)

	h, ok := mb.GetHandler("slack")
	if !ok {
		t.Fatal("GetHandler(\"slack\") returned false; expected handler")
	}
	if h == nil {
		t.Fatal("Handler is nil")
	}
}

func TestMessageBus_GetHandler_Missing(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	_, ok := mb.GetHandler("nonexistent")
	if ok {
		t.Error("GetHandler should return false for unregistered channel")
	}
}

func TestMessageBus_ConsumeAfterClose(t *testing.T) {
	mb := NewMessageBus()
	mb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, ok := mb.ConsumeInbound(ctx)
	if ok {
		t.Error("ConsumeInbound on closed bus should return false")
	}
}

func TestMessageBus_SubscribeAfterClose(t *testing.T) {
	mb := NewMessageBus()
	mb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, ok := mb.SubscribeOutbound(ctx)
	if ok {
		t.Error("SubscribeOutbound on closed bus should return false")
	}
}

func TestMessageBus_DoubleClose(t *testing.T) {
	mb := NewMessageBus()
	mb.Close()
	// Second close should not panic thanks to sync.Once
	mb.Close()
}

func TestMessageBus_ConsumeContextCancel(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, ok := mb.ConsumeInbound(ctx)
	if ok {
		t.Error("ConsumeInbound with cancelled context should return false")
	}
}

func TestMessageBus_MultipleMessages(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	// Publish 3 messages
	for i := 0; i < 3; i++ {
		mb.PublishInbound(InboundMessage{Content: "msg"})
	}

	// Consume all 3
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for i := 0; i < 3; i++ {
		_, ok := mb.ConsumeInbound(ctx)
		if !ok {
			t.Fatalf("ConsumeInbound returned false on message %d", i)
		}
	}
}
