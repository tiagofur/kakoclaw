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
	GetAnalytics(ctx context.Context, postID string) (*SocialAnalytics, error)
}

// MultiPlatformProvider implements SocialMediaProvider by routing to native platform providers.
type MultiPlatformProvider struct {
	providers map[string]PlatformProvider
}

// NewMultiPlatformProvider creates a provider that routes to configured native platforms.
// Each account is registered as "platform:alias" (e.g. "twitter:personal", "facebook:brand").
func NewMultiPlatformProvider(cfg config.SocialMediaConfig) *MultiPlatformProvider {
	m := &MultiPlatformProvider{
		providers: make(map[string]PlatformProvider),
	}

	for alias, twCfg := range cfg.Twitter {
		if twCfg.APIKey != "" && twCfg.AccessToken != "" {
			m.providers["twitter:"+alias] = NewTwitterProvider(twCfg)
		}
	}
	for alias, bsCfg := range cfg.Bluesky {
		if bsCfg.Handle != "" && bsCfg.AppPassword != "" {
			m.providers["bluesky:"+alias] = NewBlueskyProvider(bsCfg)
		}
	}
	for alias, liCfg := range cfg.LinkedIn {
		if liCfg.AccessToken != "" && liCfg.AuthorURN != "" {
			m.providers["linkedin:"+alias] = NewLinkedInProvider(liCfg)
		}
	}
	for alias, fbCfg := range cfg.Facebook {
		if fbCfg.PageAccessToken != "" && fbCfg.PageID != "" {
			m.providers["facebook:"+alias] = NewFacebookProvider(fbCfg)
		}
	}
	for alias, igCfg := range cfg.Instagram {
		if igCfg.AccessToken != "" && igCfg.AccountID != "" {
			m.providers["instagram:"+alias] = NewInstagramProvider(igCfg)
		}
	}
	for alias, tkCfg := range cfg.TikTok {
		if tkCfg.AccessToken != "" {
			m.providers["tiktok:"+alias] = NewTikTokProvider(tkCfg)
		}
	}

	return m
}

// Name returns the provider name for the SocialMediaProvider interface.
func (m *MultiPlatformProvider) Name() string { return "multi-platform" }

// PlatformCount returns how many platforms are configured.
func (m *MultiPlatformProvider) PlatformCount() int {
	return len(m.providers)
}

// ConfiguredPlatforms returns a list of configured platform:alias keys.
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
	var opts map[string]any
	if req.ScheduleAt != nil {
		opts = map[string]any{"schedule_at": req.ScheduleAt.Unix()}
	}

	for _, platform := range req.Platforms {
		platform = strings.ToLower(platform)

		provider, ok := m.providers[platform]
		if !ok {
			// Fallback: if no alias given (e.g. "twitter"), use first matching provider
			if !strings.Contains(platform, ":") {
				for key, p := range m.providers {
					if strings.HasPrefix(key, platform+":") {
						provider = p
						ok = true
						break
					}
				}
			}
		}
		if !ok {
			results = append(results, SocialPostResult{
				Platform: platform,
				Status:   "failed",
				Error:    fmt.Sprintf("platform %q not configured -- add credentials in Settings or config.json", platform),
			})
			continue
		}

		result, err := provider.Post(ctx, content, req.MediaURLs, opts)
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

func (m *MultiPlatformProvider) GetAnalytics(ctx context.Context, postID, platform string) (*SocialAnalytics, error) {
	if len(m.providers) == 0 {
		return nil, fmt.Errorf("no social media platforms configured -- add credentials in Settings or config.json")
	}

	platform = strings.ToLower(platform)
	provider, ok := m.providers[platform]
	if !ok && !strings.Contains(platform, ":") {
		for key, p := range m.providers {
			if strings.HasPrefix(key, platform+":") {
				provider = p
				ok = true
				break
			}
		}
	}
	if !ok {
		return nil, fmt.Errorf("platform %q not configured -- add credentials in Settings or config.json", platform)
	}

	return provider.GetAnalytics(ctx, postID)
}
