package attachments

import "testing"

func TestIsAllowedContentType(t *testing.T) {
	tests := []struct {
		ct      string
		allowed bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"application/pdf", true},
		{"text/plain", true},
		{"text/csv", true},
		{"application/json", true},
		{"application/octet-stream", false},
		{"video/mp4", false},
		{"application/x-executable", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ct, func(t *testing.T) {
			result := IsAllowedContentType(tt.ct)
			if result != tt.allowed {
				t.Errorf("IsAllowedContentType(%q) = %v, want %v", tt.ct, result, tt.allowed)
			}
		})
	}
}
