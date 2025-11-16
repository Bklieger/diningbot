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
echo -e "${YELLOW}Note:${NC} Start the HTTP wrapper first: ${CYAN}go run http_wrapper.go${NC}"
echo ""

# 1. Initialize
echo -e "${BOLD}${GREEN}1. Initialize${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "curl-client",
        "version": "1.0.0"
      }
    }
  }' | jq '.'
echo ""
echo ""

# 2. List Tools
echo -e "${BOLD}${GREEN}2. List Tools${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }' | jq '.'
echo ""
echo ""

# 3. Get Menu
echo -e "${BOLD}${GREEN}3. Get Menu${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
TODAY=$(date +%-m/%-d/%Y)
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
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
  }" | jq '.'
echo ""
echo ""

# 4. Get Menus Range
echo -e "${BOLD}${GREEN}4. Get Menus Range${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
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
  }' | jq '.'
echo ""
echo ""
echo -e "${BOLD}${GREEN}========================================${NC}"
echo -e "${BOLD}${GREEN}  Complete!${NC}"
echo -e "${BOLD}${GREEN}========================================${NC}"

