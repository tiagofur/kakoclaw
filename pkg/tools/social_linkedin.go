package tools

// LinkedInProvider posts to LinkedIn via the REST API.
// Auth: OAuth 2.0 Bearer token (user obtains externally)
// Posting: POST https://api.linkedin.com/rest/posts
// Images: POST https://api.linkedin.com/rest/images?action=initializeUpload

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
)

// LinkedInProvider implements PlatformProvider for LinkedIn using the
// Community Management REST API with OAuth 2.0 Bearer token authentication.
type LinkedInProvider struct {
	accessToken string
	authorURN   string // e.g. "urn:li:person:xxxxx" or "urn:li:organization:xxxxx"
	httpClient  *http.Client
}

// NewLinkedInProvider creates a LinkedInProvider from the given config.
func NewLinkedInProvider(cfg config.LinkedInSocialConfig) *LinkedInProvider {
	return &LinkedInProvider{
		accessToken: cfg.AccessToken,
		authorURN:   cfg.AuthorURN,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Platform returns the platform identifier.
func (l *LinkedInProvider) Platform() string { return "linkedin" }

// uploadImage initializes an image upload, uploads the binary data, and returns
// the image URN that can be embedded in a post.
func (l *LinkedInProvider) uploadImage(ctx context.Context, imagePath string) (string, error) {
	// Step 1: Initialize upload
	initBody := map[string]any{
		"initializeUploadRequest": map[string]any{
			"owner": l.authorURN,
		},
	}
	jsonBody, err := json.Marshal(initBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal init body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.linkedin.com/rest/images?action=initializeUpload", bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	l.setHeaders(req)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("linkedin image init failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var initResp struct {
		Value struct {
			UploadURL string `json:"uploadUrl"`
			Image     string `json:"image"`
		} `json:"value"`
	}
	if err := json.Unmarshal(respBody, &initResp); err != nil {
		return "", fmt.Errorf("failed to parse init response: %w", err)
	}

	// Step 2: Upload the image data
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	uploadReq, err := http.NewRequestWithContext(ctx, "PUT", initResp.Value.UploadURL, bytes.NewReader(imageData))
	if err != nil {
		return "", err
	}
	uploadReq.Header.Set("Authorization", "Bearer "+l.accessToken)

	uploadResp, err := l.httpClient.Do(uploadReq)
	if err != nil {
		return "", err
	}
	defer uploadResp.Body.Close()
	io.ReadAll(uploadResp.Body) // drain

	if uploadResp.StatusCode < 200 || uploadResp.StatusCode >= 300 {
		return "", fmt.Errorf("linkedin image upload failed (status %d)", uploadResp.StatusCode)
	}

	return initResp.Value.Image, nil // returns urn:li:image:xxx
}

// setHeaders applies the standard LinkedIn REST API headers to a request.
func (l *LinkedInProvider) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+l.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")
	// LinkedIn-Version: YYYYMM format
	req.Header.Set("LinkedIn-Version", time.Now().Format("200601"))
}

// Post creates a LinkedIn post. mediaURLs are treated as local file paths;
// only the first image is uploaded (LinkedIn single-image posts).
func (l *LinkedInProvider) Post(ctx context.Context, content string, mediaURLs []string, opts map[string]any) (*SocialPostResult, error) {
	postBody := map[string]any{
		"author":     l.authorURN,
		"commentary": content,
		"visibility": "PUBLIC",
		"distribution": map[string]any{
			"feedDistribution":               "MAIN_FEED",
			"targetEntities":                 []any{},
			"thirdPartyDistributionChannels": []any{},
		},
		"lifecycleState":            "PUBLISHED",
		"isReshareDisabledByAuthor": false,
	}

	// Upload first image if provided
	if len(mediaURLs) > 0 {
		imageURN, err := l.uploadImage(ctx, mediaURLs[0])
		if err == nil {
			postBody["content"] = map[string]any{
				"media": map[string]any{
					"id": imageURN,
				},
			}
		}
	}

	jsonBody, err := json.Marshal(postBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal post body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.linkedin.com/rest/posts", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	l.setHeaders(req)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))

	// LinkedIn returns 201 Created on success
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linkedin post failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Post ID from x-restli-id header
	postID := resp.Header.Get("x-restli-id")

	return &SocialPostResult{
		Platform: "linkedin",
		PostID:   postID,
		PostURL:  "", // LinkedIn doesn't return a direct URL in the response
		Status:   "posted",
	}, nil
}
