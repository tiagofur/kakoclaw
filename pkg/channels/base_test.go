package channels

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/storage"
)

func TestBaseChannel_IsAllowed(t *testing.T) {
	tests := []struct {
		name      string
		allowList []string
		senderID  string
		want      bool
	}{
		{
			name:      "empty allowlist allows all",
			allowList: []string{},
			senderID:  "123456",
			want:      true,
		},
		{
			name:      "numeric ID match",
			allowList: []string{"123456"},
			senderID:  "123456",
			want:      true,
		},
		{
			name:      "numeric ID no match",
			allowList: []string{"123456"},
			senderID:  "999999",
			want:      false,
		},
		{
			name:      "username with @ match",
			allowList: []string{"@testuser"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "username without @ match",
			allowList: []string{"testuser"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "username no match - different user",
			allowList: []string{"@otheruser"},
			senderID:  "123456|testuser",
			want:      false,
		},
		{
			name:      "numeric ID in compound senderID",
			allowList: []string{"123456"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "multiple entries - numeric match",
			allowList: []string{"123456", "@admin"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "multiple entries - username match",
			allowList: []string{"999999", "@testuser"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "issue #62 - username without @ should work",
			allowList: []string{"testuser"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "issue #62 - @username should work",
			allowList: []string{"@testuser"},
			senderID:  "123456|testuser",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBaseChannel("test", nil, nil, tt.allowList)
			got := bc.IsAllowed(tt.senderID)
			if got != tt.want {
				t.Errorf("IsAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewBaseChannel(t *testing.T) {
	bc := NewBaseChannel("test", nil, nil, []string{"user1"})
	if bc == nil {
		t.Fatal("NewBaseChannel returned nil")
	}
	if got := bc.Name(); got != "test" {
		t.Errorf("Name() = %q, want %q", got, "test")
	}
}

func TestBaseChannel_IsRunning(t *testing.T) {
	bc := NewBaseChannel("test", nil, nil, nil)
	if bc.IsRunning() {
		t.Error("IsRunning() = true for new channel, want false")
	}
}

func TestBaseChannel_GetUserIDForSender_Default(t *testing.T) {
	bc := NewBaseChannel("test", nil, nil, nil)

	// Without a user resolver set, should return (0, nil)
	userID, err := bc.GetUserIDForSender("any-sender")
	if err != nil {
		t.Errorf("GetUserIDForSender() err = %v, want nil", err)
	}
	if userID != 0 {
		t.Errorf("GetUserIDForSender() userID = %d, want 0", userID)
	}
}

func TestBaseChannel_GetUserIDForSender_WithResolver(t *testing.T) {
	bc := NewBaseChannel("test", nil, nil, nil)

	bc.SetUserResolver(func(senderID string) (int64, error) {
		if senderID == "known" {
			return 42, nil
		}
		return 0, fmt.Errorf("unknown sender: %s", senderID)
	})

	// Known sender
	userID, err := bc.GetUserIDForSender("known")
	if err != nil {
		t.Errorf("GetUserIDForSender(known) err = %v, want nil", err)
	}
	if userID != 42 {
		t.Errorf("GetUserIDForSender(known) userID = %d, want 42", userID)
	}

	// Unknown sender
	_, err = bc.GetUserIDForSender("unknown-sender")
	if err == nil {
		t.Error("GetUserIDForSender(unknown-sender) err = nil, want error")
	}
}

func TestBaseChannel_HandleMessage_NotAllowed(t *testing.T) {
	// Create a channel with an allow list but nil bus.
	// If the sender is not in the allow list, HandleMessage returns nil
	// early before touching the bus, so nil bus is safe here.
	bc := NewBaseChannel("test", nil, nil, []string{"allowed-user"})

	err := bc.HandleMessage("disallowed-user", "chat1", "hello", nil, nil)
	if err != nil {
		t.Errorf("HandleMessage() with disallowed sender returned err = %v, want nil", err)
	}
}

func TestBaseChannel_SetCommandHandler(t *testing.T) {
	bc := NewBaseChannel("test", nil, nil, nil)

	// SetCommandHandler is a no-op on BaseChannel; just verify no panic
	bc.SetCommandHandler(nil)

	handler := NewCommandHandler(nil)
	bc.SetCommandHandler(handler)
}

func TestBaseChannel_DMPolicy_Open(t *testing.T) {
	b := NewBaseChannel("test", nil, nil, nil)
	b.SetDMPolicy("open", nil)
	if !b.ShouldDispatch("unknown-sender") {
		t.Error("ShouldDispatch() = false with open policy, want true")
	}
}

func TestBaseChannel_DMPolicy_Disabled(t *testing.T) {
	b := NewBaseChannel("test", nil, nil, nil)
	b.SetDMPolicy("disabled", nil)
	if b.ShouldDispatch("unknown-sender") {
		t.Error("ShouldDispatch() = true with disabled policy, want false")
	}
}

func TestBaseChannel_DMPolicy_Allowlist(t *testing.T) {
	b := NewBaseChannel("test", nil, nil, []string{"known"})
	b.SetDMPolicy("allowlist", nil)
	if b.ShouldDispatch("unknown") {
		t.Error("ShouldDispatch(unknown) = true with allowlist policy, want false")
	}
	if !b.ShouldDispatch("known") {
		t.Error("ShouldDispatch(known) = false with allowlist policy, want true")
	}
}

func TestBaseChannel_DMPolicy_Pairing_Approved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := storage.NewUserStorage(path)
	if err != nil {
		t.Fatalf("NewUserStorage: %v", err)
	}
	defer s.Close()
	ps := s.Pairing()

	// Pre-approve the sender
	if err := ps.InsertPending("test", "approved-user", "code1"); err != nil {
		t.Fatalf("InsertPending: %v", err)
	}
	if _, err := ps.Approve("test", "code1"); err != nil {
		t.Fatalf("Approve: %v", err)
	}

	b := NewBaseChannel("test", nil, nil, nil)
	b.SetDMPolicy("pairing", ps)

	if !b.ShouldDispatch("approved-user") {
		t.Error("ShouldDispatch(approved-user) = false, want true")
	}
	if b.ShouldDispatch("unknown-user") {
		t.Error("ShouldDispatch(unknown-user) = true, want false")
	}
}

func TestBaseChannel_HandleMessage_DMPolicy_Blocks(t *testing.T) {
	b := NewBaseChannel("test", nil, nil, nil)
	b.SetDMPolicy("disabled", nil)

	// disabled policy returns before touching the bus, so nil bus is safe.
	err := b.HandleMessage("any-sender", "chat1", "hello", nil, nil)
	if err != nil {
		t.Errorf("HandleMessage() returned err = %v, want nil", err)
	}
}

func TestBaseChannel_HandleMessage_DMPolicy_Open_Passes(t *testing.T) {
	mb := bus.NewMessageBus()
	defer mb.Close()

	b := NewBaseChannel("test", nil, mb, nil)
	b.SetDMPolicy("open", nil)

	err := b.HandleMessage("any-sender", "chat1", "hello", nil, nil)
	if err != nil {
		t.Errorf("HandleMessage() returned err = %v, want nil", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so ConsumeInbound returns quickly
	msg, ok := mb.ConsumeInbound(ctx)
	if !ok {
		// The message was published but context was already cancelled before we consumed.
		// Try a non-blocking read via a fresh background context with a tiny deadline.
		ctx2, cancel2 := context.WithCancel(context.Background())
		defer cancel2()
		msg, ok = mb.ConsumeInbound(ctx2)
		cancel2()
	}
	if !ok || msg.SenderID != "any-sender" {
		t.Error("open policy must dispatch the message onto the bus")
	}
}
