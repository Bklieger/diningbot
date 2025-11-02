//go:build integration
// +build integration

package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bklieger/diningbot/client"
	"github.com/bklieger/diningbot/config"
	"github.com/bklieger/diningbot/utils"
)

// TestIntegrationFullFlow tests the complete flow from initialization to fetching menu
func TestIntegrationFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	diningClient, err := client.NewDiningHallClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test with a known location and meal type
	location := "Arrillaga Family Dining Commons"
	mealType := "Lunch"
	date := utils.FormatDate(time.Now())

	foods, err := diningClient.GetMenu(location, date, mealType)
	if err != nil {
		t.Fatalf("Failed to get menu: %v", err)
	}

	// Menu might be empty on some days, so just verify we got a response
	t.Logf("Found %d menu items for %s at %s on %s", len(foods), mealType, location, date)

	// If we got items, verify they don't contain filtered text
	for _, food := range foods {
		if len(food) == 0 {
			t.Error("Got empty food item")
		}
		// Verify filtering is working
		lowerFood := food
		if contains(lowerFood, "ingredient") || contains(lowerFood, "allergen") {
			t.Errorf("Food item contains filtered text: %s", food)
		}
	}
}

// TestIntegrationMultipleLocations tests fetching from different locations
func TestIntegrationMultipleLocations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	diningClient, err := client.NewDiningHallClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	locations := []string{
		"Arrillaga Family Dining Commons",
		"Branner Dining",
		"Florence Moore Dining",
	}

	date := utils.FormatDate(time.Now())
	mealType := "Lunch"

	for _, location := range locations {
		t.Run(location, func(t *testing.T) {
			foods, err := diningClient.GetMenu(location, date, mealType)
			if err != nil {
				t.Errorf("Failed to get menu for %s: %v", location, err)
				return
			}
			t.Logf("%s: %d items", location, len(foods))
		})
	}
}

// TestIntegrationMultipleMealTypes tests different meal types
func TestIntegrationMultipleMealTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	diningClient, err := client.NewDiningHallClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	location := "Arrillaga Family Dining Commons"
	date := utils.FormatDate(time.Now())

	mealTypes := []string{"Breakfast", "Lunch", "Dinner"}

	for _, mealType := range mealTypes {
		t.Run(mealType, func(t *testing.T) {
			foods, err := diningClient.GetMenu(location, date, mealType)
			if err != nil {
				t.Errorf("Failed to get menu for %s: %v", mealType, err)
				return
			}
			t.Logf("%s: %d items", mealType, len(foods))
		})
	}
}

// TestIntegrationSessionPersistence tests that session persists across multiple requests
func TestIntegrationSessionPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	diningClient, err := client.NewDiningHallClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	location := "Branner Dining"
	mealType := "Lunch"

	// Make multiple requests with the same client
	for i := 0; i < 3; i++ {
		date := utils.FormatDate(time.Now().AddDate(0, 0, i))
		foods, err := diningClient.GetMenu(location, date, mealType)
		if err != nil {
			t.Errorf("Request %d failed: %v", i+1, err)
			continue
		}
		t.Logf("Request %d: %d items", i+1, len(foods))
	}
}

// TestIntegrationInvalidInputs tests error handling with invalid inputs
func TestIntegrationInvalidInputs(t *testing.T) {
	diningClient, err := client.NewDiningHallClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name     string
		location string
		date     string
		mealType string
		wantErr  bool
	}{
		{
			name:     "invalid location",
			location: "Invalid Dining Hall",
			date:     utils.FormatDate(time.Now()),
			mealType: "Lunch",
			wantErr:  true,
		},
		{
			name:     "invalid meal type",
			location: "Branner Dining",
			date:     utils.FormatDate(time.Now()),
			mealType: "Invalid",
			wantErr:  true,
		},
		{
			name:     "valid inputs",
			location: "Branner Dining",
			date:     utils.FormatDate(time.Now()),
			mealType: "Lunch",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := diningClient.GetMenu(tt.location, tt.date, tt.mealType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMenu() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIntegrationAllLocations tests all configured locations
func TestIntegrationAllLocations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	diningClient, err := client.NewDiningHallClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	date := utils.FormatDate(time.Now())
	mealType := "Lunch"

	successCount := 0
	for _, location := range config.ValidLocations {
		t.Run(location, func(t *testing.T) {
			foods, err := diningClient.GetMenu(location, date, mealType)
			if err != nil {
				t.Logf("Failed to get menu for %s: %v", location, err)
				return
			}
			t.Logf("%s: %d items", location, len(foods))
			successCount++
		})
	}

	if successCount == 0 {
		t.Error("No locations returned successful results")
	}
}

// TestIntegrationDebugMode tests debug mode functionality
func TestIntegrationDebugMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	diningClient, err := client.NewDiningHallClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Enable debug mode
	diningClient.Debug = true

	location := "Branner Dining"
	date := utils.FormatDate(time.Now())
	mealType := "Lunch"

	_, err = diningClient.GetMenu(location, date, mealType)
	if err != nil {
		t.Errorf("GetMenu() with debug failed: %v", err)
	}

	// Check if debug file was created
	expectedFile := "debug_response_Branner_Dining_" + strings.ReplaceAll(date, "/", "_") + ".html"
	if _, err := os.Stat(expectedFile); err == nil {
		// Clean up
		os.Remove(expectedFile)
		t.Logf("Debug file created and cleaned up: %s", expectedFile)
	}

	// Clean up other potential debug files
	os.Remove("debug_initial_page.html")
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
