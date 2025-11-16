# diningbot

A Go application to fetch daily menus from Stanford dining halls, available as both a CLI tool and an MCP (Model Context Protocol) server with full test coverage.

## Features

- **MCP Server**: Provides dining hall menu data via Model Context Protocol
- **HTTP Wrapper**: Accepts curl requests for easy testing
- **Caching Layer**: 1-hour cache to reduce server load and improve response times
- Fetches menus for any valid dining hall location
- Supports all meal types (Breakfast, Lunch, Dinner, Brunch)
- Full test coverage across all packages
- Organized into packages for maintainability
- Validates location and meal type inputs with enum schemas

## Structure

```
.
├── cache/          # In-memory caching with TTL
├── client/         # HTTP client and session management
├── config/         # Configuration and validation
├── parser/         # HTML parsing utilities
├── utils/          # Utility functions
├── main.go         # MCP server entry point
└── http_wrapper.go # HTTP wrapper for curl testing
```

## Installation

```bash
# Run setup script
./setup.sh

# Or manually
go mod download
go build -o diningbot
```

## Usage

### MCP Server

The primary interface is an MCP server that communicates over stdio:

```bash
# Run the MCP server
go run main.go

# Or build and run
go build -o diningbot
./diningbot
```

### HTTP Wrapper (for curl testing)

To test with curl, use the HTTP wrapper:

```bash
# Terminal 1: Start the HTTP wrapper
go run http_wrapper.go

# Terminal 2: Use curl commands (see test_mcp_wrapper.sh)
./test_mcp_wrapper.sh
```

### MCP Tools

The server exposes two tools:

1. **`get_menu`** - Get menu for a specific location, date, and meal type
   - Parameters:
     - `location` (required, enum): Dining hall location
     - `date` (optional): Date in M/D/YYYY format (defaults to today)
     - `mealType` (required, enum): Meal type

2. **`get_menus_range`** - Get menus for multiple days
   - Parameters:
     - `location` (required, enum): Dining hall location
     - `mealType` (required, enum): Meal type
     - `days` (optional): Number of days (default: 7, max: 30)
     - `startDate` (optional): Start date in M/D/YYYY format (defaults to today)

### Valid Locations (enum)

- Arrillaga Family Dining Commons
- Branner Dining
- EVGR Dining
- Florence Moore Dining
- Gerhard Casper Dining
- Lakeside Dining
- Ricker Dining
- Stern Dining
- Wilbur Dining

### Valid Meal Types (enum)

- Breakfast
- Lunch
- Dinner
- Brunch

## Caching

The application includes a **1-hour cache** to:
- Reduce load on Stanford servers
- Improve response times for repeated queries
- Cache is shared across all MCP tool calls
- Automatic expiry after 1 hour

Cache keys are based on: `location|date|mealType`

## Testing

### Unit Tests

```bash
go test ./... -cover
```

### Integration Tests

```bash
# All integration tests (includes MCP tests)
go test -tags=integration -v

# Only MCP integration tests
go test -tags=integration -v -run TestMCP

# Specific test
go test -tags=integration -v -run TestMCPGetMenu
```

### Test Coverage

- `cache`: 100.0%
- `client`: 56.6%
- `config`: 85.7%
- `parser`: 73.8%
- `utils`: 100.0%

### Integration Test Coverage

- Full flow from initialization to menu fetching
- Multiple locations and meal types
- Session persistence across requests
- Error handling with invalid inputs
- Debug mode functionality
- MCP server initialization and tool listing
- MCP tool execution with schema validation
- Cache functionality

## Using with MCP Clients

Configure your MCP client (e.g., Claude Desktop) to connect to:

```json
{
  "mcpServers": {
    "diningbot": {
      "command": "go",
      "args": ["run", "main.go"],
      "cwd": "/path/to/diningbot"
    }
  }
}
```

Or use the built binary:

```json
{
  "mcpServers": {
    "diningbot": {
      "command": "./diningbot",
      "cwd": "/path/to/diningbot"
    }
  }
}
```

## Testing with curl

See `test_mcp_wrapper.sh` for example curl commands. The HTTP wrapper runs on `http://localhost:8080/mcp`.

## Dependencies

- `github.com/modelcontextprotocol/go-sdk/mcp` - MCP server SDK
- `golang.org/x/net/html` - HTML parsing
- `golang.org/x/net/publicsuffix` - Cookie jar support

## License

See LICENSE file for details.
