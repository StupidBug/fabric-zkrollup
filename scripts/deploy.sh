#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default port
PORT=8080

# Function to kill zkrollup processes
kill_zkrollup() {
    echo -e "${GREEN}Cleaning up zkrollup processes...${NC}"
    pkill -f zkrollup || true
    sleep 1
}

# Main deployment process
echo -e "${GREEN}Starting deployment...${NC}"

# Kill existing processes
kill_zkrollup

# Build project
echo -e "${GREEN}Building project...${NC}"
go build ./cmd/zkrollup || {
    echo -e "${RED}Build failed${NC}"
    exit 1
}

# Start server
echo -e "${GREEN}Starting server on port $PORT...${NC}"
./zkrollup

# Wait for server to start
echo -e "${GREEN}Waiting for server to start...${NC}"
for i in {1..5}; do
    if curl -s "http://localhost:$PORT/api/v1/state/root" > /dev/null; then
        echo -e "${GREEN}Server is running on port $PORT${NC}"
        exit 0
    fi
    sleep 1
done

echo -e "${RED}Server failed to start${NC}"
kill_zkrollup
exit 1 