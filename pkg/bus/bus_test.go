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

// ---------------------------------------------------------------------------
// PublishInbound on closed bus should not block or panic
// ---------------------------------------------------------------------------

func TestMessageBus_PublishInboundAfterClose(t *testing.T) {
	mb := NewMessageBus()
	mb.Close()

	// Should not panic or block
	done := make(chan struct{})
	go func() {
		mb.PublishInbound(InboundMessage{Content: "after close"})
		close(done)
	}()

	select {
	case <-done:
		// success: did not block
	case <-time.After(time.Second):
		t.Fatal("PublishInbound on closed bus blocked")
	}
}

func TestMessageBus_PublishOutboundAfterClose(t *testing.T) {
	mb := NewMessageBus()
	mb.Close()

	done := make(chan struct{})
	go func() {
		mb.PublishOutbound(OutboundMessage{Content: "after close"})
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("PublishOutbound on closed bus blocked")
	}
}

// ---------------------------------------------------------------------------
// RegisterHandler overwrites existing handler
// ---------------------------------------------------------------------------

func TestMessageBus_RegisterHandler_Overwrites(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	called := false
	handler1 := func(msg InboundMessage) error { return nil }
	handler2 := func(msg InboundMessage) error {
		called = true
		return nil
	}

	mb.RegisterHandler("test", handler1)
	mb.RegisterHandler("test", handler2)

	h, ok := mb.GetHandler("test")
	if !ok {
		t.Fatal("GetHandler returned false after registering handler")
	}

	// Call the handler to verify it's the second one
	_ = h(InboundMessage{})
	if !called {
		t.Error("Expected second handler to be called, but it was not")
	}
}

// ---------------------------------------------------------------------------
// Inbound and outbound channels are independent
// ---------------------------------------------------------------------------

func TestMessageBus_InboundOutboundIndependent(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	// Publish to different channels
	mb.PublishInbound(InboundMessage{Content: "inbound"})
	mb.PublishOutbound(OutboundMessage{Content: "outbound"})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Consume inbound should get inbound message
	inMsg, ok := mb.ConsumeInbound(ctx)
	if !ok {
		t.Fatal("ConsumeInbound failed")
	}
	if inMsg.Content != "inbound" {
		t.Errorf("inbound content = %q; want %q", inMsg.Content, "inbound")
	}

	// Subscribe outbound should get outbound message
	outMsg, ok := mb.SubscribeOutbound(ctx)
	if !ok {
		t.Fatal("SubscribeOutbound failed")
	}
	if outMsg.Content != "outbound" {
		t.Errorf("outbound content = %q; want %q", outMsg.Content, "outbound")
	}
}

// ---------------------------------------------------------------------------
// Message field preservation
// ---------------------------------------------------------------------------

func TestMessageBus_InboundFieldPreservation(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	msg := InboundMessage{
		UserID:     42,
		Channel:    "telegram",
		SenderID:   "sender123",
		ChatID:     "chat456",
		Content:    "test message",
		Media:      []string{"photo.jpg"},
		SessionKey: "telegram:chat456",
		Metadata:   map[string]string{"key": "val"},
	}

	mb.PublishInbound(msg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, ok := mb.ConsumeInbound(ctx)
	if !ok {
		t.Fatal("ConsumeInbound failed")
	}

	if got.UserID != 42 {
		t.Errorf("UserID = %d; want 42", got.UserID)
	}
	if got.Channel != "telegram" {
		t.Errorf("Channel = %q; want %q", got.Channel, "telegram")
	}
	if got.SenderID != "sender123" {
		t.Errorf("SenderID = %q; want %q", got.SenderID, "sender123")
	}
	if got.ChatID != "chat456" {
		t.Errorf("ChatID = %q; want %q", got.ChatID, "chat456")
	}
	if got.Content != "test message" {
		t.Errorf("Content = %q; want %q", got.Content, "test message")
	}
	if len(got.Media) != 1 || got.Media[0] != "photo.jpg" {
		t.Errorf("Media = %v; want [photo.jpg]", got.Media)
	}
	if got.SessionKey != "telegram:chat456" {
		t.Errorf("SessionKey = %q; want %q", got.SessionKey, "telegram:chat456")
	}
	if got.Metadata["key"] != "val" {
		t.Errorf("Metadata[key] = %q; want %q", got.Metadata["key"], "val")
	}
}

func TestMessageBus_OutboundFieldPreservation(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	msg := OutboundMessage{
		UserID:   7,
		Channel:  "discord",
		ChatID:   "ch789",
		Content:  "reply",
		Metadata: map[string]string{"agent": "researcher"},
	}

	mb.PublishOutbound(msg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, ok := mb.SubscribeOutbound(ctx)
	if !ok {
		t.Fatal("SubscribeOutbound failed")
	}

	if got.UserID != 7 {
		t.Errorf("UserID = %d; want 7", got.UserID)
	}
	if got.Channel != "discord" {
		t.Errorf("Channel = %q; want %q", got.Channel, "discord")
	}
	if got.ChatID != "ch789" {
		t.Errorf("ChatID = %q; want %q", got.ChatID, "ch789")
	}
	if got.Content != "reply" {
		t.Errorf("Content = %q; want %q", got.Content, "reply")
	}
	if got.Metadata["agent"] != "researcher" {
		t.Errorf("Metadata[agent] = %q; want %q", got.Metadata["agent"], "researcher")
	}
}

// ---------------------------------------------------------------------------
// SubscribeOutbound with context cancellation
// ---------------------------------------------------------------------------

func TestMessageBus_SubscribeOutboundContextCancel(t *testing.T) {
	mb := NewMessageBus()
	defer mb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, ok := mb.SubscribeOutbound(ctx)
	if ok {
		t.Error("SubscribeOutbound with cancelled context should return false")
	}
}
