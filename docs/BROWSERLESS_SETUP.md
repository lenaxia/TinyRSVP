# Browserless Chrome Setup for Roo Code

## Overview

This document describes how to set up and use browserless Chrome with remote debugging for Roo Code integration.

## Quick Start

### 1. Start Browserless

```bash
docker-compose -f docker-compose.browserless.yml up -d
```

### 2. Verify It's Running

```bash
# Check container status
docker ps | grep browserless

# Check health
curl http://localhost:3000/pressure

# Get WebSocket endpoint
curl http://localhost:3000/json/version
```

### 3. Configure Roo Code

Roo Code can connect to the Chrome browser through the browserless API:

**Primary Connection:** Use the browserless HTTP API at `http://localhost:3000`

**WebSocket Connection:** Get the dynamic WebSocket URL via:
```bash
curl http://localhost:3000/json | jq -r '.[0].webSocketDebuggerUrl'
```

The browserless instance provides:
- **HTTP API Port:** 3000 (for browserless API and WebSocket proxy)
- **Remote Debugging Port:** 9222 (internal Chrome DevTools Protocol)

## Remote Debugging Configuration

The browserless container is launched with the following Chrome flags:

```
--remote-debugging-address=0.0.0.0
--remote-debugging-port=9222
--no-sandbox
--disable-setuid-sandbox
--disable-dev-shm-usage
--disable-gpu
```

## Connection Methods

### Method 1: Direct Chrome DevTools Protocol

Connect directly to the Chrome instance:

```javascript
// Example using puppeteer-core
const puppeteer = require('puppeteer-core');

const browser = await puppeteer.connect({
  browserWSEndpoint: 'ws://localhost:9222/devtools/browser/<id>'
});
```

### Method 2: Browserless HTTP API

Use the browserless HTTP API:

```bash
# Create a new page
curl -X POST http://localhost:3000/chrome \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

### Method 3: Chrome DevTools UI

1. Open Chrome browser
2. Navigate to: `chrome://inspect`
3. Click "Configure..."
4. Add: `localhost:9222`
5. Your browserless instance will appear under "Remote Target"

## Environment Variables

The browserless container supports these key environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `ENABLE_DEBUGGER` | `true` | Enable remote debugging |
| `CONNECTION_TIMEOUT` | `60000` | Connection timeout in ms |
| `MAX_CONCURRENT_SESSIONS` | `10` | Max concurrent browser sessions |
| `TOKEN` | _(empty)_ | Optional API token for security |
| `CORS` | `true` | Enable CORS for API access |

## Resource Limits

The container is configured with:
- **Shared Memory:** 2GB (`shm_size: 2gb`)
- **CPU:** 99% max
- **Memory:** 99% max

Adjust these in [`docker-compose.browserless.yml`](../docker-compose.browserless.yml) if needed.

## Common Commands

```bash
# Start browserless
docker-compose -f docker-compose.browserless.yml up -d

# Stop browserless
docker-compose -f docker-compose.browserless.yml down

# View logs
docker-compose -f docker-compose.browserless.yml logs -f

# Restart browserless
docker-compose -f docker-compose.browserless.yml restart

# Check resource usage
docker stats browserless-chrome
```

## Troubleshooting

### Connection Refused

If you get connection refused errors:

1. Check container is running: `docker ps | grep browserless`
2. Check logs: `docker-compose -f docker-compose.browserless.yml logs`
3. Verify port is exposed: `docker port browserless-chrome`

### Out of Memory

If Chrome crashes due to memory:

1. Increase `shm_size` in docker-compose file
2. Reduce `MAX_CONCURRENT_SESSIONS`
3. Add memory limits to container

### Slow Performance

1. Check resource usage: `docker stats browserless-chrome`
2. Reduce concurrent sessions
3. Enable `PREBOOT_CHROME=true` (already enabled)

## Security Considerations

### For Development

The current configuration is suitable for local development with no authentication.

### For Production

If exposing browserless to a network:

1. **Enable Token Authentication:**
   ```yaml
   environment:
     - TOKEN=your-secure-token-here
   ```

2. **Restrict Network Access:**
   ```yaml
   ports:
     - "127.0.0.1:3000:3000"
     - "127.0.0.1:9222:9222"
   ```

3. **Use Reverse Proxy:**
   - Add Traefik/Nginx in front
   - Enable TLS
   - Add rate limiting

## Integration with TinyRSVP

To integrate browserless with the TinyRSVP test environment:

### Option 1: Separate Compose File (Current)

Keep browserless separate and start it independently:

```bash
docker-compose -f docker-compose.browserless.yml up -d
docker-compose -f docker-compose.test.yml up -d
```

### Option 2: Merge into Test Environment

Add the browserless service to [`docker-compose.test.yml`](../docker-compose.test.yml):

```yaml
services:
  # ... existing services ...
  
  browserless:
    image: browserless/chrome:latest
    container_name: browserless-chrome
    ports:
      - "3000:3000"
      - "9222:9222"
    # ... rest of configuration ...
    networks:
      - tinyrsvp-test
```

## API Examples

### Get Browser Version

```bash
curl http://localhost:3000/json/version
```

### Take Screenshot

```bash
curl -X POST http://localhost:3000/screenshot \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}' \
  --output screenshot.png
```

### Generate PDF

```bash
curl -X POST http://localhost:3000/pdf \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}' \
  --output page.pdf
```

### Execute Script

```bash
curl -X POST http://localhost:3000/function \
  -H "Content-Type: application/json" \
  -d '{
    "code": "async ({ page }) => { await page.goto(\"https://example.com\"); return await page.title(); }"
  }'
```

## References

- [Browserless Documentation](https://docs.browserless.io/)
- [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/)
- [Puppeteer Documentation](https://pptr.dev/)

## Support

For issues with:
- **Browserless:** https://github.com/browserless/chrome/issues
- **TinyRSVP Integration:** See project README
