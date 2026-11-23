package main

import "testing"

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://EXAMPLE.COM/about",
			expected: "https://example.com/about",
		},
		{
			input:    "https://example.com:443/about",
			expected: "https://example.com/about",
		},
		{
			input:    "https://example.com/about#team",
			expected: "https://example.com/about",
		},
		{
			input:    "https://example.com",
			expected: "https://example.com/",
		},
		{
			input:    "http://example.com:80/",
			expected: "http://example.com/",
		},
	}

	for _, tt := range tests {
		actual, err := normalizeURL(tt.input)
		if err != nil {
			t.Fatalf("normalizeURL(%q) returned error: %v", tt.input, err)
		}

		if actual != tt.expected {
			t.Errorf(
				"normalizeURL(%q) = %q, want %q",
				tt.input,
				actual,
				tt.expected,
			)
		}
	}
}
