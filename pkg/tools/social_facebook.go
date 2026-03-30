package tools

// FacebookProvider posts to Facebook Pages via the Graph API.
// Auth: Page Access Token
// Posting: POST https://graph.facebook.com/{page_id}/feed
// Images: POST https://graph.facebook.com/{page_id}/photos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
)

// FacebookProvider implements PlatformProvider for Facebook Pages using the
// Graph API with a Page Access Token for authentication.
type FacebookProvider struct {
	pageAccessToken string
	pageID          string
	httpClient      *http.Client
}

// NewFacebookProvider creates a FacebookProvider from the given config.
func NewFacebookProvider(cfg config.FacebookSocialConfig) *FacebookProvider {
	return &FacebookProvider{
		pageAccessToken: cfg.PageAccessToken,
		pageID:          cfg.PageID,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}

// Platform returns the platform identifier.
func (f *FacebookProvider) Platform() string { return "facebook" }

func (f *FacebookProvider) GetAnalytics(ctx context.Context, postID string) (*SocialAnalytics, error) {
	insightsURL := fmt.Sprintf("https://graph.facebook.com/%s/insights?metric=post_impressions,post_engaged_users,post_reactions_by_type_total&access_token=%s", url.PathEscape(postID), url.QueryEscape(f.pageAccessToken))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, insightsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		if strings.Contains(string(respBody), "pages_read_engagement") {
			return nil, fmt.Errorf("facebook analytics require the pages_read_engagement permission: %s", string(respBody))
		}
		return nil, fmt.Errorf("facebook analytics failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			Name   string `json:"name"`
			Values []struct {
				Value any `json:"value"`
			} `json:"values"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse facebook analytics response: %w", err)
	}

	analytics := &SocialAnalytics{PostID: postID, Platform: f.Platform()}
	for _, metric := range result.Data {
		if len(metric.Values) == 0 {
			continue
		}
		value := metric.Values[0].Value
		switch metric.Name {
		case "post_impressions":
			analytics.Impressions = anyToInt(value)
		case "post_engaged_users":
			analytics.Clicks = anyToInt(value)
		case "post_reactions_by_type_total":
			if reactions, ok := value.(map[string]any); ok {
				for _, reactionValue := range reactions {
					analytics.Likes += anyToInt(reactionValue)
				}
			}
		}
	}

	return analytics, nil
}

// Post creates a Facebook page post. If mediaURLs are provided, the first
// image is uploaded as a photo post; otherwise a text-only post is created.
func (f *FacebookProvider) Post(ctx context.Context, content string, mediaURLs []string, opts map[string]any) (*SocialPostResult, error) {
	var scheduleAt *time.Time
	if scheduleAtUnix, ok := scheduleUnix(opts); ok {
		scheduledAt, err := requireMinimumSchedule(scheduleAtUnix, 10*time.Minute)
		if err != nil {
			return nil, err
		}
		scheduleAt = &scheduledAt
	}

	var postID string
	var err error

	if len(mediaURLs) > 0 {
		postID, err = f.postWithPhoto(ctx, content, mediaURLs[0], scheduleAt)
	} else {
		postID, err = f.postText(ctx, content, scheduleAt)
	}

	if err != nil {
		return nil, err
	}

	return &SocialPostResult{
		Platform: "facebook",
		PostID:   postID,
		PostURL:  fmt.Sprintf("https://facebook.com/%s", postID),
		Status:   socialPostStatus(scheduleAt),
	}, nil
}

// postText publishes a text-only post to the page feed.
func (f *FacebookProvider) postText(ctx context.Context, content string, scheduleAt *time.Time) (string, error) {
	body := map[string]any{
		"message":      content,
		"access_token": f.pageAccessToken,
	}
	applyFacebookSchedule(body, scheduleAt)
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal post body: %w", err)
	}

	url := fmt.Sprintf("https://graph.facebook.com/%s/feed", f.pageID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("facebook post failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return result.ID, nil
}

// postWithPhoto publishes a photo post with a caption to the page.
func (f *FacebookProvider) postWithPhoto(ctx context.Context, content, imagePath string, scheduleAt *time.Time) (string, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("message", content)
	writer.WriteField("access_token", f.pageAccessToken)
	if scheduleAt != nil {
		writer.WriteField("published", "false")
		writer.WriteField("scheduled_publish_time", fmt.Sprintf("%d", scheduleAt.Unix()))
	}

	part, err := writer.CreateFormFile("source", "image.png")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(imageData); err != nil {
		return "", fmt.Errorf("failed to write image data: %w", err)
	}
	writer.Close()

	url := fmt.Sprintf("https://graph.facebook.com/%s/photos", f.pageID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("facebook photo post failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID     string `json:"id"`
		PostID string `json:"post_id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	if result.PostID != "" {
		return result.PostID, nil
	}
	return result.ID, nil
}

func applyFacebookSchedule(body map[string]any, scheduleAt *time.Time) {
	if scheduleAt == nil {
		return
	}
	body["published"] = false
	body["scheduled_publish_time"] = scheduleAt.Unix()
}

func socialPostStatus(scheduleAt *time.Time) string {
	if scheduleAt != nil {
		return "scheduled"
	}
	return "posted"
}

func anyToInt(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}
