package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type record struct {
	Question    string
	Choices     []string
	RightAnswer int
}

func main() {
	var records = []record{
		{
			Question:    "What is the capital of United Arab Emirates?",
			Choices:     []string{"Dubai", "Abu Dhabi", "Sharjah", "Ajman"},
			RightAnswer: 2,
		},
		{
			Question:    "Which language is used to build this quiz?",
			Choices:     []string{"Python", "Java", "Go", "C++"},
			RightAnswer: 3,
		},
		{
			Question:    "What is the result of 5 + 3 * 2?",
			Choices:     []string{"16", "11", "10", "15"},
			RightAnswer: 2,
		},
		{
			Question:    "This is a broken question with no choices. What will you do?",
			Choices:     []string{},
			RightAnswer: 1,
		},
		{
			Question:    "Who developed the Go programming language?",
			Choices:     []string{"Apple", "Microsoft", "Google", "Amazon"},
			RightAnswer: 3,
		},
	}

	reader := bufio.NewReader(os.Stdin)

	playGame(reader, os.Stdout, records)
}

func playGame(reader *bufio.Reader, writer io.Writer, records []record) {
	score := 0
	questionCount := 0
	if len(records) == 0 {
		fmt.Fprintln(writer, "Error: The question bank is empty.")
		return
	}
	for _, record := range records {
		if len(record.Choices) == 0 {
			fmt.Fprintln(writer, "Warning: Skipping a malformed question with no choices.")
			continue
		}
		fmt.Fprintf(writer, "%v\n", record.Question)
		for i, choice := range record.Choices {
			fmt.Fprintf(writer, "%v. %v\n", i+1, choice)
		}

		input, err := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(writer, err)
				return
			} else if err == io.EOF && len(input) == 0 {
				fmt.Fprintln(writer, "Quiz terminated early. Goodbye!")
				return
			}
		}
		answer, err := strconv.ParseUint(input, 10, 8)
		if err != nil || int(answer) < 1 || int(answer) > len(record.Choices) {
			fmt.Fprintln(writer, "Invalid or unrecognized choice. Marked as wrong.")
			questionCount++
			continue
		}
		if int(answer) == record.RightAnswer {
			score++
		}
		questionCount++
	}
	if questionCount == 0 {
		fmt.Fprintln(writer, "Error: No valid questions to process.")
		return
	}
	percentage := (float64(score) / float64(questionCount)) * 100
	fmt.Fprintf(writer, "You scored %v out of %v = %.0f%%.\n", score, questionCount, percentage)
}
