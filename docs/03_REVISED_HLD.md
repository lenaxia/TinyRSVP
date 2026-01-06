# TinyRSVP - High-Level Design (Revised)

**Version:** 2.0  
**Date:** 2026-01-06  
**Status:** Authoritative Specification  
**Supersedes:** [`docs/00_INITIAL_HLD.md`](00_INITIAL_HLD.md)

---

## Document Control

**Change History:**

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-01-06 | Initial HLD | Human |
| 2.0 | 2026-01-06 | Comprehensive revision addressing design review findings | AI Assistant |

**Review Status:**
- Design Review: [`docs/02_HLD_DESIGN_REVIEW.md`](02_HLD_DESIGN_REVIEW.md)
- Critical Issues Addressed: 25/25
- High Priority Issues Addressed: 25/25

---

## Table of Contents

1. [Project Overview](#1-project-overview)
2. [Core Principles](#2-core-principles)
3. [User Roles & Permissions](#3-user-roles--permissions)
4. [Authentication & Authorization](#4-authentication--authorization)
5. [Event Model](#5-event-model)
6. [Invite & Guest Access Model](#6-invite--guest-access-model)
7. [RSVP Model](#7-rsvp-model)
8. [Preference Questions](#8-preference-questions)
9. [Email System](#9-email-system)
10. [Calendar Integration (ICS)](#10-calendar-integration-ics)
11. [Templates & Customization](#11-templates--customization)
12. [Asset Storage](#12-asset-storage)
13. [Database Schema](#13-database-schema)
14. [Validation Rules](#14-validation-rules)
15. [Error Handling](#15-error-handling)
16. [Security](#16-security)
17. [Operations](#17-operations)
18. [API Routes](#18-api-routes)
19. [Request Flow](#19-request-flow)
20. [Deployment Model](#20-deployment-model)
21. [v0 Scope](#21-v0-scope)
22. [Success Criteria](#22-success-criteria)

---

## 1. Project Overview

### 1.1 Purpose

TinyRSVP is a **self-hosted, small-scale RSVP and invitation platform** designed for homelab environments, family events, clubs, and private gatherings. It provides a privacy-focused alternative to commercial services like Evite.

**Target Users:**
- Families hosting events (birthdays, holidays, reunions)
- Clubs and organizations (book clubs, sports teams)
- Private gatherings (dinner parties, game nights)
- Homelab enthusiasts wanting self-hosted solutions

**Not Intended For:**
- Enterprise event management
- Mass marketing campaigns
- Public ticketing systems
- Multi-tenant SaaS offerings

### 1.2 Key Features

**For Event Organizers:**
- Create and manage events with full details
- Send personalized email invitations with calendar attachments
- Track RSVPs in real-time
- Collect guest preferences (dietary restrictions, +1s, etc.)
- Export guest lists
- Customize invitation templates

**For Guests:**
- RSVP via unique link (no account required)
- Add event to calendar automatically
- Update RSVP anytime before deadline
- Specify number of +1s
- Answer preference questions

---

## 2. Core Principles

These principles are **non-negotiable** and guide all design decisions:

1. **Guests Never Need Accounts** - Token-based access only
2. **Self-Hosted First** - Designed for single-node homelab deployment
3. **Privacy-Focused** - No data sharing, no tracking, no third parties
4. **Docker-First** - Primary deployment via single container
5. **Feature Completeness > Feature Breadth** - Do fewer things well
6. **Simplicity** - Prefer simple solutions over complex ones
7. **Predictability** - Behavior should be obvious and consistent
8. **Low Operational Overhead** - Minimal maintenance required

---

## 3. User Roles & Permissions

### 3.1 Role Definitions

#### Admin

**Purpose:** Full system control and configuration

**Permissions:**
- ✅ All Event Manager permissions
- ✅ Manage users (create, edit, delete, assign roles)
- ✅ Configure system settings (SMTP, storage, etc.)
- ✅ View system health and logs
- ✅ Manage global templates
- ✅ Permanently delete events (including archived)
- ✅ Access all administrative endpoints

**Restrictions:**
- None

#### Event Manager

**Purpose:** Create and manage events

**Permissions:**
- ✅ Create events
- ✅ Edit own events
- ✅ Delete own events (moves to archived state)
- ✅ View all events (including archived)
- ✅ Manage invites for own events
- ✅ Send emails for own events
- ✅ View RSVPs for own events
- ✅ Export guest lists for own events
- ✅ Create custom templates for own events

**Restrictions:**
- ❌ Cannot edit other users' events
- ❌ Cannot delete other users' events
- ❌ Cannot access system settings
- ❌ Cannot manage users
- ❌ Cannot permanently delete events
- ❌ Cannot view system logs

#### Guest

**Purpose:** RSVP to events

**Permissions:**
- ✅ View event details via invite token
- ✅ Submit RSVP (yes/no/maybe)
- ✅ Specify number of +1s (within limits)
- ✅ Answer preference questions
- ✅ Update RSVP until deadline

**Restrictions:**
- ❌ No system account
- ❌ Cannot create events
- ❌ Cannot see other guests' RSVPs
- ❌ Cannot access admin functions
- ❌ Access limited to specific event via token

### 3.2 Permission Matrix

| Action | Admin | Event Manager | Guest |
|--------|-------|---------------|-------|
| Create event | ✅ | ✅ | ❌ |
| Edit own event | ✅ | ✅ | ❌ |
| Edit other's event | ✅ | ❌ | ❌ |
| Delete own event | ✅ | ✅ (archive) | ❌ |
| Delete other's event | ✅ | ❌ | ❌ |
| Permanently delete event | ✅ | ❌ | ❌ |
| View all events | ✅ | ✅ | ❌ |
| View archived events | ✅ | ✅ | ❌ |
| Create invites | ✅ | ✅ (own events) | ❌ |
| Send emails | ✅ | ✅ (own events) | ❌ |
| View RSVPs | ✅ | ✅ (own events) | ❌ |
| Submit RSVP | ✅ | ✅ | ✅ (with token) |
| Manage users | ✅ | ❌ | ❌ |
| Configure SMTP | ✅ | ❌ | ❌ |
| View system health | ✅ | ❌ | ❌ |
| Access logs | ✅ | ❌ | ❌ |

---

## 4. Authentication & Authorization

### 4.1 Admin/Manager Authentication

**Supported Modes:** (mutually exclusive, configured at startup)

#### Mode 1: OIDC (OpenID Connect)

**Configuration:**
```
OIDC_ENABLED=true
OIDC_ISSUER_URL=https://auth.example.com
OIDC_CLIENT_ID=tinyrsvp
OIDC_CLIENT_SECRET=<secret>
OIDC_REDIRECT_URL=https://rsvp.example.com/auth/callback
```

**Flow:**
1. User clicks "Login"
2. App redirects to OIDC provider authorization URL
3. User authenticates at provider
4. Provider redirects to `/auth/callback` with authorization code
5. App exchanges code for ID token and access token
6. App validates ID token signature and claims
7. App extracts `sub` (subject) and `email` claims
8. App creates/updates user in database
9. App creates session and sets secure cookie
10. App redirects to dashboard

**Bootstrap Admin Creation:**
- First user to successfully authenticate becomes Admin automatically
- Subsequent users are created with Event Manager role by default
- Admins can promote Event Managers to Admin

**Required Claims:**
- `sub` (subject) - unique user identifier
- `email` - user email address

**Optional Claims:**
- `name` - display name
- `picture` - profile picture URL

**Error Handling:**
- Missing `email` claim → reject authentication, show error
- Invalid ID token → reject authentication, show error
- OIDC provider unreachable → show error, retry button
- Token validation failure → reject authentication, log details

#### Mode 2: Forward Auth

**Configuration:**
```
FORWARD_AUTH_ENABLED=true
FORWARD_AUTH_USER_HEADER=X-Forwarded-User
FORWARD_AUTH_EMAIL_HEADER=X-Forwarded-Email
```

**Flow:**
1. Reverse proxy (Traefik/Nginx) enforces authentication
2. Proxy sets headers: `X-Forwarded-User`, `X-Forwarded-Email`
3. App validates headers are present and non-empty
4. App creates/updates user in database
5. App creates session and sets secure cookie
6. App serves request

**Bootstrap Admin Creation:**
- First user to access system becomes Admin automatically
- Subsequent users are created with Event Manager role by default

**Header Validation:**
- Headers must be present
- Headers must be non-empty strings
- Email must match basic email format (contains @)
- Headers are trusted (proxy must prevent spoofing)

**Security Requirements:**
- App must only be accessible via trusted reverse proxy
- Proxy must strip/override forwarded headers from clients
- App should validate requests come from known proxy IP (optional)

**Error Handling:**
- Missing headers → HTTP 401, show error page
- Empty headers → HTTP 401, show error page
- Invalid email format → HTTP 400, show error page

### 4.2 Session Management

**Storage:** Database-backed sessions

**Session Table:**
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_accessed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    INDEX idx_user_id (user_id),
    INDEX idx_expires_at (expires_at)
);
```

**Session Cookie:**
- Name: `tinyrsvp_session`
- Value: Session ID (cryptographically random, 32 bytes, base64)
- HttpOnly: `true` (prevent JavaScript access)
- Secure: `true` (HTTPS only)
- SameSite: `Lax` (CSRF protection)
- Max-Age: 7 days (604800 seconds)

**Session Lifecycle:**
1. **Creation:** On successful authentication
2. **Validation:** On every request, check session exists and not expired
3. **Refresh:** Update `last_accessed_at` on each request
4. **Expiration:** 7 days from creation (not sliding)
5. **Cleanup:** Background job deletes expired sessions every hour

**Session Invalidation:**
- Explicit logout → delete session immediately
- User deletion → cascade delete all user sessions
- Role change → no automatic invalidation (takes effect on next request)

### 4.3 Guest Authentication

**Method:** Token-based access (no accounts)

**Token Properties:**
- 256-bit cryptographically secure random value
- URL-safe Base64 encoding (43 characters)
- Stored as HMAC-SHA256 hash in database
- Scoped to single invite/event

**Token Generation:**
```
1. Generate 32 random bytes using crypto/rand
2. Base64-URL encode → token string
3. Compute HMAC-SHA256(secret_key, token) → hash
4. Store hash in database
5. Return token to caller (never stored in plain text)
```

**Token Validation:**
```
1. Receive token from URL parameter
2. Compute HMAC-SHA256(secret_key, token) → hash
3. Constant-time compare hash with database value
4. Check invite not revoked
5. Check token not expired
6. Grant access to event
```

**Token Expiration:**
- Tokens expire 30 days after event date
- Expired tokens return "Event has ended" message
- Cleanup job deletes expired tokens weekly

**Token Revocation:**
- Admin/Manager can revoke invite
- Revoked token returns "Invite has been cancelled" message
- Guest cannot request new token (must contact organizer)

**Token Regeneration:**
- Admin/Manager can regenerate token for invite
- Old token immediately invalidated
- New token sent via email
- Use case: Token leaked or shared inappropriately

**Security:**
- HMAC secret key generated on first startup, stored in database
- Constant-time comparison prevents timing attacks
- Tokens never logged or displayed in admin UI (only last 6 chars shown)

---

## 5. Event Model

### 5.1 Event Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| id | INTEGER | Auto | Primary key |
| title | TEXT | Yes | Event name (max 200 chars) |
| description | TEXT | No | Event details (max 5000 chars, Markdown) |
| location | TEXT | No | Address or URL (max 500 chars) |
| start_time | TIMESTAMP | Yes | Event start (ISO 8601 with timezone) |
| end_time | TIMESTAMP | No | Event end (ISO 8601 with timezone) |
| timezone | TEXT | Yes | IANA timezone (e.g., "America/Los_Angeles") |
| rsvp_deadline | TIMESTAMP | No | Last date to RSVP (ISO 8601 with timezone) |
| max_plus_ones | INTEGER | Yes | Default max +1s per invite (0-10, default 0) |
| created_by | INTEGER | Yes | User ID of creator |
| status | TEXT | Yes | Event status (see lifecycle below) |
| template_id | INTEGER | No | Assigned template for invites |
| created_at | TIMESTAMP | Auto | Creation timestamp |
| updated_at | TIMESTAMP | Auto | Last update timestamp |
| version | INTEGER | Auto | Optimistic locking version |

### 5.2 Event Lifecycle States

```
┌─────────┐
│  DRAFT  │ Initial state, not visible to guests
└────┬────┘
     │ publish
     ▼
┌───────────┐
│ PUBLISHED │ Active, invites can be sent, guests can RSVP
└────┬──────┘
     │ cancel
     ▼
┌───────────┐
│ CANCELLED │ Event cancelled, guests notified, no more RSVPs
└────┬──────┘
     │ (automatic 30 days after event_date)
     ▼
┌──────────┐
│ ARCHIVED │ Read-only, visible to managers, can be permanently deleted by admin
└──────────┘
```

**State Transitions:**

| From | To | Trigger | Effect |
|------|----| --------|--------|
| DRAFT | PUBLISHED | Manual publish | Invites can be sent, event visible |
| DRAFT | CANCELLED | Manual cancel | Event cancelled before publishing |
| PUBLISHED | CANCELLED | Manual cancel | Guests notified, RSVPs locked |
| PUBLISHED | ARCHIVED | Auto (30 days after event) | Read-only, cleanup eligible |
| CANCELLED | ARCHIVED | Auto (30 days after event) | Read-only, cleanup eligible |

**State Rules:**
- DRAFT: Can edit all fields, can delete
- PUBLISHED: Can edit most fields, cannot change start_time if <24hrs away
- CANCELLED: Cannot edit, cannot un-cancel
- ARCHIVED: Read-only, Event Managers can view, only Admin can permanently delete

### 5.3 Event Validation Rules

**Title:**
- Required
- Min length: 3 characters
- Max length: 200 characters
- No leading/trailing whitespace

**Description:**
- Optional
- Max length: 5000 characters
- Supports Markdown (sanitized on render)

**Location:**
- Optional
- Max length: 500 characters
- Can be address or URL

**Start Time:**
- Required
- Must be valid ISO 8601 timestamp
- Must include timezone
- Cannot be in the past (at creation)
- Must be before end_time (if end_time provided)

**End Time:**
- Optional
- Must be valid ISO 8601 timestamp
- Must be after start_time
- Must be same day or within 7 days of start_time

**Timezone:**
- Required
- Must be valid IANA timezone name
- Validated against IANA timezone database
- Examples: "America/Los_Angeles", "Europe/London", "Asia/Tokyo"

**RSVP Deadline:**
- Optional
- Must be before start_time
- Must be in the future (at creation)
- Recommended: At least 24 hours before start_time

**Max Plus Ones:**
- Required
- Integer between 0 and 10
- Default: 0 (no +1s allowed)

### 5.4 Event Capacity

**v0 Scope:** No total event capacity limit

**Rationale:** Small events (family/friends) rarely need capacity limits. Capacity per invite (+1 limits) is sufficient.

**Future (v1+):** May add optional total capacity with waitlist support.

---

## 6. Invite & Guest Access Model

### 6.1 Invite Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| id | INTEGER | Auto | Primary key |
| event_id | INTEGER | Yes | Foreign key to events |
| name | TEXT | No | Guest name (max 100 chars) |
| email | TEXT | No | Guest email (max 255 chars) |
| token_hash | TEXT | Yes | HMAC-SHA256 hash of token |
| max_plus_ones | INTEGER | Yes | Override event default (0-10) |
| status | TEXT | Yes | Invite status (see below) |
| sent_at | TIMESTAMP | No | When invite email was sent |
| created_at | TIMESTAMP | Auto | Creation timestamp |
| updated_at | TIMESTAMP | Auto | Last update timestamp |
| expires_at | TIMESTAMP | Auto | Token expiration (event_date + 30 days) |

### 6.2 Invite Status

| Status | Description |
|--------|-------------|
| DRAFT | Created but not sent |
| SENT | Email sent to guest |
| VIEWED | Guest viewed RSVP page |
| RESPONDED | Guest submitted RSVP |
| REVOKED | Invite cancelled by organizer |

**Status Transitions:**
- DRAFT → SENT (when email sent)
- SENT → VIEWED (when guest opens RSVP link)
- VIEWED → RESPONDED (when guest submits RSVP)
- Any → REVOKED (manual revocation)

### 6.3 Invite Creation Methods

#### Method 1: Individual Email Invite

**Use Case:** Sending personalized invites to known guests

**Process:**
1. Admin/Manager enters guest name and email
2. System generates unique token
3. System creates invite record
4. System queues email with RSVP link
5. Status: DRAFT → SENT (when email sent)

#### Method 2: Bulk CSV Import

**Use Case:** Importing large guest lists

**CSV Format:**
```csv
name,email,max_plus_ones
John Doe,john@example.com,2
Jane Smith,jane@example.com,1
```

**Process:**
1. Admin/Manager uploads CSV file
2. System validates CSV format and data
3. System creates invite records in batch
4. System queues emails in batch
5. System shows summary (created, failed, duplicates)

**Validation:**
- CSV must have header row
- Email column required, name and max_plus_ones optional
- Max 500 rows per upload
- Duplicate emails within same event rejected
- Invalid emails rejected with error report

#### Method 3: Manual Creation (No Email)

**Use Case:** In-person distribution, phone invites

**Process:**
1. Admin/Manager creates invite without email
2. System generates token and RSVP URL
3. System displays URL for manual distribution
4. Status remains DRAFT until guest accesses link

### 6.4 Token Security

**Secret Key Management:**
- Generated on first startup using crypto/rand (32 bytes)
- Stored in database `config` table
- Never logged or exposed via API
- Rotated manually via admin command (invalidates all tokens)

**HMAC-SHA256 Hashing:**
```go
func HashToken(token string, secret []byte) string {
    h := hmac.New(sha256.New, secret)
    h.Write([]byte(token))
    return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func ValidateToken(token string, hash string, secret []byte) bool {
    expected := HashToken(token, secret)
    return hmac.Equal([]byte(expected), []byte(hash))
}
```

**Why HMAC-SHA256 vs Bcrypt:**
- Tokens are random (not user-chosen passwords)
- HMAC provides authentication + integrity
- Constant-time comparison prevents timing attacks
- Faster than bcrypt (important for guest experience)
- Appropriate for high-entropy random tokens

### 6.5 Guest RSVP URL Format

```
https://rsvp.example.com/rsvp/{token}
```

**Example:**
```
https://rsvp.example.com/rsvp/a3F8kL9mN2pQ5rT7vW0xY4zA6bC8dE1fG3hJ5k
```

**Token in URL:**
- 43 characters (256 bits base64-URL encoded)
- URL-safe (no special characters needing encoding)
- Unguessable (2^256 possible values)

---

## 7. RSVP Model

### 7.1 RSVP Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| id | INTEGER | Auto | Primary key |
| invite_id | INTEGER | Yes | Foreign key to invites |
| response | TEXT | Yes | yes/no/maybe |
| plus_ones | INTEGER | Yes | Number of +1s (0 to invite.max_plus_ones) |
| created_at | TIMESTAMP | Auto | First RSVP timestamp |
| updated_at | TIMESTAMP | Auto | Last update timestamp |

**Constraints:**
- One RSVP per invite (1:1 relationship)
- plus_ones must be <= invite.max_plus_ones
- Cannot create RSVP if invite.status = REVOKED
- Cannot update RSVP after event.rsvp_deadline (if set)

### 7.2 RSVP Response Values

| Value | Meaning | Counted in Attendance |
|-------|---------|----------------------|
| yes | Attending | Yes |
| no | Not attending | No |
| maybe | Unsure | No (but tracked separately) |

### 7.3 RSVP State Transitions

**Allowed Transitions:**
- None → yes/no/maybe (initial RSVP)
- yes ↔ no ↔ maybe (unlimited changes before deadline)

**Restrictions:**
- Cannot change after RSVP deadline
- Cannot RSVP if invite revoked
- Cannot RSVP if event cancelled

**Deadline Enforcement:**
- Deadline checked on RSVP submission
- If past deadline: Show error "RSVP deadline has passed"
- Event details still visible, just cannot submit/change RSVP
- No grace period in v0 (strict enforcement)

### 7.4 Plus Ones Validation

**Rules:**
- plus_ones must be integer >= 0
- plus_ones must be <= invite.max_plus_ones
- plus_ones must be <= event.max_plus_ones (if invite override not set)
- If response = "no", plus_ones should be 0 (enforced on save)

**Error Messages:**
- "You can bring up to {max} guest(s)" (if exceeded)
- "Plus ones not allowed for this event" (if max = 0)

### 7.5 RSVP Confirmation

**On Successful RSVP:**
1. Save RSVP to database
2. Update invite.status to RESPONDED
3. Show confirmation page with:
   - Event details
   - RSVP summary (response + plus_ones)
   - "Add to Calendar" button (downloads ICS)
   - "Update RSVP" link (same token)
4. Send confirmation email (optional, configurable)

**Confirmation Email:**
- Subject: "RSVP Confirmed: {Event Title}"
- Body: Event details, RSVP summary, update link
- Attachment: ICS calendar file

---

## 8. Preference Questions

### 8.1 Question Types

| Type | Description | Validation |
|------|-------------|------------|
| text | Free-form text input | Max 500 chars |
| select | Single choice from options | Must select one option |
| boolean | Yes/No question | true/false |

### 8.2 Question Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| id | INTEGER | Auto | Primary key |
| event_id | INTEGER | Yes | Foreign key to events |
| question_text | TEXT | Yes | Question prompt (max 500 chars) |
| question_type | TEXT | Yes | text/select/boolean |
| options | JSON | Conditional | Required for select type |
| required | BOOLEAN | Yes | Must be answered |
| display_order | INTEGER | Yes | Sort order (0-based) |
| created_at | TIMESTAMP | Auto | Creation timestamp |

**Options JSON Format (for select type):**
```json
[
  {"value": "vegetarian", "label": "Vegetarian"},
  {"value": "vegan", "label": "Vegan"},
  {"value": "gluten_free", "label": "Gluten-Free"},
  {"value": "none", "label": "No restrictions"}
]
```

### 8.3 Question Validation

**Question Text:**
- Required
- Min length: 5 characters
- Max length: 500 characters

**Question Type:**
- Must be one of: text, select, boolean

**Options:**
- Required if type = select
- Must be valid JSON array
- Min 2 options, max 20 options
- Each option must have "value" and "label"
- Values must be unique within question

**Display Order:**
- Integer >= 0
- Auto-assigned if not provided (max + 1)
- Can be reordered by admin

### 8.4 Answer Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| id | INTEGER | Auto | Primary key |
| rsvp_id | INTEGER | Yes | Foreign key to rsvps |
| question_id | INTEGER | Yes | Foreign key to preference_questions |
| answer_text | TEXT | Conditional | For text type |
| answer_option | TEXT | Conditional | For select type (value) |
| answer_boolean | BOOLEAN | Conditional | For boolean type |
| created_at | TIMESTAMP | Auto | Creation timestamp |
| updated_at | TIMESTAMP | Auto | Last update timestamp |

**Constraints:**
- One answer per RSVP per question
- Answer type must match question type
- Required questions must have answer

### 8.5 Question Lifecycle

**Creation:**
- Can add questions anytime (even after invites sent)
- New questions apply to future RSVPs
- Existing RSVPs not required to answer new questions

**Editing:**
- Can edit question_text anytime
- Cannot change question_type (would invalidate existing answers)
- Can add/remove options for select type (existing answers preserved if valid)

**Deletion:**
- Can delete question anytime
- Cascade deletes all answers
- Warning shown if answers exist

**Answer Editing:**
- Guests can edit answers anytime before RSVP deadline
- Answer history not tracked in v0

---

## 9. Email System

### 9.1 SMTP Configuration

**Required Settings:**
```
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=user@gmail.com
SMTP_PASSWORD=app-password
SMTP_FROM=noreply@example.com
SMTP_FROM_NAME=TinyRSVP
```

**Optional Settings:**
```
SMTP_TLS=true                    # Use STARTTLS (default: true)
SMTP_SKIP_VERIFY=false           # Skip cert verification (default: false)
SMTP_TIMEOUT=30                  # Connection timeout in seconds (default: 30)
```

**Configuration Validation:**
- On startup: Validate all required settings present
- On startup: Attempt SMTP connection (fail fast if invalid)
- Admin UI: "Test Email" button sends test email
- Invalid config: App fails to start with clear error message

### 9.2 Email Queue

**Queue Table:**
```sql
CREATE TABLE email_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    to_email TEXT NOT NULL,
    to_name TEXT,
    subject TEXT NOT NULL,
    body_text TEXT NOT NULL,
    body_html TEXT,
    attachments JSON,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP,
    last_error TEXT,
    scheduled_for TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_status_scheduled (status, scheduled_for)
);
```

**Status Values:**
- `pending` - Waiting to be sent
- `sending` - Currently being sent
- `sent` - Successfully delivered
- `failed` - Permanently failed after retries
- `cancelled` - Manually cancelled

### 9.3 Email Sending Flow

**Immediate Send (Hybrid Approach):**
```
1. Email queued with status='pending', scheduled_for=NOW
2. Immediate send attempt:
   a. Update status='sending'
   b. Attempt SMTP send
   c. If success: status='sent', done
   d. If failure: status='pending', schedule retry
3. Return to caller (don't wait for send)
```

**Background Retry Processing:**
```
1. Every 60 seconds, background goroutine runs
2. Query: status='pending' AND scheduled_for <= NOW
3. For each email:
   a. Update status='sending'
   b. Attempt SMTP send
   c. If success: status='sent'
   d. If failure: schedule next retry
4. Sleep 60 seconds, repeat
```

**Retry Policy:**
- Attempt 1: Immediate
- Attempt 2: +1 minute (exponential backoff)
- Attempt 3: +5 minutes
- Attempt 4: +15 minutes
- After 4 attempts: status='failed', notify admin

**Rate Limiting:**
- Configurable via `EMAIL_RATE_LIMIT` (default: 50 per minute)
- Enforced by background worker
- Tracks sends in rolling 60-second window
- If limit reached: delay next batch

### 9.4 Email Templates

**Invite Email:**
```
Subject: You're invited: {event.title}

Hi {invite.name},

You're invited to {event.title}!

When: {event.start_time} ({event.timezone})
Where: {event.location}

{event.description}

Please RSVP by {event.rsvp_deadline}:
{rsvp_url}

See you there!

---
Sent via TinyRSVP
```

**RSVP Confirmation Email:**
```
Subject: RSVP Confirmed: {event.title}

Hi {invite.name},

Your RSVP has been confirmed!

Event: {event.title}
Your Response: {rsvp.response}
Plus Ones: {rsvp.plus_ones}

When: {event.start_time}
Where:
**Single-Node Deployment:**
- Not designed for horizontal scaling in v0
- Single container handles all requests
- Database is single SQLite file
- No distributed locking needed

**Email Sending:**
- Rate limited to prevent SMTP throttling
- Large batches may take time to process
- Consider SMTP provider limits

**Storage:**
- Local filesystem only in v0
- No CDN or distributed storage
- Image serving from app (not optimized)

---

## 25. Appendix

### 25.1 Glossary

| Term | Definition |
|------|------------|
| Admin | User with full system control |
| Event Manager | User who can create and manage events |
| Guest | Person invited to event (no account) |
| Invite | Record linking guest to event with unique token |
| RSVP | Guest's response to invitation |
| Token | Cryptographically secure random value for guest access |
| HMAC | Hash-based Message Authentication Code |
| ICS | iCalendar file format for calendar events |
| IANA Timezone | Standard timezone database (e.g., America/Los_Angeles) |
| Optimistic Locking | Concurrency control using version numbers |

### 25.2 References

**Standards:**
- RFC 5545: iCalendar specification
- RFC 6749: OAuth 2.0 Authorization Framework
- OpenID Connect Core 1.0
- CAN-SPAM Act compliance

**Libraries:**
- github.com/coreos/go-oidc - OIDC client
- golang.org/x/oauth2 - OAuth2 helper
- github.com/mattn/go-sqlite3 - SQLite driver
- github.com/golang-migrate/migrate - Database migrations

**Documentation:**
- Go html/template: https://pkg.go.dev/html/template
- IANA Timezone Database: https://www.iana.org/time-zones
- iCalendar Validator: https://icalendar.org/validator.html

### 25.3 Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-06 | Human | Initial HLD |
| 2.0 | 2026-01-06 | AI Assistant | Comprehensive revision addressing 50+ gaps from design review |

**Major Changes in v2.0:**
- Added complete database schema with all 9 tables
- Defined explicit permission matrix for roles
- Specified bootstrap admin creation mechanism
- Defined token lifecycle and expiration policy
- Specified timezone handling (IANA format)
- Defined event lifecycle states and transitions
- Specified email retry policy and rate limiting
- Defined session management strategy
- Added validation rules section
- Added error handling section
- Added security section
- Added operations section
- Added request flow diagrams
- Resolved contradictions (passphrase, public events, guest OIDC)
- Clarified v0 scope exclusions

---

**This document is the authoritative specification for TinyRSVP v0 implementation.**

**Next Steps:**
1. Review and approve this HLD
2. Create implementation backlog stories
3. Begin implementation following TDD approach
4. Reference this document for all design decisions

---

**Document Status:** ✅ Complete and Ready for Implementation
