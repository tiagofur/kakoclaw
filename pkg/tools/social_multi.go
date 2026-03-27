package tools

// MultiPlatformProvider routes social media requests to native platform providers.
// Zero cost -- uses free platform APIs directly.

import (
	"context"
	"fmt"
	"strings"

	"github.com/sipeed/makoclaw/pkg/config"
)

// PlatformProvider is the interface each platform-specific provider implements.
type PlatformProvider interface {
	Platform() string
	Post(ctx context.Context, content string, mediaURLs []string, opts map[string]any) (*SocialPostResult, error)
}

// MultiPlatformProvider implements SocialMediaProvider by routing to native platform providers.
type MultiPlatformProvider struct {
	providers map[string]PlatformProvider
}

// NewMultiPlatformProvider creates a provider that routes to configured native platforms.
func NewMultiPlatformProvider(cfg config.SocialMediaConfig) *MultiPlatformProvider {
	m := &MultiPlatformProvider{
		providers: make(map[string]PlatformProvider),
	}

	// Register configured platforms
	if cfg.Twitter.APIKey != "" && cfg.Twitter.AccessToken != "" {
		m.providers["twitter"] = NewTwitterProvider(cfg.Twitter)
	}
	if cfg.Bluesky.Handle != "" && cfg.Bluesky.AppPassword != "" {
		m.providers["bluesky"] = NewBlueskyProvider(cfg.Bluesky)
	}
	if cfg.LinkedIn.AccessToken != "" && cfg.LinkedIn.AuthorURN != "" {
		m.providers["linkedin"] = NewLinkedInProvider(cfg.LinkedIn)
	}
	if cfg.Facebook.PageAccessToken != "" && cfg.Facebook.PageID != "" {
		m.providers["facebook"] = NewFacebookProvider(cfg.Facebook)
	}

	return m
}

// Name returns the provider name for the SocialMediaProvider interface.
func (m *MultiPlatformProvider) Name() string { return "multi-platform" }

// PlatformCount returns how many platforms are configured.
func (m *MultiPlatformProvider) PlatformCount() int {
	return len(m.providers)
}

// ConfiguredPlatforms returns a list of configured platform names.
func (m *MultiPlatformProvider) ConfiguredPlatforms() []string {
	platforms := make([]string, 0, len(m.providers))
	for name := range m.providers {
		platforms = append(platforms, name)
	}
	return platforms
}

// Post sends the content to each requested platform. Hashtags from the request
// are appended to the content once, then each platform provider receives the
// full text -- individual providers do NOT handle hashtag logic.
func (m *MultiPlatformProvider) Post(ctx context.Context, req SocialPostRequest) ([]SocialPostResult, error) {
	if len(m.providers) == 0 {
		return nil, fmt.Errorf("no social media platforms configured -- add credentials in Settings or config.json")
	}

	// Build content with hashtags
	content := req.Content
	if len(req.Hashtags) > 0 {
		content += " " + strings.Join(req.Hashtags, " ")
	}

	var results []SocialPostResult
	var lastErr error

	for _, platform := range req.Platforms {
		platform = strings.ToLower(platform)

		provider, ok := m.providers[platform]
		if !ok {
			results = append(results, SocialPostResult{
				Platform: platform,
				Status:   "failed",
				Error:    fmt.Sprintf("platform %q not configured -- add credentials in config.json", platform),
			})
			continue
		}

		result, err := provider.Post(ctx, content, req.MediaURLs, nil)
		if err != nil {
			lastErr = err
			results = append(results, SocialPostResult{
				Platform: platform,
				Status:   "failed",
				Error:    err.Error(),
			})
			continue
		}

		results = append(results, *result)
	}

	// If ALL platforms failed, return error
	if len(results) > 0 {
		allFailed := true
		for _, r := range results {
			if r.Status != "failed" {
				allFailed = false
				break
			}
		}
		if allFailed && lastErr != nil {
			return results, fmt.Errorf("all platforms failed: %w", lastErr)
		}
	}

	return results, nil
}

// GetAnalytics is not supported by native APIs in a unified way.
// Each platform has different analytics APIs with varying permission requirements.
func (m *MultiPlatformProvider) GetAnalytics(ctx context.Context, postID, platform string) (*SocialAnalytics, error) {
	return nil, fmt.Errorf("analytics for %s: use the platform's native dashboard -- native API analytics require additional permissions that vary per platform", platform)
}
