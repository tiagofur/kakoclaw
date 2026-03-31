package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
)

type InstagramProvider struct {
	accessToken string
	accountID   string
	httpClient  *http.Client
}

func NewInstagramProvider(cfg config.InstagramSocialConfig) *InstagramProvider {
	return &InstagramProvider{
		accessToken: cfg.AccessToken,
		accountID:   cfg.AccountID,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (ig *InstagramProvider) Platform() string { return "instagram" }

func (ig *InstagramProvider) GetAnalytics(ctx context.Context, postID string) (*SocialAnalytics, error) {
	insightsURL := fmt.Sprintf(
		"https://graph.facebook.com/v21.0/%s/insights?metric=impressions,reach,likes,comments,shares,saved&access_token=%s",
		url.PathEscape(postID),
		url.QueryEscape(ig.accessToken),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, insightsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ig.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("instagram analytics failed (status %d): %s", resp.StatusCode, string(respBody))
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
		return nil, fmt.Errorf("failed to parse instagram analytics response: %w", err)
	}

	analytics := &SocialAnalytics{PostID: postID, Platform: ig.Platform()}
	for _, metric := range result.Data {
		if len(metric.Values) == 0 {
			continue
		}
		value := metric.Values[0].Value
		switch metric.Name {
		case "impressions":
			analytics.Impressions = anyToInt(value)
		case "likes":
			analytics.Likes = anyToInt(value)
		case "comments":
			analytics.Comments = anyToInt(value)
		case "shares":
			analytics.Shares = anyToInt(value)
		case "reach":
		case "saved":
		}
	}

	return analytics, nil
}

func (ig *InstagramProvider) Post(ctx context.Context, content string, mediaURLs []string, opts map[string]any) (*SocialPostResult, error) {
	if _, ok := scheduleUnix(opts); ok {
		return nil, ErrSchedulingNotSupported
	}

	if len(mediaURLs) > 0 {
		return ig.postWithImage(ctx, content, mediaURLs[0])
	}
	return ig.postText(ctx, content)
}

func (ig *InstagramProvider) postText(ctx context.Context, content string) (*SocialPostResult, error) {
	body := map[string]any{
		"message":      content,
		"access_token": ig.accessToken,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal post body: %w", err)
	}

	apiURL := fmt.Sprintf("https://graph.facebook.com/v21.0/%s/media?media_type=TEXT", ig.accountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	containerID, err := ig.createMediaContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	return ig.publishContainer(ctx, containerID)
}

func (ig *InstagramProvider) postWithImage(ctx context.Context, content, imagePath string) (*SocialPostResult, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}

	uploadURL, err := ig.initImageUpload(ctx)
	if err != nil {
		return nil, err
	}

	if err := ig.uploadImageBinary(ctx, uploadURL, imageData); err != nil {
		return nil, err
	}

	return ig.publishContainer(ctx, uploadURL)
}

func (ig *InstagramProvider) initImageUpload(ctx context.Context) (string, error) {
	body := map[string]any{
		"access_token": ig.accessToken,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	apiURL := fmt.Sprintf("https://graph.facebook.com/v21.0/%s/media", ig.accountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	return ig.createMediaContainer(ctx, req)
}

func (ig *InstagramProvider) createMediaContainer(ctx context.Context, req *http.Request) (string, error) {
	resp, err := ig.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("instagram container creation failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse instagram response: %w", err)
	}
	return result.ID, nil
}

func (ig *InstagramProvider) uploadImageBinary(ctx context.Context, uploadURL string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := ig.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
		return fmt.Errorf("instagram image upload failed (status %d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (ig *InstagramProvider) publishContainer(ctx context.Context, containerID string) (*SocialPostResult, error) {
	body := map[string]any{
		"creation_id":  containerID,
		"access_token": ig.accessToken,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://graph.facebook.com/v21.0/%s/media_publish", ig.accountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ig.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("instagram publish failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse instagram publish response: %w", err)
	}

	return &SocialPostResult{
		Platform: "instagram",
		PostID:   result.ID,
		PostURL:  fmt.Sprintf("https://instagram.com/p/%s", result.ID),
		Status:   "posted",
	}, nil
}
