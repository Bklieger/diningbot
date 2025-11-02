package config

import "testing"

func TestIsValidLocation(t *testing.T) {
	tests := []struct {
		name     string
		location string
		want     bool
	}{
		{"valid location", "Arrillaga Family Dining Commons", true},
		{"valid location 2", "Branner Dining", true},
		{"invalid location", "Invalid Location", false},
		{"empty location", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidLocation(tt.location)
			if got != tt.want {
				t.Errorf("IsValidLocation(%q) = %v, want %v", tt.location, got, tt.want)
			}
		})
	}
}

func TestIsValidMealType(t *testing.T) {
	tests := []struct {
		name     string
		mealType string
		want     bool
	}{
		{"valid meal type", "Breakfast", true},
		{"valid meal type 2", "Lunch", true},
		{"valid meal type 3", "Dinner", true},
		{"valid meal type 4", "Brunch", true},
		{"invalid meal type", "Invalid", false},
		{"empty meal type", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidMealType(tt.mealType)
			if got != tt.want {
				t.Errorf("IsValidMealType(%q) = %v, want %v", tt.mealType, got, tt.want)
			}
		})
	}
}
