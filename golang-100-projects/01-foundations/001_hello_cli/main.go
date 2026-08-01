package hellocli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("What is your name? ")
	name, nameErr := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	if nameErr != nil || name == "" {
		name = "User"
	}
	fmt.Print("How old are you? ")
	age, ageInErr := reader.ReadString('\n')
	if ageInErr != nil {
		fmt.Println("sorry you entered an invalid number")
		os.Exit(1)
	}
	age = strings.TrimSpace(age)
	ageNum, ageErr := strconv.ParseUint(age, 10, 8)
	if ageErr != nil || ageNum <= 0 {
		fmt.Println("sorry you entered an invalid number")
		os.Exit(1)
	} else if ageNum == 100 {
		fmt.Printf("Hello, %s! Wow you are already 100 years old.\n", name)
		os.Exit(0)
	} else if ageNum > 100 {
		fmt.Println("sorry you can't be more than 100 years old")
		os.Exit(1)
	}
	ageNum = 100 - ageNum
	fmt.Printf("Hello, %s! You are going to be 100 years old in %v years.\n", name, ageNum)
}
