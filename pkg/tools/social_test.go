package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// mockSocialProvider implements SocialMediaProvider for testing.
type mockSocialProvider struct {
	postResults []SocialPostResult
	postErr     error
	analytics   *SocialAnalytics
	analyticsErr error
	lastRequest *SocialPostRequest
}

func (m *mockSocialProvider) Name() string { return "mock" }

func (m *mockSocialProvider) Post(ctx context.Context, req SocialPostRequest) ([]SocialPostResult, error) {
	m.lastRequest = &req
	if m.postErr != nil {
		return nil, m.postErr
	}
	return m.postResults, nil
}

func (m *mockSocialProvider) GetAnalytics(ctx context.Context, postID, platform string) (*SocialAnalytics, error) {
	if m.analyticsErr != nil {
		return nil, m.analyticsErr
	}
	return m.analytics, nil
}

func TestSocialPostTool_Name(t *testing.T) {
	tool := NewSocialPostTool(nil)
	if tool.Name() != "social_post" {
		t.Fatalf("expected name 'social_post', got %q", tool.Name())
	}
}

func TestSocialPostTool_NilProvider(t *testing.T) {
	tool := NewSocialPostTool(nil)
	result, err := tool.Execute(context.Background(), map[string]any{
		"action":    "post",
		"platforms": []any{"twitter"},
		"content":   "hello",
	})
	if err != nil {
		t.Fatalf("expected nil error for nil provider, got: %v", err)
	}
	if !strings.Contains(result, "No social media platforms configured") {
		t.Fatalf("expected helpful error message, got: %s", result)
	}
}

func TestSocialPostTool_Preview(t *testing.T) {
	mock := &mockSocialProvider{}
	tool := NewSocialPostTool(mock)

	result, err := tool.Execute(context.Background(), map[string]any{
		"action":    "preview",
		"platforms": []any{"twitter", "linkedin"},
		"content":   "Check out MakoClaw!",
		"hashtags":  []any{"#golang", "#ai"},
	})
	if err != nil {
		t.Fatalf("preview failed: %v", err)
	}

	// Preview should show content
	if !strings.Contains(result, "Check out MakoClaw!") {
		t.Fatalf("expected content in preview, got: %s", result)
	}

	// Preview should show hashtags
	if !strings.Contains(result, "#golang") {
		t.Fatalf("expected hashtags in preview, got: %s", result)
	}

	// Preview should show character counts
	if !strings.Contains(result, "Characters:") {
		t.Fatalf("expected character count in preview, got: %s", result)
	}

	// Preview should show both platforms
	if !strings.Contains(result, "Twitter") || !strings.Contains(result, "Linkedin") {
		t.Fatalf("expected both platform names in preview, got: %s", result)
	}

	// Preview should NOT require confirmation
	if strings.Contains(result, "Confirmation required") {
		t.Fatalf("preview should not require confirmation, got: %s", result)
	}

	// Preview should hint at how to proceed
	if !strings.Contains(result, "confirmed=true") {
		t.Fatalf("expected instructions to confirm, got: %s", result)
	}
}

func TestSocialPostTool_PostRequiresConfirmation(t *testing.T) {
	mock := &mockSocialProvider{
		postResults: []SocialPostResult{{Platform: "twitter", PostID: "123", Status: "success"}},
	}
	tool := NewSocialPostTool(mock)

	// Post without confirmed=true should return confirmation message
	result, err := tool.Execute(context.Background(), map[string]any{
		"action":    "post",
		"platforms": []any{"twitter"},
		"content":   "Hello world!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "Confirmation required") {
		t.Fatalf("expected confirmation message, got: %s", result)
	}
	if !strings.Contains(result, "confirmed=true") {
		t.Fatalf("expected instructions to confirm, got: %s", result)
	}

	// Provider should NOT have been called
	if mock.lastRequest != nil {
		t.Fatalf("provider should not be called without confirmation")
	}
}

func TestSocialPostTool_PostWithConfirmation(t *testing.T) {
	mock := &mockSocialProvider{
		postResults: []SocialPostResult{
			{Platform: "twitter", PostID: "tw_123", PostURL: "https://twitter.com/post/123", Status: "success"},
			{Platform: "linkedin", PostID: "li_456", PostURL: "https://linkedin.com/post/456", Status: "success"},
		},
	}
	tool := NewSocialPostTool(mock)

	result, err := tool.Execute(context.Background(), map[string]any{
		"action":    "post",
		"platforms": []any{"twitter", "linkedin"},
		"content":   "Hello world!",
		"hashtags":  []any{"#test"},
		"confirmed": true,
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}

	// Verify provider was called
	if mock.lastRequest == nil {
		t.Fatalf("expected provider to be called")
	}
	if len(mock.lastRequest.Platforms) != 2 {
		t.Fatalf("expected 2 platforms, got %d", len(mock.lastRequest.Platforms))
	}
	if mock.lastRequest.Content != "Hello world!" {
		t.Fatalf("expected content 'Hello world!', got %q", mock.lastRequest.Content)
	}

	// Verify result contains post IDs
	if !strings.Contains(result, "tw_123") {
		t.Fatalf("expected twitter post ID in result, got: %s", result)
	}
	if !strings.Contains(result, "li_456") {
		t.Fatalf("expected linkedin post ID in result, got: %s", result)
	}

	// Verify result is valid JSON
	var results []SocialPostResult
	if err := json.Unmarshal([]byte(result), &results); err != nil {
		t.Fatalf("result should be valid JSON: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestSocialPostTool_Schedule(t *testing.T) {
	mock := &mockSocialProvider{
		postResults: []SocialPostResult{
			{Platform: "twitter", PostID: "tw_sched", Status: "scheduled"},
		},
	}
	tool := NewSocialPostTool(mock)

	result, err := tool.Execute(context.Background(), map[string]any{
		"action":        "schedule",
		"platforms":     []any{"twitter"},
		"content":       "Scheduled post!",
		"schedule_time": "2026-04-01T10:00:00Z",
		"confirmed":     true,
	})
	if err != nil {
		t.Fatalf("schedule failed: %v", err)
	}

	// Verify provider received schedule time
	if mock.lastRequest == nil {
		t.Fatalf("expected provider to be called")
	}
	if mock.lastRequest.ScheduleAt == nil {
		t.Fatalf("expected ScheduleAt to be set")
	}

	// Verify result
	if !strings.Contains(result, "tw_sched") {
		t.Fatalf("expected scheduled post ID, got: %s", result)
	}

	// Schedule without schedule_time should fail
	_, err = tool.Execute(context.Background(), map[string]any{
		"action":    "schedule",
		"platforms": []any{"twitter"},
		"content":   "No time set!",
		"confirmed": true,
	})
	if err == nil {
		t.Fatalf("expected error when schedule_time is missing")
	}
}

func TestSocialAnalyticsTool_GetMetrics(t *testing.T) {
	mock := &mockSocialProvider{
		analytics: &SocialAnalytics{
			PostID:      "tw_123",
			Platform:    "twitter",
			Likes:       42,
			Comments:    7,
			Shares:      15,
			Impressions: 1200,
			Clicks:      89,
		},
	}
	tool := NewSocialAnalyticsTool(mock)

	result, err := tool.Execute(context.Background(), map[string]any{
		"post_id":  "tw_123",
		"platform": "twitter",
	})
	if err != nil {
		t.Fatalf("analytics failed: %v", err)
	}

	var analytics SocialAnalytics
	if err := json.Unmarshal([]byte(result), &analytics); err != nil {
		t.Fatalf("result should be valid JSON: %v", err)
	}
	if analytics.Likes != 42 {
		t.Fatalf("expected 42 likes, got %d", analytics.Likes)
	}
	if analytics.Impressions != 1200 {
		t.Fatalf("expected 1200 impressions, got %d", analytics.Impressions)
	}
	if analytics.Clicks != 89 {
		t.Fatalf("expected 89 clicks, got %d", analytics.Clicks)
	}
}

// mockPlatformProvider is a fake PlatformProvider for testing MultiPlatformProvider.
type mockPlatformProvider struct {
	platform string
	postErr  error
}

func (m *mockPlatformProvider) Platform() string { return m.platform }
func (m *mockPlatformProvider) Post(ctx context.Context, content string, mediaURLs []string, opts map[string]any) (*SocialPostResult, error) {
	if m.postErr != nil {
		return nil, m.postErr
	}
	return &SocialPostResult{
		Platform: m.platform,
		PostID:   "mock-123",
		PostURL:  "https://example.com/post/123",
		Status:   "posted",
	}, nil
}

func TestMultiPlatformProvider_Post(t *testing.T) {
	mp := &MultiPlatformProvider{
		providers: map[string]PlatformProvider{
			"twitter": &mockPlatformProvider{platform: "twitter"},
		},
	}

	results, err := mp.Post(context.Background(), SocialPostRequest{
		Platforms: []string{"twitter"},
		Content:   "Test post",
		Hashtags:  []string{"#test"},
	})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Platform != "twitter" {
		t.Fatalf("expected platform 'twitter', got %q", results[0].Platform)
	}
	if results[0].PostID != "mock-123" {
		t.Fatalf("expected post ID 'mock-123', got %q", results[0].PostID)
	}
}

func TestMultiPlatformProvider_NoPlatformsConfigured(t *testing.T) {
	mp := &MultiPlatformProvider{
		providers: map[string]PlatformProvider{},
	}

	_, err := mp.Post(context.Background(), SocialPostRequest{
		Platforms: []string{"twitter"},
		Content:   "Test",
	})
	if err == nil {
		t.Fatal("expected error when no platforms configured")
	}
	if !strings.Contains(err.Error(), "no social media platforms configured") {
		t.Fatalf("expected helpful error, got: %v", err)
	}
}

func TestMultiPlatformProvider_MixedResults(t *testing.T) {
	mp := &MultiPlatformProvider{
		providers: map[string]PlatformProvider{
			"twitter": &mockPlatformProvider{platform: "twitter"},
			// bluesky not configured
		},
	}

	results, err := mp.Post(context.Background(), SocialPostRequest{
		Platforms: []string{"twitter", "bluesky"},
		Content:   "Test",
	})
	if err != nil {
		t.Fatalf("expected nil error for partial success, got: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// twitter should succeed, bluesky should fail
	if results[0].Status != "posted" {
		t.Fatalf("expected twitter posted, got %q", results[0].Status)
	}
	if results[1].Status != "failed" {
		t.Fatalf("expected bluesky failed, got %q", results[1].Status)
	}
}

func TestSocialAnalyticsTool_NilProvider(t *testing.T) {
	tool := NewSocialAnalyticsTool(nil)
	result, err := tool.Execute(context.Background(), map[string]any{
		"post_id":  "123",
		"platform": "twitter",
	})
	if err != nil {
		t.Fatalf("expected nil error for nil provider, got: %v", err)
	}
	if !strings.Contains(result, "No social media platforms configured") {
		t.Fatalf("expected helpful error message, got: %s", result)
	}
}

func TestSocialPostTool_PreviewCharLimitWarning(t *testing.T) {
	mock := &mockSocialProvider{}
	tool := NewSocialPostTool(mock)

	// Create content that exceeds Twitter's 280 char limit
	longContent := strings.Repeat("a", 300)

	result, err := tool.Execute(context.Background(), map[string]any{
		"action":    "preview",
		"platforms": []any{"twitter"},
		"content":   longContent,
	})
	if err != nil {
		t.Fatalf("preview failed: %v", err)
	}

	if !strings.Contains(result, "WARNING") {
		t.Fatalf("expected character limit warning for Twitter, got: %s", result)
	}
	if !strings.Contains(result, fmt.Sprintf("300 / 280")) {
		t.Fatalf("expected '300 / 280' in preview, got: %s", result)
	}
}
