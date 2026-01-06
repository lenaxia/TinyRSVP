# User Story: Docker Setup

**Epic:** [00_EPIC_foundation.md](00_EPIC_foundation.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 4 hours

---

## User Story

As an **operator**, I want **Docker containerization** so that **the application can be easily deployed in any environment**.

---

## Acceptance Criteria

- [ ] Dockerfile creates optimized multi-stage build
- [ ] Docker image builds successfully
- [ ] docker-compose.yml for local development
- [ ] Application runs in container
- [ ] Database persists via volume mount
- [ ] Health checks integrated in Docker
- [ ] Image size optimized (< 50MB)
- [ ] All tests pass with timeout

---

## Technical Details

### Multi-Stage Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o tinyrsvp ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/tinyrsvp .

# Copy migrations
COPY --from=builder /build/migrations ./migrations

# Copy templates and static files
COPY --from=builder /build/templates ./templates
COPY --from=builder /build/static ./static

# Create data directory
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run as non-root user
RUN addgroup -g 1000 tinyrsvp && \
    adduser -D -u 1000 -G tinyrsvp tinyrsvp && \
    chown -R tinyrsvp:tinyrsvp /app /data

USER tinyrsvp

# Start application
CMD ["./tinyrsvp"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  tinyrsvp:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: tinyrsvp
    ports:
      - "8080:8080"
    environment:
      - SERVER_HOST=0.0.0.0
      - SERVER_PORT=8080
      - DATABASE_TYPE=sqlite
      - DATABASE_PATH=/data/tinyrsvp.db
      - SMTP_HOST=${SMTP_HOST:-localhost}
      - SMTP_PORT=${SMTP_PORT:-587}
      - EMAIL_FROM=${EMAIL_FROM:-noreply@example.com}
      - STORAGE_TYPE=local
      - STORAGE_LOCAL_PATH=/data/uploads
    volumes:
      - tinyrsvp-data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s

volumes:
  tinyrsvp-data:
    driver: local
```

### .dockerignore

```
# Git
.git
.gitignore

# Documentation
*.md
docs/

# Development
.vscode/
.idea/

# Build artifacts
bin/
*.exe
*.test
*.out

# Data
data/
*.db
*.db-shm
*.db-wal

# Environment
.env
.env.local

# Temporary
tmp/
temp/
```

---

## Tasks

### Phase 1: Dockerfile Creation (TDD)
- [ ] Write test for Docker image build
- [ ] Write test for binary exists in image
- [ ] Write test for migrations directory exists
- [ ] Create multi-stage Dockerfile
- [ ] Run tests (should pass)

### Phase 2: Docker Compose (TDD)
- [ ] Write test for docker-compose up
- [ ] Write test for health check passes
- [ ] Write test for volume persistence
- [ ] Create docker-compose.yml
- [ ] Run tests (should pass)

### Phase 3: Optimization
- [ ] Measure image size
- [ ] Optimize layers
- [ ] Verify CGO enabled for SQLite
- [ ] Test with minimal Alpine base

### Phase 4: Integration
- [ ] Build and run container locally
- [ ] Test health checks
- [ ] Test volume persistence
- [ ] Document Docker usage

---

## Testing Requirements

### Build Tests

```bash
#!/bin/bash
# test_docker_build.sh

set -e

echo "Building Docker image..."
docker build -t tinyrsvp:test .

echo "Checking image size..."
SIZE=$(docker images tinyrsvp:test --format "{{.Size}}")
echo "Image size: $SIZE"

echo "Verifying binary exists..."
docker run --rm tinyrsvp:test ls -lh /app/tinyrsvp

echo "Verifying migrations exist..."
docker run --rm tinyrsvp:test ls -lh /app/migrations

echo "Verifying non-root user..."
docker run --rm tinyrsvp:test id

echo "All build tests passed!"
```

### Runtime Tests

```bash
#!/bin/bash
# test_docker_run.sh

set -e

echo "Starting container..."
docker-compose up -d

echo "Waiting for health check..."
sleep 10

echo "Checking health endpoint..."
curl -f http://localhost:8080/health || exit 1

echo "Checking readiness endpoint..."
curl -f http://localhost:8080/ready || exit 1

echo "Checking logs..."
docker-compose logs tinyrsvp

echo "Stopping container..."
docker-compose down

echo "All runtime tests passed!"
```

---

## Build Commands

### Local Development

```bash
# Build image
docker build -t tinyrsvp:latest .

# Run with docker-compose
docker-compose up -d

# View logs
docker-compose logs -f tinyrsvp

# Stop
docker-compose down

# Clean up
docker-compose down -v
```

### Production Build

```bash
# Build with version tag
docker build -t tinyrsvp:v0.1.0 .

# Tag for registry
docker tag tinyrsvp:v0.1.0 registry.example.com/tinyrsvp:v0.1.0

# Push to registry
docker push registry.example.com/tinyrsvp:v0.1.0
```

---

## Environment Variables

All configuration via environment variables (12-factor app):

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_BASE_URL=https://rsvp.example.com

# Database
DATABASE_TYPE=sqlite
DATABASE_PATH=/data/tinyrsvp.db

# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
EMAIL_FROM=noreply@example.com

# Storage
STORAGE_TYPE=local
STORAGE_LOCAL_PATH=/data/uploads

# Security
SECURITY_SESSION_DURATION=168h
SECURITY_TOKEN_EXPIRY=720h
```

---

## Volume Mounts

### Data Persistence

```yaml
volumes:
  - tinyrsvp-data:/data
```

**Contains:**
- SQLite database file
- Uploaded files
- Generated files

### Configuration (Optional)

```yaml
volumes:
  - ./config:/config:ro
```

---

## Dependencies

**Depends on:** 
- [00_STORY_go_module_setup.md](00_STORY_go_module_setup.md)
- [00_STORY_config_management.md](00_STORY_config_management.md)
- [00_STORY_database_connection.md](00_STORY_database_connection.md)
- [00_STORY_database_migrations.md](00_STORY_database_migrations.md)
- [00_STORY_health_checks.md](00_STORY_health_checks.md)

**Blocks:** 
- Production deployment
- All other epics (deployment dependency)

---

## Definition of Done

- [ ] All acceptance criteria met
- [ ] All tasks completed
- [ ] Docker image builds successfully
- [ ] Image size < 50MB
- [ ] docker-compose up works
- [ ] Health checks pass in container
- [ ] Volume persistence verified
- [ ] Non-root user verified
- [ ] Documentation complete
- [ ] Changes committed to git

---

## Notes

### Multi-Stage Build Benefits
- Smaller final image (no build tools)
- Faster deployment
- More secure (minimal attack surface)
- Cached layers for faster rebuilds

### Security Considerations
- Run as non-root user
- Minimal base image (Alpine)
- No unnecessary packages
- Health checks for monitoring
- Read-only filesystem where possible

### Optimization Tips
- Order Dockerfile commands by change frequency
- Combine RUN commands to reduce layers
- Use .dockerignore to exclude unnecessary files
- Leverage build cache effectively

### Homelab Deployment
- Use docker-compose for simplicity
- Persistent volumes for data
- Traefik/Nginx reverse proxy
- Let's Encrypt for TLS

---

## References

- **README-LLM.md:** Docker-first experience
- **HLD:** Section 20 (Deployment)
- **Docker Best Practices:** https://docs.docker.com/develop/dev-best-practices/
- **Multi-Stage Builds:** https://docs.docker.com/build/building/multi-stage/
