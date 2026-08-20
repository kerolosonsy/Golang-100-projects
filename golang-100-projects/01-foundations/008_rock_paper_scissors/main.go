package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

//go:generate stringer -type=Choice
type Choice int

const (
	Rock Choice = iota
	Paper
	Scissors
)

//go:generate stringer -type=Result
type Result int

const (
	Win Result = iota
	Lose
	Draw
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	playGame(reader, os.Stdout, realRandomChoice)
}

func playGame(reader *bufio.Reader, writer io.Writer, getComputerChoice func() Choice) {
	fmt.Fprintln(writer, "Enter your choice: \n1 rock\n2 paper\n3 scissors")
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Fprintln(writer, err)
		return
	} else if err == io.EOF {
		if input == "" {
			fmt.Fprintln(writer, err)
			return
		}
		fmt.Fprintln(writer, "")
	}

	userChoice, err := parseChoice(input)
	if err != nil {
		fmt.Fprintln(writer, err)
		return
	}

	computerChoice := getComputerChoice()

	result := decideWinner(userChoice, computerChoice)

	fmt.Fprintf(writer, "User chose %v, Computer chose %v, Result: %v\n", userChoice, computerChoice, result)
}

func realRandomChoice() Choice {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	return Choice(rng.Intn(3))
}

func parseChoice(input string) (Choice, error) {
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)
	switch input {
	case "r", "rock", "1":
		return Rock, nil
	case "p", "paper", "2":
		return Paper, nil
	case "s", "scissors", "3":
		return Scissors, nil
	default:
		return 0, errors.New("Invalid input 'dynamite'. Please choose rock, paper, or scissors (or 1, 2, 3).")
	}
}

func decideWinner(user, computer Choice) Result {
	if user == computer {
		return Draw
	} else if user == Rock && computer == Scissors {
		return Win
	} else if user == Paper && computer == Rock {
		return Win
	} else if user == Scissors && computer == Paper {
		return Win
	} else {
		return Lose
	}
}
