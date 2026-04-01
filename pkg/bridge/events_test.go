package bridge

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEventSessionResetConstantAndTokensUsedJSON(t *testing.T) {
	ev := Event{
		Type:       EventSessionReset,
		TokensUsed: 123,
	}

	data, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	jsonText := string(data)
	if !strings.Contains(jsonText, `"event":"session_reset"`) {
		t.Fatalf("expected session_reset event type, got %s", jsonText)
	}
	if !strings.Contains(jsonText, `"tokens_used":123`) {
		t.Fatalf("expected tokens_used field, got %s", jsonText)
	}
}
