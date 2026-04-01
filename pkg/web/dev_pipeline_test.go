package web

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/devmemory"
)

type mockEmbedder struct{}

func (m *mockEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	vecs := make([][]float32, len(texts))
	for i, text := range texts {
		vec := []float32{0, 0, 0}
		switch {
		case strings.Contains(text, "bug"):
			vec[0] = 1
		case strings.Contains(text, "tests"):
			vec[1] = 1
		default:
			vec[2] = 1
		}
		vecs[i] = vec
	}
	return vecs, nil
}

func (m *mockEmbedder) Dimensions() int { return 3 }

func TestDevPipeline_RequestAndResponseTypes(t *testing.T) {
	req := &DevRequest{
		UserUUID:        "user-123",
		Prompt:          "hello",
		PromptInjection: "context",
	}
	resp := &DevResponse{
		UserUUID:  "user-123",
		SessionID: "session-123",
	}

	if req.UserUUID != "user-123" {
		t.Fatalf("expected UserUUID to be accessible, got %q", req.UserUUID)
	}
	if req.Prompt != "hello" {
		t.Fatalf("expected Prompt to be accessible, got %q", req.Prompt)
	}
	if req.PromptInjection != "context" {
		t.Fatalf("expected PromptInjection to be accessible, got %q", req.PromptInjection)
	}
	if resp.UserUUID != "user-123" {
		t.Fatalf("expected response UserUUID to be accessible, got %q", resp.UserUUID)
	}
	if resp.SessionID != "session-123" {
		t.Fatalf("expected SessionID to be accessible, got %q", resp.SessionID)
	}
}

func TestDevPipeline_RunPre(t *testing.T) {
	t.Run("runs middlewares in registration order", func(t *testing.T) {
		pipeline := NewDevPipeline()
		var order []string

		pipeline.Use(func(ctx context.Context, req *DevRequest) (*DevRequest, error) {
			order = append(order, "a")
			req.Prompt += "-a"
			return req, nil
		})
		pipeline.Use(func(ctx context.Context, req *DevRequest) (*DevRequest, error) {
			order = append(order, "b")
			req.Prompt += "-b"
			return req, nil
		})
		pipeline.Use(func(ctx context.Context, req *DevRequest) (*DevRequest, error) {
			order = append(order, "c")
			req.Prompt += "-c"
			return req, nil
		})

		result, err := pipeline.RunPre(context.Background(), &DevRequest{Prompt: "prompt"})
		if err != nil {
			t.Fatalf("RunPre returned unexpected error: %v", err)
		}

		if !reflect.DeepEqual(order, []string{"a", "b", "c"}) {
			t.Fatalf("expected middleware order [a b c], got %v", order)
		}
		if result == nil || result.Prompt != "prompt-a-b-c" {
			t.Fatalf("expected final prompt to be propagated, got %+v", result)
		}
	})

	t.Run("empty chain returns original request", func(t *testing.T) {
		pipeline := NewDevPipeline()
		req := &DevRequest{Prompt: "prompt"}

		result, err := pipeline.RunPre(context.Background(), req)
		if err != nil {
			t.Fatalf("RunPre returned unexpected error: %v", err)
		}
		if result != req {
			t.Fatalf("expected original request pointer to be returned")
		}
	})

	t.Run("short circuits on first error", func(t *testing.T) {
		pipeline := NewDevPipeline()
		boom := errors.New("boom")
		called := false

		pipeline.Use(func(ctx context.Context, req *DevRequest) (*DevRequest, error) {
			return req, boom
		})
		pipeline.Use(func(ctx context.Context, req *DevRequest) (*DevRequest, error) {
			called = true
			return req, nil
		})

		result, err := pipeline.RunPre(context.Background(), &DevRequest{Prompt: "prompt"})
		if !errors.Is(err, boom) {
			t.Fatalf("expected boom error, got %v", err)
		}
		if result != nil {
			t.Fatalf("expected nil request on error, got %+v", result)
		}
		if called {
			t.Fatalf("expected later middleware to be skipped")
		}
	})
}

func TestDevPipeline_RunPost(t *testing.T) {
	t.Run("runs all middlewares even when one errors", func(t *testing.T) {
		pipeline := NewDevPipeline()
		var order []string

		pipeline.UsePost(func(ctx context.Context, resp *DevResponse) error {
			order = append(order, "x")
			return errors.New("warn")
		})
		pipeline.UsePost(func(ctx context.Context, resp *DevResponse) error {
			order = append(order, "y")
			return nil
		})

		if err := pipeline.RunPost(context.Background(), &DevResponse{UserUUID: "user-123"}); err != nil {
			t.Fatalf("expected RunPost to swallow errors, got %v", err)
		}
		if !reflect.DeepEqual(order, []string{"x", "y"}) {
			t.Fatalf("expected both post middlewares to run, got %v", order)
		}
	})

	t.Run("empty chain returns nil without panic", func(t *testing.T) {
		pipeline := NewDevPipeline()
		if err := pipeline.RunPost(context.Background(), &DevResponse{}); err != nil {
			t.Fatalf("expected nil error for empty post chain, got %v", err)
		}
	})
}

func TestDevPipeline_MemoryInjectionMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		getStore        func(string) (*devmemory.Store, error)
		wantContains    string
		wantErr         bool
		wantStoreCalled bool
	}{
		{
			name: "sets prompt injection when memory returns value",
			getStore: func(userUUID string) (*devmemory.Store, error) {
				store, err := devmemory.NewStore(filepath.Join(t.TempDir(), "dev-memory.db"), &mockEmbedder{})
				if err != nil {
					return nil, err
				}
				if err := store.Add(context.Background(), "existing context about bug fix", ""); err != nil {
					return nil, err
				}
				return store, nil
			},
			wantContains:    "existing context about bug fix",
			wantStoreCalled: true,
		},
		{
			name: "no-op when memory returns empty string",
			getStore: func(userUUID string) (*devmemory.Store, error) {
				return devmemory.NewStore(filepath.Join(t.TempDir(), "dev-memory.db"), &mockEmbedder{})
			},
			wantContains:    "",
			wantStoreCalled: true,
		},
		{
			name: "swallows get store error",
			getStore: func(userUUID string) (*devmemory.Store, error) {
				return nil, errors.New("memory unavailable")
			},
			wantContains:    "",
			wantStoreCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			var store *devmemory.Store
			middleware := MemoryInjectionMiddleware(func(userUUID string) (*devmemory.Store, error) {
				called = true
				var err error
				store, err = tt.getStore(userUUID)
				return store, err
			})
			if store != nil {
				defer store.Close()
			}

			req := &DevRequest{UserUUID: "user-123", Prompt: "bug fix"}
			result, err := middleware(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}
			if result != req {
				t.Fatalf("expected middleware to return same request pointer")
			}
			if tt.wantContains == "" {
				if req.PromptInjection != "" {
					t.Fatalf("expected empty PromptInjection, got %q", req.PromptInjection)
				}
			} else if !strings.Contains(req.PromptInjection, tt.wantContains) {
				t.Fatalf("expected PromptInjection to contain %q, got %q", tt.wantContains, req.PromptInjection)
			}
			if called != tt.wantStoreCalled {
				t.Fatalf("expected getStore called=%v, got %v", tt.wantStoreCalled, called)
			}
			if store != nil {
				_ = store.Close()
			}
		})
	}
}
