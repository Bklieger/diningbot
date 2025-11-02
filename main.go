package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bklieger/diningbot/client"
	"github.com/bklieger/diningbot/config"
	"github.com/bklieger/diningbot/utils"
)

func main() {
	diningClient, err := client.NewDiningHallClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	// Parse command line arguments
	location := config.ValidLocations[0] // Default to first location
	mealType := "Breakfast"              // Default meal type
	outputJSON := false
	debug := false

	args := os.Args[1:]
	var filteredArgs []string
	for _, arg := range args {
		if arg == "--json" {
			outputJSON = true
		} else if arg == "--debug" {
			debug = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

	// Enable debug if requested
	if debug {
		diningClient.Debug = true
	}

	if len(args) > 0 {
		// Check if first arg is a valid location
		if config.IsValidLocation(args[0]) {
			location = args[0]
		} else {
			fmt.Fprintf(os.Stderr, "Warning: '%s' is not a valid location. Using default: %s\n", args[0], location)
		}
	}

	if len(args) > 1 {
		// Check if second arg is a valid meal type
		if config.IsValidMealType(args[1]) {
			mealType = args[1]
		} else {
			fmt.Fprintf(os.Stderr, "Warning: '%s' is not a valid meal type. Using default: %s\n", args[1], mealType)
		}
	}

	// Get menu for the next 7 days
	now := time.Now()
	results := make(map[string][]string)

	fmt.Printf("Fetching %s menus for %s...\n\n", mealType, location)
	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, i)
		dateStr := utils.FormatDate(date)

		foods, err := diningClient.GetMenu(location, dateStr, mealType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching menu for %s: %v\n", dateStr, err)
			results[dateStr] = []string{}
			continue
		}

		results[dateStr] = foods
		fmt.Printf("%s for %s (%s):\n", mealType, dateStr, date.Format("Monday"))
		if len(foods) == 0 {
			fmt.Println("  No items found")
		} else {
			for _, food := range foods {
				fmt.Printf("  - %s\n", food)
			}
		}
		fmt.Println()
	}

	// Optionally output as JSON
	if outputJSON {
		jsonOutput, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(jsonOutput))
	}
}
