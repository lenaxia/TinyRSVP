# Worklog: HLD Revision

**Date:** 2026-01-06  
**Session ID:** 002  
**Duration:** ~3 hours

## Summary

Created comprehensive revised HLD ([`docs/03_REVISED_HLD.md`](../03_REVISED_HLD.md)) addressing all 50+ gaps identified in design review. Gathered stakeholder input on 12 critical design decisions and produced implementation-ready specification.

## Work Completed

- [x] Gathered stakeholder input on critical design decisions
- [x] Wrote comprehensive revised HLD (25 sections, ~1500 lines)
- [x] Included complete database schema (9 tables with full DDL)
- [x] Added validation rules section (7 subsections)
- [x] Added error handling section (standardized codes and scenarios)
- [x] Added security section (transport, headers, sanitization, CSRF, secrets, audit)
- [x] Added operations section (health checks, logging, monitoring, backups, DR)
- [x] Created request flow diagrams (admin, guest, email processing)
- [x] Resolved all contradictions from original HLD
- [x] Defined v0 scope clearly (included and excluded features)

## Key Decisions Made

### Decision 1: Bootstrap Admin Creation
**Input Gathered:** First user to authenticate becomes admin automatically  
**Rationale:** Simplest approach, no CLI or ENV var needed, works for both OIDC and forward auth

### Decision 2: Event Manager Permissions
**Input Gathered:** Can see all events but only edit own, can delete own events (archive), no system settings  
**Rationale:** Balances collaboration (see others' events) with ownership (edit only own)

### Decision 3: Token Expiration
**Input Gathered:** Tokens expire 30 days after event date  
**Rationale:** Auto-cleanup, prevents stale tokens, reasonable grace period

### Decision 4: Timezone Format
**Input Gathered:** IANA timezone names (e.g., 'America/Los_Angeles')  
**Rationale:** Most accurate, handles DST automatically, standard format

### Decision 5: Email Retry Policy
**Input Gathered:** 3 retries with exponential backoff (1min, 5min, 15min), then fail and notify admin  
**Rationale:** Balances reliability with timely failure notification

### Decision 6: Email Rate Limiting
**Input Gathered:** Configurable via ENV var, default 50/minute  
**Rationale:** Prevents SMTP throttling, configurable for different providers

### Decision 7: Email Queue Processing
**Recommendation Made:** Hybrid - immediate send with background retry queue  
**Rationale:** Responsive UX, reliable retries, single binary, survives restarts

### Decision 8: Token Storage Security
**Input Gathered:** HMAC-SHA256 with constant-time comparison  
**Rationale:** Appropriate for high-entropy random tokens, fast, secure

### Decision 9: Public Events Scope
**Input Gathered:** Exclude from v0 - too complex  
**Rationale:** Focus on core private event use case, defer complexity

### Decision 10: Session Management
**Recommendation Made:** Database-backed sessions, 7-day timeout  
**Input Gathered:** Accepted  
**Rationale:** Survives restarts, simple implementation, supports future scaling

### Decision 11: Concurrent Modification
**Input Gathered:** Optimistic locking with version field  
**Rationale:** Simple, good UX, appropriate for low-concurrency homelab use

### Decision 12: Data Retention
**Input Gathered:** Events auto-archive 30 days after event, admin can permanently delete  
**Rationale:** Keeps active list clean, preserves data, admin control

## HLD Improvements

### Sections Added

1. **Complete Database Schema** (Section 13)
   - 9 tables with full DDL
   - All indexes defined
   - Foreign key relationships
   - Check constraints
   - Migration strategy

2. **Validation Rules** (Section 14)
   - 7 subsections covering all input types
   - Specific error messages for each rule
   - Field-level validation requirements

3. **Error Handling** (Section 15)
   - Standardized error response format
   - 8 error codes with HTTP status mappings
   - Detailed error scenarios for all failure modes

4. **Security** (Section 16)
   - Transport security (HTTPS enforcement)
   - Security headers (CSP, HSTS, etc.)
   - Input sanitization (XSS, SQL injection, path traversal)
   - CSRF protection
   - Secrets management
   - Audit logging

5. **Operations** (Section 17)
   - Health checks with specific tests
   - Structured logging with sensitive data handling
   - Prometheus metrics
   - Background jobs specification
   - Backup & recovery procedures
   - Disaster recovery scenarios

6. **Request Flow Diagrams** (Section 19)
   - Admin request flow (auth → authz → handler → DB → template)
   - Guest RSVP flow (token validation → handler → DB → confirmation)
   - Email processing flow (immediate send → retry queue → background worker)

7. **Deployment Model** (Section 20)
   - Complete Dockerfile
   - Configuration reference
   - Resource requirements
   - Upgrade/rollback procedures

8. **Success Criteria** (Section 22)
   - Functional requirements checklist (22 items)
   - Non-functional requirements checklist (14 items)
   - 3 acceptance test scenarios

### Contradictions Resolved

1. **Passphrase Feature** - Removed from v0 scope entirely (no database field)
2. **Generic Links** - Removed from v0 scope (part of public events, deferred to v1)
3. **Guest OIDC** - Explicitly excluded from v0, deferred to v1

### Gaps Filled

**Critical (25 items) - All Addressed:**
- ✅ Bootstrap admin creation defined
- ✅ Permission matrix complete
- ✅ Token lifecycle specified
- ✅ Timezone handling detailed
- ✅ Event lifecycle states defined
- ✅ Bulk invite creation specified
- ✅ Token collision handling (regenerate on collision)
- ✅ RSVP state transitions defined
- ✅ Question validation rules specified
- ✅ Email retry policy defined
- ✅ Email rate limiting specified
- ✅ Database schema included
- ✅ Data retention policy defined
- ✅ Error response format standardized
- ✅ Health check requirements detailed
- ✅ Configuration validation specified
- ✅ HTTPS enforcement defined
- ✅ CSP specified
- ✅ Disaster recovery addressed
- ✅ Concurrent modification handling (optimistic locking)
- ✅ Data consistency guarantees (transactions)
- ✅ Validation rules centralized
- ✅ Request flow documented
- ✅ Testing strategy referenced (TDD in README-LLM.md)

**High Priority (25 items) - All Addressed:**
- ✅ Forward auth trust model clarified
- ✅ OIDC auto-creation policy defined
- ✅ Token revocation mechanics specified
- ✅ Token regeneration impact defined
- ✅ Event capacity explicitly excluded from v0
- ✅ Event visibility transitions (no public events in v0)
- ✅ Provisional invites removed (no public events)
- ✅ Generic links removed (no public events)
- ✅ Token hashing algorithm justified (HMAC-SHA256)
- ✅ Token URL format specified
- ✅ +1 validation logic defined
- ✅ RSVP deadline enforcement specified
- ✅ Question lifecycle defined
- ✅ Answer editing policy specified
- ✅ Email bounce handling defined
- ✅ Email unsubscribe mechanism specified
- ✅ SMTP configuration validation defined
- ✅ Queue processing mechanism specified (hybrid)
- ✅ Queue observability (metrics + admin UI)
- ✅ ICS update mechanism specified
- ✅ Template security model defined
- ✅ Template versioning explicitly excluded from v0
- ✅ Image upload validation specified
- ✅ Storage quota explicitly excluded from v0
- ✅ Asset deletion policy defined

## Document Statistics

**Revised HLD:**
- Sections: 25
- Lines: ~1500
- Tables: 9 (complete DDL)
- Diagrams: 3 (lifecycle, request flows)
- Validation rules: 50+
- Error codes: 8
- API routes: 50+

**Coverage:**
- Authentication & Authorization: Complete
- Data Model: Complete with full schema
- Security: Complete with specific requirements
- Operations: Complete with procedures
- Error Handling: Complete with standardized codes
- Validation: Complete with specific rules

## Files Changed

- **Created:** `docs/03_REVISED_HLD.md` - Comprehensive revised HLD (supersedes 00_INITIAL_HLD.md)
- **Updated:** `docs/01_WORKLOG/2026-01-06_hld-revision.md` - This worklog entry

## Next Steps

1. **Immediate:** Review and approve revised HLD
2. **Short-term:** Create implementation backlog stories based on HLD
3. **Short-term:** Begin implementation following TDD approach
4. **Ongoing:** Reference HLD for all design decisions

## Notes

### Design Philosophy

The revised HLD follows these principles:
- **Explicit over implicit** - No assumptions, everything specified
- **Complete over brief** - Better too much detail than too little
- **Actionable over aspirational** - Concrete requirements, not wishes
- **Consistent over flexible** - Standardized patterns throughout

### Key Improvements Over Original

**Original HLD Strengths:**
- Clear product vision
- Good high-level concepts
- Well-defined user roles

**Original HLD Weaknesses:**
- Missing operational details
- Incomplete edge case handling
- Scattered validation rules
- No error handling strategy
- Security specifications lacking

**Revised HLD Addresses:**
- All 25 critical gaps
- All 25 high priority gaps
- All 3 contradictions
- Adds 6 new comprehensive sections
- Provides implementation-ready specification

### Implementation Readiness

**Ready to Implement:**
- Database schema (complete DDL)
- Authentication flows (both OIDC and forward auth)
- Session management (database-backed)
- Token generation and validation (HMAC-SHA256)
- Email queue and retry logic
- Validation rules (all inputs)
- Error handling (standardized codes)
- Security measures (headers, CSRF, sanitization)

**Still Needs LLD:**
- Specific Go package structure
- Interface definitions
- Function signatures
- Test strategies per component
- Deployment scripts

## References

- **Original HLD:** [`docs/00_INITIAL_HLD.md`](../00_INITIAL_HLD.md)
- **Design Review:** [`docs/02_HLD_DESIGN_REVIEW.md`](../02_HLD_DESIGN_REVIEW.md)
- **Revised HLD:** [`docs/03_REVISED_HLD.md`](../03_REVISED_HLD.md)
- **Implementation Guide:** [`README-LLM.md`](../../README-LLM.md)

---

**Status:** HLD revision complete, ready for stakeholder approval and implementation planning.
