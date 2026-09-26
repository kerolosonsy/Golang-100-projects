package main

import (
	"os"
	"strings"
	"testing"
)

func TestGetEnvStatus(t *testing.T) {
	t.Setenv("USER", "Keroonsy")
	t.Setenv("HOME", "")
	os.Unsetenv("PATH")
	tests := []struct {
		name      string
		userInput string
		expected  string
	}{
		{"V2: Allowed and Present", "USER", "Keroonsy"},
		{"V3: Allowed but Absent", "PATH", "(unset)"},
		{"V4: Allowed but Empty", "HOME", "(empty)"},
		{"V9: Not in Allowlist", "SECRET_KEY", "(not allowed)"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			output := getEnvStatus(tc.userInput)
			printedText := strings.TrimSpace(output)
			if !strings.Contains(printedText, tc.expected) {
				t.Errorf("failed:\nExpected to contain: %q\nActually got: %q", tc.expected, printedText)
			}
		})
	}
}
