#!/bin/bash

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${BOLD}${BLUE}========================================${NC}"
echo -e "${BOLD}${BLUE}  MCP Server CURL Commands${NC}"
echo -e "${BOLD}${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}Note:${NC} Start the MCP server first: ${CYAN}PORT=8080 go run main.go${NC}"
echo -e "${YELLOW}      ${NC}Or use Docker: ${CYAN}docker run -p 8080:8080 diningbot-mcp${NC}"
echo ""

# Helper function to extract JSON from SSE response
extract_json() {
    grep "^data:" | sed 's/^data: //' | jq '.'
}

# Helper function to extract session ID from headers
get_session_id() {
    grep -i "mcp-session-id" | cut -d' ' -f2 | tr -d '\r'
}

# 1. Initialize and get session ID
echo -e "${BOLD}${GREEN}1. Initialize${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
INIT_RESPONSE=$(curl -s -i -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "MCP-Protocol-Version: 2025-06-18" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-06-18",
      "capabilities": {},
      "clientInfo": {
        "name": "curl-client",
        "version": "1.0.0"
      }
    }
  }')

SESSION_ID=$(echo "$INIT_RESPONSE" | get_session_id)
echo "$INIT_RESPONSE" | extract_json
echo ""
echo "Session ID: $SESSION_ID"
echo ""

# 2. Send initialized notification
echo -e "${BOLD}${GREEN}2. Send Initialized Notification${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "MCP-Protocol-Version: 2025-06-18" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "method": "notifications/initialized"
  }' > /dev/null
echo "✓ Initialized"
echo ""
echo ""

# 3. List Tools
echo -e "${BOLD}${GREEN}3. List Tools${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "MCP-Protocol-Version: 2025-06-18" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }' | extract_json
echo ""
echo ""

# 4. Get Menu
echo -e "${BOLD}${GREEN}4. Get Menu${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
TODAY=$(date +%-m/%-d/%Y)
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "MCP-Protocol-Version: 2025-06-18" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d "{
    \"jsonrpc\": \"2.0\",
    \"id\": 3,
    \"method\": \"tools/call\",
    \"params\": {
      \"name\": \"get_menu\",
      \"arguments\": {
        \"location\": \"Arrillaga Family Dining Commons\",
        \"date\": \"$TODAY\",
        \"mealType\": \"Lunch\"
      }
    }
  }" | extract_json
echo ""
echo ""

# 5. Get Menus Range
echo -e "${BOLD}${GREEN}5. Get Menus Range${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -H "MCP-Protocol-Version: 2025-06-18" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "get_menus_range",
      "arguments": {
        "location": "Branner Dining",
        "mealType": "Lunch",
        "days": 3
      }
    }
  }' | extract_json
echo ""
echo ""
echo -e "${BOLD}${GREEN}========================================${NC}"
echo -e "${BOLD}${GREEN}  Complete!${NC}"
echo -e "${BOLD}${GREEN}========================================${NC}"
