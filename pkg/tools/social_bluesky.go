package tools

// BlueskyProvider posts to Bluesky via the AT Protocol.
// Auth: handle + app password -> JWT session token
// Posting: com.atproto.repo.createRecord
// Images: com.atproto.repo.uploadBlob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
)

// BlueskyProvider implements PlatformProvider for the Bluesky social network
// using the AT Protocol (atproto) for authentication and posting.
type BlueskyProvider struct {
	handle      string
	appPassword string
	pdsURL      string // default: "https://bsky.social"
	httpClient  *http.Client
	// session state
	accessJwt string
	did       string // user's DID (did:plc:xxx)
}

// NewBlueskyProvider creates a BlueskyProvider from the given config.
func NewBlueskyProvider(cfg config.BlueskySocialConfig) *BlueskyProvider {
	pdsURL := cfg.PDSURL
	if pdsURL == "" {
		pdsURL = "https://bsky.social"
	}
	return &BlueskyProvider{
		handle:      cfg.Handle,
		appPassword: cfg.AppPassword,
		pdsURL:      pdsURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Platform returns the platform identifier.
func (b *BlueskyProvider) Platform() string { return "bluesky" }

func (b *BlueskyProvider) GetAnalytics(ctx context.Context, postID string) (*SocialAnalytics, error) {
	if err := b.createSession(ctx); err != nil {
		return nil, fmt.Errorf("bluesky authentication failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.pdsURL+"/xrpc/app.bsky.feed.getPostThread?uri="+url.QueryEscape(postID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+b.accessJwt)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bluesky analytics failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Thread struct {
			Post struct {
				URI         string `json:"uri"`
				LikeCount   int    `json:"likeCount"`
				ReplyCount  int    `json:"replyCount"`
				RepostCount int    `json:"repostCount"`
			} `json:"post"`
		} `json:"thread"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse bluesky analytics response: %w", err)
	}

	post := result.Thread.Post
	if post.URI != "" {
		postID = post.URI
	}
	return &SocialAnalytics{
		PostID:   postID,
		Platform: b.Platform(),
		Likes:    post.LikeCount,
		Comments: post.ReplyCount,
		Shares:   post.RepostCount,
	}, nil
}

// createSession authenticates with the PDS and obtains a JWT access token.
func (b *BlueskyProvider) createSession(ctx context.Context) error {
	body := map[string]string{
		"identifier": b.handle,
		"password":   b.appPassword,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal session body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.pdsURL+"/xrpc/com.atproto.server.createSession", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bluesky auth failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var session struct {
		AccessJwt string `json:"accessJwt"`
		DID       string `json:"did"`
	}
	if err := json.Unmarshal(respBody, &session); err != nil {
		return err
	}

	b.accessJwt = session.AccessJwt
	b.did = session.DID
	return nil
}

// uploadBlob uploads image data to the PDS and returns the blob reference
// suitable for embedding in a post record.
func (b *BlueskyProvider) uploadBlob(ctx context.Context, imageData []byte, mimeType string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", b.pdsURL+"/xrpc/com.atproto.repo.uploadBlob", bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+b.accessJwt)
	req.Header.Set("Content-Type", mimeType)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("blob upload failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Blob map[string]any `json:"blob"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result.Blob, nil
}

// Post creates a Bluesky post with optional image attachments (max 4).
// mediaURLs are treated as local file paths.
func (b *BlueskyProvider) Post(ctx context.Context, content string, mediaURLs []string, opts map[string]any) (*SocialPostResult, error) {
	if _, ok := scheduleUnix(opts); ok {
		return nil, ErrSchedulingNotSupported
	}

	// Authenticate
	if err := b.createSession(ctx); err != nil {
		return nil, fmt.Errorf("bluesky authentication failed: %w", err)
	}

	// Build post record
	record := map[string]any{
		"$type":     "app.bsky.feed.post",
		"text":      content,
		"createdAt": time.Now().UTC().Format(time.RFC3339),
	}

	// Upload images if any (max 4 per Bluesky)
	if len(mediaURLs) > 0 {
		var images []map[string]any
		for i, mediaURL := range mediaURLs {
			if i >= 4 {
				break // Bluesky limit
			}

			// Read image from local path
			imageData, err := os.ReadFile(mediaURL)
			if err != nil {
				// Skip files that can't be read
				continue
			}

			mimeType := "image/png" // default
			if len(imageData) > 2 && imageData[0] == 0xFF && imageData[1] == 0xD8 {
				mimeType = "image/jpeg"
			}

			blob, err := b.uploadBlob(ctx, imageData, mimeType)
			if err != nil {
				continue
			}

			images = append(images, map[string]any{
				"alt":   "",
				"image": blob,
			})
		}
		if len(images) > 0 {
			record["embed"] = map[string]any{
				"$type":  "app.bsky.embed.images",
				"images": images,
			}
		}
	}

	// Create the post record
	createBody := map[string]any{
		"repo":       b.did,
		"collection": "app.bsky.feed.post",
		"record":     record,
	}

	jsonBody, err := json.Marshal(createBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.pdsURL+"/xrpc/com.atproto.repo.createRecord", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+b.accessJwt)
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bluesky post failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse post response: %w", err)
	}

	// Convert AT URI to web URL: at://did/app.bsky.feed.post/rkey -> https://bsky.app/profile/handle/post/rkey
	postURL := ""
	rkey := splitATURI(result.URI)
	if rkey != "" {
		postURL = fmt.Sprintf("https://bsky.app/profile/%s/post/%s", b.handle, rkey)
	}

	return &SocialPostResult{
		Platform: "bluesky",
		PostID:   result.URI,
		PostURL:  postURL,
		Status:   "posted",
	}, nil
}

// splitATURI extracts the rkey (last segment) from an AT Protocol URI.
// URI format: at://did:plc:xxx/app.bsky.feed.post/rkey
func splitATURI(uri string) string {
	parts := strings.Split(uri, "/")
	if len(parts) >= 5 {
		return parts[len(parts)-1]
	}
	return ""
}
