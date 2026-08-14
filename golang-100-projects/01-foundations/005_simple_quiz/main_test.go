package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestPlayGame(t *testing.T) {
	validBank := []record{
		{Question: "1+1", Choices: []string{"2", "3"}, RightAnswer: 1},
		{Question: "2+2", Choices: []string{"3", "4"}, RightAnswer: 2},
	}

	malformedBank := []record{
		{Question: "Broken", Choices: []string{}, RightAnswer: 1},
	}

	emptyBank := []record{}

	tests := []struct {
		name               string
		bank               []record
		input              string
		expectedSubstrings []string
	}{
		{"All correct", validBank, "1\n2\n", []string{"You scored 2 out of 2 = 100%."}},
		{"All wrong", validBank, "2\n1\n", []string{"You scored 0 out of 2 = 0%."}},
		{"Mixed answers", validBank, "1\n1\n", []string{"You scored 1 out of 2 = 50%."}},
		{"Empty bank", emptyBank, "", []string{"Error: The question bank is empty."}},
		{"All malformed", malformedBank, "", []string{"Warning: Skipping a malformed question", "Error: No valid questions to process."}},
		{"Invalid input", validBank, "abc\n1\n", []string{"Invalid or unrecognized choice", "You scored 0 out of 2 = 0%."}},
		{"Immediate EOF", validBank, "", []string{"Quiz terminated early. Goodbye!"}},
		{"EOF without newline", validBank, "1\n2", []string{"You scored 2 out of 2 = 100%."}}, 
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			var output strings.Builder

			playGame(reader, &output, tc.bank)

			printedText := output.String()
			for _, expected := range tc.expectedSubstrings {
				if !strings.Contains(printedText, expected) {
					t.Errorf("Expected output to contain %q, but got:\n%s", expected, printedText)
				}
			}
		})
	}
}
