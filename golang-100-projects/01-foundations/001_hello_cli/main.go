package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("What is your name? ")
	name, nameErr := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if errors.Is(nameErr, io.EOF) {
		println("")
	} else if nameErr != nil {
		name = "User"
	}
	if name == "" {
		name = "User"
	}
	fmt.Print("How old are you? ")
	age, ageInErr := reader.ReadString('\n')
	if ageInErr != nil {
		if ageInErr != io.EOF {
			fmt.Println("sorry you entered an invalid number")
			os.Exit(1)
		}
		println("")
	}
	age = strings.TrimSpace(age)
	ageNum, ageErr := strconv.ParseUint(age, 10, 8)
	// if the user input wrong number i'll exit the app.
	if errors.Is(ageErr, strconv.ErrRange) || ageNum > 100 {
		fmt.Printf("Hello, %s! Wow You have lived more than a century!\n", name)
		os.Exit(0)
	} else if ageNum == 100 {
		fmt.Printf("Hello, %s! Wow you are already 100 years old.\n", name)
		os.Exit(0)
	} else if ageErr != nil || ageNum <= 0 {
		fmt.Println("sorry you entered an invalid number")
		os.Exit(1)
	}
	ageNum = 100 - ageNum
	fmt.Printf("Hello, %s! You are going to be 100 years old in %v years.\n", name, ageNum)
}
