package service

import (
	"testing"
)

func TestNewLuhnOrderNumberValidator(t *testing.T) {
	validator := NewLuhnOrderNumberValidator()
	if validator == nil {
		t.Error("NewLuhnOrderNumberValidator() returned nil")
	}

	_, ok := validator.(*luhnOrderNumberValidator)
	if !ok {
		t.Error("NewLuhnOrderNumberValidator() does not return *luhnOrderNumberValidator")
	}
}

func Test_luhnOrderNumberValidator_Validate(t *testing.T) {
	v := &luhnOrderNumberValidator{}

	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"valid number 1", "79927398713", true},
		{"valid number 2", "12345678903", true},
		{"valid number 3", "4532015112830366", true},
		{"invalid number 1", "79927398714", false},
		{"invalid number 2", "12345678904", false},
		{"empty string", "", false},
		{"single digit", "1", false},
		{"non-numeric", "abc", false},
		{"mixed alphanumeric", "7992739871a", false},
		{"valid with spaces", "7992 7398 713", false},
		{"valid number 4", "4532015112830366", true},
		{"invalid checksum", "4532015112830367", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := v.Validate(tt.number); got != tt.want {
				t.Errorf("Validate(%q) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}
