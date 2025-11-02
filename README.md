# diningbot

A Go application to fetch daily menus from Stanford dining halls with full test coverage.

## Features

- Fetches menus for any valid dining hall location
- Supports all meal types (Breakfast, Lunch, Dinner, Brunch)
- Full test coverage across all packages
- Organized into packages for maintainability
- Validates location and meal type inputs

## Structure

```
.
├── client/       # HTTP client and session management
├── config/       # Configuration and validation
├── parser/       # HTML parsing utilities
├── utils/        # Utility functions
└── main.go       # Application entry point
```

## Installation

```bash
go mod download
```

## Usage

```bash
# Get breakfast for next 7 days at default location
go run main.go

# Get breakfast for a specific location
go run main.go "Arrillaga Family Dining Commons"

# Get lunch for a specific location
go run main.go "Arrillaga Family Dining Commons" "Lunch"

# Output as JSON
go run main.go "Arrillaga Family Dining Commons" "Breakfast" --json
```

## Valid Locations

- Arrillaga Family Dining Commons
- Branner Dining
- EVGR Dining
- Florence Moore Dining
- Gerhard Casper Dining
- Lakeside Dining
- Ricker Dining
- Stern Dining
- Wilbur Dining

## Valid Meal Types

- Breakfast
- Lunch
- Dinner
- Brunch

## Testing

Run unit tests:

```bash
go test ./... -cover
```

Run integration tests (requires network access):

```bash
go test -v -tags=integration -timeout 120s
```

Test coverage:
- `client`: 56.6%
- `config`: 85.7%
- `parser`: 73.8%
- `utils`: 100.0%

Integration tests include:
- Full flow from initialization to menu fetching
- Multiple locations and meal types
- Session persistence across requests
- Error handling with invalid inputs
- Debug mode functionality

## Dependencies

- `golang.org/x/net/html` - HTML parsing
- `golang.org/x/net/publicsuffix` - Cookie jar support
