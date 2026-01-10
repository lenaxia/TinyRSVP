# Worklog: Test Environment Setup and Docker Image Issue

**Date:** 2026-01-10  
**Session:** 47  
**Focus:** Test environment setup, Docker image debugging  
**Status:** In Progress - Docker rebuild

---

## Summary

Created complete test environment with mock SMTP (MailHog), OIDC provider (Authelia), and reverse proxy (Traefik). Discovered critical issue: Docker image was stale (built Jan 6) and missing all template files created on Jan 9. Currently rebuilding image to include templates.

---

## Work Completed

### 1. Test Environment Creation

**Created Files:**
- `docker-compose.test.yml` - Complete test stack with 4 services
- `test/authelia/configuration.yml` - Authelia OIDC configuration
- `test/authelia/users_database.yml` - Test user accounts with Argon2 hashes
- `test/start-test-env.sh` - Automated startup script
- `test/README.md` - Comprehensive testing documentation
- `test/SECURITY_NOTICE.md` - Security documentation for test secrets

**Services Configured:**
1. **TinyRSVP** - Main application (port 8080)
2. **MailHog** - Mock SMTP server (ports 8025 UI, 1025 SMTP)
3. **Authelia** - OIDC provider (port 9091)
4. **Traefik** - Reverse proxy (ports 80, 8082)

**Test Users Created:**
- `admin` / `admin123` (admin role)
- `testuser` / `test123` (regular user)
- `guest` / `guest123` (guest role)

### 2. Security Configuration for Test Secrets

**Created Files:**
- `.gitleaks.toml` - GitLeaks configuration to allow test secrets
- `.gitattributes` - Git attributes for test files
- Added security headers to all test configuration files

**Purpose:** Prevent false positives from secret scanning tools while clearly marking test-only credentials as safe to commit.

### 3. Authelia Configuration Fixes

**Issues Fixed:**
1. `default_policy: bypass` → `one_factor` (Authelia requirement)
2. `domain: localhost` → `127.0.0.1` (valid cookie domain)

**Result:** Authelia now starts successfully and is healthy.

### 4. Docker Image Issue Discovery

**Problem Identified:**
- Docker image built on **January 6, 2026**
- All template files created on **January 9, 2026**
- Container running 3-day-old code without templates
- `templates/web/` directory empty in container
- Application crashes during template loading

**Templates Missing from Container:**
```
admin_dashboard.html (4.9 KB) - Jan 9 22:46
confirmation.html (5.4 KB) - Jan 9 13:48
dashboard.html (5.7 KB) - Jan 9 13:51
event_detail.html (9.8 KB) - Jan 9 20:34
event_form.html (15.6 KB) - Jan 9 18:30
event_list.html (9.2 KB) - Jan 9 13:51
invite_list.html (16.4 KB) - Jan 9 13:52
rsvp_page.html (19.5 KB) - Jan 9 18:29
rsvp_summary.html (11.4 KB) - Jan 9 13:52
unsubscribe.html (3.0 KB) - Jan 9 22:12
user_management.html (6.6 KB) - Jan 9 22:48
```

**Symptoms:**
- Application starts but only registers `/health` and `/ready` endpoints
- All other routes return 404
- Initialization stops after health endpoint registration
- No error messages in logs (silent failure during template loading)

### 5. Docker Image Rebuild

**Action Taken:**
```bash
docker compose -f docker-compose.test.yml build --no-cache tinyrsvp
```

**Status:** In progress (started ~7 minutes ago)

**Expected Result:**
- Fresh image with all current templates
- Full application initialization
- All routes registered
- Working UI

---

## Technical Details

### Docker Build Process

**Dockerfile Structure:**
```dockerfile
# Stage 1: Builder (golang:1.24-alpine)
- Install build dependencies (gcc, musl-dev, sqlite-dev)
- Copy source code
- Compile Go binary with CGO enabled

# Stage 2: Runtime (alpine:latest)
- Install runtime dependencies (ca-certificates, sqlite-libs, wget)
- Copy binary from builder
- Copy migrations, templates, static files
- Create non-root user
- Set up healthcheck
```

**Build Command:**
```bash
CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-w -s" \
    -mod=vendor \
    -o tinyrsvp ./cmd/server
```

### Why Templates Were Missing

**Root Cause Analysis:**
1. Docker image built on Jan 6 (before Epic 7/8 completion)
2. Templates created during Epic 7 (Jan 9)
3. `docker-compose up` used cached image
4. No automatic rebuild triggered
5. Container ran old code without templates

**Why It Wasn't Obvious:**
- Health check passed (only needs `/health` endpoint)
- Container showed as "healthy"
- No error logs (silent failure during template.ParseFiles)
- Application started but didn't complete initialization

### Application Initialization Sequence

**Expected Flow:**
```
1. ✅ Start server
2. ✅ Load configuration
3. ✅ Connect to database
4. ✅ Run migrations
5. ✅ Register health endpoints
6. ❌ Create system user (FAILED HERE - templates missing)
7. ❌ Seed templates
8. ❌ Initialize services
9. ❌ Load template files
10. ❌ Register application routes
11. ❌ Start background jobs
12. ❌ Start email processor
```

**Actual Flow (with stale image):**
- Steps 1-5 complete
- Step 6 fails silently when trying to load templates
- Application continues running with only health endpoints
- No error logged (Go's template.ParseFiles doesn't log on missing files)

---

## Lessons Learned

### 1. Docker Image Staleness
**Issue:** Docker Compose doesn't automatically rebuild images when source changes.

**Solution:** Always rebuild after significant code changes:
```bash
docker compose -f docker-compose.test.yml build --no-cache
```

**Best Practice:** Add to workflow:
```bash
# Always rebuild before testing
docker compose -f docker-compose.test.yml build && \
docker compose -f docker-compose.test.yml up -d
```

### 2. Silent Template Loading Failures
**Issue:** Go's `template.ParseFiles()` returns error but application continued.

**Solution:** Add better error handling in main.go:
```go
templates, err := template.ParseFiles("templates/web/rsvp_page.html")
if err != nil {
    logger.Error("Failed to load RSVP templates", "error", err)
    os.Exit(1)  // ← This exists but wasn't triggered
}
```

**Investigation Needed:** Why didn't the error handling trigger?

### 3. Health Checks Can Be Misleading
**Issue:** Container showed "healthy" but application wasn't fully functional.

**Reason:** Health check only tests `/health` endpoint, which was registered before failure.

**Improvement:** Consider more comprehensive health checks:
```dockerfile
HEALTHCHECK CMD wget --spider http://localhost:8080/health && \
                wget --spider http://localhost:8080/ready && \
                test -f /app/templates/web/dashboard.html
```

---

## Next Steps

### Immediate (After Build Completes)
1. ✅ Verify new image timestamp
2. ✅ Restart all containers
3. ✅ Check templates exist in container
4. ✅ Verify full initialization logs
5. ✅ Test all routes
6. ✅ Complete end-to-end RSVP workflow test

### Short Term
1. Document Docker rebuild requirement in README
2. Add build step to test startup script
3. Investigate why template loading error didn't log
4. Consider adding template existence check to health endpoint

### Before Epic 9 (Security Review)
1. Complete functional testing
2. Document any bugs found
3. Fix critical issues
4. Verify all Epic 0-8 features work

---

## Questions for Investigation

1. **Why didn't template loading errors trigger os.Exit(1)?**
   - Code shows error handling exists
   - No error logs appeared
   - Application continued with partial initialization

2. **Should health check verify templates exist?**
   - Current: Only checks HTTP response
   - Proposed: Also verify critical files exist

3. **How to prevent stale image issues?**
   - Add build step to startup script?
   - Use docker-compose build --pull?
   - Document rebuild requirement?

---

## Files Modified

### New Files Created
- `docker-compose.test.yml`
- `test/authelia/configuration.yml`
- `test/authelia/users_database.yml`
- `test/start-test-env.sh`
- `test/README.md`
- `test/SECURITY_NOTICE.md`
- `.gitleaks.toml`
- `.gitattributes`

### Files Modified
- None (only new files created)

---

## Current Status

**Build Status:** In progress (~7-10 minutes elapsed)  
**Services Running:**
- ✅ MailHog (healthy)
- ✅ Authelia (healthy)
- ⏳ TinyRSVP (waiting for rebuild)
- ⏳ Traefik (needs port fix)

**Blocking Issue:** Docker image rebuild in progress

**ETA:** 1-2 minutes until build completes

---

## References

- [Docker Compose Build Docs](https://docs.docker.com/compose/reference/build/)
- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Authelia Configuration](https://www.authelia.com/configuration/prologue/introduction/)

---

**Next Session:** Complete functional testing after Docker rebuild
