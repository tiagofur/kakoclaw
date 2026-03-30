package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// mockSocialProvider implements SocialMediaProvider for testing.
type mockSocialProvider struct {
	postResults  []SocialPostResult
	postErr      error
	analytics    *SocialAnalytics
	analyticsErr error
	lastRequest  *SocialPostRequest
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

func TestSocialPostTool_ScheduleUnsupported(t *testing.T) {
	mock := &mockSocialProvider{postErr: ErrSchedulingNotSupported}
	tool := NewSocialPostTool(mock)

	_, err := tool.Execute(context.Background(), map[string]any{
		"action":        "schedule",
		"platforms":     []any{"twitter"},
		"content":       "Scheduled post!",
		"schedule_time": "2026-04-01T10:00:00Z",
		"confirmed":     true,
	})
	if err == nil {
		t.Fatal("expected scheduling error")
	}
	if !strings.Contains(err.Error(), "only supported for Facebook") {
		t.Fatalf("expected clear scheduling error, got: %v", err)
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
	platform      string
	postErr       error
	analytics     *SocialAnalytics
	analyticsErr  error
	lastContent   string
	lastMediaURLs []string
	lastOpts      map[string]any
}

func (m *mockPlatformProvider) Platform() string { return m.platform }
func (m *mockPlatformProvider) Post(ctx context.Context, content string, mediaURLs []string, opts map[string]any) (*SocialPostResult, error) {
	if m.postErr != nil {
		return nil, m.postErr
	}
	m.lastContent = content
	m.lastMediaURLs = append([]string(nil), mediaURLs...)
	if opts != nil {
		m.lastOpts = make(map[string]any, len(opts))
		for k, v := range opts {
			m.lastOpts[k] = v
		}
	}
	return &SocialPostResult{
		Platform: m.platform,
		PostID:   "mock-123",
		PostURL:  "https://example.com/post/123",
		Status:   "posted",
	}, nil
}

func (m *mockPlatformProvider) GetAnalytics(ctx context.Context, postID string) (*SocialAnalytics, error) {
	if m.analyticsErr != nil {
		return nil, m.analyticsErr
	}
	if m.analytics != nil {
		return m.analytics, nil
	}
	return &SocialAnalytics{PostID: postID, Platform: m.platform}, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
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

func TestMultiPlatformProvider_PostPassesScheduleAtOption(t *testing.T) {
	provider := &mockPlatformProvider{platform: "facebook"}
	mp := &MultiPlatformProvider{
		providers: map[string]PlatformProvider{
			"facebook:brand": provider,
		},
	}

	scheduleAt := time.Unix(1767225600, 0).UTC()
	_, err := mp.Post(context.Background(), SocialPostRequest{
		Platforms:  []string{"facebook"},
		Content:    "Test post",
		ScheduleAt: &scheduleAt,
	})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if provider.lastOpts == nil {
		t.Fatal("expected opts to be passed to provider")
	}
	if got := provider.lastOpts["schedule_at"]; got != scheduleAt.Unix() {
		t.Fatalf("expected schedule_at %d, got %#v", scheduleAt.Unix(), got)
	}
}

func TestMultiPlatformProvider_GetAnalyticsRoutesToProvider(t *testing.T) {
	expected := &SocialAnalytics{PostID: "abc", Platform: "twitter", Likes: 10}
	mp := &MultiPlatformProvider{
		providers: map[string]PlatformProvider{
			"twitter:personal": &mockPlatformProvider{platform: "twitter", analytics: expected},
		},
	}

	analytics, err := mp.GetAnalytics(context.Background(), "abc", "twitter")
	if err != nil {
		t.Fatalf("GetAnalytics failed: %v", err)
	}
	if analytics.Likes != expected.Likes {
		t.Fatalf("expected %d likes, got %d", expected.Likes, analytics.Likes)
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

func TestTwitterProvider_GetAnalytics(t *testing.T) {
	provider := &TwitterProvider{
		apiKey:            "api-key",
		apiSecret:         "api-secret",
		accessToken:       "access-token",
		accessTokenSecret: "access-secret",
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", req.Method)
			}
			if req.URL.String() != "https://api.twitter.com/2/tweets/123?tweet.fields=public_metrics" {
				t.Fatalf("unexpected URL: %s", req.URL.String())
			}
			if req.Header.Get("Authorization") == "" {
				t.Fatal("expected Authorization header")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"data":{"public_metrics":{"like_count":10,"reply_count":2,"retweet_count":3,"quote_count":4}}}`)),
			}, nil
		})},
	}

	analytics, err := provider.GetAnalytics(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetAnalytics failed: %v", err)
	}
	if analytics.Likes != 10 || analytics.Comments != 2 || analytics.Shares != 7 || analytics.Impressions != 0 {
		t.Fatalf("unexpected analytics: %+v", analytics)
	}
}

func TestBlueskyProvider_PostUsesATURIAsPostID(t *testing.T) {
	provider := &BlueskyProvider{
		handle:      "alice.test",
		appPassword: "app-pass",
		pdsURL:      "https://bsky.social",
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/xrpc/com.atproto.server.createSession":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"accessJwt":"jwt","did":"did:plc:alice"}`)),
				}, nil
			case "/xrpc/com.atproto.repo.createRecord":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"uri":"at://did:plc:alice/app.bsky.feed.post/rkey123","cid":"cid123"}`)),
				}, nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		})},
	}

	result, err := provider.Post(context.Background(), "hello", nil, nil)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if result.PostID != "at://did:plc:alice/app.bsky.feed.post/rkey123" {
		t.Fatalf("expected AT URI post ID, got %q", result.PostID)
	}
}

func TestBlueskyProvider_GetAnalytics(t *testing.T) {
	provider := &BlueskyProvider{
		handle:      "alice.test",
		appPassword: "app-pass",
		pdsURL:      "https://bsky.social",
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Path {
			case "/xrpc/com.atproto.server.createSession":
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"accessJwt":"jwt","did":"did:plc:alice"}`)),
				}, nil
			case "/xrpc/app.bsky.feed.getPostThread":
				if got := req.URL.Query().Get("uri"); got != "at://did:plc:alice/app.bsky.feed.post/rkey123" {
					t.Fatalf("unexpected URI query: %s", got)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"thread":{"post":{"uri":"at://did:plc:alice/app.bsky.feed.post/rkey123","likeCount":5,"replyCount":2,"repostCount":3}}}`)),
				}, nil
			default:
				t.Fatalf("unexpected request path: %s", req.URL.Path)
				return nil, nil
			}
		})},
	}

	analytics, err := provider.GetAnalytics(context.Background(), "at://did:plc:alice/app.bsky.feed.post/rkey123")
	if err != nil {
		t.Fatalf("GetAnalytics failed: %v", err)
	}
	if analytics.Likes != 5 || analytics.Comments != 2 || analytics.Shares != 3 {
		t.Fatalf("unexpected analytics: %+v", analytics)
	}
}

func TestFacebookProvider_PostSchedulesWhenRequested(t *testing.T) {
	var requestBody map[string]any
	scheduleAt := time.Now().Add(11 * time.Minute).UTC().Truncate(time.Second)
	provider := &FacebookProvider{
		pageAccessToken: "token",
		pageID:          "page-1",
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", req.Method)
			}
			defer req.Body.Close()
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("failed reading request body: %v", err)
			}
			if err := json.Unmarshal(body, &requestBody); err != nil {
				t.Fatalf("failed parsing request body: %v", err)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"id":"page-1_123"}`)),
			}, nil
		})},
	}

	result, err := provider.Post(context.Background(), "hello", nil, map[string]any{"schedule_at": scheduleAt.Unix()})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if requestBody["published"] != false {
		t.Fatalf("expected published=false, got %#v", requestBody["published"])
	}
	if int64(requestBody["scheduled_publish_time"].(float64)) != scheduleAt.Unix() {
		t.Fatalf("expected scheduled_publish_time %d, got %#v", scheduleAt.Unix(), requestBody["scheduled_publish_time"])
	}
	if result.Status != "scheduled" {
		t.Fatalf("expected scheduled status, got %q", result.Status)
	}
}

func TestFacebookProvider_RejectsScheduleTooSoon(t *testing.T) {
	provider := &FacebookProvider{}
	_, err := provider.Post(context.Background(), "hello", nil, map[string]any{"schedule_at": time.Now().Add(9 * time.Minute).Unix()})
	if err == nil {
		t.Fatal("expected scheduling validation error")
	}
	if !strings.Contains(err.Error(), "at least 10 minutes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFacebookProvider_GetAnalytics(t *testing.T) {
	provider := &FacebookProvider{
		pageAccessToken: "token",
		pageID:          "page-1",
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", req.Method)
			}
			if req.URL.Query().Get("metric") != "post_impressions,post_engaged_users,post_reactions_by_type_total" {
				t.Fatalf("unexpected metrics query: %s", req.URL.Query().Get("metric"))
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"data":[{"name":"post_impressions","values":[{"value":1000}]},{"name":"post_engaged_users","values":[{"value":25}]},{"name":"post_reactions_by_type_total","values":[{"value":{"like":5,"love":3}}]}]}`)),
			}, nil
		})},
	}

	analytics, err := provider.GetAnalytics(context.Background(), "page-1_123")
	if err != nil {
		t.Fatalf("GetAnalytics failed: %v", err)
	}
	if analytics.Impressions != 1000 || analytics.Clicks != 25 || analytics.Likes != 8 {
		t.Fatalf("unexpected analytics: %+v", analytics)
	}
}

func TestFacebookProvider_GetAnalyticsPermissionError(t *testing.T) {
	provider := &FacebookProvider{
		pageAccessToken: "token",
		pageID:          "page-1",
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusForbidden,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"error":{"message":"Missing pages_read_engagement permission"}}`)),
			}, nil
		})},
	}

	_, err := provider.GetAnalytics(context.Background(), "page-1_123")
	if err == nil {
		t.Fatal("expected permission error")
	}
	if !strings.Contains(err.Error(), "pages_read_engagement") {
		t.Fatalf("expected permission guidance, got: %v", err)
	}
}

func TestLinkedInProvider_GetAnalyticsUnsupported(t *testing.T) {
	provider := &LinkedInProvider{}
	_, err := provider.GetAnalytics(context.Background(), "abc")
	if !errors.Is(err, ErrAnalyticsNotSupported) {
		t.Fatalf("expected ErrAnalyticsNotSupported, got %v", err)
	}
}
