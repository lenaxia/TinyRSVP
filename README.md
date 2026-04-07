# TinyRSVP

> **Note for LLMs:** If you are an AI assistant working on this project, you MUST read [`README-LLM.md`](README-LLM.md) before doing anything else. This file is for human users.

A self-hosted, privacy-focused RSVP and invitation platform for family events, clubs, and private gatherings.

---

## Overview

TinyRSVP is a lightweight alternative to services like Evite, designed for self-hosting in homelab environments. It prioritizes simplicity, privacy, and ease of deployment.

**Key Features:**
- **Privacy-First** — self-hosted, no data sharing with third parties
- **No Guest Accounts** — token-based access, guests just click a link
- **Email Invitations** — personalized invites with .ics calendar attachments
- **Theme Picker** — 7 built-in themes, custom colors and images
- **Mobile-Friendly** — responsive design for all devices
- **Docker-Ready** — single container, multi-arch (amd64 + arm64)
- **Flexible Auth** — Forward Auth (Traefik/Caddy) or OIDC for admins

---

## Quick Start

### Using the published image (recommended)

```bash
docker run -d \
  --name tinyrsvp \
  -p 8080:8080 \
  -v tinyrsvp-data:/data \
  -e TOKEN_SECRET=$(openssl rand -hex 32) \
  -e SERVER_BASE_URL=http://localhost:8080 \
  -e DATABASE_PATH=/data/tinyrsvp.db \
  -e STORAGE_TYPE=local \
  -e STORAGE_LOCAL_PATH=/data/uploads \
  -e SMTP_HOST=your-smtp-host \
  -e SMTP_PORT=587 \
  -e EMAIL_FROM=noreply@example.com \
  -e FORWARD_AUTH_ENABLED=true \
  -e FORWARD_AUTH_USER_HEADER=X-Forwarded-User \
  -e FORWARD_AUTH_EMAIL_HEADER=X-Forwarded-Email \
  -e FORWARD_AUTH_TRUSTED_IPS=172.16.0.0/12 \
  ghcr.io/lenaxia/tinyrsvp:latest
```

### Using Docker Compose

```bash
git clone https://github.com/lenaxia/TinyRSVP.git
cd TinyRSVP
cp .env.example .env
# Edit .env with your settings
docker compose up -d
```

---

## Configuration

### Required Variables

| Variable | Description |
|----------|-------------|
| `TOKEN_SECRET` | 256-bit secret for signing invite tokens. **Required — app refuses to start without it.** Generate with `openssl rand -hex 32`. |
| `SERVER_BASE_URL` | Public URL guests use to reach TinyRSVP. Used in all email links and .ics attachments. |
| `SMTP_HOST` | SMTP server hostname. |
| `EMAIL_FROM` | From address for outbound email. |
| One of: `FORWARD_AUTH_ENABLED=true` or `OIDC_ENABLED=true` | Authentication method (see below). |

### Authentication

TinyRSVP has no local user accounts. Admins authenticate via a reverse proxy (Forward Auth) or an OIDC provider. You must configure one.

#### Option A — Forward Auth (tested, recommended for beta)

Your reverse proxy injects `X-Forwarded-User` and `X-Forwarded-Email` headers after authenticating the user. Works with Traefik + Authelia, Caddy + authz middleware, nginx, etc.

```bash
FORWARD_AUTH_ENABLED=true
FORWARD_AUTH_USER_HEADER=X-Forwarded-User
FORWARD_AUTH_EMAIL_HEADER=X-Forwarded-Email
# Individual IPs or CIDR ranges are both accepted:
FORWARD_AUTH_TRUSTED_IPS=127.0.0.1,172.16.0.0/12
```

`FORWARD_AUTH_TRUSTED_IPS` must include the IP of your reverse proxy container. Using a CIDR like `172.16.0.0/12` covers all Docker bridge subnets (`172.16.x.x`–`172.31.x.x`) regardless of which subnet Docker assigns.

#### Option B — OIDC (implemented, not yet integration-tested in beta)

```bash
OIDC_ENABLED=true
OIDC_ISSUER_URL=https://auth.example.com/application/o/tinyrsvp/
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_REDIRECT_URL=https://rsvp.example.com/auth/oidc/callback
```

OIDC is fully implemented (go-oidc library, full login/callback/logout flow) but has not been integration-tested against a real provider in this beta. Unit tests pass with a mock provider. If you try it, please report results via GitHub Issues.

### SMTP

```bash
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-username
SMTP_PASSWORD=your-password
EMAIL_FROM=noreply@example.com
# SMTP_TLS=true     # default: true (STARTTLS). Set false for port 25 or MailHog.
```

### Full Reference

See [`.env.example`](.env.example) for the complete annotated list of all environment variables.

### Docker Compose Example

```yaml
services:
  tinyrsvp:
    image: ghcr.io/lenaxia/tinyrsvp:latest
    ports:
      - "8080:8080"
    volumes:
      - tinyrsvp-data:/data
    environment:
      - SERVER_BASE_URL=https://rsvp.yourdomain.com
      - TOKEN_SECRET=${TOKEN_SECRET}               # openssl rand -hex 32
      - DATABASE_PATH=/data/tinyrsvp.db
      - STORAGE_TYPE=local
      - STORAGE_LOCAL_PATH=/data/uploads
      - FORWARD_AUTH_ENABLED=true
      - FORWARD_AUTH_USER_HEADER=X-Forwarded-User
      - FORWARD_AUTH_EMAIL_HEADER=X-Forwarded-Email
      - FORWARD_AUTH_TRUSTED_IPS=172.16.0.0/12
      - SMTP_HOST=${SMTP_HOST}
      - SMTP_PORT=${SMTP_PORT:-587}
      - EMAIL_FROM=${EMAIL_FROM}
      - SMTP_USERNAME=${SMTP_USERNAME}
      - SMTP_PASSWORD=${SMTP_PASSWORD}
    restart: unless-stopped

volumes:
  tinyrsvp-data:
```

---

## Features

### For Event Organizers

- **Create Events** — set title, date/time, location, RSVP deadline, timezone
- **Manage Invites** — send personalized email invitations individually or via CSV bulk import
- **Track Responses** — see who's coming, who declined, who hasn't replied
- **Preference Questions** — ask custom questions (dietary restrictions, t-shirt size, song requests, etc.)
- **Theme Picker** — 7 built-in themes (Wedding Elegance, Birthday Celebration, Garden Party, Corporate Professional, Holiday Festive, Modern Minimalist, and more), color overrides, custom header images
- **Lifecycle Management** — draft → published → cancelled → archived, with automatic archiving of past events

### For Guests

- **No Account Required** — click the link in the email, done
- **Calendar Integration** — .ics attachment adds the event to any calendar app
- **Update Anytime** — change RSVP response until the deadline
- **Plus Ones** — specify number of additional guests
- **Unsubscribe** — opt out of reminder emails at any time

---

## Architecture

```
Browser / Mobile
      │
      ▼
Reverse Proxy  (Traefik, Caddy, nginx)
  - TLS termination
  - Authentication  (Forward Auth or OIDC)
      │
      ▼
TinyRSVP Application  :8080
  - Event management
  - Invite & token system
  - RSVP handling
  - Theme / template rendering
  - Email queue processor
      │              │
      ▼              ▼
   SQLite         Local storage
  /data/          /data/uploads/
  tinyrsvp.db
      │
      ▼
   SMTP server
```

---

## Technology Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24 |
| Frontend | Plain CSS + Vanilla JS (no build step) |
| Templates | Go `html/template` |
| Database | SQLite (PostgreSQL planned for v1) |
| Authentication | Forward Auth (tested) or OIDC (implemented) |
| Storage | Local filesystem (S3-compatible planned for v1) |
| Container | Docker, multi-arch (linux/amd64, linux/arm64) |

---

## Deployment

### Homelab (Recommended)

Works great on:
- Raspberry Pi 4/5 (arm64 image available)
- NAS devices (Synology, QNAP) with Docker
- Home servers
- Any Docker host

Minimum requirements: 256MB RAM, 1GB disk, Docker.

### Cloud / VPS

Works on any VPS that runs Docker. Typical deployment: VPS + Traefik as reverse proxy + Authelia for authentication.

### Typical Traefik + Authelia Setup

Traefik handles TLS termination and authentication. After a successful login, Traefik injects `X-Forwarded-User` and `X-Forwarded-Email` headers. TinyRSVP trusts those headers from the Traefik container IP.

Set `FORWARD_AUTH_TRUSTED_IPS` to include Traefik's container IP or the Docker bridge CIDR (`172.16.0.0/12`).

---

## Building from Source

```bash
git clone https://github.com/lenaxia/TinyRSVP.git
cd TinyRSVP

# Download dependencies
go mod download

# Run tests
go test -timeout 30s ./...

# Build binary (requires CGO for SQLite)
CGO_ENABLED=1 go build -o bin/tinyrsvp ./cmd/server

# Run (set required env vars first)
export TOKEN_SECRET=$(openssl rand -hex 32)
export SERVER_BASE_URL=http://localhost:8080
export DATABASE_PATH=/tmp/tinyrsvp.db
export STORAGE_TYPE=local
export STORAGE_LOCAL_PATH=/tmp/tinyrsvp-uploads
export SMTP_HOST=localhost
export SMTP_PORT=1025
export EMAIL_FROM=noreply@localhost
export FORWARD_AUTH_ENABLED=true
export FORWARD_AUTH_USER_HEADER=X-Forwarded-User
export FORWARD_AUTH_EMAIL_HEADER=X-Forwarded-Email
export "FORWARD_AUTH_TRUSTED_IPS=127.0.0.1,::1"
./bin/tinyrsvp
```

### Project Structure

```
TinyRSVP/
├── cmd/server/         # Application entrypoint
├── internal/           # Core application packages
│   ├── auth/           # OIDC + Forward Auth
│   ├── email/          # SMTP sender, queue processor
│   ├── events/         # Event service
│   ├── handlers/       # HTTP handlers and router
│   ├── invites/        # Invite + token management
│   ├── rsvp/           # RSVP service
│   └── templates/      # Theme/template engine
├── pkg/                # Shared packages (ICS, token generation)
├── templates/          # HTML templates (web + email)
├── static/             # CSS, JS, images
├── migrations/sqlite/  # Database migrations
├── docs/               # Backlog, worklogs, design docs
└── tests/              # e2e and UX test suites
```

---

## Roadmap

### v0.1 (Current — beta)
- [x] Event CRUD with lifecycle management
- [x] Cryptographic invite tokens (256-bit, HMAC-hashed)
- [x] Bulk CSV invite import
- [x] RSVP (yes/no/maybe, plus ones, preference questions)
- [x] Email queue with retry, ICS attachments, unsubscribe
- [x] 7 built-in themes, color overrides, custom images
- [x] Forward Auth + OIDC admin authentication
- [x] Mobile-responsive UI
- [x] Docker / Docker Compose, multi-arch image (amd64 + arm64)
- [x] Published to GHCR: `ghcr.io/lenaxia/tinyrsvp`

### v1 (Planned)
- [ ] PostgreSQL support
- [ ] S3-compatible storage
- [ ] Reminder scheduling UI
- [ ] Security audit (OWASP Top 10)
- [ ] Multi-language support

### v2 (Future)
- [ ] Guest OIDC (optional passwordless accounts)
- [ ] SMS notifications
- [ ] CalDAV calendar sync
- [ ] Advanced template editor (UI)
- [ ] API for integrations

---

## FAQ

**Q: Do guests need accounts?**
No. Guests access their invite via a unique link. No signup, no password.

**Q: How do I create the first admin user?**
TinyRSVP has no local accounts. The first person to log in via your auth provider (Forward Auth or OIDC) is automatically created as an admin. Just set up your reverse proxy or OIDC provider and log in.

**Q: What SMTP providers work?**
Any SMTP server: Gmail (app password), SendGrid, Mailgun, Fastmail, Zoho, self-hosted Postfix, MailHog for local testing, etc.

**Q: Is this production-ready?**
v0.1 is feature-complete and suitable for homelab / private use behind a reverse proxy. A formal security audit is planned before recommending public internet exposure to untrusted users.

**Q: Can I use this on a Raspberry Pi?**
Yes — the `arm64` image is published alongside `amd64` in the same manifest.

**Q: How do I back up my data?**
Copy the SQLite database file (`/data/tinyrsvp.db`) and the uploads directory (`/data/uploads/`). For Docker: `docker cp tinyrsvp:/data ./backup`.

**Q: What if I lose TOKEN_SECRET?**
All existing invite links will stop working. Store `TOKEN_SECRET` in a password manager or secrets vault alongside your other deployment credentials.

**Q: Can I customize templates?**
Yes — templates live in `templates/` and can be bind-mounted into the container. The theme picker also supports per-event color and image overrides through the UI.

---

## Security & Privacy

- Guests access events via unique 256-bit cryptographically secure tokens — no passwords, no accounts
- Tokens are HMAC-hashed in the database; the plain token is never stored after send
- Admin authentication is fully delegated to your existing auth infrastructure (no local passwords to compromise)
- All data stays on your server — nothing is sent to third parties
- TLS is handled at the reverse proxy layer; run TinyRSVP behind Traefik/Caddy/nginx in production

---

## Support & Contributing

- **Issues / Bug Reports**: [GitHub Issues](https://github.com/lenaxia/TinyRSVP/issues)
- **Discussions**: [GitHub Discussions](https://github.com/lenaxia/TinyRSVP/discussions)
- **Contributing**: See [`README-LLM.md`](README-LLM.md) for project structure and development workflow

---

## License

TinyRSVP is released under the [GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0).

If you modify and run this software as a network service, you must make the modified source available to users of that service under the same license.

---

*Built with Go. Designed for self-hosters.*
