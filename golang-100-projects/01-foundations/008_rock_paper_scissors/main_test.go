package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestPlayGame(t *testing.T) {
	tests := []struct {
		name           string
		userInput      string
		computerChoice Choice
		expected       string
	}{

		{"User Rock vs Comp Scissors", "rock\n", Scissors, "Win"},
		{"User Paper vs Comp Rock", "paper\n", Rock, "Win"},
		{"User Scissors vs Comp Paper", "scissors\n", Paper, "Win"},

		{"User Scissors vs Comp Rock", "scissors\n", Rock, "Lose"},
		{"User Rock vs Comp Paper", "rock\n", Paper, "Lose"},
		{"User Paper vs Comp Scissors", "paper\n", Scissors, "Lose"},

		{"Draw Rock", "rock\n", Rock, "Draw"},
		{"Draw Paper", "paper\n", Paper, "Draw"},
		{"Draw Scissors", "scissors\n", Scissors, "Draw"},

		{"Mixed Case Input", "RoCk\n", Scissors, "Win"},
		{"Whitespace Input", "  paper  \n", Rock, "Win"},
		{"Numeric Input", "3\n", Paper, "Win"},

		{"Invalid Input Dynamite", "dynamite\n", Rock, "Invalid input"},
		{"Empty Input (EOF)", "", Rock, "EOF"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.userInput))
			var output strings.Builder
			playGame(reader, &output, func() Choice {
				return tc.computerChoice
			})
			printedText := strings.TrimSpace(output.String())
			if !strings.Contains(printedText, tc.expected) {
				t.Errorf("failed:\nExpected to contain: %q\nActually got: %q", tc.expected, printedText)
			}
		})
	}
}
