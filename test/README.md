# TinyRSVP Test Environment

Complete test environment with mock SMTP server (MailHog), OIDC provider (Authelia), and reverse proxy (Traefik) for testing the full RSVP workflow.

## Quick Start

```bash
# Start the test environment
./test/start-test-env.sh

# Or manually
docker-compose -f docker-compose.test.yml up -d
```

## Services

### TinyRSVP Application
- **URL**: http://localhost:8080
- **Purpose**: Main RSVP application
- **Database**: SQLite (in Docker volume)
- **Storage**: Local filesystem (in Docker volume)

### MailHog (Mock SMTP Server)
- **Web UI**: http://localhost:8025
- **SMTP Port**: 1025
- **Purpose**: Captures all outgoing emails for testing
- **Features**:
  - View sent emails in web interface
  - No actual emails sent to real addresses
  - Perfect for testing invite/confirmation emails
  - Inspect email headers, HTML, and attachments

### Authelia (OIDC Provider)
- **URL**: http://localhost:9091
- **Purpose**: Provides OIDC authentication for admin users
- **Test Users**:
  - `admin` / `admin123` (Admin role)
  - `testuser` / `test123` (Regular user)
  - `guest` / `guest123` (Guest role)

### Traefik (Reverse Proxy)
- **Dashboard**: http://localhost:8082
- **Purpose**: Optional reverse proxy for advanced testing
- **Features**: Request routing, load balancing

## Test Workflow

### 1. Admin Login
```
1. Navigate to http://localhost:8080
2. Click "Login" button
3. You'll be redirected to Authelia (http://localhost:9091)
4. Enter credentials: admin / admin123
5. You'll be redirected back to TinyRSVP as an authenticated admin
```

### 2. Create an Event
```
1. From the dashboard, click "Create Event"
2. Fill in event details:
   - Name: "Test Birthday Party"
   - Date/Time: Future date
   - Location: "123 Main St"
   - RSVP Deadline: Before event date
3. Add preference questions (optional):
   - "Dietary restrictions?"
   - "Will you bring a plus one?"
4. Save the event
```

### 3. Send Invites
```
1. Navigate to the event's invite page
2. Option A - Individual Invite:
   - Enter guest name and email
   - Click "Send Invite"
   
3. Option B - Bulk CSV Import:
   - Prepare CSV file with columns: name, email
   - Upload CSV file
   - Review and confirm
   - Click "Send All Invites"
```

### 4. Check Emails in MailHog
```
1. Open http://localhost:8025
2. You'll see all sent invitation emails
3. Click on an email to view:
   - Email subject and body
   - HTML rendering
   - Unique RSVP token link
   - .ics calendar attachment
4. Copy the RSVP link from the email
```

### 5. Guest RSVP (No Login Required)
```
1. Open the RSVP link from the email
   (e.g., http://localhost:8080/rsvp?token=abc123...)
2. Guest sees event details without logging in
3. Select RSVP response: Yes / No / Maybe
4. Answer preference questions
5. Specify number of plus ones (if allowed)
6. Submit RSVP
7. See confirmation page
```

### 6. Check Confirmation Email
```
1. Return to MailHog (http://localhost:8025)
2. Find the confirmation email
3. Verify it contains:
   - RSVP confirmation
   - Event details
   - .ics calendar attachment
   - Link to update RSVP
```

### 7. View RSVP Summary (Admin)
```
1. Log back into TinyRSVP as admin
2. Navigate to event dashboard
3. View RSVP summary:
   - Total responses (Yes/No/Maybe)
   - Guest list with responses
   - Preference question answers
   - Plus one counts
4. Export guest list as CSV (optional)
```

## Testing Scenarios

### Scenario 1: Basic RSVP Flow
Test the complete workflow from event creation to guest RSVP.

### Scenario 2: Plus Ones
1. Create event with plus ones allowed (max 2)
2. Send invite
3. Guest RSVPs with 2 plus ones
4. Verify count in admin dashboard

### Scenario 3: RSVP Updates
1. Guest submits initial RSVP (Yes)
2. Guest uses update link to change to No
3. Verify updated response in admin dashboard
4. Check update confirmation email in MailHog

### Scenario 4: Deadline Enforcement
1. Create event with past RSVP deadline
2. Try to RSVP via token link
3. Verify deadline message is shown
4. Verify RSVP submission is blocked

### Scenario 5: Token Security
1. Try accessing RSVP page without token
2. Try accessing with invalid token
3. Try accessing with revoked token
4. Verify proper error messages

### Scenario 6: Email Attachments
1. Send invite
2. Check MailHog for .ics attachment
3. Download and verify .ics file format
4. Import into calendar app (optional)

## Troubleshooting

### Services Won't Start
```bash
# Check Docker is running
docker info

# View service logs
docker-compose -f docker-compose.test.yml logs -f

# Restart specific service
docker-compose -f docker-compose.test.yml restart tinyrsvp
```

### Can't Access Services
```bash
# Check if ports are already in use
lsof -i :8080  # TinyRSVP
lsof -i :8025  # MailHog
lsof -i :9091  # Authelia

# Stop conflicting services or change ports in docker-compose.test.yml
```

### Authelia Login Issues
```bash
# Check Authelia logs
docker-compose -f docker-compose.test.yml logs authelia

# Verify configuration
cat test/authelia/configuration.yml
cat test/authelia/users_database.yml

# Reset Authelia data
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d
```

### Emails Not Appearing in MailHog
```bash
# Check TinyRSVP logs for SMTP errors
docker-compose -f docker-compose.test.yml logs tinyrsvp | grep -i smtp

# Verify MailHog is running
curl http://localhost:8025

# Check SMTP connection
docker-compose -f docker-compose.test.yml exec tinyrsvp nc -zv mailhog 1025
```

### Database Issues
```bash
# Reset database (WARNING: Deletes all data)
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d

# Access database directly
docker-compose -f docker-compose.test.yml exec tinyrsvp sqlite3 /data/tinyrsvp.db
```

## Advanced Testing

### Testing with Traefik
```bash
# Enable Traefik routing (edit docker-compose.test.yml)
# Add labels to tinyrsvp service:
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.tinyrsvp.rule=Host(`rsvp.localhost`)"
  - "traefik.http.services.tinyrsvp.loadbalancer.server.port=8080"

# Access via Traefik
http://rsvp.localhost
```

### Load Testing
```bash
# Install Apache Bench
sudo apt-get install apache2-utils

# Test RSVP endpoint
ab -n 100 -c 10 http://localhost:8080/health

# Test with authentication
# (requires valid session cookie)
```

### Integration Testing
```bash
# Run Go integration tests against test environment
go test -v ./tests/e2e/... -tags=integration

# Run specific test
go test -v ./tests/e2e/rsvp_test.go -run TestRSVPFlow
```

## Cleanup

### Stop Services (Keep Data)
```bash
docker-compose -f docker-compose.test.yml down
```

### Stop Services (Delete All Data)
```bash
docker-compose -f docker-compose.test.yml down -v
```

### Remove Images
```bash
docker-compose -f docker-compose.test.yml down --rmi all -v
```

## Configuration Files

### docker-compose.test.yml
Main orchestration file defining all services and their configuration.

### test/authelia/configuration.yml
Authelia OIDC provider configuration with test settings.

### test/authelia/users_database.yml
Test user accounts for Authelia authentication.

### test/start-test-env.sh
Automated startup script with health checks.

## Security Notes

⚠️ **WARNING**: This is a TEST ENVIRONMENT ONLY

- All secrets are hardcoded and insecure
- Passwords are simple and publicly known
- No TLS/HTTPS encryption
- No rate limiting or security hardening
- **NEVER use these configurations in production**

## Next Steps

After testing in this environment:

1. Review Epic 9 (Security Review) requirements
2. Run security scans (OWASP ZAP, gosec, etc.)
3. Perform penetration testing
4. Set up production environment with:
   - Real OIDC provider (Authentik, Keycloak, etc.)
   - Real SMTP server (Gmail, SendGrid, etc.)
   - TLS certificates
   - Strong secrets and passwords
   - Security hardening
   - Monitoring and logging

## Support

For issues or questions:
- Check logs: `docker-compose -f docker-compose.test.yml logs -f`
- Review documentation: `docs/`
- Check GitHub issues: https://github.com/lenaxia/tinyrsvp/issues
