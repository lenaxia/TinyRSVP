# TinyRSVP

> **Note for LLMs:** If you are an AI assistant working on this project, you MUST read [`README-LLM.md`](README-LLM.md) instead before moving on to any other task.. This file is for human users.

A self-hosted, privacy-focused RSVP and invitation platform for family events, clubs, and private gatherings.

---

## Overview

TinyRSVP is a lightweight alternative to services like Evite, designed specifically for self-hosting in homelab environments. It prioritizes simplicity, privacy, and ease of deployment.

**Key Features:**
- 🔒 **Privacy-First**: Self-hosted, no data sharing with third parties
- 👥 **No Guest Accounts Required**: Token-based access for guests
- 📧 **Email Invitations**: Send invites with calendar attachments
- 📱 **Mobile-Friendly**: Responsive design for all devices
- 🐳 **Docker-Ready**: Single container deployment
- 🔐 **Flexible Auth**: Forward auth (tested) or OIDC (implemented, not yet integration-tested) for admins

---

## Quick Start

### Prerequisites

- Docker and Docker Compose
- SMTP server for sending emails
- (Optional) OIDC provider (Authentik, Keycloak, etc.)

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/lenaxia/tinyrsvp.git
   cd tinyrsvp
   ```

2. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

3. **Start the application:**
   ```bash
   docker-compose up -d
   ```

4. **Access the application:**
   ```
   http://localhost:8080
   ```

---

## Configuration

### Environment Variables

```bash
# Server
SERVER_PORT=8080
SERVER_BASE_URL=https://rsvp.yourdomain.com   # Required — used in emails and ICS attachments

# Database (SQLite only in v0)
DATABASE_PATH=/data/tinyrsvp.db

# Token Security (REQUIRED)
TOKEN_SECRET=<generate with: openssl rand -hex 32>
TOKEN_HASHING_ENABLED=true   # default: true

# Authentication — choose one
FORWARD_AUTH_ENABLED=true
FORWARD_AUTH_USER_HEADER=X-Forwarded-User
FORWARD_AUTH_EMAIL_HEADER=X-Forwarded-Email
FORWARD_AUTH_TRUSTED_IPS=127.0.0.1,172.17.0.1

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
EMAIL_FROM=noreply@yourdomain.com

# Storage (local filesystem only in v0)
STORAGE_TYPE=local
STORAGE_LOCAL_PATH=/data/uploads
```

See [`.env.example`](.env.example) for the full annotated list of all environment variables.

### Token Security

**`TOKEN_SECRET` is required.** The app will refuse to start without it.

**Generate a secure secret:**
```bash
openssl rand -hex 32
```

| Mode | Configuration | Behavior |
|------|---------------|----------|
| **Production** | `TOKEN_SECRET=<secret>`<br>`TOKEN_HASHING_ENABLED=true` | Tokens HMAC-hashed with your secret |
| **Plain Token** | `TOKEN_HASHING_ENABLED=false` | Tokens stored in plain text (simpler but less secure) |

### Docker Compose Example

```yaml
services:
  tinyrsvp:
    image: tinyrsvp:latest
    ports:
      - "8080:8080"
    volumes:
      - tinyrsvp-data:/data
    environment:
      - SERVER_BASE_URL=https://rsvp.yourdomain.com   # Set this — used in all email links
      - TOKEN_SECRET=${TOKEN_SECRET}                   # Required: openssl rand -hex 32
      - DATABASE_PATH=/data/tinyrsvp.db
      - FORWARD_AUTH_ENABLED=true
      - FORWARD_AUTH_USER_HEADER=X-Forwarded-User
      - FORWARD_AUTH_EMAIL_HEADER=X-Forwarded-Email
      - FORWARD_AUTH_TRUSTED_IPS=172.17.0.1,172.18.0.1
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

- **Create Events**: Set date, time, location, and RSVP deadline
- **Manage Invites**: Send personalized invitations via email
- **Track Responses**: View who's coming, who's not, and who hasn't responded
- **Preference Questions**: Ask custom questions (dietary restrictions, +1s, etc.)
- **Theme Picker**: Choose from 7 built-in themes or customize colors and images

### For Guests

- **Easy RSVP**: Click link in email, no account needed
- **Calendar Integration**: Add event to calendar with .ics attachment
- **Update Anytime**: Change RSVP until deadline
- **Bring Guests**: Specify number of +1s
- **Answer Questions**: Provide dietary preferences, song requests, etc.

---

## Architecture

```
┌─────────────────────────────────────────┐
│     Reverse Proxy (Traefik/Nginx)      │
│         - TLS Termination               │
│         - Forward Auth (Optional)       │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│          TinyRSVP Application           │
│                                         │
│  ┌──────────┐  ┌──────────┐           │
│  │  Events  │  │ Invites  │           │
│  │  Manager │  │  System  │           │
│  └──────────┘  └──────────┘           │
│         │            │                 │
│         └────────────┘                 │
│              │                         │
│              ▼                         │
│     ┌─────────────────┐               │
│     │     SQLite      │               │
│     └─────────────────┘               │
└─────────────────────────────────────────┘
         │              │
         ▼              ▼
  ┌──────────┐   ┌──────────┐
   │   SMTP   │   │  Storage │
   │  Server  │   │  (local) │
  └──────────┘   └──────────┘
```

---

## Technology Stack

- **Backend**: Go
- **Frontend**: Plain CSS + Vanilla JavaScript (mobile-first)
- **Templates**: Go `html/template`
- **Database**: SQLite (PostgreSQL planned for v1)
- **Authentication**: Forward Auth (tested) or OIDC (implemented, not integration-tested in beta)
- **Storage**: Local filesystem (S3-compatible planned for v1)

---

## Use Cases

### Family Events
- Birthday parties
- Holiday gatherings
- Family reunions
- Graduation celebrations

### Clubs & Organizations
- Book club meetings
- Sports team events
- Community gatherings
- Volunteer activities

### Private Events
- Wedding receptions
- Baby showers
- Dinner parties
- Game nights

---

## Security & Privacy

- **No Guest Accounts**: Guests access via unique, unguessable tokens
- **Token Security**: 256-bit cryptographically secure tokens, hashed in database
- **Admin Authentication**: Forward auth (tested) or OIDC (implemented, not integration-tested — no local passwords)
- **Self-Hosted**: Your data stays on your server
- **Optional Passphrases**: Add extra security to invites
- **HTTPS Required**: TLS termination at reverse proxy

---

## Deployment Options

### Homelab (Recommended)

Perfect for:
- Raspberry Pi
- NAS devices (Synology, QNAP)
- Home servers
- Docker hosts

Requirements:
- 512MB RAM minimum
- 1GB disk space
- Docker support

### Cloud Hosting

Works with:
- DigitalOcean Droplets
- AWS EC2
- Google Cloud Compute
- Linode
- Any VPS provider

---

## Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/lenaxia/tinyrsvp.git
cd tinyrsvp

# Install dependencies
go mod download

# Run tests
go test -timeout 30s ./...

# Build binary
go build -o bin/tinyrsvp ./cmd/server

# Run locally
./bin/tinyrsvp
```

### Project Structure

```
TinyRSVP/
├── cmd/                # Application entrypoints
├── internal/           # Private application code
├── pkg/                # Public packages
├── templates/          # HTML templates
├── static/             # CSS, JS, images
├── migrations/         # Database migrations
├── docs/               # Documentation
└── tests/              # Integration tests
```

### Contributing

This project is 100% LLM-implemented with human oversight. If you're an LLM working on this project, please read [`README-LLM.md`](README-LLM.md) for detailed guidelines.

For humans interested in contributing:
1. Read [`docs/00_INITIAL_HLD.md`](docs/00_INITIAL_HLD.md) for the complete specification
2. Check [`docs/00_BACKLOG/`](docs/00_BACKLOG/) for current work items
3. Follow the development workflow in [`README-LLM.md`](README-LLM.md)

---

## Documentation

- **[README-LLM.md](README-LLM.md)** - LLM implementation guide (for AI assistants)
- **[docs/00_INITIAL_HLD.md](docs/00_INITIAL_HLD.md)** - High-level design specification
- **[docs/00_BACKLOG/](docs/00_BACKLOG/)** - Sprint stories and epics
- **[docs/01_WORKLOG/](docs/01_WORKLOG/)** - Progress updates and handoffs
- **[llm-workflows/](llm-workflows/)** - LLM workflow templates

---

## Roadmap

### v0 (Current — feature complete, beta)
- [x] Project setup and infrastructure
- [x] Core event management (create, edit, publish, cancel, archive)
- [x] Invite system with cryptographic tokens (individual, bulk CSV)
- [x] RSVP handling (yes/no/maybe, plus ones, preference questions)
- [x] Email sending with ICS calendar attachments
- [x] Template system with theme picker (7 themes, custom images, color overrides)
- [x] SQLite database
- [x] OIDC and forward-auth admin authentication
- [x] Mobile-responsive UI
- [x] Docker / Docker Compose deployment

### v1 (Planned)
- [ ] PostgreSQL support
- [ ] S3-compatible storage
- [ ] Reminder scheduling UI
- [ ] Multi-language support
- [ ] Advanced template editor

### v2 (Future)
- [ ] Guest OIDC (optional passwordless accounts)
- [ ] SMS notifications
- [ ] CalDAV calendar sync
- [ ] Event analytics dashboard
- [ ] API for integrations

---

## FAQ

**Q: Do guests need to create accounts?**  
A: No, guests access events via unique links sent in email invitations.

**Q: Can I use this for large events?**  
A: TinyRSVP is designed for small to medium events (up to a few hundred guests). For larger events, consider a dedicated service.

**Q: What email providers are supported?**  
A: Any SMTP server works (Gmail, SendGrid, Mailgun, self-hosted, etc.).

**Q: Can I customize the look and feel?**  
A: Yes, templates and CSS can be customized. See documentation for details.

**Q: Is this production-ready?**  
A: v0 is feature-complete and suitable for homelab / private use behind a reverse proxy. A full security audit (Epic 09) is planned before recommending public internet exposure.

**Q: How do I backup my data?**  
A: Backup the database file and uploads directory. For SQLite, copy `/data/tinyrsvp.db`.

**Q: Can I migrate from Evite/Paperless Post?**  
A: Not currently, but manual import is possible by creating events and invites.

---

## Support

- **Issues**: [GitHub Issues](https://github.com/lenaxia/tinyrsvp/issues)
- **Documentation**: [docs/](docs/)
- **Discussions**: [GitHub Discussions](https://github.com/lenaxia/tinyrsvp/discussions)

---

## License

TinyRSVP is released under the [GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0).

If you modify and run this software as a network service, you must make the modified source available to users of that service under the same license.

---

## Acknowledgments

- Built with Go and love for self-hosting
- Inspired by Evite, Paperless Post, and the homelab community
- 100% LLM-implemented with human oversight

---

**Made with ❤️ for the self-hosting community**
