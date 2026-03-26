package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// platformCharLimit returns the maximum character count for the given platform.
func platformCharLimit(platform string) int {
	switch strings.ToLower(platform) {
	case "twitter":
		return 280
	case "linkedin":
		return 3000
	case "instagram":
		return 2200
	case "tiktok":
		return 4000
	case "facebook":
		return 63206
	default:
		return 0
	}
}

// SocialPostTool allows the agent to publish content to social media platforms.
type SocialPostTool struct {
	provider SocialMediaProvider
	userID   int64
	userUUID string
	mu       sync.RWMutex
}

// NewSocialPostTool creates a SocialPostTool backed by the given provider.
func NewSocialPostTool(provider SocialMediaProvider) *SocialPostTool {
	return &SocialPostTool{provider: provider}
}

func (t *SocialPostTool) Name() string {
	return "social_post"
}

func (t *SocialPostTool) Description() string {
	return "Post content to social media platforms. Supports Twitter/X, LinkedIn, Instagram, Facebook, and TikTok. Use action=preview first to review before posting."
}

func (t *SocialPostTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "preview to check content, post to publish now, schedule to publish later",
				"enum":        []string{"post", "schedule", "preview"},
			},
			"platforms": map[string]any{
				"type":        "array",
				"description": "Target platforms: twitter, linkedin, instagram, facebook, tiktok",
				"items": map[string]any{
					"type": "string",
				},
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The post text content",
			},
			"media_urls": map[string]any{
				"type":        "array",
				"description": "Image/video URLs or local workspace paths",
				"items": map[string]any{
					"type": "string",
				},
			},
			"hashtags": map[string]any{
				"type":        "array",
				"description": "Hashtags to append",
				"items": map[string]any{
					"type": "string",
				},
			},
			"schedule_time": map[string]any{
				"type":        "string",
				"description": "ISO 8601 datetime for scheduling, e.g. 2026-04-01T10:00:00Z",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "Title for platforms that support it (LinkedIn articles)",
			},
			"confirmed": map[string]any{
				"type":        "boolean",
				"description": "Must be true to actually post or schedule",
			},
		},
		"required": []string{"action", "platforms", "content"},
	}
}

// SetUserID implements the UserAwareTool interface.
func (t *SocialPostTool) SetUserID(userID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.userID = userID
}

// SetUserContext implements the UserConfigTool interface.
func (t *SocialPostTool) SetUserContext(userID int64, userUUID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.userID = userID
	t.userUUID = userUUID
}

func (t *SocialPostTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	if t.provider == nil {
		return "Error: No social media provider configured. Set MAKOCLAW_TOOLS_SOCIAL_API_KEY to enable social posting.", nil
	}

	action, _ := args["action"].(string)
	if action == "" {
		return "", fmt.Errorf("action is required (post, schedule, or preview)")
	}

	content, _ := args["content"].(string)
	if content == "" {
		return "", fmt.Errorf("content is required")
	}

	// Extract platforms from args (handle both []any and []string)
	platforms := extractStringSlice(args, "platforms")
	if len(platforms) == 0 {
		return "", fmt.Errorf("at least one platform is required")
	}

	hashtags := extractStringSlice(args, "hashtags")
	mediaURLs := extractStringSlice(args, "media_urls")
	title, _ := args["title"].(string)

	// Build the full content with hashtags for preview/validation
	fullContent := content
	if len(hashtags) > 0 {
		fullContent += " " + strings.Join(hashtags, " ")
	}

	if action == "preview" {
		return t.buildPreview(platforms, fullContent, mediaURLs), nil
	}

	// Post and schedule actions require confirmation
	if !isConfirmed(args) {
		return ConfirmationMessage("social_post", "posting to social media is public and irreversible"), nil
	}

	req := SocialPostRequest{
		Platforms: platforms,
		Content:   content,
		MediaURLs: mediaURLs,
		Hashtags:  hashtags,
		Title:     title,
	}

	if action == "schedule" {
		scheduleStr, _ := args["schedule_time"].(string)
		if scheduleStr == "" {
			return "", fmt.Errorf("schedule_time is required for schedule action")
		}
		scheduleAt, err := time.Parse(time.RFC3339, scheduleStr)
		if err != nil {
			return "", fmt.Errorf("invalid schedule_time format (expected ISO 8601): %w", err)
		}
		req.ScheduleAt = &scheduleAt
	}

	results, err := t.provider.Post(ctx, req)
	if err != nil {
		return "", fmt.Errorf("social post failed: %w", err)
	}

	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	return string(resultJSON), nil
}

func (t *SocialPostTool) buildPreview(platforms []string, fullContent string, mediaURLs []string) string {
	var lines []string
	lines = append(lines, "=== Social Media Post Preview ===")
	lines = append(lines, "")

	for _, platform := range platforms {
		limit := platformCharLimit(platform)
		charCount := len(fullContent)

		lines = append(lines, fmt.Sprintf("--- %s ---", strings.ToUpper(platform[:1])+platform[1:]))
		lines = append(lines, fmt.Sprintf("Content: %s", fullContent))

		if limit > 0 {
			lines = append(lines, fmt.Sprintf("Characters: %d / %d", charCount, limit))
			if charCount > limit {
				lines = append(lines, fmt.Sprintf("WARNING: Content exceeds %s character limit by %d characters!", platform, charCount-limit))
			}
		} else {
			lines = append(lines, fmt.Sprintf("Characters: %d", charCount))
		}

		if len(mediaURLs) > 0 {
			lines = append(lines, fmt.Sprintf("Media: %d file(s)", len(mediaURLs)))
		}
		lines = append(lines, "")
	}

	lines = append(lines, "To publish, re-call with the same arguments plus confirmed=true.")
	return strings.Join(lines, "\n")
}

// extractStringSlice extracts a []string from args[key], handling []any from JSON.
func extractStringSlice(args map[string]any, key string) []string {
	val, ok := args[key]
	if !ok {
		return nil
	}

	// Already a []string (from tests or direct Go calls)
	if ss, ok := val.([]string); ok {
		return ss
	}

	// From JSON unmarshaling: []any
	if arr, ok := val.([]any); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}

	return nil
}
