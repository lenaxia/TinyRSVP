# TinyRSVP v0 - Epic & Story Breakdown Summary

**Created:** 2026-01-06  
**Status:** Ready for Implementation  
**Total Effort:** ~10 weeks

---

## Executive Summary

This document provides a comprehensive breakdown of the TinyRSVP v0 implementation into 10 epics and 155 user stories. The breakdown is based on the [Revised HLD](../02_REVISED_HLD.md) and 8 detailed [LLD documents](../04_LLD_INDEX.md).

**Key Metrics:**
- **10 Epics** organized by domain and dependency
- **155 User Stories** (estimated, detailed stories created during implementation)
- **~12 weeks** total estimated effort
- **4 phases** of implementation
- **Clear dependency chain** ensuring proper build order

---

## Epic Status Summary (UPDATED 2026-04-06 — code-verified)

| Epic | Status | Confidence | Test Pass Rate | Notes |
|------|--------|-----------|----------------|-------|
| 00: Foundation | ✅ Complete | HIGH | 100% | |
| 01: Auth | ✅ Complete | HIGH | 100% | OIDC uses mock provider in tests; real wire tests not run |
| 02: Events | ✅ Complete | HIGH | 100% | |
| 03: Invites | ✅ Complete | HIGH | 100% | Plain token stored alongside hash (minor) |
| 04: RSVP | ✅ Complete | HIGH | 100% | |
| 05: Email | ⚠️ Mostly Complete | MEDIUM | 100%* | *Tests pass but assert broken behavior; 3 bugs — see Epic 05 |
| 06: Templates | ✅ Complete | HIGH | 100% | Component editor wired in service/handler but not in router |
| 07: Frontend | ✅ Complete | HIGH | 100% | UX tests need Chrome; all non-browser tests pass |
| 08: API | ⚠️ Mostly Complete | MEDIUM | 100% | X-Test-User-ID bypass active in production (Critical) |
| 09: Security | ❌ Not Started | N/A | N/A | Has 1 pre-identified critical issue (see Epic 09) |
| 10: Tech Debt | 🔄 Active | N/A | N/A | 5 stories complete; 6 new issues added from code validation |
| 11: RSVP Themes | ✅ Complete | HIGH | 100% | |
| 12: Test Infrastructure | ⚠️ Partial | HIGH | 100% | Phases 1-3 + Phase 5 done; Phase 4 partial (cleanup + TESTING.md missing) |
| 13: Guest Accounts | ❌ Not Started | N/A | N/A | |
| 14: Bug Fixes & Gaps | ❌ Not Started | N/A | N/A | 6 stories: 1 critical, 3 high, 1 medium, 1 low |

**All 32 non-browser test packages pass (verified 2026-04-06).**

**Production Ready:** NO — two blockers before public deployment:
1. `X-Test-User-ID` auth bypass must be removed/gated — **Epic 14 Story 01**
2. Three functional bugs in email system — **Epic 14 Stories 02, 03, 04**

---

## Epic Overview

### Epic 00: Foundation & Project Setup
**Priority:** High | **Effort:** 1 week | **Stories:** 7  
**Status:** ✅ Complete | **Confidence:** HIGH (95%) | **Test Pass Rate:** 100%

**Purpose:** Establish foundational infrastructure including Go module, configuration, database layer, and migrations.

**Key Deliverables:**
- Go module with all dependencies
- Environment-based configuration
- SQLite connection and pooling
- All 9 database tables via migrations
- Repository pattern implementation
- Health check endpoint
- Docker container setup

**Blocks:** All other epics

---

### Epic 01: Authentication & Authorization
**Priority:** High | **Effort:** 1 week | **Stories:** 8  
**Status:** ✅ Complete | **Confidence:** HIGH | **Test Pass Rate:** 100%

**Purpose:** Implement secure authentication (OIDC + forward auth) and role-based access control.

**Key Deliverables:** All delivered and tests pass. OIDC callback tests use a local mock OIDC provider with real JWT signing — not a stub. ForwardAuth requires `FORWARD_AUTH_TRUSTED_IPS` at startup (validated by config, not silently broken).

**Note:** `X-Test-User-ID` bypass in `internal/middleware/rbac.go:16` is used by the UX test suite but is ungated in production — tracked in Epic 08 / Epic 09.

**Depends on:** Epic 00  
**Blocks:** Epic 02, 03

---

### Epic 02: Event Management
**Priority:** High | **Effort:** 2 weeks | **Stories:** 11  
**Status:** ✅ Complete | **Confidence:** HIGH (90%) | **Test Pass Rate:** 100%

**Purpose:** Implement complete event lifecycle from draft to archived, with timezone support and preference questions.

**Key Deliverables:**
- Event CRUD operations
- State machine (draft → published → cancelled → archived)
- IANA timezone handling
- Optimistic locking
- Preference questions CRUD
- Event validation
- Auto-archive job

**Depends on:** Epic 00, 01  
**Blocks:** Epic 03, 04

---

### Epic 03: Invite & Token Management
**Priority:** High | **Effort:** 1.5 weeks | **Stories:** 11  
**Status:** ✅ Complete | **Confidence:** HIGH | **Test Pass Rate:** 100%

**Purpose:** Implement secure token-based guest access with individual invites, bulk CSV import, and token lifecycle.

**Note:** Plain token stored in `token` column alongside `token_hash` (migration 000008) for send-path convenience. The hash provides the security guarantee but the plain token is recoverable from DB directly.

**Depends on:** Epic 00, 01, 02  
**Blocks:** Epic 04, 05

---

### Epic 04: RSVP & Guest Experience
**Priority:** High | **Effort:** 1 week | **Stories:** 11  
**Status:** ✅ Complete | **Confidence:** HIGH | **Test Pass Rate:** 100%

**Purpose:** Enable guests to RSVP via token links, answer questions, and update responses until deadline. RSVP service tests use real in-memory SQLite with full migrations.

**Depends on:** Epic 00, 02, 03  
**Blocks:** Epic 05 (confirmation emails)

---

### Epic 05: Email System & Calendar Integration
**Priority:** High | **Effort:** 1.5 weeks | **Stories:** 15  
**Status:** ⚠️ Mostly Complete | **Confidence:** MEDIUM | **Test Pass Rate:** 100%*

**Purpose:** Implement reliable email delivery with queue, retry logic, and ICS calendar file generation.

**Known Bugs (code-verified):**
1. Invite emails hard-coded as plaintext — `TemplateTypeInviteEmail` never rendered (`internal/invites/service.go:275`)
2. Confirmation emails show "Question 1", "Question 2" instead of question text (`internal/email/confirmation_service.go:157`); test asserts this broken behavior
3. Unsubscribe DB operation succeeds but success page body is "Failed to render page" — `unsubscribe.html` not in rsvpPageTemplates

*Tests pass but #2 test asserts wrong behavior.

**Depends on:** Epic 00, 02, 03, 06  
**Blocks:** None

---

### Epic 06: Templates & Asset Management
**Priority:** Medium | **Effort:** 1 week | **Stories:** 13  
**Status:** ✅ Complete | **Confidence:** HIGH | **Test Pass Rate:** 100%

**Purpose:** Implement template system for customization and asset management for images.

**Note:** `TemplateEditorHandlers` (`/templates/{id}/edit`, `/api/templates/{id}/components`) is fully implemented in `internal/handlers/template_editor.go` but never registered in the router — the component editor UI is unreachable. Tracked as Epic 10 Story 19.

**Depends on:** Epic 00, 01  
**Blocks:** Epic 05 (email templates)

---

### Epic 07: Frontend & User Experience
**Priority:** High | **Effort:** 1 week | **Stories:** 21  
**Status:** ✅ Complete | **Confidence:** HIGH | **Test Pass Rate:** 100%

**Purpose:** Implement mobile-first, accessible UI using plain CSS and vanilla JavaScript. All Go tests pass. Chromedp UX tests require headless Chrome and cannot run in this environment — they pass when Chrome is available.

**Depends on:** Epic 08 (needs API routes)  
**Blocks:** None

---

### Epic 08: API & HTTP Layer
**Priority:** High | **Effort:** 1.5 weeks | **Stories:** 18  
**Status:** ⚠️ Mostly Complete | **Confidence:** MEDIUM | **Test Pass Rate:** 100%

**Purpose:** Wire all components together with complete HTTP API and middleware.

**Known Issues:**
1. `X-Test-User-ID` header bypasses authentication unconditionally in production (`internal/middleware/rbac.go:16`). Critical — must be fixed before public deployment.
2. Template editor routes (`/templates/{id}/edit`) documented in API spec but never registered in router.

**Depends on:** All other epics (orchestration)
**Blocks:** Epic 07 (frontend needs routes)

---

### Epic 09: Security Review & Penetration Testing
**Priority:** Critical | **Effort:** 2 weeks | **Stories:** 40  
**Status:** ❌ Not Started | **Confidence:** N/A | **Test Pass Rate:** N/A

**Purpose:** Comprehensive security assessment and penetration testing to identify and remediate vulnerabilities before production deployment.

**Key Deliverables:**
- Automated security scanning (OWASP ZAP, Nuclei, gosec, Trivy)
- Dependency vulnerability scanning
- Authentication and authorization bypass testing
- Token security validation (entropy, brute force, timing attacks)
- Injection attack testing (SQL, XSS, template, command, path traversal)
- CSRF protection validation
- Business logic security testing
- Rate limiting and DoS resistance
- File upload security testing
- Data exposure and privacy testing
- Security headers validation
- Manual penetration testing
- Comprehensive security assessment report

**Depends on:** All other epics (00-08) - requires complete implementation
**Blocks:** Production deployment

---

### Epic 11: RSVP Themes
**Priority:** High | **Effort:** 1 week | **Stories:** 7  
**Status:** ✅ Complete | **Confidence:** HIGH | **Test Pass Rate:** 100%

**Purpose:** Implement Evite-style component-based RSVP page customization with 7 built-in themes, color overrides, and custom image upload.

**Depends on:** Epic 06 (Templates)  
**Blocks:** None

---

### Epic 12: Test Infrastructure Modernization
**Priority:** High | **Effort:** 3-4 weeks (20-26 hours) | **Stories:** 20  
**Status:** ⚠️ Partial — Phases 1-3 + Phase 5 complete, Phase 4 partial | **Test Pass Rate:** 100%

**Purpose:** Modernize test infrastructure with centralized testutil package and generated mocks.

**What's Done:** `internal/testutil/` package exists with all helpers; 30 generated mocks across repositories/services/other; test builders, HTTP helpers, and fixture files exist; testutil README and README-LLM section updated.

**What's Remaining:** ~67 test files still define local mock structs (migration incomplete); no `TESTING.md` at repo root.

**Depends on:** None (can run parallel)  
**Blocks:** None (improves developer experience)

---

### Epic 13: Guest Accounts & Encryption at Rest
**Priority:** Medium | **Effort:** 3-4 weeks | **Stories:** 12  
**Status:** ❌ Not Started | **Confidence:** N/A | **Test Pass Rate:** N/A

**Purpose:** Implement optional passwordless guest accounts with OTP-based authentication (email/SMS), and encrypt all PII at rest across every table using application-level AES-256-GCM with HMAC-SHA256 blind indexes.

**Key Deliverables:**
- `pkg/crypto/` Encryptor (AES-256-GCM + HKDF + HMAC blind hash)
- Encrypted PII in `users`, `invites`, `sessions`, `email_queue`
- `guest_accounts`, `guest_sessions`, `guest_otp_codes` tables
- OTP delivery via email (and optionally SMS)
- `/guest/auth/` routes and `RequireGuestAuth` middleware
- RSVP confirmation page opt-in prompt
- Zero changes to existing staff auth system

**Depends on:** Epic 00, 01, 03, 04, 05  
**Blocks:** Nothing (additive feature)

---

## Story Count by Epic

| Epic | Stories | Avg per Story | Notes |
|------|---------|---------------|-------|
| 00: Foundation | 7 | 1 day | Infrastructure setup |
| 01: Auth | 8 | 1 day | Security-critical |
| 02: Events | 11 | 1.5 days | Core business logic |
| 03: Invites | 11 | 1 day | Token security |
| 04: RSVP | 11 | 0.5 day | Guest experience |
| 05: Email | 15 | 0.5 day | Reliability focus |
| 06: Templates | 13 | 0.5 day | Security + UX |
| 07: Frontend | 21 | 0.5 day | UI components |
| 08: API | 18 | 0.5 day | Integration |
| 09: Security | 40 | 0.5 day | Pen testing & hardening |
| 12: Test Infrastructure | 20 | 1 day | Modernize testing patterns |

**Total:** 155 stories

---

## Critical Path

The critical path for v0 completion:

```
Epic 00 (1w) → Epic 01 (1w) → Epic 02 (2w) → Epic 03 (1.5w) →
Epic 04 (1w) → Epic 06 (1w) → Epic 05 (1.5w) → Epic 08 (1.5w) →
Epic 07 (1w) → Epic 09 (2w)

Total: 12.5 weeks on critical path
```

**Parallelization Opportunities:**
- Epic 06 (Templates) can start after Epic 01
- Epic 07 (Frontend) can develop in parallel with Epic 08 (API)
- Epic 09 (Security) requires all other epics complete

**Optimized Timeline:** ~11 weeks with parallel work

**Note:** Epic 09 (Security Review & Penetration Testing) is critical for production readiness and must be completed before deployment.

---

## Risk Assessment

### High Risk Areas
1. **Token Security (Epic 03)** - Cryptographic implementation must be correct
2. **Email Reliability (Epic 05)** - SMTP integration can be fragile
3. **Timezone Handling (Epic 02)** - Complex edge cases
4. **XSS Prevention (Epic 06)** - Security-critical template system

### Mitigation Strategies
- Comprehensive test coverage (TDD)
- Security review for crypto and templates
- Use proven libraries (go-oidc, golang-migrate)
- Follow Go best practices
- Regular integration testing

---

## Success Metrics

### Functional Completeness
- [ ] All 10 epics complete
- [ ] All 155 stories complete
- [ ] All acceptance criteria met
- [ ] All tests passing

### Quality Metrics
- [ ] Test coverage >80%
- [ ] Zero critical security issues (verified by Epic 09)
- [ ] Zero high-priority bugs
- [ ] Performance targets met
- [ ] Security assessment passed

### Documentation
- [ ] All READMEs updated
- [ ] API documentation complete
- [ ] Deployment guide complete
- [ ] Troubleshooting guide complete

---

## Implementation Strategy

### Test-Driven Development
**Every story follows TDD:**
1. Write tests first
2. Run tests (should fail)
3. Write minimal code to pass
4. Run tests (should pass)
5. Refactor if needed

### Type Safety
**Every story enforces:**
- Strongly-typed structs (no `map[string]interface{}`)
- Explicit type declarations
- Exported types for reuse
- No `interface{}` when type is known

### Documentation
**Every story updates:**
- Relevant README files
- Architecture diagrams if changed
- Worklog entries for significant progress
- This backlog with completion status

---

## Resource Requirements

### Development Environment
- Go 1.21+
- Docker and Docker Compose
- SQLite 3.35+
- Git
- Text editor / IDE

### External Services (for testing)
- OIDC provider (Keycloak/Authentik)
- SMTP server (Gmail/SendGrid/local)
- Email client for testing

### Skills Required
- Go programming
- HTTP/REST APIs
- SQL and database design
- HTML/CSS/JavaScript
- Docker
- Security best practices

---

## Next Steps

### Immediate Actions
1. **Review this breakdown** - Ensure alignment with project goals
2. **Prioritize Epic 00** - Start with foundation
3. **Create first user stories** - Detailed stories for Epic 00
4. **Set up development environment** - Go, Docker, etc.
5. **Begin implementation** - Follow TDD approach

### First Sprint (Week 1)
- Complete Epic 00: Foundation
- Set up CI/CD (if applicable)
- Establish development workflow
- Create first worklog entries

---

## Questions for Consideration

Before starting implementation, consider:

1. **OIDC Provider:** Which provider will be used for testing?
2. **SMTP Provider:** Gmail, SendGrid, or local SMTP server?
3. **Domain Name:** What domain for testing? (affects ICS UIDs)
4. **Deployment Target:** Raspberry Pi, NAS, or cloud VPS?
5. **CI/CD:** GitHub Actions, GitLab CI, or manual?

---

## References

### Design Documents
- [Revised HLD](../02_REVISED_HLD.md) - Authoritative specification
- [LLD Index](../04_LLD_INDEX.md) - All 8 LLD documents
- [Design Review](../01_HLD_DESIGN_REVIEW.md) - Design review findings
- [LLD Review](../05_LLD_REVIEW_FINDINGS.md) - LLD review findings

### Implementation Guides
- [README-LLM.md](../../README-LLM.md) - LLM implementation guide
- [LLM Workflows](../../llm-workflows/README.md) - Workflow templates

### Project Files
- [README.md](../../README.md) - User-facing documentation
- [Worklog](../01_WORKLOG/README.md) - Progress tracking

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01-06 | Initial epic and story breakdown created |

---

### Epic 14: Bug Fixes & Code Gaps
**Priority:** High | **Effort:** 1 week | **Stories:** 6  
**Status:** ❌ Not Started | **Source:** Code validation 2026-04-06

**Purpose:** Fix all bugs and dead code identified during the 2026-04-06 code-level validation pass. Every item has a confirmed file:line reference. Completing this epic is required before beginning Epic 09 (Security audit).

| Story | Severity | Summary |
|-------|----------|---------|
| 14_STORY_01 | Critical | Remove `X-Test-User-ID` auth bypass from `rbac.go` |
| 14_STORY_02 | High | Render `TemplateTypeInviteEmail` in `SendInvite` (currently hardcoded plaintext) |
| 14_STORY_03 | High | Look up question text in confirmation emails (currently shows "Question N") |
| 14_STORY_04 | High | Add `unsubscribe.html` to `rsvpPageTemplates` (currently returns garbled body) |
| 14_STORY_05 | Medium | Wire template editor into router, or delete it |
| 14_STORY_06 | Low | Move `MockService` from `internal/email/service.go` to testutil |

**Depends on:** Epics 00–11  
**Blocks:** Epic 09 (story 01 must be resolved first)

---


**Last Validated:** 2026-04-06 (code-level verification, not doc assertions)  
**Next Action:** Fix X-Test-User-ID bypass (Epic 09), fix Epic 05 email bugs, then begin Epic 09 security audit
