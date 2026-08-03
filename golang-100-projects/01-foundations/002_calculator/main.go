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
	fmt.Print("Please enter your equation: ")
	equ, equErr := reader.ReadString('\n')
	if errors.Is(equErr, io.EOF) {
		println("")
	} else if equErr != nil {
		println("please try again later.")
	}
	equ = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(equ, "=", ""), " ", ""))

	str, err := getEquationResult(getEquationTypes(equ), equ)
	equ += " = " + str
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	fmt.Println(equ)
}

func getEquationTypes(equ1 string) []string {
	strType := []string{}
	equ := []rune(equ1)
	for i := 0; i < len(equ); i++ {
		v := string(equ[i])
		switch v {
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			strType = append(strType, "number")
		case "+", "-", "*", "/":
			if len(strType) == 0 && v == "-" {
				strType = append(strType, "negative")
			} else if len(strType) == 0 {
				strType = append(strType, "invalid")
			} else if strType[i-1] == "operator" && v == "-" {
				strType = append(strType, "negative")
			} else if strType[i-1] == "operator" {
				strType = append(strType, "invalid")
			} else if strType[i-1] == "negative" {
				strType = append(strType, "invalid")
			} else if strType[i-1] == "invalid" {
				strType = append(strType, "invalid")
			} else {
				strType = append(strType, "operator")
			}
		case ".":
			strType = append(strType, "decimal")
		default:
			strType = append(strType, "unknown")
		}

	}
	return strType
}
func getEquationResult(strType []string, equ1 string) (string, error) {
	var num1, num2 float64
	var operator string
	var num1True bool = true
	var num2True bool = true
	var num1Err, num2Err error
	var opCount = 0
	equ := []rune(equ1)
	for i := 0; i < len(strType); i++ {
		v := string(equ[i])
		if strType[i] == "number" || strType[i] == "negative" || strType[i] == "decimal" {
			numStr := v
			for j := i + 1; j < len(strType); j++ {
				v := string(equ[j])
				if strType[j] == "operator" || strType[j] == "negative" || strType[j] == "unknown" || strType[j] == "invalid" {
					i = j - 1
					break
				} else if strings.Contains(numStr, ".") && v == "." {
					return "0", errors.New("non-numeric operand")
				} else {
					numStr += v
				}
				if j == len(strType)-1 {
					i = j
					break
				}
			}
			if num1True {
				num1, num1Err = strconv.ParseFloat(numStr, 64)
				num1True = false
			} else {
				num2, num2Err = strconv.ParseFloat(numStr, 64)
				num2True = false
			}
			if num1Err != nil || num2Err != nil {
				return "0", errors.New("you have entered invalid numbers")
			}
		} else if strType[i] == "operator" {
			operator = v
		} else if strType[i] == "invalid" {
			return "0", errors.New("non-numeric operand")
		} else if strType[i] == "unknown" {
			return "0", errors.New("unrecognized operator")
		} else {
			continue
		}

		if !num1True && !num2True && operator != "" {
			opCount++
			switch operator {
			case "+":
				num1 = num1 + num2
				num2True = true
				operator = ""
			case "-":
				num1 = num1 - num2
				num2True = true
				operator = ""
			case "*":
				num1 = num1 * num2
				num2True = true
				operator = ""
			case "/":
				if num2 == 0 {
					return "0", errors.New("division by zero is not allowed")
				}
				num1 = num1 / num2
				num2True = true
				operator = ""
			}
		}
	}
	if !num1True && opCount > 0 {
		return strconv.FormatFloat(num1, 'f', -1, 64), nil

	} else {
		return "0", errors.New("Missing operator")
	}
}
