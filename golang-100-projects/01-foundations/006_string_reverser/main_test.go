package main

import "testing"

func TestReverse(t *testing.T) {

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"Empty string", "", ""},
		{"Single character", "a", "a"},
		{"Two characters", "ab", "ba"},
		{"Sentence with spaces", "hello world", "dlrow olleh"},
		{"Arabic string", "مرحبا", "ابحرم"},
		{"Emojis", "🚗💨", "💨🚗"},
		{"Mixed Unicode", "Hello 🌍", "🌍 olleH"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := reverse(tc.input)
			if result != tc.output {
				t.Errorf("Expected %q, got %q", tc.output, result)
			}
		})
	}
}
func TestPalindrome(t *testing.T) {

	tests := []struct {
		name   string
		input  string
		output bool
	}{
		{"Empty string", "", true},
		{"Single character", "a", true},
		{"Two characters", "ab", false},
		// تم تعديل الاسم هنا كمان
		{"Sentence with spaces", "hello world", false},
		{"Palindrome", "racecar", true},
		// اختبار الـ Palindrome مع الحروف الكبيرة والصغيرة
		{"Mixed case Palindrome", "RaceCar", true},
		// اختبارات الـ Unicode للـ Palindrome
		{"Arabic Palindrome", "خوخ", true},
		{"Emoji Palindrome", "😊😂😊", true},
		// اختبار الحالة الطرفية الخاصة بالـ Unicode (İi) عشان نأكد إننا شغالين بـ ToLower مش EqualFold
		{"Unicode strict case", "İi", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := palindrome(tc.input)
			if result != tc.output {
				t.Errorf("Expected %t, got %t", tc.output, result)
			}
		})
	}
}
