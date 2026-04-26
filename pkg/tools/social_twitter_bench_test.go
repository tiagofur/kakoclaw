package tools

import (
	"testing"
)

func BenchmarkOauthSign(b *testing.B) {
	t := &TwitterProvider{
		apiKey:            "key",
		apiSecret:         "secret",
		accessToken:       "token",
		accessTokenSecret: "token_secret",
	}
	params := map[string]string{
		"status": "Hello World",
		"lat":    "37.7821120598956",
		"long":   "-122.400612831116",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.oauthSign("POST", "https://api.twitter.com/1.1/statuses/update.json", params)
	}
}
