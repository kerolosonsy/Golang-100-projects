package main

import (
	"math"
	"testing"
)

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func TestConvertValue(t *testing.T) {
	tests := []struct {
		name          string
		category      string
		source        string
		target        string
		amount        float64
		expectedValue float64
		expectError   bool
	}{
		{"Meter to Foot", "1", "1", "2", 5, 16.4042, false},
		{"Foot to Meter", "1", "2", "1", 16.4042, 5.0, false},
		{"Celsius to Fahrenheit", "3", "1", "2", 0, 32.0, false},
		{"Fahrenheit to Celsius", "3", "2", "1", 32.0, 0.0, false},
		{"Kg to Pound", "2", "1", "2", 10, 22.0462, false},
		{"Kg to Pound", "2", "2", "1", 22.0462, 10, false},
		{"Meter to Meter", "1", "1", "1", 5.5, 5.5, false},
		{"Negative Length", "1", "1", "2", -5, 0, true},
		{"Negative Weight", "2", "1", "2", -10, 0, true},
		{"Below Absolute Zero (Celsius)", "3", "1", "2", -300, 0, true},
		{"Unknown Category", "99", "1", "2", 10, 0, true},
		{"Unknown Unit", "1", "9", "2", 10, 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := convertValue(tc.category, tc.source, tc.target, tc.amount)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error for %s but got none", tc.name)
				}
				return
			}

			if !tc.expectError && err != nil {
				t.Errorf("Did not expect an error for %s but got: %v", tc.name, err)
			}

			tolerance := 0.001
			if !almostEqual(result, tc.expectedValue, tolerance) {
				t.Errorf("%s failed: expected %v, got %v", tc.name, tc.expectedValue, result)
			}
		})
	}
}
