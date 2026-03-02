# High-Level Design Review - TinyRSVP

**Date:** 2026-01-06  
**Reviewer:** AI Assistant  
**Review Type:** Adversarial Design Review  
**HLD Version:** Initial (from docs/00_INITIAL_HLD.md)

---

## Executive Summary

This document presents a comprehensive adversarial design review of the TinyRSVP High-Level Design. The review identifies **critical gaps**, **design inconsistencies**, **security concerns**, and **missing requirements** that must be addressed before implementation.

**Overall Assessment:** The HLD provides a solid foundation but has **significant gaps** in operational requirements, error handling, data lifecycle, and edge case handling. Several areas require clarification or expansion.

**Severity Levels:**
- 🔴 **CRITICAL** - Must be addressed before implementation
- 🟡 **HIGH** - Should be addressed in v0
- 🟢 **MEDIUM** - Should be addressed but can be deferred
- ⚪ **LOW** - Nice to have, can be deferred to v1+

---

## 1. Authentication & Authorization Model

### 1.1 Admin/Manager Authentication

#### Issues Identified

🔴 **CRITICAL: Role Assignment Process Undefined**
- HLD states "First login may auto-create user (role must be assigned by admin)"
- **Gap:** How does the FIRST admin get created? Bootstrap problem.
- **Gap:** Who assigns the role to auto-created users? If no admin exists yet, system is locked.
- **Recommendation:** Define bootstrap admin creation mechanism (environment variable, CLI command, or first-user-is-admin pattern)

🔴 **CRITICAL: Authorization Model Incomplete**
- HLD defines roles (Admin, Event Manager) but not permissions
- **Gap:** What specific actions can Event Manager do vs Admin?
- **Gap:** Can Event Manager see/edit other managers' events?
- **Gap:** Can Event Manager delete events? Revoke invites?
- **Gap:** Can Event Manager access system settings (SMTP config)?
- **Recommendation:** Define explicit permission matrix for each role

🟡 **HIGH: Forward Auth Trust Model Unclear**
- HLD states app "trusts headers set by reverse proxy"
- **Gap:** What happens if headers are missing?
- **Gap:** What happens if headers are malformed?
- **Gap:** How does app verify headers are from trusted proxy (not spoofed)?
- **Gap:** Should app validate header format/content?
- **Recommendation:** Define header validation rules and fallback behavior

🟡 **HIGH: OIDC Auto-Creation Policy Undefined**
- "First login may auto-create user" - when exactly?
- **Gap:** What if OIDC provider returns no email claim?
- **Gap:** What if email claim is empty string?
- **Gap:** What if same email exists with different OIDC subject?
- **Gap:** What role is assigned to auto-created users?
- **Recommendation:** Define explicit auto-creation rules and edge case handling

🟢 **MEDIUM: Session Management Not Specified**
- HLD mentions "set session cookie" but no details
- **Gap:** Session duration/timeout not specified
- **Gap:** Session storage mechanism not specified (in-memory, database, Redis?)
- **Gap:** Session invalidation on logout not specified
- **Gap:** Concurrent session handling not specified
- **Recommendation:** Add session management requirements to HLD

🟢 **MEDIUM: Multi-Tenancy Considerations**
- HLD assumes single organization deployment
- **Gap:** What if multiple families want to use same instance?
- **Gap:** Can events be isolated by organization/tenant?
- **Gap:** Can admins be scoped to specific tenants?
- **Recommendation:** Clarify if multi-tenancy is explicitly excluded or should be considered

### 1.2 Guest Authentication

#### Issues Identified

🔴 **CRITICAL: Token Lifecycle Undefined**
- HLD states tokens "can optionally expire (future feature)"
- **Gap:** What is the default behavior in v0? Never expire?
- **Gap:** If tokens never expire, what are security implications?
- **Gap:** Can expired tokens be renewed?
- **Recommendation:** Define v0 token expiration policy (even if "never expires")

🟡 **HIGH: Token Revocation Mechanics Unclear**
- HLD states tokens "can be revoked"
- **Gap:** What happens to guest with revoked token?
- **Gap:** Can guest request new token?
- **Gap:** Is revocation immediate or eventual?
- **Gap:** How is revocation communicated to guest?
- **Recommendation:** Define revocation workflow and guest notification

🟡 **HIGH: Token Regeneration Impact Undefined**
- HLD states tokens "can be regenerated"
- **Gap:** Does regeneration invalidate old token immediately?
- **Gap:** What happens if guest has old token bookmarked?
- **Gap:** Should old token redirect to error or new token?
- **Gap:** How is regeneration communicated to guest?
- **Recommendation:** Define regeneration behavior and communication strategy

🟡 **HIGH: Passphrase Feature Incomplete**
- HLD mentions "optional passphrase (hashed)" but marked as "Explicitly Excluded" from v0
- **Gap:** Inconsistency - is it in data model or not?
- **Gap:** If in data model but not implemented, what's the migration path?
- **Recommendation:** Remove from data model if excluded, or clarify v0 scope

🟢 **MEDIUM: Guest OIDC Scope Unclear**
- "Optional guest auth (off by default)" via OIDC
- **Gap:** What's the use case for guest OIDC?
- **Gap:** Does it replace token or supplement it?
- **Gap:** If guest authenticates via OIDC, do they still need token?
- **Gap:** How does guest OIDC interact with invite token?
- **Recommendation:** Clarify guest OIDC purpose and interaction model

---

## 2. Event Model

### 2.1 Event Attributes

#### Issues Identified

🔴 **CRITICAL: Timezone Handling Incomplete**
- HLD states "Timezone (mandatory)" but no details
- **Gap:** What timezone format? IANA (America/Los_Angeles) or offset (UTC-8)?
- **Gap:** How are timezones displayed to guests in different zones?
- **Gap:** Does ICS file include timezone or convert to UTC?
- **Gap:** What happens if timezone is invalid?
- **Recommendation:** Specify timezone format and conversion rules

🟡 **HIGH: Event Lifecycle States Missing**
- HLD defines event attributes but not states
- **Gap:** Can events be draft, published, cancelled, completed?
- **Gap:** Can published events be unpublished?
- **Gap:** What happens to invites when event is cancelled?
- **Gap:** Can past events be edited?
- **Recommendation:** Define event lifecycle states and transitions

🟡 **HIGH: Event Capacity Not Addressed**
- HLD has "Max allowed +1s" per invite but no total capacity
- **Gap:** Can event have max total attendees?
- **Gap:** What happens when capacity is reached?
- **Gap:** Are +1s counted toward capacity?
- **Gap:** Can capacity be changed after invites sent?
- **Recommendation:** Clarify if event capacity is in scope or explicitly excluded

🟡 **HIGH: Event Visibility Transition Unclear**
- HLD defines "Public or private" but not transitions
- **Gap:** Can private event become public?
- **Gap:** Can public event become private?
- **Gap:** What happens to existing invites during transition?
- **Gap:** What happens to public RSVPs if event becomes private?
- **Recommendation:** Define visibility change rules and impact

🟢 **MEDIUM: Event Recurrence Not Addressed**
- HLD assumes single-occurrence events
- **Gap:** Are recurring events explicitly excluded?
- **Gap:** What if user wants weekly book club meetings?
- **Recommendation:** Explicitly state recurring events are out of scope for v0

🟢 **MEDIUM: Event Ownership Transfer**
- HLD states "Created by (user)" but no transfer mechanism
- **Gap:** Can event ownership be transferred?
- **Gap:** What happens if event creator leaves organization?
- **Gap:** Can multiple users co-own an event?
- **Recommendation:** Define ownership model and transfer rules

### 2.2 Event Visibility Modes

#### Issues Identified

🔴 **CRITICAL: Public Event Access Control Undefined**
- HLD states public events have "generic link" but details missing
- **Gap:** Can anyone with link RSVP unlimited times?
- **Gap:** How is abuse prevented (spam RSVPs)?
- **Gap:** Can public event require email verification?
- **Gap:** What's the difference between "require" vs "optionally request" email?
- **Recommendation:** Define public event access control and abuse prevention

🟡 **HIGH: Provisional Invite Lifecycle Unclear**
- HLD mentions "generates provisional invites" for public events
- **Gap:** What is a provisional invite?
- **Gap:** How does it differ from regular invite?
- **Gap:** When does provisional become permanent?
- **Gap:** Can provisional invites be revoked?
- **Recommendation:** Define provisional invite concept and lifecycle

🟢 **MEDIUM: Public Event Discovery**
- HLD doesn't address event discovery
- **Gap:** Can public events be listed/searched?
- **Gap:** Is there a public event directory?
- **Gap:** How do guests find public events?
- **Recommendation:** Clarify if discovery is in scope or out of scope

---

## 3. Invites & Guest Access Model

### 3.1 Invite Creation

#### Issues Identified

🔴 **CRITICAL: Bulk Invite Creation Not Specified**
- HLD mentions "explicit email invite" but not bulk operations
- **Gap:** Can admin upload CSV of 100 guests?
- **Gap:** How are bulk invite errors handled?
- **Gap:** Can admin send invites to all at once or individually?
- **Gap:** What's the rate limit for invite creation?
- **Recommendation:** Define bulk invite creation requirements and limits

🟡 **HIGH: Generic Link Mechanics Unclear**
- HLD mentions "generic link (public or private)" but details missing
- **Gap:** What's a "private generic link"? Isn't that contradictory?
- **Gap:** Does generic link create invite per use or single shared invite?
- **Gap:** Can generic link be disabled?
- **Gap:** How is generic link different from public event?
- **Recommendation:** Clarify generic link concept and mechanics

🟡 **HIGH: Manual Invite Creation Use Case Unclear**
- HLD mentions "manual creation (no email)"
- **Gap:** Why create invite without sending email?
- **Gap:** How does guest get the token?
- **Gap:** Is this for in-person distribution?
- **Recommendation:** Clarify manual invite use case and workflow

🟢 **MEDIUM: Invite Deduplication Not Addressed**
- **Gap:** Can same email be invited twice?
- **Gap:** What happens if duplicate invite is created?
- **Gap:** Should system warn or prevent duplicates?
- **Recommendation:** Define duplicate invite handling policy

### 3.2 Invite Token Design

#### Issues Identified

🔴 **CRITICAL: Token Collision Handling Missing**
- HLD specifies 256-bit tokens but not collision handling
- **Gap:** What happens if token collision occurs (astronomically unlikely but possible)?
- **Gap:** Should system detect and regenerate?
- **Gap:** What's the retry limit?
- **Recommendation:** Define collision detection and handling

🟡 **HIGH: Token Storage Security Incomplete**
- HLD states "stored hashed in database (SHA-256)"
- **Gap:** Is SHA-256 appropriate for token hashing? (It's fast, not slow like bcrypt)
- **Gap:** Should tokens be salted?
- **Gap:** What about timing attacks on token validation?
- **Recommendation:** Reconsider hashing algorithm (SHA-256 is for integrity, not password-like hashing)

🟡 **HIGH: Token URL Format Not Specified**
- HLD mentions "URL-safe Base64 encoding"
- **Gap:** What's the full URL format? `/rsvp/{token}`?
- **Gap:** How long is the encoded token?
- **Gap:** Does URL include any metadata (event ID)?
- **Recommendation:** Specify complete token URL format

🟢 **MEDIUM: Token Rotation Policy**
- **Gap:** Should tokens be rotated periodically?
- **Gap:** What's the security vs usability tradeoff?
- **Recommendation:** Define token rotation policy or explicitly exclude

### 3.3 Guest RSVP Capabilities

#### Issues Identified

🔴 **CRITICAL: RSVP State Transitions Undefined**
- HLD lists "yes / no / maybe" but not transition rules
- **Gap:** Can guest change from "yes" to "no" after deadline?
- **Gap:** Can guest change unlimited times?
- **Gap:** Should system track RSVP history?
- **Gap:** What happens if guest changes from "yes" to "no" day before event?
- **Recommendation:** Define RSVP state transition rules and deadline enforcement

🟡 **HIGH: +1 Validation Logic Missing**
- HLD states "specify number of +1s (within limits)"
- **Gap:** What if guest specifies 5 but limit is 2?
- **Gap:** Is this client-side or server-side validation?
- **Gap:** Can guest reduce +1s after increasing?
- **Gap:** What if event capacity is reached?
- **Recommendation:** Define +1 validation and error handling

🟡 **HIGH: RSVP Deadline Enforcement Unclear**
- HLD mentions "RSVP deadline (optional)" but not enforcement
- **Gap:** What happens if guest tries to RSVP after deadline?
- **Gap:** Can admin extend deadline?
- **Gap:** Can guest view event details after deadline?
- **Gap:** Should system send deadline reminders?
- **Recommendation:** Define deadline enforcement and grace period policy

🟢 **MEDIUM: RSVP Confirmation Not Specified**
- **Gap:** Does guest receive confirmation email after RSVP?
- **Gap:** Can guest view their RSVP later?
- **Gap:** Is there a confirmation page?
- **Recommendation:** Define RSVP confirmation workflow

---

## 4. Preference Questions

### 4.1 Question Types

#### Issues Identified

🔴 **CRITICAL: Question Validation Rules Missing**
- HLD defines types (Text, Select, Boolean) but not validation
- **Gap:** Text: max length? allowed characters? multiline?
- **Gap:** Select: min/max options? single vs multi-select?
- **Gap:** Boolean: default value? required vs optional?
- **Gap:** What happens if validation fails?
- **Recommendation:** Define validation rules for each question type

🟡 **HIGH: Question Lifecycle Unclear**
- **Gap:** Can questions be added after invites sent?
- **Gap:** Can questions be deleted after some guests answered?
- **Gap:** Can question type be changed?
- **Gap:** What happens to existing answers if question deleted?
- **Recommendation:** Define question lifecycle and impact on existing answers

🟡 **HIGH: Answer Editing Policy Undefined**
- HLD states "edit their RSVP at any time" but not answers
- **Gap:** Can guest edit answers after submission?
- **Gap:** Should system track answer history?
- **Gap:** Can guest skip optional questions?
- **Recommendation:** Define answer editing and history policy

🟢 **MEDIUM: Question Ordering Not Addressed**
- **Gap:** Can admin specify question display order?
- **Gap:** Is order preserved?
- **Recommendation:** Clarify if ordering is in scope

🟢 **MEDIUM: Conditional Questions Not Addressed**
- **Gap:** Can questions be conditional (show if RSVP=yes)?
- **Gap:** Is this explicitly out of scope?
- **Recommendation:** Explicitly exclude conditional logic for v0

---

## 5. Email System

### 5.1 Email Sending

#### Issues Identified

🔴 **CRITICAL: Email Delivery Failure Handling Missing**
- HLD mentions "retries" but no details
- **Gap:** How many retries? What intervals?
- **Gap:** What happens after max retries?
- **Gap:** How is admin notified of failures?
- **Gap:** Can admin manually retry failed emails?
- **Gap:** What about permanent failures (invalid email)?
- **Recommendation:** Define retry policy and failure notification

🔴 **CRITICAL: Email Rate Limiting Not Specified**
- **Gap:** What if admin sends 1000 invites at once?
- **Gap:** Will SMTP server rate limit?
- **Gap:** Should app throttle sending?
- **Gap:** What's the sending queue processing rate?
- **Recommendation:** Define email rate limiting and throttling strategy

🟡 **HIGH: Email Bounce Handling Missing**
- **Gap:** How are bounces detected?
- **Gap:** Should system mark email as invalid?
- **Gap:** Should admin be notified?
- **Gap:** Can guest update their email?
- **Recommendation:** Define bounce detection and handling

🟡 **HIGH: Email Unsubscribe Not Addressed**
- **Gap:** Can guests unsubscribe from reminders?
- **Gap:** Is unsubscribe link required (CAN-SPAM)?
- **Gap:** Does unsubscribe affect RSVP ability?
- **Recommendation:** Define unsubscribe mechanism and compliance

🟡 **HIGH: SMTP Configuration Validation Missing**
- HLD mentions "user-provided SMTP configuration"
- **Gap:** How is SMTP config validated?
- **Gap:** What if credentials are wrong?
- **Gap:** Is there a test email function?
- **Gap:** Can admin update SMTP config without restart?
- **Recommendation:** Define SMTP validation and testing requirements

🟢 **MEDIUM: Email Templates Not Detailed**
- HLD mentions "custom subject/body" but not template system
- **Gap:** What template variables are available?
- **Gap:** Can admin preview emails before sending?
- **Gap:** Are templates versioned?
- **Recommendation:** Define template system requirements

### 5.2 Email Queue

#### Issues Identified

🔴 **CRITICAL: Queue Processing Not Specified**
- HLD states "queued in database" but not processing
- **Gap:** Is queue processed by cron job, background worker, or on-demand?
- **Gap:** What's the processing interval?
- **Gap:** Can queue be paused?
- **Gap:** What happens if app restarts during processing?
- **Recommendation:** Define queue processing mechanism and reliability

🟡 **HIGH: Queue Observability Missing**
- HLD mentions "observability" but no details
- **Gap:** Can admin view queue status?
- **Gap:** Can admin see pending/failed emails?
- **Gap:** Are there metrics (sent, failed, pending)?
- **Gap:** Is there a queue dashboard?
- **Recommendation:** Define queue monitoring and admin visibility

🟢 **MEDIUM: Queue Prioritization Not Addressed**
- **Gap:** Are all emails equal priority?
- **Gap:** Should reminders be lower priority than invites?
- **Gap:** Can admin prioritize certain emails?
- **Recommendation:** Clarify if prioritization is needed

---

## 6. Calendar Attachments (ICS)

### Issues Identified

🔴 **CRITICAL: ICS Generation Edge Cases Missing**
- **Gap:** What if event has no end time?
- **Gap:** What if timezone is invalid?
- **Gap:** What if location is very long (>255 chars)?
- **Gap:** What if description contains special characters?
- **Recommendation:** Define ICS generation validation and limits

🟡 **HIGH: ICS Update Mechanism Unclear**
- **Gap:** If event details change, how do guests get updated ICS?
- **Gap:** Should system send update emails with new ICS?
- **Gap:** Does ICS include sequence number for updates?
- **Gap:** How are cancellations communicated via ICS?
- **Recommendation:** Define ICS update and cancellation workflow

🟡 **HIGH: ICS Compatibility Testing Not Mentioned**
- HLD claims compatibility with Google/Apple/Outlook
- **Gap:** How is compatibility verified?
- **Gap:** What if calendar client doesn't support feature?
- **Gap:** Are there fallbacks?
- **Recommendation:** Define ICS testing and compatibility requirements

🟢 **MEDIUM: ICS Attachment vs Link**
- **Gap:** Should ICS be attachment or downloadable link?
- **Gap:** What if email client blocks attachments?
- **Gap:** Should both options be available?
- **Recommendation:** Clarify ICS delivery mechanism

---

## 7. Templates & Customization

### Issues Identified

🔴 **CRITICAL: Template Security Not Addressed**
- HLD mentions "safe variable interpolation" but no details
- **Gap:** What template engine is used?
- **Gap:** How is XSS prevented?
- **Gap:** Can templates execute code?
- **Gap:** What variables are available?
- **Gap:** How are user-uploaded templates validated?
- **Recommendation:** Define template security model and validation

🟡 **HIGH: Template Versioning Missing**
- **Gap:** What happens if template is updated after invites sent?
- **Gap:** Should system keep template versions?
- **Gap:** Can admin revert to previous template?
- **Gap:** Do existing invites use old or new template?
- **Recommendation:** Define template versioning strategy

🟡 **HIGH: Image Upload Validation Missing**
- HLD mentions "upload images" but no validation
- **Gap:** What file types are allowed?
- **Gap:** What's the max file size?
- **Gap:** How are malicious files prevented?
- **Gap:** Are images scanned for malware?
- **Gap:** What if image is extremely large (10000x10000)?
- **Recommendation:** Define image upload validation and limits

🟢 **MEDIUM: Template Sharing Not Addressed**
- **Gap:** Can templates be shared between events?
- **Gap:** Are templates global or per-user?
- **Gap:** Can templates be exported/imported?
- **Recommendation:** Clarify template scope and sharing

---

## 8. Asset Storage

### Issues Identified

🔴 **CRITICAL: Storage Migration Not Addressed**
- HLD supports local FS and S3 but not migration
- **Gap:** Can storage provider be changed after deployment?
- **Gap:** How are existing assets migrated?
- **Gap:** What happens if migration fails?
- **Gap:** Is there a migration tool?
- **Recommendation:** Define storage migration strategy or explicitly exclude

🟡 **HIGH: Storage Quota Not Specified**
- **Gap:** Is there a storage limit per event/user/system?
- **Gap:** What happens when quota is reached?
- **Gap:** Can admin view storage usage?
- **Gap:** Is there automatic cleanup of old assets?
- **Recommendation:** Define storage quota and management

🟡 **HIGH: Asset Deletion Policy Missing**
- **Gap:** When are assets deleted?
- **Gap:** Are assets deleted when event is deleted?
- **Gap:** Are assets deleted when template is deleted?
- **Gap:** Is there a grace period?
- **Gap:** Can deleted assets be recovered?
- **Recommendation:** Define asset lifecycle and deletion policy

🟢 **MEDIUM: Asset Access Control Unclear**
- **Gap:** Are uploaded assets public or private?
- **Gap:** Can anyone with URL access assets?
- **Gap:** Should assets require authentication?
- **Gap:** What about assets in email templates?
- **Recommendation:** Define asset access control model

---

## 9. Database Schema

### Issues Identified

🔴 **CRITICAL: Schema Provided But Not in HLD**
- HLD states "Exact SQL previously defined and considered final" but SQL not included
- **Gap:** Cannot review schema without seeing it
- **Gap:** Cannot verify relationships and constraints
- **Gap:** Cannot assess normalization and indexes
- **Recommendation:** Include complete SQL schema in HLD or reference separate document

🔴 **CRITICAL: Data Retention Policy Missing**
- **Gap:** How long is data kept?
- **Gap:** Are old events automatically deleted?
- **Gap:** Can admin configure retention?
- **Gap:** What about GDPR/privacy compliance?
- **Gap:** Can guests request data deletion?
- **Recommendation:** Define data retention and privacy compliance requirements

🟡 **HIGH: Database Backup Strategy Missing**
- **Gap:** How is database backed up?
- **Gap:** What's the backup frequency?
- **Gap:** Where are backups stored?
- **Gap:** How is restore tested?
- **Gap:** Is backup admin's responsibility or app's?
- **Recommendation:** Define backup requirements or explicitly make it admin responsibility

🟡 **HIGH: Database Migration Strategy Unclear**
- HLD mentions "migrations/" folder but no strategy
- **Gap:** How are migrations applied?
- **Gap:** Can migrations be rolled back?
- **Gap:** What happens if migration fails?
- **Gap:** Are migrations tested?
- **Recommendation:** Define migration strategy and tooling

🟢 **MEDIUM: Database Performance Not Addressed**
- **Gap:** What's the expected data volume?
- **Gap:** Are indexes defined?
- **Gap:** What queries need optimization?
- **Gap:** Is query performance monitored?
- **Recommendation:** Define performance requirements and monitoring

---

## 10. API & Page Routes

### Issues Identified

🔴 **CRITICAL: API Versioning Not Addressed**
- HLD lists routes but no versioning strategy
- **Gap:** What happens when API changes?
- **Gap:** Are routes versioned (/v1/events)?
- **Gap:** How is backwards compatibility maintained?
- **Recommendation:** Define API versioning strategy or explicitly exclude

🔴 **CRITICAL: Error Response Format Not Specified**
- **Gap:** What format are errors returned in?
- **Gap:** Are error codes standardized?
- **Gap:** What HTTP status codes are used?
- **Gap:** Are errors user-friendly or technical?
- **Recommendation:** Define error response format and codes

🟡 **HIGH: Rate Limiting Not Specified**
- **Gap:** Are API endpoints rate limited?
- **Gap:** What's the limit per user/IP?
- **Gap:** How are rate limit violations handled?
- **Gap:** Is rate limiting per-endpoint or global?
- **Recommendation:** Define rate limiting strategy

🟡 **HIGH: Input Validation Not Specified**
- **Gap:** How are inputs validated?
- **Gap:** What's the max request size?
- **Gap:** How are malicious inputs prevented?
- **Gap:** Are there input sanitization rules?
- **Recommendation:** Define input validation requirements

🟡 **HIGH: CSRF Protection Not Mentioned**
- **Gap:** How is CSRF prevented?
- **Gap:** Are CSRF tokens required?
- **Gap:** What about API endpoints?
- **Recommendation:** Define CSRF protection strategy

🟢 **MEDIUM: API Documentation Not Addressed**
- **Gap:** Will API be documented?
- **Gap:** Is OpenAPI/Swagger used?
- **Gap:** Are examples provided?
- **Recommendation:** Clarify if API documentation is in scope

---

## 11. Deployment Model

### Issues Identified

🔴 **CRITICAL: Health Check Endpoint Missing**
- HLD mentions `/admin/health` but no details
- **Gap:** What does health check verify?
- **Gap:** Does it check database connectivity?
- **Gap:** Does it check SMTP connectivity?
- **Gap:** What's the response format?
- **Recommendation:** Define health check requirements

🔴 **CRITICAL: Configuration Validation Missing**
- HLD states "config via environment variables" but no validation
- **Gap:** What happens if required config is missing?
- **Gap:** What happens if config is invalid?
- **Gap:** Does app fail to start or log error?
- **Gap:** Can config be reloaded without restart?
- **Recommendation:** Define configuration validation and error handling

🟡 **HIGH: Logging Strategy Not Specified**
- **Gap:** What logging level is used?
- **Gap:** Where are logs written?
- **Gap:** What's the log format?
- **Gap:** Are logs structured (JSON)?
- **Gap:** How are sensitive data (tokens, passwords) handled in logs?
- **Recommendation:** Define logging strategy and sensitive data handling

🟡 **HIGH: Monitoring and Metrics Missing**
- **Gap:** What metrics are exposed?
- **Gap:** Is Prometheus/metrics endpoint available?
- **Gap:** What should be monitored?
- **Gap:** Are there alerting recommendations?
- **Recommendation:** Define monitoring and metrics requirements

🟡 **HIGH: Upgrade Strategy Not Defined**
- **Gap:** How are upgrades performed?
- **Gap:** Is zero-downtime upgrade possible?
- **Gap:** What's the rollback procedure?
- **Gap:** Are database migrations automatic?
- **Recommendation:** Define upgrade and rollback procedures

🟢 **MEDIUM: Resource Requirements Not Specified**
- HLD mentions "512MB RAM minimum" in README but not in HLD
- **Gap:** What's the expected CPU usage?
- **Gap:** What's the expected disk I/O?
- **Gap:** What's the expected network bandwidth?
- **Recommendation:** Define resource requirements and limits

---

## 12. Security Considerations

### Issues Identified

🔴 **CRITICAL: TLS/HTTPS Enforcement Not Specified**
- HLD mentions "HTTPS Required: TLS termination at reverse proxy"
- **Gap:** Does app enforce HTTPS?
- **Gap:** What happens if accessed via HTTP?
- **Gap:** Are secure cookies used?
- **Gap:** Is HSTS header set?
- **Recommendation:** Define HTTPS enforcement and secure cookie policy

🔴 **CRITICAL: Content Security Policy Not Mentioned**
- **Gap:** Is CSP header set?
- **Gap:** What's the CSP policy?
- **Gap:** How is XSS prevented?
- **Recommendation:** Define CSP and XSS prevention strategy

🟡 **HIGH: SQL Injection Prevention Not Mentioned**
- **Gap:** How is SQL injection prevented?
- **Gap:** Are prepared statements used?
- **Gap:** Is ORM used?
- **Recommendation:** Define SQL injection prevention strategy

🟡 **HIGH: Secrets Management Not Addressed**
- HLD mentions SMTP password, OIDC secret in environment variables
- **Gap:** How are secrets stored securely?
- **Gap:** Should secrets be encrypted at rest?
- **Gap:** Can secrets be rotated?
- **Gap:** Are secrets logged?
- **Recommendation:** Define secrets management strategy

🟡 **HIGH: Audit Logging Not Specified**
- **Gap:** Are admin actions logged?
- **Gap:** Are RSVP changes logged?
- **Gap:** Are login attempts logged?
- **Gap:** Can audit logs be tampered with?
- **Recommendation:** Define audit logging requirements

🟢 **MEDIUM: Security Headers Not Specified**
- **Gap:** What security headers are set?
- **Gap:** X-Frame-Options, X-Content-Type-Options?
- **Gap:** Referrer-Policy?
- **Recommendation:** Define security headers policy

---

## 13. Operational Concerns

### Issues Identified

🔴 **CRITICAL: Disaster Recovery Not Addressed**
- **Gap:** What's the recovery time objective (RTO)?
- **Gap:** What's the recovery point objective (RPO)?
- **Gap:** How is data restored after failure?
- **Gap:** Is there a DR plan?
- **Recommendation:** Define disaster recovery requirements

🟡 **HIGH: Observability Gaps**
- **Gap:** How are errors tracked?
- **Gap:** Is there error aggregation (Sentry)?
- **Gap:** Can admin see system health?
- **Gap:** Are there dashboards?
- **Recommendation:** Define observability and error tracking

🟡 **HIGH: Capacity Planning Not Addressed**
- **Gap:** What's the max number of events?
- **Gap:** What's the max number of invites per event?
- **Gap:** What's the max number of guests?
- **Gap:** What are the scaling limits?
- **Recommendation:** Define capacity limits and scaling considerations

🟢 **MEDIUM: Maintenance Mode Not Specified**
- **Gap:** Can app be put in maintenance mode?
- **Gap:** How are users notified?
- **Gap:** Can admin still access during maintenance?
- **Recommendation:** Define maintenance mode requirements

---

## 14. Edge Cases and Error Scenarios

### Critical Edge Cases Not Addressed

🔴 **CRITICAL: Concurrent Modification Conflicts**
- **Gap:** What if two admins edit same event simultaneously?
- **Gap:** What if guest submits RSVP twice (double-click)?
- **Gap:** How are race conditions prevented?
- **Recommendation:** Define concurrency control strategy

🔴 **CRITICAL: Data Consistency Scenarios**
- **Gap:** What if email queue fails mid-batch?
- **Gap:** What if database transaction fails?
- **Gap:** Are operations idempotent?
- **Recommendation:** Define transaction boundaries and consistency guarantees

🟡 **HIGH: Network Failure Scenarios**
- **Gap:** What if SMTP server is unreachable?
- **Gap:** What if OIDC provider is down?
- **Gap:** What if S3 storage is unavailable?
- **Gap:** Are there timeouts and retries?
- **Recommendation:** Define network failure handling and timeouts

🟡 **HIGH: Invalid Data Scenarios**
- **Gap:** What if event date is in the past?
- **Gap:** What if RSVP deadline is after event date?- **Gap:** What if email address is malformed?
- **Gap:** What if timezone doesn't exist?
- **Recommendation:** Define input validation rules and error messages

---

## 15. Consistency and Contradiction Analysis

### Contradictions Found

🔴 **CRITICAL: Passphrase Feature Inconsistency**
- **Location:** Section 5.1 (Invite Attributes) vs Section 14 (v0 Scope)
- **Contradiction:** Invite model includes "optional passphrase (hashed)" but v0 explicitly excludes "Passphrases"
- **Impact:** Unclear if database schema should include passphrase field
- **Recommendation:** Remove from data model if excluded, or move to v1 scope with clear migration path

🟡 **HIGH: Generic Link Terminology Confusion**
- **Location:** Section 5.1 (Invite Creation) and Section 4.2 (Event Visibility)
- **Issue:** "Generic link (public or private)" is confusing - how can a generic link be private?
- **Impact:** Implementation ambiguity
- **Recommendation:** Clarify terminology - perhaps "shareable link" vs "public event page"

🟡 **HIGH: Guest OIDC Scope Ambiguity**
- **Location:** Section 3.2 (Guest Authentication) vs Section 14 (v0 Scope)
- **Issue:** Guest OIDC mentioned as "optional" but v0 excludes "Guest OIDC"
- **Impact:** Unclear if any guest OIDC code should be written
- **Recommendation:** Clarify if basic infrastructure should exist but be disabled, or completely excluded

### Missing Cross-References

🟡 **HIGH: Template Assignment Not Linked to Events**
- Section 4.1 mentions "Assigned invite template" but Section 9 doesn't explain assignment mechanism
- **Recommendation:** Add template assignment workflow to Section 9

🟡 **HIGH: Email Queue Not Linked to Reminders**
- Section 7.3 mentions reminders but doesn't reference email queue (Section 7.4)
- **Recommendation:** Clarify how reminders use email queue

---

## 16. Missing Requirements Categories

### User Experience Requirements

🔴 **CRITICAL: Error Messages Not Specified**
- **Gap:** What error messages do users see?
- **Gap:** Are errors user-friendly or technical?
- **Gap:** Are errors internationalized?
- **Gap:** What's the error message format?
- **Recommendation:** Define error message standards and examples

🟡 **HIGH: Loading States Not Addressed**
- **Gap:** What happens during long operations?
- **Gap:** Are there loading indicators?
- **Gap:** What's the timeout for operations?
- **Recommendation:** Define loading state requirements

🟡 **HIGH: Accessibility Not Mentioned**
- **Gap:** Is WCAG compliance required?
- **Gap:** Are keyboard shortcuts supported?
- **Gap:** Is screen reader support required?
- **Recommendation:** Define accessibility requirements or explicitly exclude

🟢 **MEDIUM: Mobile Experience Not Detailed**
- README mentions "mobile-friendly" but HLD doesn't specify requirements
- **Gap:** What's the minimum supported screen size?
- **Gap:** Are touch gestures supported?
- **Gap:** Is there a mobile-specific layout?
- **Recommendation:** Define mobile requirements in HLD

### Data Validation Requirements

🔴 **CRITICAL: Validation Rules Not Centralized**
- Validation mentioned throughout but not consolidated
- **Gap:** What's the max event title length?
- **Gap:** What's the max description length?
- **Gap:** What's the max number of preference questions?
- **Gap:** What's the max number of invites per event?
- **Recommendation:** Create validation rules section in HLD

### Internationalization Requirements

🟡 **HIGH: i18n Not Addressed**
- **Gap:** Is multi-language support required?
- **Gap:** What's the default language?
- **Gap:** Can users choose language?
- **Gap:** Are emails localized?
- **Recommendation:** Define i18n requirements or explicitly exclude for v0

### Compliance Requirements

🟡 **HIGH: Privacy Compliance Not Detailed**
- HLD mentions "privacy-focused" but no specific requirements
- **Gap:** Is GDPR compliance required?
- **Gap:** Is CCPA compliance required?
- **Gap:** Are privacy policies required?
- **Gap:** Can users export their data?
- **Gap:** Can users delete their data?
- **Recommendation:** Define privacy compliance requirements

🟡 **HIGH: Email Compliance Not Addressed**
- **Gap:** Is CAN-SPAM compliance required?
- **Gap:** Are unsubscribe links required?
- **Gap:** Is sender identification required?
- **Recommendation:** Define email compliance requirements

### Performance Requirements

🟢 **MEDIUM: Response Time Not Specified**
- **Gap:** What's the acceptable page load time?
- **Gap:** What's the acceptable API response time?
- **Gap:** What's the acceptable email send time?
- **Recommendation:** Define performance SLOs

🟢 **MEDIUM: Concurrent User Support Not Specified**
- **Gap:** How many concurrent users should be supported?
- **Gap:** How many concurrent RSVP submissions?
- **Gap:** What's the expected load?
- **Recommendation:** Define concurrency requirements

---

## 17. Architecture Gaps

### Component Interaction Not Detailed

🔴 **CRITICAL: Request Flow Not Documented**
- **Gap:** What's the flow from HTTP request to response?
- **Gap:** How do middleware components interact?
- **Gap:** What's the error propagation path?
- **Recommendation:** Add request flow diagram and description

🟡 **HIGH: Background Job Processing Unclear**
- Email queue mentioned but processing mechanism not specified
- **Gap:** Is there a background worker?
- **Gap:** Is processing synchronous or asynchronous?
- **Gap:** How are jobs scheduled?
- **Recommendation:** Define background job architecture

🟡 **HIGH: Caching Strategy Not Mentioned**
- **Gap:** Is caching used?
- **Gap:** What's cached (templates, config, sessions)?
- **Gap:** What's the cache invalidation strategy?
- **Recommendation:** Define caching strategy or explicitly exclude

### State Management Not Addressed

🟡 **HIGH: Session State Storage Unclear**
- **Gap:** Where are sessions stored?
- **Gap:** In-memory, database, Redis?
- **Gap:** What happens on app restart?
- **Gap:** How are sessions shared across instances (if scaled)?
- **Recommendation:** Define session storage mechanism

🟢 **MEDIUM: Application State Not Defined**
- **Gap:** What application state exists?
- **Gap:** Is state shared across requests?
- **Gap:** How is state initialized?
- **Recommendation:** Define application state management

---

## 18. Testing and Quality Assurance

### Testing Requirements Missing

🔴 **CRITICAL: Testing Strategy Not Defined**
- README-LLM.md emphasizes TDD but HLD doesn't mention testing
- **Gap:** What types of tests are required?
- **Gap:** What's the minimum test coverage?
- **Gap:** Are integration tests required?
- **Gap:** Are E2E tests required?
- **Recommendation:** Add testing requirements to HLD

🟡 **HIGH: Test Data Management Not Addressed**
- **Gap:** How is test data created?
- **Gap:** Are there test fixtures?
- **Gap:** How is test data cleaned up?
- **Recommendation:** Define test data management strategy

### Quality Gates Not Defined

🟡 **HIGH: Acceptance Criteria Not Specified**
- Section 15 mentions "success criteria" but not detailed
- **Gap:** What must work for v0 to be complete?
- **Gap:** What's the definition of done?
- **Gap:** What's the acceptance test suite?
- **Recommendation:** Define detailed acceptance criteria

---

## 19. Documentation Gaps

### User Documentation Not Mentioned

🟡 **HIGH: User Guide Not Specified**
- **Gap:** Is user documentation required?
- **Gap:** What should be documented?
- **Gap:** Who writes documentation?
- **Recommendation:** Define documentation requirements

### API Documentation Not Mentioned

🟡 **HIGH: API Documentation Strategy Missing**
- **Gap:** How are APIs documented?
- **Gap:** Is OpenAPI spec generated?
- **Gap:** Are examples provided?
- **Recommendation:** Define API documentation strategy

### Operational Documentation Not Specified

🟡 **HIGH: Runbook Not Mentioned**
- **Gap:** Is operational runbook required?
- **Gap:** What troubleshooting guides are needed?
- **Gap:** What's the incident response process?
- **Recommendation:** Define operational documentation requirements

---

## 20. Prioritized Recommendations

### Must Address Before Implementation (CRITICAL - 25 Items)

1. **Define Bootstrap Admin Creation** - How does first admin get created?
2. **Specify Authorization Permission Matrix** - What can each role do?
3. **Define Token Lifecycle Policy** - Do tokens expire in v0?
4. **Specify Timezone Handling** - Format and conversion rules
5. **Define Event Lifecycle States** - Draft, published, cancelled, etc.
6. **Specify Public Event Access Control** - Prevent abuse
7. **Define Bulk Invite Creation** - CSV upload, error handling
8. **Specify Token Collision Handling** - Detection and retry
9. **Define RSVP State Transitions** - Change rules and deadline enforcement
10. **Specify Question Validation Rules** - Max length, allowed values
11. **Define Email Retry Policy** - Intervals, max attempts, failure handling
12. **Specify Email Rate Limiting** - Throttling strategy
13. **Include Database Schema in HLD** - Cannot review without it
14. **Define Data Retention Policy** - GDPR/privacy compliance
15. **Specify Error Response Format** - Standardized error codes
16. **Define Health Check Requirements** - What to verify
17. **Specify Configuration Validation** - Startup behavior
18. **Define TLS/HTTPS Enforcement** - Secure cookie policy
19. **Specify Content Security Policy** - XSS prevention
20. **Define Disaster Recovery Requirements** - RTO/RPO
21. **Specify Concurrent Modification Handling** - Race condition prevention
22. **Define Data Consistency Guarantees** - Transaction boundaries
23. **Specify Validation Rules** - Centralized validation section
24. **Define Request Flow** - HTTP request to response path
25. **Specify Testing Strategy** - Required test types and coverage

### Should Address in v0 (HIGH Priority - 25 Items)

26. **Clarify Forward Auth Trust Model** - Header validation
27. **Define OIDC Auto-Creation Policy** - Edge cases
28. **Specify Token Revocation Mechanics** - Guest notification
29. **Define Token Regeneration Impact** - Old token behavior
30. **Specify Event Capacity Requirements** - Total attendee limits
31. **Define Event Visibility Transitions** - Public to private rules
32. **Clarify Provisional Invite Concept** - Lifecycle and differences
33. **Define Generic Link Mechanics** - Private vs public clarification
34. **Reconsider Token Hashing Algorithm** - SHA-256 vs bcrypt
35. **Specify Token URL Format** - Complete URL structure
36. **Define +1 Validation Logic** - Error handling
37. **Specify RSVP Deadline Enforcement** - Grace period
38. **Define Question Lifecycle** - Impact on existing answers
39. **Specify Answer Editing Policy** - History tracking
40. **Define Email Bounce Handling** - Detection and notification
41. **Specify Email Unsubscribe Mechanism** - CAN-SPAM compliance
42. **Define SMTP Configuration Validation** - Test email function
43. **Specify Queue Processing Mechanism** - Background worker
44. **Define Queue Observability** - Admin dashboard
45. **Specify ICS Update Mechanism** - Event change notifications
46. **Define Template Security Model** - XSS prevention
47. **Specify Template Versioning** - Impact on existing invites
48. **Define Image Upload Validation** - File type, size limits
49. **Specify Storage Quota** - Limits and management
50. **Define Asset Deletion Policy** - Lifecycle and grace period

---

## 21. Summary and Next Steps

### Critical Findings Summary

The HLD provides a solid conceptual foundation but has **significant implementation gaps** that must be addressed:

**Authentication & Authorization:**
- Bootstrap admin creation undefined
- Permission matrix incomplete
- Forward auth trust model unclear

**Data Model:**
- Database schema not included in HLD
- Data retention policy missing
- Lifecycle states not defined

**Security:**
- Token hashing algorithm questionable (SHA-256 vs bcrypt)
- HTTPS enforcement not specified
- CSP and security headers missing
- Secrets management not addressed

**Operations:**
- Email retry and rate limiting not specified
- Queue processing mechanism unclear
- Health checks not detailed
- Disaster recovery not addressed

**Edge Cases:**
- Concurrent modification handling missing
- Network failure scenarios not addressed
- Invalid data handling incomplete
- Error response format not standardized

### Inconsistencies Found

1. **Passphrase feature** - included in data model but excluded from v0
2. **Generic link terminology** - "private generic link" is contradictory
3. **Guest OIDC** - mentioned as optional but excluded from v0

### Recommended Actions

**Immediate (Before Implementation):**
1. Address all 25 CRITICAL items listed in Section 20
2. Resolve 3 contradictions in Section 15
3. Include complete database schema in HLD
4. Define validation rules section
5. Add request flow diagram

**Short-term (During v0 Development):**
1. Address HIGH priority items (26-50)
2. Create detailed API documentation
3. Define testing strategy
4. Document operational procedures

**Long-term (Post-v0):**
1. Address MEDIUM/LOW priority items
2. Consider multi-tenancy
3. Plan for internationalization
4. Define scaling strategy

### HLD Revision Recommendations

**Structural Improvements:**
1. Add "Validation Rules" section consolidating all validation requirements
2. Add "Error Handling" section with standardized error codes
3. Add "Security" section consolidating security requirements
4. Add "Operations" section with monitoring, logging, backup requirements
5. Include complete database schema with relationships diagram
6. Add request flow diagrams for key workflows

**Content Additions:**
1. Define explicit permission matrix for roles
2. Specify all lifecycle states and transitions
3. Define all edge case handling
4. Specify all error scenarios and responses
5. Define all operational requirements

**Clarity Improvements:**
1. Resolve terminology inconsistencies
2. Remove contradictions between sections
3. Add cross-references between related sections
4. Provide concrete examples for abstract concepts

---

## 22. Conclusion

The TinyRSVP HLD demonstrates **good high-level thinking** but requires **substantial elaboration** before implementation can begin safely. The document successfully defines the product vision and core concepts but lacks the operational detail, edge case handling, and security specifications needed for production deployment.

**Recommendation:** **Do not proceed with implementation** until at minimum the 25 CRITICAL items are addressed. The current HLD would lead to significant rework, security vulnerabilities, and operational issues.

**Estimated Effort to Address Findings:**
- CRITICAL items: 3-5 days of design work
- HIGH priority items: 2-3 days of design work
- HLD revision and documentation: 1-2 days

**Total:** Approximately 1-2 weeks of design work before implementation should begin.

---

**Review Completed:** 2026-01-06  
**Reviewer:** AI Assistant  
**Next Review:** After HLD revision addressing critical items
