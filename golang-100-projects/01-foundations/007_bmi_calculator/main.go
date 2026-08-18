package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	runBMICalculator(reader, os.Stdout)
}

func calculateBMI(height, weight float64) float64 {
	return weight / (height * height)
}

func getCategory(bmi float64) string {
	switch {
	case bmi < 18.5:
		return "underweight"
	case bmi < 25:
		return "normal weight"
	case bmi < 30:
		return "overweight"
	default:
		return "obese"
	}
}
func runBMICalculator(reader *bufio.Reader, writer io.Writer) {
	fmt.Fprintln(writer, "Welcome to the BMI Calculator!")
	fmt.Fprint(writer, "Please enter your height in meters: ")
	heightInput, err := reader.ReadString('\n')
	heightInput = strings.TrimSpace(heightInput)
	if (err != nil && err != io.EOF) || (heightInput == "" && err == io.EOF) {
		fmt.Fprintln(writer, "Error reading input.")
		return
	}
	fmt.Fprint(writer, "Please enter your weight in KG: ")
	weightInput, err := reader.ReadString('\n')
	weightInput = strings.TrimSpace(weightInput)
	if (err != nil && err != io.EOF) || (weightInput == "" && err == io.EOF) {
		fmt.Fprintln(writer, "Error reading input.")
		return
	}
	height, err := strconv.ParseFloat(heightInput, 64)
	if err != nil {
		fmt.Fprintln(writer, "Invalid input: Height must be a valid number.")
		return
	}
	weight, err := strconv.ParseFloat(weightInput, 64)
	if err != nil {
		fmt.Fprintln(writer, "Invalid input: Weight must be a valid number.")
		return
	}
	if height <= 0 || weight <= 0 || math.IsNaN(height) || math.IsInf(height, 0) || math.IsNaN(weight) || math.IsInf(weight, 0) {
		fmt.Fprintln(writer, errors.New("Height and Weight must be positive numbers"))
		return
	} else if height >= 3 {
		fmt.Fprintln(writer, errors.New("Height must be less than 3 meters"))
		return
	} else if weight >= 500 {
		fmt.Fprintln(writer, errors.New("Weight must be less than 500 KG"))
		return
	}
	bmi := calculateBMI(height, weight)
	category := getCategory(bmi)
	fmt.Fprintf(writer, "\nThe computed BMI is %.2f, which falls into the %s category.\nThis is an educational tool, not medical advice\n", bmi, category)
}
