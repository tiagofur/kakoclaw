package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// SocialAnalyticsTool retrieves engagement metrics for social media posts.
type SocialAnalyticsTool struct {
	provider SocialMediaProvider
	userID   int64
	mu       sync.RWMutex
}

// NewSocialAnalyticsTool creates a SocialAnalyticsTool backed by the given provider.
func NewSocialAnalyticsTool(provider SocialMediaProvider) *SocialAnalyticsTool {
	return &SocialAnalyticsTool{provider: provider}
}

func (t *SocialAnalyticsTool) Name() string {
	return "social_analytics"
}

func (t *SocialAnalyticsTool) Description() string {
	return "Get engagement metrics for social media posts. Returns likes, comments, shares, impressions, and clicks."
}

func (t *SocialAnalyticsTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"post_id": map[string]any{
				"type":        "string",
				"description": "The post ID to get analytics for",
			},
			"platform": map[string]any{
				"type":        "string",
				"description": "The platform: twitter, linkedin, instagram, facebook, tiktok",
			},
		},
		"required": []string{"post_id", "platform"},
	}
}

// SetUserID implements the UserAwareTool interface.
func (t *SocialAnalyticsTool) SetUserID(userID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.userID = userID
}

func (t *SocialAnalyticsTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	if t.provider == nil {
		return "Error: No social media platforms configured. Add credentials in Settings or config.json under tools.social_media.", nil
	}

	postID, _ := args["post_id"].(string)
	if postID == "" {
		return "", fmt.Errorf("post_id is required")
	}

	platform, _ := args["platform"].(string)
	if platform == "" {
		return "", fmt.Errorf("platform is required")
	}

	analytics, err := t.provider.GetAnalytics(ctx, postID, platform)
	if err != nil {
		return "", fmt.Errorf("analytics request failed: %w", err)
	}

	resultJSON, err := json.MarshalIndent(analytics, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal analytics: %w", err)
	}

	return string(resultJSON), nil
}
