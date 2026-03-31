package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
)

type TikTokProvider struct {
	accessToken string
	httpClient  *http.Client
}

func NewTikTokProvider(cfg config.TikTokSocialConfig) *TikTokProvider {
	return &TikTokProvider{
		accessToken: cfg.AccessToken,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (tk *TikTokProvider) Platform() string { return "tiktok" }

func (tk *TikTokProvider) GetAnalytics(ctx context.Context, postID string) (*SocialAnalytics, error) {
	apiURL := fmt.Sprintf(
		"https://open.tiktokapis.com/v2/video/query/?fields=like_count,comment_count,share_count,view_count",
	)
	body := map[string]any{
		"filters": map[string]any{
			"video_ids": []string{postID},
		},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tiktok analytics request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tk.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := tk.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tiktok analytics failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			Videos []struct {
				ID           string `json:"id"`
				LikeCount    int    `json:"like_count"`
				CommentCount int    `json:"comment_count"`
				ShareCount   int    `json:"share_count"`
				ViewCount    int    `json:"view_count"`
			} `json:"videos"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tiktok analytics response: %w", err)
	}

	if len(result.Data.Videos) == 0 {
		return &SocialAnalytics{PostID: postID, Platform: tk.Platform()}, nil
	}

	video := result.Data.Videos[0]
	return &SocialAnalytics{
		PostID:      video.ID,
		Platform:    tk.Platform(),
		Likes:       video.LikeCount,
		Comments:    video.CommentCount,
		Shares:      video.ShareCount,
		Impressions: video.ViewCount,
	}, nil
}

func (tk *TikTokProvider) Post(ctx context.Context, content string, mediaURLs []string, opts map[string]any) (*SocialPostResult, error) {
	return nil, ErrSchedulingNotSupported
}
