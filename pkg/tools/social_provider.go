package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SocialPostResult represents the outcome of posting to a single platform.
type SocialPostResult struct {
	Platform string `json:"platform"`
	PostID   string `json:"post_id"`
	PostURL  string `json:"post_url"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

// SocialPostRequest contains all parameters needed to publish a social media post.
type SocialPostRequest struct {
	Platforms  []string   `json:"platforms"`
	Content    string     `json:"content"`
	MediaURLs  []string   `json:"media_urls,omitempty"`
	Hashtags   []string   `json:"hashtags,omitempty"`
	ScheduleAt *time.Time `json:"schedule_at,omitempty"`
	Title      string     `json:"title,omitempty"`
}

// SocialAnalytics holds engagement metrics for a single post on a single platform.
type SocialAnalytics struct {
	PostID      string `json:"post_id"`
	Platform    string `json:"platform"`
	Likes       int    `json:"likes"`
	Comments    int    `json:"comments"`
	Shares      int    `json:"shares"`
	Impressions int    `json:"impressions"`
	Clicks      int    `json:"clicks"`
}

// SocialMediaProvider abstracts social media posting backends (Ayrshare, etc.).
type SocialMediaProvider interface {
	Post(ctx context.Context, req SocialPostRequest) ([]SocialPostResult, error)
	GetAnalytics(ctx context.Context, postID, platform string) (*SocialAnalytics, error)
	Name() string
}

// --- Ayrshare Provider ---

var socialHTTPClient = &http.Client{Timeout: 30 * time.Second}

// AyrshareProvider implements SocialMediaProvider using the Ayrshare API.
type AyrshareProvider struct {
	apiKey     string
	httpClient *http.Client
}

// NewAyrshareProvider creates an AyrshareProvider with the given API key.
func NewAyrshareProvider(apiKey string) *AyrshareProvider {
	return &AyrshareProvider{
		apiKey:     apiKey,
		httpClient: socialHTTPClient,
	}
}

func (a *AyrshareProvider) Name() string { return "ayrshare" }

func (a *AyrshareProvider) Post(ctx context.Context, req SocialPostRequest) ([]SocialPostResult, error) {
	postContent := req.Content
	if len(req.Hashtags) > 0 {
		postContent += " " + strings.Join(req.Hashtags, " ")
	}

	body := map[string]any{
		"post":      postContent,
		"platforms": req.Platforms,
	}

	if len(req.MediaURLs) > 0 {
		body["mediaUrls"] = req.MediaURLs
	}

	if req.ScheduleAt != nil {
		body["scheduleDate"] = req.ScheduleAt.UTC().Format(time.RFC3339)
	}

	if req.Title != "" {
		body["title"] = req.Title
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://app.ayrshare.com/api/post", strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5 MB cap
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ayrshare API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		ID       string `json:"id"`
		PostIds  []struct {
			Platform string `json:"platform"`
			PostID   string `json:"id"`
			PostURL  string `json:"postUrl"`
			Status   string `json:"status"`
		} `json:"postIds"`
		Status string `json:"status"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := make([]SocialPostResult, 0, len(req.Platforms))
	for _, p := range apiResp.PostIds {
		results = append(results, SocialPostResult{
			Platform: p.Platform,
			PostID:   p.PostID,
			PostURL:  p.PostURL,
			Status:   p.Status,
		})
	}

	// If no postIds returned, create a single result from top-level response
	if len(results) == 0 && apiResp.Status != "" {
		for _, platform := range req.Platforms {
			results = append(results, SocialPostResult{
				Platform: platform,
				PostID:   apiResp.ID,
				Status:   apiResp.Status,
			})
		}
	}

	return results, nil
}

func (a *AyrshareProvider) GetAnalytics(ctx context.Context, postID, platform string) (*SocialAnalytics, error) {
	analyticsURL := fmt.Sprintf("https://app.ayrshare.com/api/analytics/post?id=%s&platforms[]=%s", postID, platform)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", analyticsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024)) // 5 MB cap
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ayrshare API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Analytics struct {
			Likes       int `json:"likes"`
			Comments    int `json:"comments"`
			Shares      int `json:"shares"`
			Impressions int `json:"impressions"`
			Clicks      int `json:"clicks"`
		} `json:"analytics"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &SocialAnalytics{
		PostID:      postID,
		Platform:    platform,
		Likes:       apiResp.Analytics.Likes,
		Comments:    apiResp.Analytics.Comments,
		Shares:      apiResp.Analytics.Shares,
		Impressions: apiResp.Analytics.Impressions,
		Clicks:      apiResp.Analytics.Clicks,
	}, nil
}
