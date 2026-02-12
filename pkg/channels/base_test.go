package channels

import (
	"testing"
)

func TestBaseChannel_IsAllowed(t *testing.T) {
	tests := []struct {
		name      string
		allowList []string
		senderID  string
		want      bool
	}{
		{
			name:      "empty allowlist allows all",
			allowList: []string{},
			senderID:  "123456",
			want:      true,
		},
		{
			name:      "numeric ID match",
			allowList: []string{"123456"},
			senderID:  "123456",
			want:      true,
		},
		{
			name:      "numeric ID no match",
			allowList: []string{"123456"},
			senderID:  "999999",
			want:      false,
		},
		{
			name:      "username with @ match",
			allowList: []string{"@testuser"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "username without @ match",
			allowList: []string{"testuser"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "username no match - different user",
			allowList: []string{"@otheruser"},
			senderID:  "123456|testuser",
			want:      false,
		},
		{
			name:      "numeric ID in compound senderID",
			allowList: []string{"123456"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "multiple entries - numeric match",
			allowList: []string{"123456", "@admin"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "multiple entries - username match",
			allowList: []string{"999999", "@testuser"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "issue #62 - username without @ should work",
			allowList: []string{"testuser"},
			senderID:  "123456|testuser",
			want:      true,
		},
		{
			name:      "issue #62 - @username should work",
			allowList: []string{"@testuser"},
			senderID:  "123456|testuser",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBaseChannel("test", nil, nil, tt.allowList)
			got := bc.IsAllowed(tt.senderID)
			if got != tt.want {
				t.Errorf("IsAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
