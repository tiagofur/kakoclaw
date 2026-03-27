package tools

import (
	"context"
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

// SocialMediaProvider abstracts social media posting backends.
type SocialMediaProvider interface {
	Post(ctx context.Context, req SocialPostRequest) ([]SocialPostResult, error)
	GetAnalytics(ctx context.Context, postID, platform string) (*SocialAnalytics, error)
	Name() string
}
