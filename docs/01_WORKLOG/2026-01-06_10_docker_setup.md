# Worklog: Docker Setup

**Date:** 2026-01-06  
**Story:** [00_STORY_07_docker_setup.md](../00_BACKLOG/00_STORY_07_docker_setup.md)  
**Status:** ✅ Complete

---

## Summary

Successfully implemented Docker containerization for TinyRSVP with multi-stage builds, optimized image size, and comprehensive testing.

---

## Work Completed

### 1. Test Scripts Created
- **test_docker_build.sh**: Validates Docker image build, size, structure, and security
- **test_docker_run.sh**: Tests runtime behavior, health checks, and volume persistence

### 2. Docker Configuration Files

#### Dockerfile
- Multi-stage build using Go 1.24 Alpine
- Build stage: Compiles binary with CGO enabled for SQLite
- Runtime stage: Minimal Alpine image with only required dependencies
- Final image size: **29.7MB** (well under 50MB requirement)
- Security: Runs as non-root user (tinyrsvp:1000)
- Health check integrated using wget

#### docker-compose.yml
- Single-service configuration for local development
- Environment variables with sensible defaults
- Volume mount for data persistence
- Health check configuration
- Removed obsolete `version` field for Docker Compose v2

#### .dockerignore
- Excludes documentation, development files, and build artifacts
- Includes vendor directory for offline builds

### 3. Vendor Directory Solution
- Created `vendor/` directory with `go mod vendor`
- Resolved DNS issues in Docker build environment (WSL2 with internal DNS)
- Dockerfile uses `-mod=vendor` flag for offline builds
- Ensures reproducible builds without network dependencies

### 4. Testing Results

#### Build Tests (test_docker_build.sh)
✅ Image builds successfully  
✅ Image size: 29.7MB (< 50MB requirement)  
✅ Binary exists and is executable  
✅ Migrations directory present  
✅ Templates directory present  
✅ Static files directory present  
✅ Running as non-root user (tinyrsvp:1000)  
✅ Data directory created with correct permissions

#### Runtime Tests (test_docker_run.sh)
✅ Container starts successfully  
✅ Health check passes  
✅ Health endpoint responds (http://localhost:8080/health)  
✅ Readiness endpoint responds (http://localhost:8080/ready)  
✅ Database file created (/data/tinyrsvp.db)  
✅ No errors in container logs  
✅ Volume persistence verified across restarts

---

## Technical Decisions

### 1. Vendor Directory Approach
**Problem:** Docker build failed with DNS resolution errors in WSL2 environment  
**Solution:** Use `go mod vendor` to bundle dependencies  
**Rationale:** 
- Eliminates network dependency during build
- Ensures reproducible builds
- Common practice for containerized Go applications
- Slightly larger build context but more reliable

### 2. Multi-Stage Build
**Benefit:** Reduced final image size from ~200MB to 29.7MB  
**Approach:**
- Builder stage: Full Go toolchain + build dependencies
- Runtime stage: Minimal Alpine + runtime libraries only
- Only copy compiled binary and required assets

### 3. Non-Root User
**Security:** Container runs as tinyrsvp:1000  
**Benefit:** Follows security best practices, reduces attack surface

### 4. Health Check Integration
**Implementation:** wget-based health check every 30s  
**Benefit:** Docker and orchestrators can monitor container health

---

## Files Created/Modified

### Created
- `Dockerfile` - Multi-stage Docker build configuration
- `docker-compose.yml` - Local development compose file
- `.dockerignore` - Build context exclusions
- `test_docker_build.sh` - Build validation script
- `test_docker_run.sh` - Runtime validation script
- `vendor/` - Go module dependencies (generated)

### Modified
- None (all new files)

---

## Commands for Reference

### Build and Test
```bash
# Build image
docker build -t tinyrsvp:latest .

# Run build tests
./test_docker_build.sh

# Run runtime tests
./test_docker_run.sh
```

### Local Development
```bash
# Start services
docker compose up -d

# View logs
docker compose logs -f tinyrsvp

# Stop services
docker compose down

# Clean up including volumes
docker compose down -v
```

### Production Deployment
```bash
# Build with version tag
docker build -t tinyrsvp:v0.1.0 .

# Tag for registry
docker tag tinyrsvp:v0.1.0 registry.example.com/tinyrsvp:v0.1.0

# Push to registry
docker push registry.example.com/tinyrsvp:v0.1.0
```

---

## Acceptance Criteria Status

- [x] Dockerfile creates optimized multi-stage build
- [x] Docker image builds successfully
- [x] docker-compose.yml for local development
- [x] Application runs in container
- [x] Database persists via volume mount
- [x] Health checks integrated in Docker
- [x] Image size optimized (29.7MB < 50MB)
- [x] All tests pass with timeout

---

## Next Steps

1. ✅ Epic 00 (Foundation) is now complete
2. Ready to begin Epic 01 (Authentication)
3. Consider adding:
   - CI/CD pipeline for automated builds
   - Multi-architecture builds (amd64, arm64)
   - Docker Hub or ECR publishing workflow

---

## Notes

### WSL2 DNS Issue
Encountered DNS resolution failures during `go mod download` in Docker build. The WSL2 environment uses an internal DNS server (10.255.255.254) that Docker containers cannot reach. Resolved by using vendor directory approach instead of attempting to configure DNS in containers.

### Docker Compose v2
Updated test scripts to use `docker compose` (v2) instead of `docker-compose` (v1). Removed obsolete `version` field from docker-compose.yml.

### Environment Variables
Added `SERVER_BASE_URL` to docker-compose.yml as it's required by the configuration validation. Set default to `http://localhost:8080` for local development.

---

## References

- **Story:** [00_STORY_07_docker_setup.md](../00_BACKLOG/00_STORY_07_docker_setup.md)
- **Epic:** [00_EPIC_foundation.md](../00_BACKLOG/00_EPIC_foundation.md)
- **Docker Best Practices:** https://docs.docker.com/develop/dev-best-practices/
- **Multi-Stage Builds:** https://docs.docker.com/build/building/multi-stage/
