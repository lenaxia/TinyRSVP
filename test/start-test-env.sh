#!/bin/bash
###############################################################################
# TinyRSVP Test Environment Startup Script
###############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         TinyRSVP Test Environment Startup                 ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}✗ Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Docker is running${NC}"
echo ""

# Stop and remove existing containers
echo -e "${YELLOW}→ Stopping existing test containers...${NC}"
docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
echo ""

# Build the TinyRSVP image
echo -e "${YELLOW}→ Building TinyRSVP image...${NC}"
docker-compose -f docker-compose.test.yml build
echo ""

# Start all services
echo -e "${YELLOW}→ Starting all services...${NC}"
docker-compose -f docker-compose.test.yml up -d
echo ""

# Wait for services to be healthy
echo -e "${YELLOW}→ Waiting for services to be ready...${NC}"
echo ""

# Function to check service health
check_service() {
    local service=$1
    local url=$2
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ $service is ready${NC}"
            return 0
        fi
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo -e "${RED}✗ $service failed to start${NC}"
    return 1
}

# Check MailHog
echo -n "  Checking MailHog... "
check_service "MailHog" "http://localhost:8025"

# Check Authelia
echo -n "  Checking Authelia... "
check_service "Authelia" "http://localhost:9091/api/health"

# Check TinyRSVP
echo -n "  Checking TinyRSVP... "
check_service "TinyRSVP" "http://localhost:8080/health"

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║           Test Environment Ready!                         ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Access URLs:${NC}"
echo -e "  ${YELLOW}TinyRSVP Application:${NC}  http://localhost:8080"
echo -e "  ${YELLOW}MailHog (Email UI):${NC}    http://localhost:8025"
echo -e "  ${YELLOW}Authelia (OIDC):${NC}       http://localhost:9091"
echo -e "  ${YELLOW}Traefik Dashboard:${NC}     http://localhost:8082"
echo ""
echo -e "${BLUE}Test Credentials:${NC}"
echo -e "  ${YELLOW}Admin User:${NC}     admin / admin123"
echo -e "  ${YELLOW}Test User:${NC}      testuser / test123"
echo -e "  ${YELLOW}Guest User:${NC}     guest / guest123"
echo ""
echo -e "${BLUE}Quick Start:${NC}"
echo -e "  1. Open http://localhost:8080"
echo -e "  2. Click 'Login' - you'll be redirected to Authelia"
echo -e "  3. Login with: ${YELLOW}admin / admin123${NC}"
echo -e "  4. You'll be redirected back to TinyRSVP as an admin"
echo -e "  5. Create an event and send invites"
echo -e "  6. Check http://localhost:8025 to see sent emails"
echo ""
echo -e "${BLUE}View Logs:${NC}"
echo -e "  docker-compose -f docker-compose.test.yml logs -f [service]"
echo -e "  Services: tinyrsvp, mailhog, authelia, traefik"
echo ""
echo -e "${BLUE}Stop Environment:${NC}"
echo -e "  docker-compose -f docker-compose.test.yml down"
echo ""
echo -e "${GREEN}Happy Testing! 🎉${NC}"
