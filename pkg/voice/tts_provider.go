package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/sipeed/makoclaw/pkg/logger"
)

// TTSProvider defines the interface for text-to-speech synthesis.
type TTSProvider interface {
	Synthesize(ctx context.Context, text string) ([]byte, error)
	IsAvailable() bool
}

// ElevenLabsProvider implements TTSProvider using the ElevenLabs API.
type ElevenLabsProvider struct {
	apiKey     string
	voiceID    string
	apiBase    string
	httpClient *http.Client
}

// NewElevenLabsProvider creates a new ElevenLabs TTS provider.
func NewElevenLabsProvider(apiKey, voiceID string) *ElevenLabsProvider {
	if voiceID == "" {
		voiceID = "21m00Tcm4TlvDq8ikWAM" // default Rachel voice
	}
	return &ElevenLabsProvider{
		apiKey:  apiKey,
		voiceID: voiceID,
		apiBase: "https://api.elevenlabs.io",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// IsAvailable returns true if the provider has an API key configured.
func (p *ElevenLabsProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// Synthesize converts text to speech using the ElevenLabs API.
func (p *ElevenLabsProvider) Synthesize(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}

	url := fmt.Sprintf("%s/v1/text-to-speech/%s", p.apiBase, p.voiceID)

	reqBody := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_monolingual_v1",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", p.apiKey)

	logger.DebugCF("voice", "Sending ElevenLabs TTS request", map[string]interface{}{
		"text_length": len(text),
		"voice_id":    p.voiceID,
	})

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.ErrorCF("voice", "ElevenLabs API error", map[string]interface{}{
			"status_code": resp.StatusCode,
			"response":    string(body),
		})
		return nil, fmt.Errorf("ElevenLabs API error (status %d): %s", resp.StatusCode, string(body))
	}

	logger.InfoCF("voice", "ElevenLabs TTS synthesis completed", map[string]interface{}{
		"audio_size_bytes": len(body),
	})

	return body, nil
}

// SystemTTSProvider implements TTSProvider using the OS text-to-speech engine.
// On macOS it uses the "say" command; on other platforms it's a no-op fallback.
type SystemTTSProvider struct{}

// NewSystemTTSProvider creates a new system TTS provider.
func NewSystemTTSProvider() *SystemTTSProvider {
	return &SystemTTSProvider{}
}

// IsAvailable always returns true — the system TTS is the last-resort fallback.
func (p *SystemTTSProvider) IsAvailable() bool {
	return true
}

// Synthesize speaks the text using the OS TTS engine.
// Audio is played directly; returned bytes are empty.
func (p *SystemTTSProvider) Synthesize(ctx context.Context, text string) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}

	cmd := exec.CommandContext(ctx, "say", text)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("system TTS failed: %w", err)
	}

	return []byte{}, nil
}
