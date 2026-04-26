package tools

// TwitterProvider posts tweets via Twitter/X API v2.
// Auth: OAuth 1.0a (API Key + Secret + Access Token + Access Token Secret)
// Posting: POST https://api.twitter.com/2/tweets
// Media: POST https://upload.twitter.com/1.1/media/upload.json (multipart)

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
)

// TwitterProvider implements PlatformProvider for Twitter/X using
// OAuth 1.0a authentication and the v2 API for posting.
type TwitterProvider struct {
	apiKey            string
	apiSecret         string
	accessToken       string
	accessTokenSecret string
	httpClient        *http.Client
}

// NewTwitterProvider creates a TwitterProvider from the given config.
func NewTwitterProvider(cfg config.TwitterSocialConfig) *TwitterProvider {
	return &TwitterProvider{
		apiKey:            cfg.APIKey,
		apiSecret:         cfg.APISecret,
		accessToken:       cfg.AccessToken,
		accessTokenSecret: cfg.AccessTokenSecret,
		httpClient:        &http.Client{Timeout: 30 * time.Second},
	}
}

// Platform returns the platform identifier.
func (t *TwitterProvider) Platform() string { return "twitter" }

func (t *TwitterProvider) GetAnalytics(ctx context.Context, postID string) (*SocialAnalytics, error) {
	analyticsURL := fmt.Sprintf("https://api.twitter.com/2/tweets/%s?tweet.fields=public_metrics", url.PathEscape(postID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, analyticsURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", t.oauthSign(http.MethodGet, fmt.Sprintf("https://api.twitter.com/2/tweets/%s", url.PathEscape(postID)), map[string]string{"tweet.fields": "public_metrics"}))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("twitter analytics failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			PublicMetrics struct {
				LikeCount       int `json:"like_count"`
				ReplyCount      int `json:"reply_count"`
				RetweetCount    int `json:"retweet_count"`
				QuoteCount      int `json:"quote_count"`
				ImpressionCount int `json:"impression_count"`
			} `json:"public_metrics"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse twitter analytics response: %w", err)
	}

	metrics := result.Data.PublicMetrics
	return &SocialAnalytics{
		PostID:      postID,
		Platform:    t.Platform(),
		Likes:       metrics.LikeCount,
		Comments:    metrics.ReplyCount,
		Shares:      metrics.RetweetCount + metrics.QuoteCount,
		Impressions: metrics.ImpressionCount,
	}, nil
}

// oauthSign generates the OAuth 1.0a Authorization header value for a request.
func (t *TwitterProvider) oauthSign(method, rawURL string, params map[string]string) string {
	nonce := generateNonce()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	oauthParams := map[string]string{
		"oauth_consumer_key":     t.apiKey,
		"oauth_nonce":            nonce,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        timestamp,
		"oauth_token":            t.accessToken,
		"oauth_version":          "1.0",
	}

	// Merge all params for signature base
	allParams := make(map[string]string)
	for k, v := range oauthParams {
		allParams[k] = v
	}
	for k, v := range params {
		allParams[k] = v
	}

	// Sort and encode parameters
	keys := make([]string, 0, len(allParams))
	for k := range allParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	paramPairs := make([]string, 0, len(keys))
	for _, k := range keys {
		paramPairs = append(paramPairs, url.QueryEscape(k)+"="+url.QueryEscape(allParams[k]))
	}
	paramString := strings.Join(paramPairs, "&")

	// Signature base string
	signatureBase := strings.ToUpper(method) + "&" + url.QueryEscape(rawURL) + "&" + url.QueryEscape(paramString)

	// Signing key
	signingKey := url.QueryEscape(t.apiSecret) + "&" + url.QueryEscape(t.accessTokenSecret)

	// HMAC-SHA1
	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(signatureBase))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	oauthParams["oauth_signature"] = signature

	// Build Authorization header
	headerParts := make([]string, 0, len(oauthParams))
	for k, v := range oauthParams {
		headerParts = append(headerParts, fmt.Sprintf(`%s="%s"`, url.QueryEscape(k), url.QueryEscape(v)))
	}
	sort.Strings(headerParts) // Sort for deterministic output
	return "OAuth " + strings.Join(headerParts, ", ")
}

// generateNonce produces a random 32-character alphanumeric string for OAuth.
func generateNonce() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// uploadMedia uploads an image file to Twitter's v1.1 media upload endpoint
// and returns the media_id_string for use in tweet creation.
func (t *TwitterProvider) uploadMedia(ctx context.Context, imagePath string) (string, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	// Build multipart form with base64-encoded media_data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("media_data", "image.png")
	if err != nil {
		return "", err
	}

	// Twitter expects base64 encoded media_data
	encoded := base64.StdEncoding.EncodeToString(imageData)
	if _, err := part.Write([]byte(encoded)); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	uploadURL := "https://upload.twitter.com/1.1/media/upload.json"
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, &buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", t.oauthSign("POST", uploadURL, nil))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("twitter media upload failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		MediaIDString string `json:"media_id_string"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	return result.MediaIDString, nil
}

// Post creates a tweet with optional media attachments (max 4).
// mediaURLs are treated as local file paths that get uploaded first.
func (t *TwitterProvider) Post(ctx context.Context, content string, mediaURLs []string, opts map[string]any) (*SocialPostResult, error) {
	if _, ok := scheduleUnix(opts); ok {
		return nil, ErrSchedulingNotSupported
	}

	tweetBody := map[string]any{
		"text": content,
	}

	// Upload media if any
	if len(mediaURLs) > 0 {
		var mediaIDs []string
		for _, mediaURL := range mediaURLs {
			mediaID, err := t.uploadMedia(ctx, mediaURL)
			if err != nil {
				continue // Skip failed uploads
			}
			mediaIDs = append(mediaIDs, mediaID)
			if len(mediaIDs) >= 4 {
				break // Twitter limit
			}
		}
		if len(mediaIDs) > 0 {
			tweetBody["media"] = map[string]any{
				"media_ids": mediaIDs,
			}
		}
	}

	jsonBody, err := json.Marshal(tweetBody)
	if err != nil {
		return nil, err
	}

	tweetURL := "https://api.twitter.com/2/tweets"
	req, err := http.NewRequestWithContext(ctx, "POST", tweetURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", t.oauthSign("POST", tweetURL, nil))

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("twitter post failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tweet response: %w", err)
	}

	postURL := fmt.Sprintf("https://x.com/i/status/%s", result.Data.ID)

	return &SocialPostResult{
		Platform: "twitter",
		PostID:   result.Data.ID,
		PostURL:  postURL,
		Status:   "posted",
	}, nil
}
