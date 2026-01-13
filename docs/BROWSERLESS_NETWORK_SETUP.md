# Browserless Network Setup - Quick Reference

## Overview

This guide explains how to expose TinyRSVP (running on Traefik port 8081) to the browserless Docker container for automated browser testing.

## Network Architecture

```
Host Machine (localhost)
├── Port 8081 → Traefik (traefik-test container)
│   └── Routes to → TinyRSVP (tinyrsvp-tinyrsvp:8080)
├── Port 3000 → Browserless HTTP API
└── Port 9222 → Chrome Remote Debugging

Docker Network: tinyrsvp-test
├── traefik-test (Traefik reverse proxy)
├── tinyrsvp-tinyrsvp (TinyRSVP application)
├── browserless-chrome (Headless Chrome)
├── mailhog-test (Mock SMTP)
└── authelia-test (OIDC provider)
```

## Quick Start

### 1. Start Services in Order

```bash
# Start TinyRSVP test environment (creates the network)
docker-compose -f docker-compose.test.yml up -d

# Start browserless (joins the network)
docker-compose -f docker-compose.browserless.yml up -d
```

### 2. Verify Connectivity

```bash
# Check both containers are running
docker ps | grep -E "tinyrsvp|browserless"

# Verify browserless can reach TinyRSVP via Traefik
docker exec browserless-chrome wget -O- http://traefik-test:80/health

# Or test direct connection to app
docker exec browserless-chrome wget -O- http://tinyrsvp-tinyrsvp:8080/health
```

## Accessing TinyRSVP from Browserless

### From Host Machine (Your Code)

When making API calls to browserless from your host machine, use these URLs:

```bash
# Via Traefik (recommended - includes auth middleware)
curl -X POST http://localhost:3000/screenshot \
  -H "Content-Type: application/json" \
  -d '{"url": "http://traefik-test:80"}' \
  --output screenshot.png

# Direct to app (bypasses Traefik)
curl -X POST http://localhost:3000/screenshot \
  -H "Content-Type: application/json" \
  -d '{"url": "http://tinyrsvp-tinyrsvp:8080"}' \
  --output screenshot.png
```

### From Puppeteer/Playwright Scripts

```javascript
// Via Traefik (port 8081 on host = traefik-test:80 in Docker)
await page.goto('http://traefik-test:80');

// Direct to app
await page.goto('http://tinyrsvp-tinyrsvp:8080');
```

### URL Mapping

| Host Access | Docker Network Access | Description |
|-------------|----------------------|-------------|
| `http://localhost:8081` | `http://traefik-test:80` | Via Traefik with auth |
| `http://localhost:8080` | `http://tinyrsvp-tinyrsvp:8080` | Direct to app |
| `http://localhost:3000` | `http://browserless-chrome:3000` | Browserless API |

## Common Use Cases

### Take Screenshot of RSVP Page

```bash
# Get an event's RSVP page screenshot
curl -X POST http://localhost:3000/screenshot \
  -H "Content-Type: application/json" \
  -d '{
    "url": "http://traefik-test:80/rsvp/some-event-id",
    "options": {
      "fullPage": true,
      "type": "png"
    }
  }' \
  --output rsvp-page.png
```

### Generate PDF of Event Page

```bash
curl -X POST http://localhost:3000/pdf \
  -H "Content-Type: application/json" \
  -d '{
    "url": "http://traefik-test:80/events/123",
    "options": {
      "format": "A4",
      "printBackground": true
    }
  }' \
  --output event-page.pdf
```

### Run Automated Test Script

```bash
curl -X POST http://localhost:3000/function \
  -H "Content-Type: application/json" \
  -d '{
    "code": "async ({ page }) => {
      await page.goto(\"http://traefik-test:80\");
      const title = await page.title();
      return { title, success: true };
    }"
  }'
```

## Troubleshooting

### Error: "Network tinyrsvp-test not found"

**Cause:** TinyRSVP test environment not started.

**Solution:**
```bash
docker-compose -f docker-compose.test.yml up -d
docker-compose -f docker-compose.browserless.yml restart
```

### Error: "Could not resolve host: traefik-test"

**Cause:** Browserless not connected to tinyrsvp-test network.

**Solution:**
```bash
# Check network membership
docker network inspect tinyrsvp-test | grep browserless

# If not present, restart browserless
docker-compose -f docker-compose.browserless.yml down
docker-compose -f docker-compose.browserless.yml up -d
```

### Error: "Connection refused" or "timeout"

**Cause:** TinyRSVP or Traefik not fully started.

**Solution:**
```bash
# Check service health
docker-compose -f docker-compose.test.yml ps

# Wait for services to be healthy
docker-compose -f docker-compose.test.yml logs -f traefik
docker-compose -f docker-compose.test.yml logs -f tinyrsvp

# Test connectivity manually
docker exec browserless-chrome ping -c 3 traefik-test
docker exec browserless-chrome wget --spider http://traefik-test:80/health
```

### Verify Network Configuration

```bash
# List all networks
docker network ls

# Inspect tinyrsvp-test network
docker network inspect tinyrsvp-test

# Should show these containers:
# - tinyrsvp-tinyrsvp
# - traefik-test
# - browserless-chrome
# - mailhog-test
# - authelia-test
```

## Stopping Services

```bash
# Stop browserless
docker-compose -f docker-compose.browserless.yml down

# Stop TinyRSVP test environment
docker-compose -f docker-compose.test.yml down

# Remove network (optional)
docker network rm tinyrsvp-test
```

## Advanced: Custom Network Configuration

If you need to customize the network setup:

### Option 1: Use Host Network (Not Recommended)

```yaml
# In docker-compose.browserless.yml
services:
  browserless:
    network_mode: host
```

This allows browserless to access `localhost:8081` directly, but reduces isolation.

### Option 2: Create Shared Network Manually

```bash
# Create network first
docker network create tinyrsvp-test

# Then start services
docker-compose -f docker-compose.test.yml up -d
docker-compose -f docker-compose.browserless.yml up -d
```

## See Also

- [BROWSERLESS_SETUP.md](./BROWSERLESS_SETUP.md) - Full browserless documentation
- [docker-compose.browserless.yml](../docker-compose.browserless.yml) - Browserless configuration
- [docker-compose.test.yml](../docker-compose.test.yml) - TinyRSVP test environment
