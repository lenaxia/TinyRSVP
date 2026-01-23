#!/bin/bash
set -e

# Mock generation script for TinyRSVP
# This script generates all mocks for interfaces used in testing

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Generating mocks for TinyRSVP...${NC}"

# Ensure mockgen is installed
if ! command -v mockgen &> /dev/null; then
    echo "mockgen not found. Installing..."
    go install go.uber.org/mock/mockgen@latest
fi

# Get GOPATH bin directory
MOCKGEN="$(go env GOPATH)/bin/mockgen"
MOCK_DIR="internal/testutil/mocks"

# Create mocks directory if it doesn't exist
mkdir -p "$MOCK_DIR"

echo -e "${GREEN}Generating repository mocks...${NC}"

# Repository interfaces (will be populated in Stories 06-08)
# Validation test - generate EventRepository mock
$MOCKGEN -source=internal/db/repositories/event_repository.go \
    -destination=$MOCK_DIR/mock_event_repository.go \
    -package=mocks

echo -e "${GREEN}Mock generation complete!${NC}"
echo -e "Generated mocks are in: $MOCK_DIR"
