package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
)

// TTSSynthesizer converts text to speech audio using an external API.
type TTSSynthesizer struct {
	apiKey     string
	apiBase    string
	voice      string
	model      string
	httpClient *http.Client
}

// NewTTSSynthesizer creates a TTS synthesizer from config.
func NewTTSSynthesizer(cfg config.TTSConfig) *TTSSynthesizer {
	voice := cfg.Voice
	if voice == "" {
		voice = "alloy"
	}
	model := cfg.Model
	if model == "" {
		model = "tts-1"
	}
	apiBase := "https://api.openai.com/v1"

	logger.DebugCF("voice", "Creating TTS synthesizer", map[string]interface{}{
		"provider": cfg.Provider,
		"voice":    voice,
		"model":    model,
	})

	return &TTSSynthesizer{
		apiKey:  cfg.APIKey,
		apiBase: apiBase,
		voice:   voice,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// IsAvailable returns true if the synthesizer has an API key configured.
func (t *TTSSynthesizer) IsAvailable() bool {
	return t.apiKey != ""
}

// Synthesize converts text to speech and returns raw MP3 audio bytes.
func (t *TTSSynthesizer) Synthesize(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}

	logger.InfoCF("voice", "Starting TTS synthesis", map[string]interface{}{
		"text_length": len(text),
		"voice":       t.voice,
		"model":       t.model,
	})

	reqBody := map[string]interface{}{
		"model": t.model,
		"input": text,
		"voice": t.voice,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := t.apiBase + "/audio/speech"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.ErrorCF("voice", "TTS API error", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(body),
		})
		return nil, fmt.Errorf("TTS API error (status %d): %s", resp.StatusCode, string(body))
	}

	logger.InfoCF("voice", "TTS synthesis completed", map[string]interface{}{
		"audio_size_bytes": len(body),
	})

	return body, nil
}
