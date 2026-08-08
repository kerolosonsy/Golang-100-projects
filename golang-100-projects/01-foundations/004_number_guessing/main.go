package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	TrialNum = 5
	MaxNum   = 100
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	fmt.Printf("Welcome to the Number Guessing Game! I have picked a secret number between 0 and %v. You have %v attempts to guess it.", MaxNum, TrialNum)
	playGame(reader, os.Stdout, func() int { return rng.Intn(MaxNum + 1) })
}

// if the user entered wrong value i'll count this try
func playGame(reader *bufio.Reader, writer io.Writer, makeSecret func() int) {
	secret := makeSecret()
	for i := 0; i < TrialNum; i++ {
		num, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(writer, "I'll count this try, Try again.")
				continue
			}
			fmt.Fprintln(writer, "Goodbye!")
			return
		}
		num = strings.TrimSpace(num)
		numInt, err := strconv.ParseInt(num, 10, 64)
		if err != nil {
			fmt.Fprintln(writer, "Invalid input! Please enter only numbers. I'll count this try.")
			continue
		}
		if numInt == int64(secret) {
			if i+1 == 1 {
				fmt.Fprintf(writer, "Congratulations! You Win! The number was %v. You used %v trial\n", secret, i+1)
				return
			}
			fmt.Fprintf(writer, "Congratulations! You Win! The number was %v. You used %v trials\n", secret, i+1)
			return
		} else if numInt > MaxNum || numInt < 0 {
			fmt.Fprintf(writer, "Out of range! Please enter a number between 0 and %v. I'll count this try.\n", MaxNum)
			continue
		} else if numInt >= int64(secret) {
			fmt.Fprintln(writer, "Guess lower")
		} else if numInt <= int64(secret) {
			fmt.Fprintln(writer, "Guess higher")
		}
	}
	fmt.Fprintln(writer, "Sorry, You lose! the number was", secret)
}
