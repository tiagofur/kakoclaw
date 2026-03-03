package channels

import (
	"context"
	"testing"
)

func TestIsCommand(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "slash only",
			content: "/",
			want:    true,
		},
		{
			name:    "slash help",
			content: "/help",
			want:    true,
		},
		{
			name:    "slash setup with args",
			content: "/setup something",
			want:    true,
		},
		{
			name:    "empty string",
			content: "",
			want:    false,
		},
		{
			name:    "normal text",
			content: "normal text",
			want:    false,
		},
		{
			name:    "slash not at start",
			content: "not /a command",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCommand(tt.content)
			if got != tt.want {
				t.Errorf("IsCommand(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantCmd  string
		wantArgs string
	}{
		{
			name:     "simple command",
			content:  "/help",
			wantCmd:  "help",
			wantArgs: "",
		},
		{
			name:     "command with single arg",
			content:  "/setup wizard",
			wantCmd:  "setup",
			wantArgs: "wizard",
		},
		{
			name:     "command with multiple args",
			content:  "/cmd arg1 arg2",
			wantCmd:  "cmd",
			wantArgs: "arg1 arg2",
		},
		{
			name:     "command with no args",
			content:  "/noargs",
			wantCmd:  "noargs",
			wantArgs: "",
		},
		{
			name:     "empty string",
			content:  "",
			wantCmd:  "",
			wantArgs: "",
		},
		{
			name:     "not a command",
			content:  "hello",
			wantCmd:  "",
			wantArgs: "",
		},
		{
			name:     "command with newline separator",
			content:  "/cmd\narg on new line",
			wantCmd:  "cmd",
			wantArgs: "arg on new line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs := ParseCommand(tt.content)
			if gotCmd != tt.wantCmd {
				t.Errorf("ParseCommand(%q) cmd = %q, want %q", tt.content, gotCmd, tt.wantCmd)
			}
			if gotArgs != tt.wantArgs {
				t.Errorf("ParseCommand(%q) args = %q, want %q", tt.content, gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestHandleCommand_UnknownCommand(t *testing.T) {
	// Unknown commands should not use storage, so nil is safe here
	handler := NewCommandHandler(nil)
	ctx := context.Background()

	handled, response, err := handler.HandleCommand(ctx, "test-channel", "sender1", "/foo")
	if handled {
		t.Errorf("HandleCommand(/foo) handled = true, want false")
	}
	if response != "" {
		t.Errorf("HandleCommand(/foo) response = %q, want empty", response)
	}
	if err != nil {
		t.Errorf("HandleCommand(/foo) err = %v, want nil", err)
	}
}

func TestHandleCommand_NotACommand(t *testing.T) {
	handler := NewCommandHandler(nil)
	ctx := context.Background()

	handled, response, err := handler.HandleCommand(ctx, "test-channel", "sender1", "hello world")
	if handled {
		t.Errorf("HandleCommand(hello world) handled = true, want false")
	}
	if response != "" {
		t.Errorf("HandleCommand(hello world) response = %q, want empty", response)
	}
	if err != nil {
		t.Errorf("HandleCommand(hello world) err = %v, want nil", err)
	}
}

func TestHandleCommand_StatusWithNilStore(t *testing.T) {
	// /status does not use storage, so nil is safe
	handler := NewCommandHandler(nil)
	ctx := context.Background()

	handled, response, err := handler.HandleCommand(ctx, "test-channel", "sender1", "/status")
	if !handled {
		t.Errorf("HandleCommand(/status) handled = false, want true")
	}
	if response == "" {
		t.Error("HandleCommand(/status) response is empty, want non-empty status message")
	}
	if err != nil {
		t.Errorf("HandleCommand(/status) err = %v, want nil", err)
	}
}

func TestNewCommandHandler(t *testing.T) {
	handler := NewCommandHandler(nil)
	if handler == nil {
		t.Fatal("NewCommandHandler(nil) returned nil")
	}
}
