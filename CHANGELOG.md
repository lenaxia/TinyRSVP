# Changelog

All notable changes to TinyRSVP will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2026-03-05

Initial beta release. Feature-complete for v0 scope. Suitable for homelab / self-hosted use.

### Added

**Core event management**
- Create, edit, publish, cancel, and archive events
- Event deadlines with automatic enforcement
- Auto-archiving of past events (daily background job)
- Timezone-aware date/time handling

**Invite system**
- Cryptographically secure 256-bit invite tokens (HMAC-hashed in database)
- Individual invite creation and management
- Bulk CSV import
- Token revocation and regeneration
- Invite expiration and cleanup

**RSVP handling**
- Yes / No / Maybe responses
- Plus-ones with configurable limits
- Custom preference questions (text, boolean)
- RSVP updates after initial submission
- RSVP deadline enforcement

**Email**
- SMTP invite and confirmation emails
- ICS calendar attachment generation
- Queued email delivery with retry and rate limiting
- Configurable SMTP (supports Gmail, Mailgun, self-hosted, etc.)

**Template system**
- 7 built-in RSVP page themes (Classic, Minimal, Evite-style, and more)
- Custom header images per event
- Color overrides per event
- Go `html/template` based, XSS-safe

**Authentication**
- Forward auth support (Traefik + Authelia, Caddy, Authentik proxy, etc.)
- OIDC support via `go-oidc` (implemented; not yet integration-tested against a real provider in beta)
- Session management with secure cookies
- Role-based access control (admin / event manager)

**Infrastructure**
- SQLite database with automatic migrations
- Local filesystem storage for uploaded images
- Docker / Docker Compose single-container deployment
- Prometheus metrics endpoint
- Health and readiness endpoints
- Graceful shutdown

**UI**
- Mobile-responsive design (mobile-first CSS)
- Dashboard with event overview
- Event management UI
- Invite list and management UI
- Admin panel (user management, system stats)
- RSVP summary view per event

### Known limitations

- **Storage**: Local filesystem only. S3-compatible storage is planned for v1.
- **Database**: SQLite only. PostgreSQL is planned for v1.
- **OIDC**: Implemented but not integration-tested against a real provider. Forward auth is the tested auth path for this release.
- **Auth requirement**: An external auth provider (forward auth proxy or OIDC provider) is required. There are no local username/password accounts by design.
- **Scale**: Designed for small to medium events (up to a few hundred guests per event).

### Security notes

- `TOKEN_SECRET` is required at startup — there is no insecure default fallback.
- All invite tokens are HMAC-hashed in the database; raw tokens are never stored.
- HTTPS is required (TLS termination at reverse proxy).

---

[0.1.0]: https://github.com/lenaxia/tinyrsvp/releases/tag/v0.1.0
