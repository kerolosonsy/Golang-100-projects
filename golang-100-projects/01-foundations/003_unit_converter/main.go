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

const (
	MetersToFeet     = 3.28084
	KgToLbs          = 2.20462
	FahrenheitScale  = 1.8
	FahrenheitOffset = 32.0
)

type ConversionKey struct {
	Category string
	Source   string
	Target   string
}

type ConversionFunc func(amount float64) (float64, error)

var conversionRegistry = map[ConversionKey]ConversionFunc{
	{"1", "1", "2"}: func(amount float64) (float64, error) {
		if amount < 0 {
			return 0, errors.New("length cannot be negative")
		}
		return amount * MetersToFeet, nil
	},
	{"1", "2", "1"}: func(amount float64) (float64, error) {
		if amount < 0 {
			return 0, errors.New("length cannot be negative")
		}
		return amount / MetersToFeet, nil
	},
	{"1", "1", "1"}: func(amount float64) (float64, error) {
		if amount < 0 {
			return 0, errors.New("length cannot be negative")
		}
		return amount, nil
	},
	{"1", "2", "2"}: func(amount float64) (float64, error) {
		if amount < 0 {
			return 0, errors.New("length cannot be negative")
		}
		return amount, nil
	},
	{"2", "1", "2"}: func(amount float64) (float64, error) {
		if amount < 0 {
			return 0, errors.New("Weight cannot be negative")
		}
		return amount * KgToLbs, nil
	},
	{"2", "2", "1"}: func(amount float64) (float64, error) {
		if amount < 0 {
			return 0, errors.New("Weight cannot be negative")
		}
		return amount / KgToLbs, nil
	},
	{"2", "1", "1"}: func(amount float64) (float64, error) {
		if amount < 0 {
			return 0, errors.New("Weight cannot be negative")
		}
		return amount, nil
	},
	{"2", "2", "2"}: func(amount float64) (float64, error) {
		if amount < 0 {
			return 0, errors.New("Weight cannot be negative")
		}
		return amount, nil
	},
	{"3", "1", "2"}: func(amount float64) (float64, error) {
		if amount < -273.15 {
			return 0, errors.New("physically impossible: below absolute zero (-273.15 C)")
		}
		return (amount * FahrenheitScale) + FahrenheitOffset, nil
	},
	{"3", "2", "1"}: func(amount float64) (float64, error) {
		if amount < -459.67 {
			return 0, errors.New("physically impossible: below absolute zero (-459.67 F)")
		}
		return (amount - FahrenheitOffset) / FahrenheitScale, nil
	},
	{"3", "1", "1"}: func(amount float64) (float64, error) {
		if amount < -273.15 {
			return 0, errors.New("physically impossible: below absolute zero (-273.15 C)")
		}
		return amount, nil
	},
	{"3", "2", "2"}: func(amount float64) (float64, error) {
		if amount < -459.67 {
			return 0, errors.New("physically impossible: below absolute zero (-459.67 F)")
		}
		return amount, nil
	},
}

// 1 Length 2 Weight 3 Heat
// 1.1 Meter 1.2 Foot 2.1 Kilogram 2.2 Pound 3.1 Celsius 3.2 Fahrenheit
var categoryTypes = map[string][]string{
	"1":     {"Please select your source unit: \n1-Meter\n2-Foot\n"},
	"2":     {"Please select your source unit: \n1-Kilogram\n2-Pound\n"},
	"3":     {"Please select your source unit: \n1-Celsius\n2-Fahrenheit\n"},
	"1.1":   {"Please select your target unit: \n1-Meter\n2-Foot\n"},
	"1.2":   {"Please select your target unit: \n1-Meter\n2-Foot\n"},
	"2.1":   {"Please select your target unit: \n1-Kilogram\n2-Pound\n"},
	"2.2":   {"Please select your target unit: \n1-Kilogram\n2-Pound\n"},
	"3.1":   {"Please select your target unit: \n1-Celsius\n2-Fahrenheit\n"},
	"3.2":   {"Please select your target unit: \n1-Celsius\n2-Fahrenheit\n"},
	"1.1.1": {"meter", "meter"},
	"1.1.2": {"meter", "foot"},
	"1.2.1": {"foot", "meter"},
	"1.2.2": {"foot", "foot"},
	"2.1.1": {"Kilogram", "Kilogram"},
	"2.1.2": {"Kilogram", "Pound"},
	"2.2.1": {"Pound", "Kilogram"},
	"2.2.2": {"Pound", "Pound"},
	"3.1.1": {"Celsius", "Celsius"},
	"3.2.1": {"Fahrenheit", "Celsius"},
	"3.1.2": {"Celsius", "Fahrenheit"},
	"3.2.2": {"Fahrenheit", "Fahrenheit"},
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	var category, source_unit, target_unit, amount string
	var err error
	category, err = askInput(reader, "Please select your category: \n1-Length\n2-Weight\n3-Heat\n")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	askStr, exists := categoryTypes[category]
	if !exists {
		fmt.Println("The number you entered is not valid.")
		os.Exit(1)
	}
	source_unit, err = askInput(reader, askStr[0])
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	askStr, exists = categoryTypes[category+"."+source_unit]
	if !exists {
		fmt.Println("The number you entered is not valid.")
		os.Exit(1)
	}
	target_unit, err = askInput(reader, askStr[0])
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	amount, err = askInput(reader, "Please enter the amount: ")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	amountFl, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	result, err := convertValue(category, source_unit, target_unit, amountFl)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	resultStr, _ := categoryTypes[category+"."+source_unit+"."+target_unit]
	fmt.Printf("%v %v = %v %v\n", amount, resultStr[0], result, resultStr[1])

}

func convertValue(category, source, target string, amount float64) (float64, error) {
	key := ConversionKey{
		Category: category,
		Source:   source,
		Target:   target,
	}

	calcFunc, exists := conversionRegistry[key]
	if !exists {
		return 0, errors.New("unsupported category or unknown units")
	}

	return calcFunc(amount)
}

func askInput(reader *bufio.Reader, ask string) (string, error) {
	fmt.Print(ask)
	answer, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			fmt.Println("please try again later.")
			return "", err
		}
		fmt.Println("")
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer, nil
}
