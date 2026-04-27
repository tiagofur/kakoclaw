package web

import (
	"testing"
)

func TestSocialGenerateNonce(t *testing.T) {
	nonces := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		nonce := socialGenerateNonce()
		if len(nonce) != 32 {
			t.Errorf("expected nonce length 32, got %d", len(nonce))
		}
		if nonces[nonce] {
			t.Errorf("duplicate nonce generated: %s", nonce)
		}
		nonces[nonce] = true
	}
}
