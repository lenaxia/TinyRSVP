Below is a **single, comprehensive, implementation-grade specification document** that captures **everything agreed so far**, with **no gaps**, **no assumptions**, and **no hand-waving**.

This is written to be **directly consumable by an LLM implementing the system end-to-end**, so it is explicit, verbose, and structured. Nothing here is aspirational unless clearly marked.

---

# Self-Hosted Evite-Style RSVP Platform

**Comprehensive Project Specification & Current Status**

---

## 1. Project Overview

### 1.1 Purpose

Build a **self-hosted, small-scale RSVP and invitation platform** similar in spirit to Evite, designed for **homelab environments**, family events, clubs, and private gatherings.

The system prioritizes:

* Simplicity
* Predictability
* Low operational overhead
* No forced guest accounts
* No SaaS dependency
* Dockerized deployment
* Optional integrations only

This is **not enterprise software** and **not intended for mass marketing use**.

---

### 1.2 Core Principles (Non-Negotiable)

1. **Guests must never be required to create accounts**
2. **Admins / event managers authenticate via OIDC or forward auth**
3. **All guest access is token-based and scoped**
4. **Self-hosted first**
5. **Feature completeness > feature breadth**
6. **Everything must work in a single-node homelab**

---

## 2. User Roles & Actors

### 2.1 Roles

#### Admin

* Full system control
* Can:

  * Manage users
  * Assign roles
  * Configure SMTP
  * View system health
  * Manage templates
  * Configure storage

#### Event Manager

* Can:

  * Create and manage events
  * Manage invites
  * Send emails
  * View RSVPs
* Cannot manage system-wide settings

#### Guest

* Never has a system account
* Interacts only via:

  * Unique invite links
  * Optional email verification
  * Optional passphrase
  * Optional OIDC (if enabled)

---

## 3. Authentication & Authorization

### 3.1 Admin / Event Manager Authentication

Supported modes (mutually exclusive, configured at startup):

#### Mode 1: Forward Auth

* App trusts headers set by reverse proxy
* Expected headers:

  * `X-Forwarded-User`
  * `X-Forwarded-Email`
* Reverse proxy enforces authentication (e.g. Traefik + Authelia / Authentik)

#### Mode 2: OIDC

* Uses a standard OIDC client library
* No custom auth implementation
* Maps:

  * `sub` → `oidc_subject`
  * `email` → user email
* First login may auto-create user (role must be assigned by admin)

#### Explicitly Excluded

* Local username/password auth (unless added later)
* MFA handling inside the app

---

### 3.2 Guest Authentication

Default behavior:

* **Token-based access**
* No identity persistence beyond the invite

Optional guest auth (off by default):

* Guests may authenticate via OIDC / forward auth
* Used only to associate RSVPs with an identity
* Does **not** create a persistent user account

---

## 4. Event Model

### 4.1 Event Attributes

Each event has:

* Title
* Description
* Start time
* End time (optional)
* Timezone (mandatory)
* Location (text or URL)
* RSVP deadline (optional)
* Public or private
* Max allowed +1s (default per invite)
* Assigned invite template
* Created by (user)

---

### 4.2 Event Visibility Modes

#### Private Event

* Only accessible via invite token
* No generic access

#### Public Event

* Accessible via a public/generic link
* Admin can require or optionally request email
* Generates provisional invites

---

## 5. Invites & Guest Access Model

### 5.1 Invite Creation

Invites may be created via:

* Explicit email invite
* Generic link (public or private)
* Manual creation (no email)

Each invite contains:

* Name (optional)
* Email (optional)
* Unique token (unguessable)
* Max +1s (inherits event default unless overridden)
* Optional passphrase (hashed)
* RSVP state

---

### 5.2 Invite Token Design

* 256-bit cryptographically secure random value
* URL-safe Base64 encoding
* Stored **hashed** in database (SHA-256)
* Used as the sole credential for guest access

Invite tokens:

* Can be regenerated
* Can be revoked
* Can optionally expire (future feature)

---

### 5.3 Guest RSVP Capabilities

Via a valid invite token, a guest may:

* View event details
* RSVP: yes / no / maybe
* Specify number of +1s (within limits)
* Answer preference questions
* Edit their RSVP at any time (until deadline)

---

## 6. Preference Questions

### 6.1 Question Types

Supported types:

* Text
* Select (options stored as JSON)
* Boolean

Each question can be:

* Required or optional
* Event-scoped

Answers are stored per invite.

---

## 7. Email System

### 7.1 Email Sending

* SMTP only
* User-provided SMTP configuration
* No third-party API dependencies

Each email supports:

* Plain text body
* HTML body
* Attachments

---

### 7.2 Invite Emails

Invite emails include:

* Custom subject
* Custom body
* Unique RSVP link
* ICS calendar attachment

---

### 7.3 Reminder Emails

Admins can configure reminders:

* Relative to event date
* Sent to:

  * Non-responders
  * All invitees

---

### 7.4 Email Queue

Emails are queued in the database to allow:

* Retries
* Deferred sending
* Observability

---

## 8. Calendar Attachments (ICS)

Each invite email includes an `.ics` file containing:

* Event title
* Start/end times
* Timezone
* Location
* Description including RSVP link

Compatible with:

* Google Calendar
* Apple Calendar
* Outlook

---

## 9. Templates & Customization

### 9.1 Template Types

* Web templates (RSVP pages)
* Email templates

Each template includes:

* HTML
* Optional CSS
* Safe variable interpolation

---

### 9.2 Customization

Admins can:

* Use built-in templates
* Upload images
* Create custom templates
* Assign templates to events

---

## 10. Asset Storage

### 10.1 Storage Providers

Storage is **pluggable**.

Supported providers:

#### Local Filesystem (Default)

* Stored in mounted volume
* Single-node friendly

#### S3-Compatible Storage (Optional)

* AWS S3
* MinIO
* Ceph
* Any S3-compatible endpoint

Used for:

* Uploaded images
* Template assets
* Email assets

---

### 10.2 Storage Abstraction

Storage provider interface:

* PutObject
* GetObject
* DeleteObject
* GetPublicURL

App must not care which backend is used.

---

## 11. Database Schema (Authoritative)

SQLite by default, Postgres compatible.

### Tables

* users
* events
* preference_questions
* invites
* rsvps
* rsvp_answers
* email_queue
* templates

(Exact SQL previously defined and considered final.)

---

## 12. API & Page Routes

### 12.1 Auth & Admin

* `/login`
* `/logout`
* `/admin/users`
* `/admin/settings`
* `/admin/smtp/test`
* `/admin/health`

---

### 12.2 Events

* `/events`
* `/events/new`
* `/events/{id}`
* `/events/{id}/edit`
* `/events/{id}/summary`
* `/events/{id}/responses`
* `/events/{id}/export.csv`

---

### 12.3 Preference Questions

* `/events/{id}/questions`
* `/events/{id}/questions/{qid}`
* `/events/{id}/questions/{qid}/delete`

---

### 12.4 Invites

* `/events/{id}/invites`
* `/events/{id}/invites` (POST)
* `/invites/{id}/edit`
* `/invites/{id}/delete`
* `/invites/{id}/revoke`
* `/invites/{id}/send`
* `/invites/{id}/resend`
* `/invites/{id}/regenerate-token`
* `/invites/{id}/email-preview`

---

### 12.5 Guest RSVP

* `/rsvp/{token}` (GET/POST)
* `/public/{event_id}` (GET/POST)

---

### 12.6 Templates

* `/templates`
* `/templates/new`
* `/templates/{id}`
* `/templates/{id}/delete`

---

### 12.7 Assets

* `/uploads`
* `/assets/{path}`
* `/assets` (list)
* `/assets/{path}` (delete)
* `/events/{id}/assets`

---

### 12.8 Email Jobs

* `/emails/queue`
* `/emails/{id}/retry`
* `/jobs/send-emails`

---

## 13. Deployment Model

### 13.1 Containerization

* Docker image
* Single binary inside container
* Config via environment variables
* Volumes for:

  * Database
  * Uploads
  * Templates

---

### 13.2 Default Deployment

* Single container
* SQLite
* Local filesystem storage
* External SMTP
* Reverse proxy handles TLS & auth

---

## 14. v0 Scope (Frozen)

### Included

* Admin / manager auth
* Event creation
* Email invites
* RSVP handling
* Preference questions
* +1 limits
* SMTP + ICS
* SQLite
* Local storage
* Basic templates

### Explicitly Excluded

* SMS
* S3 storage
* Guest OIDC
* Generic links without email
* Passphrases
* Public events UI
* Reminder scheduling UI

---

## 15. Success Criteria

v0 is considered successful if:

* A single family event can be fully managed
* Guests RSVP without accounts
* Emails reliably deliver with calendar attachments
* App runs stably in a homelab for weeks
* No manual DB intervention required

---

## 16. Current Status Summary

* ✅ Product scope defined
* ✅ Auth model finalized
* ✅ Guest model finalized
* ✅ Database schema finalized
* ✅ API surface finalized
* ✅ v0 boundaries locked
* ⏳ Implementation pending

---

## 17. Next Logical Implementation Steps

1. Generate SQL migrations
2. Implement config loader
3. Implement auth middleware
4. Implement event + invite CRUD
5. Implement RSVP flow
6. Implement email + ICS
7. Implement templates
8. Dockerize

---

**This document is the authoritative source of truth for implementation.**
