package tools

import "testing"

func TestParseFromAddressSupportsDisplayName(t *testing.T) {
	envelope, header, err := parseFromAddress("KakoClaw <bot@example.com>", "")
	if err != nil {
		t.Fatalf("parseFromAddress returned error: %v", err)
	}
	if envelope != "bot@example.com" {
		t.Fatalf("expected envelope bot@example.com, got %q", envelope)
	}
	if header != "KakoClaw <bot@example.com>" && header != "\"KakoClaw\" <bot@example.com>" {
		t.Fatalf("unexpected header: %q", header)
	}
}

func TestNormalizedSMTPPasswordForGmail(t *testing.T) {
	got := normalizedSMTPPassword("smtp.gmail.com", "abcd efgh ijkl mnop")
	if got != "abcdefghijklmnop" {
		t.Fatalf("expected spaces to be removed, got %q", got)
	}
}

func TestNormalizedSMTPPasswordForNonGmail(t *testing.T) {
	original := "abcd efgh"
	got := normalizedSMTPPassword("smtp.example.com", original)
	if got != original {
		t.Fatalf("expected password to stay unchanged, got %q", got)
	}
}
