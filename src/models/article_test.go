package models

import "testing"

func TestEscapeLikePattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "plain text", input: "golang", expected: "golang"},
		{name: "percent", input: "100%", expected: "100\\%"},
		{name: "underscore", input: "foo_bar", expected: "foo\\_bar"},
		{name: "backslash", input: `c:\\temp`, expected: `c:\\\\temp`},
		{name: "mixed", input: `50%_off\\now`, expected: `50\%\_off\\\\now`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeLikePattern(tt.input); got != tt.expected {
				t.Fatalf("escapeLikePattern(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
