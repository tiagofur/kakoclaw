package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetLimit("test", 3, time.Second)

	// First 3 requests should be allowed
	if !rl.Allow("test") {
		t.Error("First request should be allowed")
	}
	if !rl.Allow("test") {
		t.Error("Second request should be allowed")
	}
	if !rl.Allow("test") {
		t.Error("Third request should be allowed")
	}

	// Fourth request should be denied
	if rl.Allow("test") {
		t.Error("Fourth request should be denied")
	}

	// Wait for window to expire
	time.Sleep(time.Second + 100*time.Millisecond)

	// Should be allowed again
	if !rl.Allow("test") {
		t.Error("Request after window should be allowed")
	}
}

func TestRateLimiter_AllowWithWait(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetLimit("test", 1, time.Second)

	// First request
	allowed, waitTime := rl.AllowWithWait("test")
	if !allowed {
		t.Error("First request should be allowed")
	}
	if waitTime != 0 {
		t.Error("Wait time should be 0 for allowed request")
	}

	// Second request should be denied with wait time
	allowed, waitTime = rl.AllowWithWait("test")
	if allowed {
		t.Error("Second request should be denied")
	}
	if waitTime <= 0 {
		t.Error("Wait time should be positive")
	}
}

func TestRateLimiter_GetRemaining(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetLimit("test", 5, time.Minute)

	// Initially should have full limit
	remaining, _ := rl.GetRemaining("test")
	if remaining != 5 {
		t.Errorf("Expected 5 remaining, got %d", remaining)
	}

	// Use some requests
	rl.Allow("test")
	rl.Allow("test")

	remaining, _ = rl.GetRemaining("test")
	if remaining != 3 {
		t.Errorf("Expected 3 remaining, got %d", remaining)
	}
}

func TestRateLimiter_NoLimit(t *testing.T) {
	rl := NewRateLimiter()
	// No limit set for "unlimited"

	// All requests should be allowed
	for i := 0; i < 100; i++ {
		if !rl.Allow("unlimited") {
			t.Error("Requests without limit should always be allowed")
		}
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetLimit("test", 1, time.Hour)

	// Use the request
	rl.Allow("test")

	// Should be denied now
	if rl.Allow("test") {
		t.Error("Should be denied after using limit")
	}

	// Reset
	rl.Reset("test")

	// Should be allowed again
	if !rl.Allow("test") {
		t.Error("Should be allowed after reset")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetLimit("test", 10, time.Millisecond)

	// Add requests
	for i := 0; i < 5; i++ {
		rl.Allow("test")
	}

	// Wait for window to expire
	time.Sleep(2 * time.Millisecond)

	// Cleanup
	rl.Cleanup()

	// Should be able to make 10 requests again
	for i := 0; i < 10; i++ {
		if !rl.Allow("test") {
			t.Errorf("Request %d should be allowed after cleanup", i)
		}
	}
}

func TestWithRateLimit(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetLimit("test", 1, time.Hour)

	// First call should succeed
	called := false
	err := WithRateLimit("test", func() error {
		called = true
		return nil
	})
	if err != nil {
		t.Errorf("First call should not error: %v", err)
	}
	if !called {
		t.Error("Function should have been called")
	}

	// Second call should fail
	called = false
	err = WithRateLimit("test", func() error {
		called = true
		return nil
	})
	if err == nil {
		t.Error("Second call should error")
	}
	if called {
		t.Error("Function should not have been called")
	}
}

func TestWithRateLimitContext(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetLimit("test", 1, time.Hour)

	// First call
	WithRateLimitContext(context.Background(), "test", func() error {
		return nil
	})

	// Second call with short timeout should fail
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := WithRateLimitContext(ctx, "test", func() error {
		return nil
	})
	if err == nil {
		t.Error("Should error with context timeout")
	}
}

func TestGlobalLimiter(t *testing.T) {
	// Get global limiter twice
	l1 := GetGlobalLimiter()
	l2 := GetGlobalLimiter()

	if l1 != l2 {
		t.Error("Global limiter should be singleton")
	}

	// Check default limits are set
	if _, exists := l1.limits["user:global"]; !exists {
		t.Error("Default user:global limit should be set")
	}
	if _, exists := l1.limits["api:openai"]; !exists {
		t.Error("Default api:openai limit should be set")
	}
}
