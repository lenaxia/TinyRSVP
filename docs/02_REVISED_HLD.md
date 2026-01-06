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
Where: {event.location}

You can update your RSVP anytime before the deadline:
{rsvp_url}

Add to your calendar (attached)

---
Sent via TinyRSVP
```

**Event Update Email:**
```
Subject: Event Updated: {event.title}

Hi {invite.name},

The event details have been updated:

{changes_summary}

Updated Details:
When: {event.start_time}
Where: {event.location}

View and update your RSVP:
{rsvp_url}

---
Sent via TinyRSVP
```

**Event Cancellation Email:**
```
Subject: Event Cancelled: {event.title}

Hi {invite.name},

Unfortunately, {event.title} has been cancelled.

{cancellation_reason}

We apologize for any inconvenience.

---
Sent via TinyRSVP
```

### 9.5 Email Bounce Handling

**v0 Scope:** Basic bounce detection

**Mechanism:**
- Monitor SMTP send errors
- Classify errors as temporary (4xx) or permanent (5xx)
- Temporary: Retry per retry policy
- Permanent: Mark email as failed, flag invite

**Bounce Types:**
- Hard bounce (mailbox doesn't exist): Mark invite.email_invalid = true
- Soft bounce (mailbox full): Retry per policy
- Block (spam filter): Mark as failed, notify admin

**Admin Notification:**
- Daily digest of failed emails
- Includes: guest name, email, error message
- Admin can manually resend or update email

**v1+ Enhancement:**
- Parse bounce emails via SMTP callback
- Automatic email validation
- Guest self-service email update

### 9.6 Email Compliance

**CAN-SPAM Compliance:**
- All emails include sender identification
- All emails include physical address (configurable)
- Reminder emails include unsubscribe link
- Unsubscribe processed within 10 business days

**Unsubscribe Mechanism:**
- Link format: `/unsubscribe/{token}`
- Sets invite.unsubscribed = true
- Stops reminder emails
- Does not affect ability to RSVP

---

## 10. Calendar Integration (ICS)

### 10.1 ICS File Generation

**Format:** iCalendar (RFC 5545)

**Required Fields:**
```
BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//TinyRSVP//EN
METHOD:REQUEST
BEGIN:VEVENT
UID:{event.id}@{domain}
DTSTAMP:{generation_time}
DTSTART;TZID={event.timezone}:{event.start_time}
DTEND;TZID={event.timezone}:{event.end_time}
SUMMARY:{event.title}
DESCRIPTION:{event.description}\n\nRSVP: {rsvp_url}
LOCATION:{event.location}
STATUS:CONFIRMED
SEQUENCE:0
BEGIN:VALARM
TRIGGER:-PT24H
ACTION:DISPLAY
DESCRIPTION:Reminder: {event.title} tomorrow
END:VALARM
END:VEVENT
END:VCALENDAR
```

**Timezone Handling:**
- Use IANA timezone name from event.timezone
- Include VTIMEZONE component for timezone definition
- Convert times to event timezone (not UTC)
- Ensures correct display in all calendar clients

**UID Generation:**
- Format: `{event.id}@{base_domain}`
- Example: `123@rsvp.example.com`
- Stable across updates (same UID for same event)

**Sequence Number:**
- Starts at 0
- Incremented on each event update
- Tells calendar clients this is an update

### 10.2 ICS Updates

**When Event Details Change:**
1. Increment event.ics_sequence
2. Generate new ICS with updated SEQUENCE
3. Queue update emails to all RESPONDED invites
4. Email subject: "Event Updated: {title}"
5. Attach updated ICS file

**When Event Cancelled:**
1. Generate ICS with STATUS:CANCELLED
2. Queue cancellation emails to all invites
3. Email subject: "Event Cancelled: {title}"
4. Attach cancellation ICS file

**Calendar Client Behavior:**
- Same UID + higher SEQUENCE = update existing event
- STATUS:CANCELLED = remove from calendar

### 10.3 ICS Validation

**Before Generation:**
- Validate timezone exists in IANA database
- Validate start_time < end_time
- Validate all required fields present
- Escape special characters per RFC 5545

**Error Handling:**
- Invalid timezone: Use UTC as fallback, log warning
- Missing end_time: Set to start_time + 2 hours
- Invalid characters: Strip or escape

---

## 11. Templates & Customization

### 11.1 Template Types

| Type | Purpose | Format |
|------|---------|--------|
| invite_email | Invitation email body | HTML + text |
| rsvp_page | Guest RSVP page | HTML |
| confirmation_page | Post-RSVP confirmation | HTML |

### 11.2 Template Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| id | INTEGER | Auto | Primary key |
| name | TEXT | Yes | Template name (max 100 chars) |
| type | TEXT | Yes | invite_email/rsvp_page/confirmation_page |
| html_content | TEXT | Yes | HTML template |
| text_content | TEXT | Conditional | Plain text (required for email) |
| css_content | TEXT | No | Inline CSS |
| is_default | BOOLEAN | Yes | System default template |
| created_by | INTEGER | No | User ID (null for system templates) |
| created_at | TIMESTAMP | Auto | Creation timestamp |
| updated_at | TIMESTAMP | Auto | Last update timestamp |

### 11.3 Template Variables

**Available in All Templates:**
- `{{.Event.Title}}`
- `{{.Event.Description}}`
- `{{.Event.StartTime}}`
- `{{.Event.EndTime}}`
- `{{.Event.Timezone}}`
- `{{.Event.Location}}`
- `{{.Event.RSVPDeadline}}`

**Available in Invite Templates:**
- `{{.Invite.Name}}`
- `{{.Invite.Email}}`
- `{{.RSVPURL}}`
- `{{.MaxPlusOnes}}`

**Available in RSVP Page Templates:**
- `{{.RSVP.Response}}`
- `{{.RSVP.PlusOnes}}`
- `{{.Questions}}` (array)

### 11.4 Template Security

**XSS Prevention:**
- Use Go `html/template` (auto-escapes by default)
- All user input escaped before rendering
- No `template.HTML` type (disables escaping)
- CSS sanitized (no `javascript:` URLs, no `expression()`)

**Template Validation:**
- Parse template on upload
- Reject if parse fails
- Reject if contains disallowed functions
- Reject if references undefined variables

**Allowed Template Functions:**
- Date formatting: `{{.StartTime.Format "Jan 2, 2006"}}`
- String operations: `{{.Title | upper}}`
- Conditionals: `{{if .RSVPDeadline}}...{{end}}`
- Loops: `{{range .Questions}}...{{end}}`

**Disallowed:**
- JavaScript execution
- External resource loading
- Form submissions to external URLs
- Arbitrary HTML attributes

### 11.5 Template Versioning

**v0 Scope:** No versioning

**Behavior:**
- Template updates apply immediately
- Existing queued emails use current template
- No template history

**Rationale:** Simplicity for v0, versioning adds complexity

**v1+ Enhancement:**
- Template versions
- Emails reference specific version
- Can revert to previous version

### 11.6 Image Upload

**Allowed File Types:**
- image/jpeg
- image/png
- image/gif
- image/webp

**Validation:**
- Max file size: 5MB
- Max dimensions: 4096x4096 pixels
- File type verified by magic bytes (not just extension)
- Image re-encoded to strip EXIF data

**Storage:**
- Stored via storage provider (local FS or S3)
- Path: `/uploads/images/{event_id}/{filename}`
- Filename: Original name sanitized + random suffix

**Access Control:**
- Images are public (no auth required)
- Used in email templates and RSVP pages
- URL format: `/assets/images/{event_id}/{filename}`

---

## 12. Asset Storage

### 12.1 Storage Provider Interface

**Required Methods:**
```go
type StorageProvider interface {
    PutObject(ctx context.Context, path string, data io.Reader, contentType string) error
    GetObject(ctx context.Context, path string) (io.ReadCloser, error)
    DeleteObject(ctx context.Context, path string) error
    GetPublicURL(ctx context.Context, path string) (string, error)
    ListObjects(ctx context.Context, prefix string) ([]string, error)
}
```

### 12.2 Local Filesystem Provider

**Configuration:**
```
STORAGE_TYPE=local
STORAGE_PATH=/data/uploads
```

**Behavior:**
- Files stored in mounted volume
- Public URLs: `/assets/{path}`
- Served by app (not external web server)

**Directory Structure:**
```
/data/uploads/
├── images/
│   ├── {event_id}/
│   │   └── {filename}
├── templates/
│   └── {template_id}/
│       └── {filename}
```

### 12.3 S3-Compatible Provider (v1+)

**Configuration:**
```
STORAGE_TYPE=s3
S3_ENDPOINT=https://s3.amazonaws.com
S3_REGION=us-west-2
S3_BUCKET=tinyrsvp-assets
S3_ACCESS_KEY=...
S3_SECRET_KEY=...
```

**Behavior:**
- Files stored in S3 bucket
- Public URLs: Pre-signed URLs or CloudFront
- Served by S3 (not app)

**v0 Scope:** Excluded (local FS only)

### 12.4 Storage Quota

**v0 Scope:** No enforced quota

**Monitoring:**
- Admin can view total storage used
- Warning if >80% of disk space used
- No automatic cleanup

**v1+ Enhancement:**
- Configurable quota per event
- Automatic cleanup of old assets
- Compression of large images

### 12.5 Asset Deletion Policy

**When Event Deleted:**
- Assets moved to trash (not immediately deleted)
- Trash emptied after 30 days
- Admin can permanently delete immediately

**When Template Deleted:**
- Assets remain if referenced by events
- Orphaned assets cleaned up monthly

**Manual Deletion:**
- Admin can delete individual assets
- Confirmation required
- No recovery after deletion

---

## 13. Database Schema

### 13.1 Schema Overview

**Database Support:**
- SQLite 3.35+ (default)
- PostgreSQL 12+ (v1+)

**Design Principles:**
- Normalized to 3NF
- Foreign keys enforced
- Indexes on common queries
- Timestamps on all tables
- Soft deletes where appropriate

### 13.2 Complete Schema

#### users

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    name TEXT,
    role TEXT NOT NULL DEFAULT 'event_manager',
    oidc_subject TEXT UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    CHECK (role IN ('admin', 'event_manager'))
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_oidc_subject ON users(oidc_subject);
CREATE INDEX idx_users_role ON users(role);
```

**Notes:**
- `oidc_subject` is NULL for forward auth users
- `email` is unique identifier
- First user gets role='admin' automatically

#### sessions

```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_accessed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

**Notes:**
- Session ID is 32-byte random value (base64-encoded)
- Expires 7 days from creation
- Cleanup job deletes expired sessions hourly

#### events

```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    location TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    timezone TEXT NOT NULL,
    rsvp_deadline TIMESTAMP,
    max_plus_ones INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'draft',
    template_id INTEGER REFERENCES templates(id) ON DELETE SET NULL,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    ics_sequence INTEGER NOT NULL DEFAULT 0,
    CHECK (status IN ('draft', 'published', 'cancelled', 'archived')),
    CHECK (max_plus_ones >= 0 AND max_plus_ones <= 10),
    CHECK (end_time IS NULL OR end_time > start_time),
    CHECK (rsvp_deadline IS NULL OR rsvp_deadline < start_time)
);

CREATE INDEX idx_events_created_by ON events(created_by);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_start_time ON events(start_time);
```

**Notes:**
- `version` used for optimistic locking
- `ics_sequence` incremented on each update for calendar clients
- `created_by` cannot be deleted (RESTRICT) - must transfer ownership first

#### invites

```sql
CREATE TABLE invites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name TEXT,
    email TEXT,
    token_hash TEXT NOT NULL UNIQUE,
    max_plus_ones INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    sent_at TIMESTAMP,
    viewed_at TIMESTAMP,
    unsubscribed BOOLEAN NOT NULL DEFAULT FALSE,
    email_invalid BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    CHECK (status IN ('draft', 'sent', 'viewed', 'responded', 'revoked')),
    CHECK (max_plus_ones >= 0 AND max_plus_ones <= 10)
);

CREATE INDEX idx_invites_event_id ON invites(event_id);
CREATE INDEX idx_invites_token_hash ON invites(token_hash);
CREATE INDEX idx_invites_email ON invites(email);
CREATE INDEX idx_invites_status ON invites(status);
CREATE INDEX idx_invites_expires_at ON invites(expires_at);
```

**Notes:**
- `token_hash` is HMAC-SHA256 of actual token
- `expires_at` set to event.start_time + 30 days
- `email` can be NULL (for manual distribution)

#### rsvps

```sql
CREATE TABLE rsvps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invite_id INTEGER NOT NULL UNIQUE REFERENCES invites(id) ON DELETE CASCADE,
    response TEXT NOT NULL,
    plus_ones INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (response IN ('yes', 'no', 'maybe')),
    CHECK (plus_ones >= 0)
);

CREATE INDEX idx_rsvps_invite_id ON rsvps(invite_id);
CREATE INDEX idx_rsvps_response ON rsvps(response);
```

**Notes:**
- One RSVP per invite (UNIQUE constraint)
- plus_ones validated against invite.max_plus_ones in application logic

#### preference_questions

```sql
CREATE TABLE preference_questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    question_type TEXT NOT NULL,
    options JSON,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (question_type IN ('text', 'select', 'boolean'))
);

CREATE INDEX idx_questions_event_id ON preference_questions(event_id);
CREATE INDEX idx_questions_display_order ON preference_questions(event_id, display_order);
```

**Notes:**
- `options` is JSON array for select type, NULL for others
- `display_order` determines question sequence

#### rsvp_answers

```sql
CREATE TABLE rsvp_answers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rsvp_id INTEGER NOT NULL REFERENCES rsvps(id) ON DELETE CASCADE,
    question_id INTEGER NOT NULL REFERENCES preference_questions(id) ON DELETE CASCADE,
    answer_text TEXT,
    answer_option TEXT,
    answer_boolean BOOLEAN,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rsvp_id, question_id)
);

CREATE INDEX idx_answers_rsvp_id ON rsvp_answers(rsvp_id);
CREATE INDEX idx_answers_question_id ON rsvp_answers(question_id);
```

**Notes:**
- One answer per RSVP per question
- Answer type must match question type (enforced in application)

#### email_queue

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
    max_attempts INTEGER NOT NULL DEFAULT 4,
    last_attempt_at TIMESTAMP,
    last_error TEXT,
    scheduled_for TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (status IN ('pending', 'sending', 'sent', 'failed', 'cancelled'))
);

CREATE INDEX idx_email_queue_status_scheduled ON email_queue(status, scheduled_for);
CREATE INDEX idx_email_queue_status ON email_queue(status);
```

**Notes:**
- `attachments` is JSON array of {filename, content_base64, content_type}
- Background worker queries by status='pending' AND scheduled_for <= NOW

#### templates

```sql
CREATE TABLE templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    html_content TEXT NOT NULL,
    text_content TEXT,
    css_content TEXT,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (type IN ('invite_email', 'rsvp_page', 'confirmation_page'))
);

CREATE INDEX idx_templates_type ON templates(type);
CREATE INDEX idx_templates_is_default ON templates(is_default);
CREATE INDEX idx_templates_created_by ON templates(created_by);
```

**Notes:**
- One default template per type
- System templates have created_by = NULL

#### config

```sql
CREATE TABLE config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Notes:**
- Stores system configuration (HMAC secret, etc.)
- Key-value pairs for flexibility

### 13.3 Database Relationships

```
users (1) ──────< (N) events
users (1) ──────< (N) sessions
users (1) ──────< (N) templates

events (1) ──────< (N) invites
events (1) ──────< (N) preference_questions
events (N) ──────> (1) templates [optional]

invites (1) ──────< (1) rsvps
invites (N) ──────> (1) events

rsvps (1) ──────< (N) rsvp_answers

preference_questions (1) ──────< (N) rsvp_answers
preference_questions (N) ──────> (1) events
```

### 13.4 Database Migrations

**Migration Tool:** golang-migrate/migrate

**Migration Files:**
- Location: `migrations/sqlite/` and `migrations/postgres/`
- Naming: `{version}_{description}.up.sql` and `{version}_{description}.down.sql`
- Example: `001_initial_schema.up.sql`

**Migration Strategy:**
- Applied automatically on startup
- Version tracked in `schema_migrations` table
- Rollback supported via `.down.sql` files
- Failed migration prevents app startup

---

## 14. Validation Rules

### 14.1 Event Validation

| Field | Rule | Error Message |
|-------|------|---------------|
| title | Required, 3-200 chars | "Event title must be between 3 and 200 characters" |
| description | Optional, max 5000 chars | "Description cannot exceed 5000 characters" |
| location | Optional, max 500 chars | "Location cannot exceed 500 characters" |
| start_time | Required, valid ISO 8601, future | "Event start time must be in the future" |
| end_time | Optional, after start_time, within 7 days | "Event end time must be after start time" |
| timezone | Required, valid IANA timezone | "Invalid timezone. Use format like 'America/Los_Angeles'" |
| rsvp_deadline | Optional, before start_time, future | "RSVP deadline must be before event start time" |
| max_plus_ones | Required, 0-10 | "Max plus ones must be between 0 and 10" |

### 14.2 Invite Validation

| Field | Rule | Error Message |
|-------|------|---------------|
| name | Optional, max 100 chars | "Guest name cannot exceed 100 characters" |
| email | Optional, valid email format, max 255 chars | "Invalid email address" |
| max_plus_ones | Required, 0-10, <= event.max_plus_ones | "Max plus ones cannot exceed event limit" |

**Email Format Validation:**
- Regex: `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
- Additional check: Not longer than 255 characters
- Additional check: Local part < 64 chars, domain < 255 chars

### 14.3 RSVP Validation

| Field | Rule | Error Message |
|-------|------|---------------|
| response | Required, yes/no/maybe | "Please select a response" |
| plus_ones | Required, 0 to invite.max_plus_ones | "You can bring up to {max} guest(s)" |
| deadline | Must be before event.rsvp_deadline | "RSVP deadline has passed" |

### 14.4 Question Validation

| Field | Rule | Error Message |
|-------|------|---------------|
| question_text | Required, 5-500 chars | "Question must be between 5 and 500 characters" |
| question_type | Required, text/select/boolean | "Invalid question type" |
| options | Required for select, 2-20 options | "Select questions must have 2-20 options" |

### 14.5 Answer Validation

| Question Type | Rule | Error Message |
|---------------|------|---------------|
| text | Max 500 chars | "Answer cannot exceed 500 characters" |
| select | Must match one option value | "Invalid selection" |
| boolean | Must be true/false | "Please answer yes or no" |
| required | Must have answer | "This question is required" |

### 14.6 Template Validation

| Field | Rule | Error Message |
|-------|------|---------------|
| name | Required, 3-100 chars | "Template name must be between 3 and 100 characters" |
| type | Required, valid type | "Invalid template type" |
| html_content | Required, valid Go template syntax | "Template syntax error: {details}" |
| text_content | Required for email templates | "Email templates must have text content" |

### 14.7 Image Upload Validation

| Rule | Error Message |
|------|---------------|
| File type must be image/jpeg, image/png, image/gif, image/webp | "Only JPEG, PNG, GIF, and WebP images are allowed" |
| Max file size: 5MB | "Image file size cannot exceed 5MB" |
| Max dimensions: 4096x4096 | "Image dimensions cannot exceed 4096x4096 pixels" |
| File must be valid image (magic bytes check) | "File is not a valid image" |

---

## 15. Error Handling

### 15.1 Error Response Format

**HTTP API Errors:**
```json
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Event title must be between 3 and 200 characters",
        "field": "title",
        "details": {}
    }
}
```

**HTML Page Errors:**
- User-friendly error page with message
- Suggested actions (e.g., "Go back and try again")
- Support contact information

### 15.2 Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| VALIDATION_ERROR | 400 | Input validation failed |
| UNAUTHORIZED | 401 | Authentication required |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| CONFLICT | 409 | Concurrent modification conflict |
| RATE_LIMITED | 429 | Too many requests |
| INTERNAL_ERROR | 500 | Unexpected server error |
| SERVICE_UNAVAILABLE | 503 | Service temporarily unavailable |

### 15.3 Error Scenarios

**Authentication Errors:**
- Missing session cookie → Redirect to /login
- Expired session → Redirect to /login with message
- Invalid session → Delete cookie, redirect to /login
- OIDC provider down → Show error page with retry button

**Authorization Errors:**
- Insufficient permissions → HTTP 403 with explanation
- Accessing other user's event → HTTP 403 "You don't have permission to edit this event"

**Validation Errors:**
- Invalid input → HTTP 400 with field-specific errors
- Multiple validation errors → Return all errors at once

**Concurrent Modification:**
- Version mismatch → HTTP 409 "Event was modified by another user. Please refresh and try again."
- Include current version in response for retry

**Resource Not Found:**
- Event not found → HTTP 404 "Event not found"
- Invite token not found → HTTP 404 "Invalid invite link"

**Rate Limiting:**
- Too many requests → HTTP 429 "Too many requests. Please try again in {seconds} seconds."
- Retry-After header included

**SMTP Errors:**
- Connection failed → Queue for retry, log error
- Authentication failed → Mark as failed, notify admin
- Recipient rejected → Mark email_invalid, notify admin

**Database Errors:**
- Connection lost → HTTP 503 "Service temporarily unavailable"
- Constraint violation → HTTP 400 with user-friendly message
- Deadlock → Retry transaction up to 3 times

---

## 16. Security

### 16.1 Transport Security

**HTTPS Enforcement:**
- App expects to run behind reverse proxy with TLS termination
- App sets Secure flag on cookies (requires HTTPS)
- App can optionally redirect HTTP to HTTPS (configurable)

**Configuration:**
```
HTTPS_REQUIRED=true          # Reject HTTP requests (default: true)
HTTPS_REDIRECT=false         # Redirect HTTP to HTTPS (default: false)
TRUSTED_PROXY_IPS=10.0.0.1   # Comma-separated list of proxy IPs
```

### 16.2 Security Headers

**All Responses Include:**
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'
```

**CSP Policy Breakdown:**
- `default-src 'self'` - Only load resources from same origin
- `img-src 'self' data:` - Images from same origin or data URLs
- `style-src 'self' 'unsafe-inline'` - Styles from same origin or inline (for email compatibility)
- `script-src 'self'` - Scripts only from same origin
- `form-action 'self'` - Forms only submit to same origin

### 16.3 Input Sanitization

**HTML Content:**
- Use Go `html/template` for automatic escaping
- Markdown rendered with sanitization (no raw HTML)
- Strip dangerous tags: `<script>`, `<iframe>`, `<object>`, `<embed>`

**SQL Injection Prevention:**
- Use parameterized queries exclusively
- Never concatenate user input into SQL
- Use database/sql package (built-in protection)

**Path Traversal Prevention:**
- Validate file paths don't contain `..`
- Use filepath.Clean() to normalize paths
- Restrict file access to designated directories

### 16.4 CSRF Protection

**Mechanism:** SameSite cookies + CSRF tokens

**CSRF Token:**
- Generated per session
- Stored in session
- Included in all forms as hidden field
- Validated on POST/PUT/DELETE requests

**Implementation:**
- Middleware validates CSRF token
- GET requests don't require token
- API endpoints require token in header or form

### 16.5 Secrets Management

**Secrets in Environment:**
- OIDC_CLIENT_SECRET
- SMTP_PASSWORD
- S3_SECRET_KEY (v1+)

**Secrets in Database:**
- HMAC secret key (for token hashing)
- Session encryption key (future)

**Protection:**
- Never logged
- Never returned in API responses
- Masked in admin UI (show only last 4 chars)
- Encrypted at rest (database encryption recommended)

**Rotation:**
- HMAC secret rotation invalidates all tokens (admin command)
- SMTP password rotation via environment variable update + restart
- OIDC secret rotation via environment variable update + restart

### 16.6 Audit Logging

**Logged Actions:**
- User login/logout
- User role changes
- Event creation/update/deletion
- Invite creation/revocation
- RSVP submissions/updates
- System configuration changes
- Failed authentication attempts

**Audit Log Table:**
```sql
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id INTEGER,
    details JSON,
    ip_address TEXT,
    user_agent TEXT
);

CREATE INDEX idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id);
```

**Retention:**
- Audit logs kept for 1 year
- Automatic cleanup of logs >1 year old

---

## 17. Operations

### 17.1 Health Checks

**Endpoint:** `/health`

**Checks Performed:**
- Database connectivity (SELECT 1)
- Database write test (INSERT/DELETE test row)
- SMTP connectivity (optional, can be slow)
- Disk space available (>10% free)

**Response Format:**
```json
{
    "status": "healthy",
    "checks": {
        "database": "ok",
        "disk_space": "ok",
        "smtp": "ok"
    },
    "version": "0.1.0",
    "uptime_seconds": 3600
}
```

**Status Codes:**
- 200: All checks passed
- 503: One or more checks failed

### 17.2 Logging

**Log Levels:**
- ERROR: Errors requiring attention
- WARN: Warnings, degraded functionality
- INFO: Normal operations, state changes
- DEBUG: Detailed debugging (disabled in production)

**Log Format:** Structured JSON
```json
{
    "timestamp": "2026-01-06T18:00:00Z",
    "level": "INFO",
    "message": "Event created",
    "event_id": 123,
    "user_id": 1,
    "ip": "10.0.0.1"
}
```

**Sensitive Data Handling:**
- Tokens never logged
- Passwords never logged
- Email addresses logged only at INFO level
- HMAC secret never logged

**Log Output:**
- stdout (captured by Docker)
- Optionally: file rotation (configurable)

### 17.3 Monitoring

**Metrics Endpoint:** `/metrics` (Prometheus format)

**Key Metrics:**
- `tinyrsvp_events_total{status}` - Events by status
- `tinyrsvp_invites_total{status}` - Invites by status
- `tinyrsvp_rsvps_total{response}` - RSVPs by response
- `tinyrsvp_emails_total{status}` - Emails by status
- `tinyrsvp_http_requests_total{method,path,status}` - HTTP requests
- `tinyrsvp_http_request_duration_seconds` - Request latency
- `tinyrsvp_db_connections` - Database connection pool
- `tinyrsvp_email_queue_size` - Pending emails

### 17.4 Background Jobs

**Jobs:**
1. **Email Queue Processor** - Every 60 seconds
2. **Session Cleanup** - Every hour
3. **Token Expiration Cleanup** - Every 24 hours
4. **Event Auto-Archive** - Every 24 hours
5. **Audit Log Cleanup** - Every 7 days

**Job Execution:**
- Single goroutine per job
- Graceful shutdown on SIGTERM
- Jobs complete before shutdown (max 30 seconds)

### 17.5 Backup & Recovery

**Backup Responsibility:** Administrator

**Recommended Backup Strategy:**
- Database: Daily backup of SQLite file
- Uploads: Daily backup of /data/uploads directory
- Config: Backup environment variables

**Recovery:**
1. Stop application
2. Restore database file
3. Restore uploads directory
4. Start application
5. Verify health check passes

**RTO/RPO:**
- RTO (Recovery Time Objective): <15 minutes
- RPO (Recovery Point Objective): 24 hours (daily backups)

### 17.6 Disaster Recovery

**Scenarios:**

**Database Corruption:**
1. Stop app
2. Restore from backup
3. Restart app
4. Verify data integrity

**Disk Full:**
1. Free disk space
2. App auto-recovers
3. Check logs for failed operations

**SMTP Outage:**
- Emails queue automatically
- Retry when SMTP recovers
- No data loss

**Complete Server Loss:**
1. Deploy new server
2. Restore database backup
3. Restore uploads backup
4. Configure environment variables
5. Start application

### 17.7 Known Limitations

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

## 18. API Routes

**See:** [Domain 8: API & HTTP Handlers LLD](lld/08_API_LLD.md) for complete route specifications

**Route Categories:**
- Authentication routes (`/auth/*`)
- Event management routes (`/events/*`)
- Invite management routes (`/invites/*`)
- RSVP routes (`/rsvp/{token}`)
- Admin routes (`/admin/*`)
- Utility routes (`/health`, `/metrics`)

**All routes documented in:** [Domain 8 LLD Section 6](lld/08_API_LLD.md#6-api-routes)

---

## 19. Request Flow

**See:** [Domain 8: API & HTTP Handlers LLD](lld/08_API_LLD.md) for complete request flow diagrams

**Middleware Chain:**
1. Recovery (panic handling)
2. Logging (request logging)
3. Security Headers (CSP, HSTS, etc.)
4. Rate Limiting (per-IP)
5. Authentication (session validation)
6. RBAC (permission checking)
7. CSRF (token validation)
8. Handler (business logic)

**Detailed flow documented in:** [Domain 8 LLD Section 5](lld/08_API_LLD.md#5-request-flow)

---

## 20. Deployment Model

### 20.1 Container Architecture

**Single Container Deployment:**
- Application binary
- SQLite database (mounted volume)
- Static assets (embedded or mounted)
- Migrations (embedded)

**Docker Image:**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o tinyrsvp cmd/server/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/tinyrsvp /usr/local/bin/
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/templates /templates
COPY --from=builder /app/static /static
EXPOSE 8080
CMD ["tinyrsvp"]
```

### 20.2 Volume Mounts

**Required:**
- `/data` - Database and uploads directory

**Optional:**
- `/config` - Configuration overrides
- `/templates` - Custom templates

### 20.3 Environment Configuration

**Minimal Configuration:**
```yaml
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
```

### 20.4 Reverse Proxy Integration

**Traefik Example:**
```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.tinyrsvp.rule=Host(`rsvp.example.com`)"
  - "traefik.http.routers.tinyrsvp.tls=true"
  - "traefik.http.routers.tinyrsvp.tls.certresolver=letsencrypt"
```

**Nginx Example:**
```nginx
server {
    listen 443 ssl http2;
    server_name rsvp.example.com;
    
    ssl_certificate /etc/ssl/certs/rsvp.crt;
    ssl_certificate_key /etc/ssl/private/rsvp.key;
    
    location / {
        proxy_pass http://tinyrsvp:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## 21. v0 Scope

### 21.1 Included Features

**Core Functionality:**
- ✅ Event creation and management
- ✅ Invite system with token-based access
- ✅ RSVP handling with preference questions
- ✅ Email sending with ICS attachments
- ✅ OIDC authentication
- ✅ Forward auth support
- ✅ SQLite database
- ✅ Local filesystem storage
- ✅ Basic templates
- ✅ Mobile-responsive UI

**Supporting Features:**
- ✅ Session management
- ✅ Role-based access control
- ✅ Email queue with retry
- ✅ Event lifecycle (draft/published/cancelled/archived)
- ✅ Optimistic locking
- ✅ Audit logging
- ✅ Health checks
- ✅ Metrics endpoint

### 21.2 Explicitly Excluded from v0

**Deferred to v1+:**
- ❌ PostgreSQL support (SQLite only in v0)
- ❌ S3-compatible storage (local FS only in v0)
- ❌ Guest OIDC authentication
- ❌ Public event links (all events require invite)
- ❌ Event passphrases
- ❌ Reminder email scheduling UI
- ❌ Template versioning
- ❌ Event capacity limits
- ❌ Waitlist functionality
- ❌ Advanced bounce handling
- ❌ SMS notifications
- ❌ CalDAV sync
- ❌ Event analytics
- ❌ Custom branding
- ❌ Multi-language support

**Rationale:** Focus on core functionality, ensure quality over quantity, maintain simplicity for v0 release.

---

## 22. Success Criteria

### 22.1 Functional Requirements

**Must Work:**
- [ ] Admin can authenticate via OIDC or forward auth
- [ ] Admin can create event with all required fields
- [ ] Admin can publish event
- [ ] Admin can create invite with email
- [ ] Admin can send invite email
- [ ] Guest receives email with RSVP link
- [ ] Guest can click link and see RSVP page
- [ ] Guest can submit RSVP (yes/no/maybe)
- [ ] Guest can specify plus ones (within limits)
- [ ] Guest can answer preference questions
- [ ] Guest receives confirmation email
- [ ] Admin can view all RSVPs for event
- [ ] Admin can export guest list
- [ ] Email includes ICS calendar attachment
- [ ] ICS file imports correctly into calendar apps

### 22.2 Non-Functional Requirements

**Performance:**
- [ ] RSVP page loads in <2 seconds
- [ ] Email sent within 5 minutes of queueing
- [ ] Supports 100 concurrent users
- [ ] Handles 1000 invites per event

**Security:**
- [ ] All cookies have Secure and HttpOnly flags
- [ ] CSRF protection on all mutations
- [ ] Tokens use HMAC-SHA256
- [ ] No XSS vulnerabilities
- [ ] No SQL injection vulnerabilities
- [ ] Security headers on all responses

**Reliability:**
- [ ] Email retry on transient failures
- [ ] Graceful degradation if SMTP down
- [ ] Database transactions for atomic operations
- [ ] No data loss on server restart

**Usability:**
- [ ] Mobile-responsive design
- [ ] Works without JavaScript
- [ ] Clear error messages
- [ ] Intuitive navigation

### 22.3 Deployment Requirements

**Must Deploy:**
- [ ] Single Docker container
- [ ] Works on Raspberry Pi 4 (ARM64)
- [ ] Uses <512MB RAM under normal load
- [ ] Starts in <10 seconds
- [ ] Health check responds correctly
- [ ] Metrics endpoint works

### 22.4 Documentation Requirements

**Must Document:**
- [ ] Installation guide
- [ ] Configuration reference
- [ ] API documentation
- [ ] Troubleshooting guide
- [ ] Backup/restore procedure

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
