package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	requests map[string][]time.Time
	limits   map[string]RateLimit
	mu       sync.RWMutex
}

// RateLimit defines the limit configuration
type RateLimit struct {
	Requests int           // Maximum number of requests
	Window   time.Duration // Time window for the limit
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limits:   make(map[string]RateLimit),
	}
}

// SetLimit configures a rate limit for a specific key
func (rl *RateLimiter) SetLimit(key string, requests int, window time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limits[key] = RateLimit{
		Requests: requests,
		Window:   window,
	}
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.limits[key]
	if !exists {
		// No limit configured, allow all
		return true
	}

	now := time.Now()
	
	// Get existing requests for this key
	requests := rl.requests[key]
	
	// Filter out old requests outside the window
	validRequests := make([]time.Time, 0)
	windowStart := now.Add(-limit.Window)
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	
	// Check if we're within the limit
	if len(validRequests) >= limit.Requests {
		return false
	}
	
	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	
	return true
}

// AllowWithWait checks if allowed and returns wait time if not
func (rl *RateLimiter) AllowWithWait(key string) (allowed bool, waitTime time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.limits[key]
	if !exists {
		return true, 0
	}

	now := time.Now()
	requests := rl.requests[key]
	
	// Filter out old requests
	validRequests := make([]time.Time, 0)
	windowStart := now.Add(-limit.Window)
	for _, t := range requests {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}
	
	if len(validRequests) >= limit.Requests {
		// Calculate wait time until oldest request expires
		if len(validRequests) > 0 {
			oldest := validRequests[0]
			waitTime = oldest.Add(limit.Window).Sub(now)
		}
		return false, waitTime
	}
	
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	
	return true, 0
}

// GetRemaining returns remaining requests and reset time
func (rl *RateLimiter) GetRemaining(key string) (remaining int, resetTime time.Time) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit, exists := rl.limits[key]
	if !exists {
		return -1, time.Time{}
	}

	now := time.Now()
	requests := rl.requests[key]
	
	// Count valid requests
	validCount := 0
	windowStart := now.Add(-limit.Window)
	var oldest time.Time
	
	for _, t := range requests {
		if t.After(windowStart) {
			validCount++
			if oldest.IsZero() || t.Before(oldest) {
				oldest = t
			}
		}
	}
	
	remaining = limit.Requests - validCount
	if remaining < 0 {
		remaining = 0
	}
	
	if !oldest.IsZero() {
		resetTime = oldest.Add(limit.Window)
	}
	
	return remaining, resetTime
}

// Cleanup removes old entries to prevent memory leaks
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	
	for key, requests := range rl.requests {
		if limit, exists := rl.limits[key]; exists {
			windowStart := now.Add(-limit.Window)
			validRequests := make([]time.Time, 0)
			
			for _, t := range requests {
				if t.After(windowStart) {
					validRequests = append(validRequests, t)
				}
			}
			
			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}
	}
}

// Reset clears all rate limit data for a key
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.requests, key)
}

// Global rate limiter instance
var globalLimiter *RateLimiter
var once sync.Once

// GetGlobalLimiter returns the singleton rate limiter instance
func GetGlobalLimiter() *RateLimiter {
	once.Do(func() {
		globalLimiter = NewRateLimiter()
		
		// Set default limits
		// Per-user limits
		globalLimiter.SetLimit("user:global", 100, time.Hour)     // 100 requests/hour per user
		globalLimiter.SetLimit("user:burst", 10, time.Minute)     // 10 requests/minute burst
		
		// API limits
		globalLimiter.SetLimit("api:openai", 60, time.Minute)     // OpenAI tier 1 limit
		globalLimiter.SetLimit("api:anthropic", 40, time.Minute)  // Anthropic limit
		globalLimiter.SetLimit("api:openrouter", 100, time.Minute) // OpenRouter limit
		
		// Tool limits
		globalLimiter.SetLimit("tool:web_search", 30, time.Hour)  // Brave Search free tier
		globalLimiter.SetLimit("tool:shell", 20, time.Hour)       // Shell execution
	})
	
	return globalLimiter
}

// WithRateLimit wraps a function with rate limiting
func WithRateLimit(key string, fn func() error) error {
	limiter := GetGlobalLimiter()
	
	allowed, waitTime := limiter.AllowWithWait(key)
	if !allowed {
		return fmt.Errorf("rate limit exceeded for %s, please wait %v", key, waitTime)
	}
	
	return fn()
}

// WithRateLimitContext wraps a function with rate limiting and context support
func WithRateLimitContext(ctx context.Context, key string, fn func() error) error {
	limiter := GetGlobalLimiter()
	
	allowed, waitTime := limiter.AllowWithWait(key)
	if !allowed {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Retry after wait
			allowed, _ = limiter.AllowWithWait(key)
			if !allowed {
				return fmt.Errorf("rate limit exceeded for %s", key)
			}
		}
	}
	
	return fn()
}
