package main

import (
	"context"
	"log"
	"time"

	"github.com/bklieger/diningbot/client"
	"github.com/bklieger/diningbot/config"
	"github.com/bklieger/diningbot/utils"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Global client instance (initialized once)
var diningClient *client.DiningHallClient

func initClient() error {
	if diningClient == nil {
		var err error
		diningClient, err = client.NewDiningHallClient()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetMenuInput defines the input for the get_menu tool
type GetMenuInput struct {
	Location string `json:"location" jsonschema:"required,description=The dining hall location name"`
	Date     string `json:"date" jsonschema:"description=Date in M/D/YYYY format (e.g., 1/15/2025). If not provided, uses today's date"`
	MealType string `json:"mealType" jsonschema:"required,description=The meal type"`
}

// GetMenuOutput defines the output for the get_menu tool
type GetMenuOutput struct {
	Location string   `json:"location"`
	Date     string   `json:"date"`
	MealType string   `json:"mealType"`
	Items    []string `json:"items"`
	Error    string   `json:"error,omitempty"`
}

// GetMenu fetches the menu for a specific location, date, and meal type
func GetMenu(ctx context.Context, req *mcp.CallToolRequest, input GetMenuInput) (
	*mcp.CallToolResult,
	GetMenuOutput,
	error,
) {
	if err := initClient(); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Failed to initialize client: " + err.Error()},
			},
		}, GetMenuOutput{}, nil
	}

	// Validate location
	if !config.IsValidLocation(input.Location) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Invalid location: " + input.Location},
			},
		}, GetMenuOutput{}, nil
	}

	// Validate meal type
	if !config.IsValidMealType(input.MealType) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Invalid meal type: " + input.MealType},
			},
		}, GetMenuOutput{}, nil
	}

	// Use provided date or default to today
	date := input.Date
	if date == "" {
		date = utils.FormatDate(time.Now())
	}

	// Fetch menu
	items, err := diningClient.GetMenu(input.Location, date, input.MealType)
	if err != nil {
		return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error fetching menu: " + err.Error()},
				},
			}, GetMenuOutput{
				Location: input.Location,
				Date:     date,
				MealType: input.MealType,
				Error:    err.Error(),
			}, nil
	}

	return nil, GetMenuOutput{
		Location: input.Location,
		Date:     date,
		MealType: input.MealType,
		Items:    items,
	}, nil
}

// GetMenusRangeInput defines the input for the get_menus_range tool
type GetMenusRangeInput struct {
	Location  string `json:"location" jsonschema:"required,description=The dining hall location name"`
	MealType  string `json:"mealType" jsonschema:"required,description=The meal type"`
	Days      int    `json:"days" jsonschema:"description=Number of days to fetch (default: 7, max: 30)"`
	StartDate string `json:"startDate" jsonschema:"description=Start date in M/D/YYYY format. If not provided, uses today's date"`
}

// GetMenusRangeOutput defines the output for the get_menus_range tool
type GetMenusRangeOutput struct {
	Location string              `json:"location"`
	MealType string              `json:"mealType"`
	Menus    map[string][]string `json:"menus"`
	Error    string              `json:"error,omitempty"`
}

// GetMenusRange fetches menus for multiple days
func GetMenusRange(ctx context.Context, req *mcp.CallToolRequest, input GetMenusRangeInput) (
	*mcp.CallToolResult,
	GetMenusRangeOutput,
	error,
) {
	if err := initClient(); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Failed to initialize client: " + err.Error()},
			},
		}, GetMenusRangeOutput{}, nil
	}

	// Validate location
	if !config.IsValidLocation(input.Location) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Invalid location: " + input.Location},
			},
		}, GetMenusRangeOutput{}, nil
	}

	// Validate meal type
	if !config.IsValidMealType(input.MealType) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Invalid meal type: " + input.MealType},
			},
		}, GetMenusRangeOutput{}, nil
	}

	// Set default days
	days := input.Days
	if days <= 0 {
		days = 7
	}
	if days > 30 {
		days = 30
	}

	// Determine start date
	var startTime time.Time
	if input.StartDate != "" {
		var err error
		startTime, err = time.Parse("1/2/2006", input.StartDate)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Invalid start date format. Use M/D/YYYY format: " + err.Error()},
				},
			}, GetMenusRangeOutput{}, nil
		}
	} else {
		startTime = time.Now()
	}

	// Fetch menus for each day
	menus := make(map[string][]string)
	for i := 0; i < days; i++ {
		date := startTime.AddDate(0, 0, i)
		dateStr := utils.FormatDate(date)

		items, err := diningClient.GetMenu(input.Location, dateStr, input.MealType)
		if err != nil {
			// Always set an empty array, never nil
			menus[dateStr] = []string{}
			continue
		}

		// Ensure items is never nil
		if items == nil {
			items = []string{}
		}
		menus[dateStr] = items
	}

	return nil, GetMenusRangeOutput{
		Location: input.Location,
		MealType: input.MealType,
		Menus:    menus,
	}, nil
}

func main() {
	// Create MCP server
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "diningbot",
			Version: "1.0.0",
		},
		nil,
	)

	// Add tools with manual schema definitions including enum values
	getMenuSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "The dining hall location name",
				"enum":        config.ValidLocations,
			},
			"date": map[string]interface{}{
				"type":        "string",
				"description": "Date in M/D/YYYY format (e.g., 1/15/2025). If not provided, uses today's date",
			},
			"mealType": map[string]interface{}{
				"type":        "string",
				"description": "The meal type",
				"enum":        config.ValidMealTypes,
			},
		},
		"required": []string{"location", "mealType"},
	}

	getMenusRangeSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type":        "string",
				"description": "The dining hall location name",
				"enum":        config.ValidLocations,
			},
			"mealType": map[string]interface{}{
				"type":        "string",
				"description": "The meal type",
				"enum":        config.ValidMealTypes,
			},
			"days": map[string]interface{}{
				"type":        "integer",
				"description": "Number of days to fetch (default: 7, max: 30)",
			},
			"startDate": map[string]interface{}{
				"type":        "string",
				"description": "Start date in M/D/YYYY format. If not provided, uses today's date",
			},
		},
		"required": []string{"location", "mealType"},
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_menu",
		Description: "Get the menu for a specific dining hall location, date, and meal type",
		InputSchema: getMenuSchema,
	}, GetMenu)

	// Define output schema for get_menus_range to handle empty arrays
	getMenusRangeOutputSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type": "string",
			},
			"mealType": map[string]interface{}{
				"type": "string",
			},
			"menus": map[string]interface{}{
				"type": "object",
				"additionalProperties": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
			},
		},
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:         "get_menus_range",
		Description:  "Get menus for multiple days for a specific dining hall location and meal type",
		InputSchema:  getMenusRangeSchema,
		OutputSchema: getMenusRangeOutputSchema,
	}, GetMenusRange)

	// Run the server over stdin/stdout
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
