package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if err != nil && err != io.EOF {
		fmt.Println("Error reading input")
		os.Exit(1)
	}

	output := reverse(input)
	fmt.Println(string(output))
	if !palindrome(input) {
		fmt.Println("No, it's not a palindrome!")
	} else {
		fmt.Println("Yes, it's a palindrome!")
	}
}

func reverse(input string) string {
	// Policy: Converting string to []rune automatically handles invalid UTF-8
	// by replacing them with the Unicode replacement character ''.
	// This prevents panics and safely preserves non-corrupted characters.
	safeRunes := []rune(input)
	output := []rune("")
	for i := len(safeRunes) - 1; i >= 0; i-- {
		output = append(output, safeRunes[i])
	}

	return string(output)
}
func palindrome(input string) bool {
	lowerInput := strings.ToLower(input)
	lowerReversed := strings.ToLower(reverse(input))

	return lowerInput == lowerReversed
}
