package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestPlayGame(t *testing.T) {
	tests := []struct {
		name               string
		secret             int
		input              string
		expectedSubstrings []string
	}{
		{"Win first try", 10, "10\n", []string{"Congratulations! You Win! The number was 10. You used 1 trial"}},

		{"Win last try", 42, "1\n52\n3\n4\n42\n", []string{"Guess higher", "Guess lower", "Guess higher", "Guess higher", "Congratulations"}},

		{"Lose all valid tries", 50, "10\n20\n30\n40\n90\n", []string{"Guess higher", "Guess higher", "Guess higher", "Guess higher", "Guess lower", "Sorry, You lose"}},

		{"String input then win", 30, "hello\n30\n", []string{"Invalid input", "Congratulations"}},

		{"Out of range upper", 70, "150\n70\n", []string{"Out of range", "Congratulations"}},

		{"Out of range lower (negative)", 15, "-10\n15\n", []string{"Out of range", "Congratulations"}},

		{"Empty input then win", 25, "\n25\n", []string{"Invalid input", "Congratulations"}},

		{"Mixed invalid inputs lose", 80, "abc\n-5\n200\n\n99\n", []string{"Invalid input", "Out of range", "Out of range", "Invalid input", "Guess lower", "Sorry, You lose"}},

		{"Immediate EOF", 42, "", []string{"Goodbye!"}},

		{"EOF after 2 tries", 42, "10\n20\n", []string{"Guess higher", "Guess higher", "Goodbye!"}},

		{"Duplicate wrong guesses", 99, "1\n1\n1\n1\n1\n", []string{"Guess higher", "Guess higher", "Guess higher", "Guess higher", "Guess higher", "Sorry, You lose"}},
	}
	for _, tc := range tests {
		var output strings.Builder
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			playGame(reader, &output, func() int { return tc.secret })
			printedText := strings.TrimSpace(output.String())

			var actualLines []string
			if printedText != "" {
				for _, line := range strings.Split(printedText, "\n") {
					actualLines = append(actualLines, strings.TrimSpace(line))
				}
			}
			if len(actualLines) != len(tc.expectedSubstrings) {
				t.Fatalf("Expected %d lines of output, got %d.\nFull Output:\n%s", len(tc.expectedSubstrings), len(actualLines), printedText)
			}

			for i, expectedStr := range tc.expectedSubstrings {
				if !strings.Contains(actualLines[i], expectedStr) {
					t.Errorf("Line %d failed:\nExpected to contain: %q\nActually got: %q", i+1, expectedStr, actualLines[i])
				}
			}
		})
	}
}
