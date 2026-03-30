package tools

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrAnalyticsNotSupported  = errors.New("analytics not supported for this platform")
	ErrSchedulingNotSupported = errors.New("scheduling not supported for this platform")
)

func scheduleUnix(opts map[string]any) (int64, bool) {
	if opts == nil {
		return 0, false
	}
	raw, ok := opts["schedule_at"]
	if !ok || raw == nil {
		return 0, false
	}
	switch v := raw.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	default:
		return 0, false
	}
}

func requireMinimumSchedule(scheduleAtUnix int64, minDelay time.Duration) (time.Time, error) {
	scheduledAt := time.Unix(scheduleAtUnix, 0).UTC()
	if time.Until(scheduledAt) < minDelay {
		return time.Time{}, fmt.Errorf("scheduled_publish_time must be at least %d minutes in the future", int(minDelay.Minutes()))
	}
	return scheduledAt, nil
}

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
	// Platforms is a list of target accounts in "platform:alias" format (e.g. "twitter:personal",
	// "facebook:brand_page"). If no alias is given (e.g. "twitter"), the first configured
	// account for that platform is used.
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

// SocialMediaProvider abstracts social media posting backends.
type SocialMediaProvider interface {
	Post(ctx context.Context, req SocialPostRequest) ([]SocialPostResult, error)
	GetAnalytics(ctx context.Context, postID, platform string) (*SocialAnalytics, error)
	Name() string
}
