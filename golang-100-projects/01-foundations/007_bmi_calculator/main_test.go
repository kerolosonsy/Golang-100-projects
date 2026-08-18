package main

import (
	"bufio"
	"strings"
	"testing"
)

func TestRunBMICalculator(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal BMI", "1.75\n70\n", "normal weight"},
		{"Underweight BMI", "1.75\n50\n", "underweight"},
		{"Overweight BMI", "1.70\n80\n", "overweight"},
		{"Obese BMI", "1.60\n90\n", "obese"},

		{"Exact Boundary 18.5 (Normal)", "1.0\n18.5\n", "normal weight"},
		{"Exact Boundary 25 (Overweight)", "1.0\n25.0\n", "overweight"},
		{"Exact Boundary 30 (Obese)", "1.0\n30.0\n", "obese"},

		{"Zero Height", "0\n70\n", "positive numbers"},
		{"Negative Weight", "1.75\n-10\n", "positive numbers"},
		{"Non-numeric Height", "one\n70\n", "Invalid input: Height must be"},
		{"Non-numeric Weight", "1.75\nseventy\n", "Invalid input: Weight must be"},

		{"Absurd Height (Range Check)", "5\n70\n", "less than 3 meters"},
		{"Absurd Weight (Range Check)", "1.75\n1000\n", "less than 500 KG"},
		{"NaN Input", "NaN\n70\n", "positive numbers"},
		{"Infinity Input", "1.75\n+Inf\n", "positive numbers"},
		{"Early Exit (EOF)", "", "Error reading input"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tc.input))
			var output strings.Builder
			runBMICalculator(reader, &output)
			printedText := strings.TrimSpace(output.String())
			if !strings.Contains(printedText, tc.expected) {
				t.Errorf("failed:\nExpected to contain: %q\nActually got: %q", tc.expected, printedText)
			}
		})
	}
}
