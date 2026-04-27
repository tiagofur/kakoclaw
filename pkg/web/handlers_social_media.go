package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
)

// handleTestSocialMediaConnection verifies credentials for a social media platform account.
// Route: POST /api/v1/social-media/{platform}/{alias}/test
func (s *Server) handleTestSocialMediaConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract platform and alias from URL: /api/v1/social-media/{platform}/{alias}/test
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "invalid path — expected /api/v1/social-media/{platform}/{alias}/test", http.StatusBadRequest)
		return
	}
	platform := parts[3]
	alias := parts[4]

	globalCfg := s.fullConfig
	if globalCfg == nil {
		http.Error(w, "config not available", http.StatusInternalServerError)
		return
	}
	userCfg, _ := config.LoadConfigForUser(userUUID)
	mergedCfg := config.MergeConfigs(globalCfg, userCfg)

	var testErr error

	switch platform {
	case "twitter":
		cfg, found := mergedCfg.Tools.SocialMedia.Twitter[alias]
		if !found {
			http.Error(w, fmt.Sprintf("twitter account %q not configured", alias), http.StatusNotFound)
			return
		}
		testErr = testTwitterConnection(cfg)
	case "bluesky":
		cfg, found := mergedCfg.Tools.SocialMedia.Bluesky[alias]
		if !found {
			http.Error(w, fmt.Sprintf("bluesky account %q not configured", alias), http.StatusNotFound)
			return
		}
		testErr = testBlueskyConnection(cfg)
	case "linkedin":
		cfg, found := mergedCfg.Tools.SocialMedia.LinkedIn[alias]
		if !found {
			http.Error(w, fmt.Sprintf("linkedin account %q not configured", alias), http.StatusNotFound)
			return
		}
		testErr = testLinkedInConnection(cfg)
	case "facebook":
		cfg, found := mergedCfg.Tools.SocialMedia.Facebook[alias]
		if !found {
			http.Error(w, fmt.Sprintf("facebook account %q not configured", alias), http.StatusNotFound)
			return
		}
		testErr = testFacebookConnection(cfg)
	default:
		http.Error(w, "unknown platform: "+platform, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if testErr != nil {
		logger.WarnCF("web", "Social media test connection failed", map[string]interface{}{
			"platform":  platform,
			"alias":     alias,
			"user_uuid": userUUID,
			"error":     testErr.Error(),
		})
		writeJSONResponse(w, map[string]interface{}{"ok": false, "error": testErr.Error()})
		return
	}

	logger.InfoCF("web", "Social media test connection succeeded", map[string]interface{}{
		"platform":  platform,
		"alias":     alias,
		"user_uuid": userUUID,
	})
	writeJSONResponse(w, map[string]interface{}{"ok": true})
}

// testTwitterConnection verifies Twitter/X credentials by calling GET /2/users/me
// using OAuth 1.0a authentication.
func testTwitterConnection(cfg config.TwitterSocialConfig) error {
	if cfg.APIKey == "" || cfg.APISecret == "" || cfg.AccessToken == "" || cfg.AccessTokenSecret == "" {
		return fmt.Errorf("Twitter credentials not fully configured")
	}

	rawURL := "https://api.twitter.com/2/users/me"
	authHeader := socialOAuthSign("GET", rawURL, nil,
		cfg.APIKey, cfg.APISecret, cfg.AccessToken, cfg.AccessTokenSecret)

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid credentials (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

// testBlueskyConnection verifies Bluesky credentials by attempting a session creation.
func testBlueskyConnection(cfg config.BlueskySocialConfig) error {
	if cfg.Handle == "" || cfg.AppPassword == "" {
		return fmt.Errorf("Bluesky handle and app password are required")
	}
	pdsURL := cfg.PDSURL
	if pdsURL == "" {
		pdsURL = "https://bsky.social"
	}

	bodyBytes, _ := json.Marshal(map[string]string{
		"identifier": cfg.Handle,
		"password":   cfg.AppPassword,
	})
	req, err := http.NewRequest("POST", pdsURL+"/xrpc/com.atproto.server.createSession", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid handle or app password")
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

// testLinkedInConnection verifies LinkedIn credentials via GET /v2/userinfo.
func testLinkedInConnection(cfg config.LinkedInSocialConfig) error {
	if cfg.AccessToken == "" {
		return fmt.Errorf("LinkedIn access token is required")
	}

	req, err := http.NewRequest("GET", "https://api.linkedin.com/v2/userinfo", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid or expired access token")
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

// testFacebookConnection verifies Facebook credentials via GET /me.
func testFacebookConnection(cfg config.FacebookSocialConfig) error {
	if cfg.PageAccessToken == "" {
		return fmt.Errorf("Facebook page access token is required")
	}

	reqURL := "https://graph.facebook.com/me?access_token=" + url.QueryEscape(cfg.PageAccessToken)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		// Facebook returns error details in JSON
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Error.Message != "" {
			return fmt.Errorf("%s", errResp.Error.Message)
		}
		return fmt.Errorf("API error %d", resp.StatusCode)
	}
	return nil
}

// socialOAuthSign generates an OAuth 1.0a Authorization header for a request.
// Used by testTwitterConnection to verify credentials without importing pkg/tools.
func socialOAuthSign(method, rawURL string, extraParams map[string]string,
	apiKey, apiSecret, accessToken, accessTokenSecret string) string {

	nonce := socialGenerateNonce()
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	oauthParams := map[string]string{
		"oauth_consumer_key":     apiKey,
		"oauth_nonce":            nonce,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        timestamp,
		"oauth_token":            accessToken,
		"oauth_version":          "1.0",
	}

	allParams := make(map[string]string)
	for k, v := range oauthParams {
		allParams[k] = v
	}
	for k, v := range extraParams {
		allParams[k] = v
	}

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

	signatureBase := strings.ToUpper(method) + "&" + url.QueryEscape(rawURL) + "&" + url.QueryEscape(paramString)
	signingKey := url.QueryEscape(apiSecret) + "&" + url.QueryEscape(accessTokenSecret)

	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(signatureBase))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	oauthParams["oauth_signature"] = signature

	headerParts := make([]string, 0, len(oauthParams))
	for k, v := range oauthParams {
		headerParts = append(headerParts, fmt.Sprintf(`%s="%s"`, url.QueryEscape(k), url.QueryEscape(v)))
	}
	sort.Strings(headerParts)
	return "OAuth " + strings.Join(headerParts, ", ")
}

func socialGenerateNonce() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
