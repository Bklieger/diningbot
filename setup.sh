#!/bin/bash

# Setup script for diningbot
# This script checks prerequisites, installs dependencies, and optionally starts the app

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== diningbot Setup Script ===${NC}\n"

# Check if Go is installed
echo -e "${YELLOW}Checking for Go installation...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed.${NC}"
    echo "Please install Go from https://go.dev/dl/"
    exit 1
fi

# Check Go version (requires 1.24.0 or higher)
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${GREEN}Found Go version: $GO_VERSION${NC}"

# Download dependencies
echo -e "\n${YELLOW}Downloading dependencies...${NC}"
go mod download
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Dependencies downloaded successfully!${NC}"
else
    echo -e "${RED}Error: Failed to download dependencies${NC}"
    exit 1
fi

# Verify dependencies
echo -e "\n${YELLOW}Verifying dependencies...${NC}"
go mod verify
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Dependencies verified!${NC}"
else
    echo -e "${RED}Error: Dependency verification failed${NC}"
    exit 1
fi

# Build the application
echo -e "\n${YELLOW}Building application...${NC}"
go build -o diningbot
if [ $? -eq 0 ]; then
    echo -e "${GREEN}Build successful!${NC}"
else
    echo -e "${RED}Error: Build failed${NC}"
    exit 1
fi

echo -e "\n${GREEN}=== Setup Complete! ===${NC}\n"

# Check if user wants to run the app
if [ "$1" == "--run" ] || [ "$1" == "-r" ]; then
    echo -e "${YELLOW}Starting application...${NC}\n"
    shift  # Remove --run/-r flag
    ./diningbot "$@"
elif [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    echo "Usage:"
    echo "  ./setup.sh              - Setup only (download deps and build)"
    echo "  ./setup.sh --run        - Setup and run with default settings"
    echo "  ./setup.sh --run [args] - Setup and run with custom arguments"
    echo ""
    echo "Examples:"
    echo "  ./setup.sh --run"
    echo "  ./setup.sh --run \"Branner Dining\" \"Lunch\""
    echo "  ./setup.sh --run \"Arrillaga Family Dining Commons\" \"Breakfast\" --json"
    echo ""
    echo "After setup, you can also run:"
    echo "  ./diningbot [location] [meal] [--json] [--debug]"
    echo "  make run"
    echo "  go run main.go [location] [meal] [--json] [--debug]"
else
    echo "Setup complete! You can now:"
    echo "  - Run: ./diningbot [location] [meal] [--json] [--debug]"
    echo "  - Or: make run"
    echo "  - Or: go run main.go [location] [meal] [--json] [--debug]"
    echo ""
    echo "To run immediately after setup, use: ./setup.sh --run"
fi

