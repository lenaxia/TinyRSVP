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
- 🔐 **Flexible Auth**: OIDC or forward auth for admins

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
PORT=8080
BASE_URL=https://rsvp.yourdomain.com

# Database
DATABASE_TYPE=sqlite  # or postgres
DATABASE_PATH=/data/tinyrsvp.db

# Authentication
AUTH_MODE=oidc  # or forward_auth
OIDC_ISSUER_URL=https://auth.yourdomain.com
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret

# SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@yourdomain.com

# Storage
STORAGE_TYPE=local  # or s3
STORAGE_PATH=/data/uploads
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  tinyrsvp:
    image: tinyrsvp:latest
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    environment:
      - DATABASE_PATH=/data/tinyrsvp.db
      - SMTP_HOST=smtp.gmail.com
      - SMTP_PORT=587
      - SMTP_USERNAME=${SMTP_USERNAME}
      - SMTP_PASSWORD=${SMTP_PASSWORD}
    restart: unless-stopped
```

---

## Features

### For Event Organizers

- **Create Events**: Set date, time, location, and RSVP deadline
- **Manage Invites**: Send personalized invitations via email
- **Track Responses**: View who's coming, who's not, and who hasn't responded
- **Preference Questions**: Ask custom questions (dietary restrictions, +1s, etc.)
- **Email Reminders**: Automatically remind non-responders
- **Export Data**: Download guest lists as CSV

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
│     │  SQLite/Postgres │               │
│     └─────────────────┘               │
└─────────────────────────────────────────┘
         │              │
         ▼              ▼
  ┌──────────┐   ┌──────────┐
  │   SMTP   │   │  Storage │
  │  Server  │   │ (FS/S3)  │
  └──────────┘   └──────────┘
```

---

## Technology Stack

- **Backend**: Go
- **Frontend**: Plain CSS + Vanilla JavaScript (mobile-first)
- **Templates**: Go `html/template`
- **Database**: SQLite (default) or PostgreSQL
- **Authentication**: OIDC or Forward Auth
- **Storage**: Local filesystem or S3-compatible

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
- **Admin Authentication**: OIDC or forward auth (no local passwords)
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
go build -o bin/tinyrsvp cmd/server/main.go

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

### v0 (Current)
- [x] Project setup and documentation
- [ ] Core event management
- [ ] Invite system with tokens
- [ ] RSVP handling
- [ ] Email sending with ICS attachments
- [ ] Basic templates
- [ ] SQLite support
- [ ] OIDC authentication

### v1 (Future)
- [ ] PostgreSQL support
- [ ] S3-compatible storage
- [ ] Guest OIDC (optional)
- [ ] Public event links
- [ ] Reminder scheduling UI
- [ ] Advanced templates
- [ ] Multi-language support

### v2 (Future)
- [ ] SMS notifications
- [ ] Calendar sync (CalDAV)
- [ ] Event analytics
- [ ] Custom branding
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
A: Currently in active development (v0). Use at your own risk.

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

[License TBD - To be determined]

---

## Acknowledgments

- Built with Go and love for self-hosting
- Inspired by Evite, Paperless Post, and the homelab community
- 100% LLM-implemented with human oversight

---

**Made with ❤️ for the self-hosting community**
