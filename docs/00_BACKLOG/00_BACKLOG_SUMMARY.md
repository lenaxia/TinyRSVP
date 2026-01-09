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

## Epic Overview

### Epic 00: Foundation & Project Setup
**Priority:** High | **Effort:** 1 week | **Stories:** 7

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

**Purpose:** Implement secure authentication (OIDC + forward auth) and role-based access control.

**Key Deliverables:**
- OIDC authentication flow
- Forward auth integration
- Database-backed sessions
- User model and repository
- Bootstrap admin creation
- RBAC middleware
- Permission checking service

**Depends on:** Epic 00  
**Blocks:** Epic 02, 03

---

### Epic 02: Event Management
**Priority:** High | **Effort:** 2 weeks | **Stories:** 11

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

**Purpose:** Implement secure token-based guest access with individual invites, bulk CSV import, and token lifecycle.

**Key Deliverables:**
- Cryptographically secure token generation (256-bit)
- HMAC-SHA256 token hashing
- Constant-time token validation
- Individual invite creation
- Bulk CSV import (500 guests)
- Token expiration and cleanup
- Token revocation and regeneration
- Invite status tracking

**Depends on:** Epic 00, 01, 02  
**Blocks:** Epic 04, 05

---

### Epic 04: RSVP & Guest Experience
**Priority:** High | **Effort:** 1 week | **Stories:** 11

**Purpose:** Enable guests to RSVP via token links, answer questions, and update responses until deadline.

**Key Deliverables:**
- RSVP submission (yes/no/maybe)
- Plus ones validation
- Preference question answering
- RSVP updates
- Deadline enforcement
- Confirmation page
- Confirmation email
- Mobile-responsive RSVP page

**Depends on:** Epic 00, 02, 03  
**Blocks:** Epic 05 (confirmation emails)

---

### Epic 05: Email System & Calendar Integration
**Priority:** High | **Effort:** 1.5 weeks | **Stories:** 15

**Purpose:** Implement reliable email delivery with queue, retry logic, and ICS calendar file generation.

**Key Deliverables:**
- SMTP configuration and validation
- Email queue with database backing
- Hybrid send strategy (immediate + retry)
- Exponential backoff retry (4 attempts)
- Rate limiting (50/minute)
- ICS file generation (RFC 5545)
- ICS updates with SEQUENCE
- Bounce handling
- Unsubscribe mechanism
- All email types (invite, confirmation, update, cancellation)

**Depends on:** Epic 00, 02, 03, 06  
**Blocks:** None

---

### Epic 06: Templates & Asset Management
**Priority:** Medium | **Effort:** 1 week | **Stories:** 13

**Purpose:** Implement template system for customization and asset management for images.

**Key Deliverables:**
- Go html/template integration
- Template CRUD operations
- Default templates (3 types)
- Template validation (XSS prevention)
- Template variable system
- Image upload with validation
- Storage provider interface
- Local filesystem implementation
- CSS sanitization
- Asset serving

**Depends on:** Epic 00, 01  
**Blocks:** Epic 05 (email templates)

---

### Epic 07: Frontend & User Experience
**Priority:** High | **Effort:** 1 week | **Stories:** 21

**Purpose:** Implement mobile-first, accessible UI using plain CSS and vanilla JavaScript.

**Key Deliverables:**
- CSS design system (variables, typography, colors)
- Responsive grid layout
- Admin dashboard UI
- Event management UI
- Invite management UI
- Guest RSVP page UI
- Confirmation page UI
- Form validation (client-side)
- Loading states
- Error display
- Keyboard navigation
- Screen reader support
- WCAG 2.1 AA compliance

**Depends on:** Epic 08 (needs API routes)  
**Blocks:** None

---

### Epic 08: API & HTTP Layer
**Priority:** High | **Effort:** 1.5 weeks | **Stories:** 18

**Purpose:** Wire all components together with complete HTTP API and middleware.

**Key Deliverables:**
- HTTP router with 50+ routes
- Middleware chain (8 layers)
- CSRF protection
- Security headers (CSP, HSTS, etc.)
- Rate limiting per IP
- Input validation and sanitization
- Error response formatting
- Health check endpoint
- Metrics endpoint (Prometheus)
- Static asset serving

**Depends on:** All other epics (orchestration)
**Blocks:** Epic 07 (frontend needs routes)

---

### Epic 09: Security Review & Penetration Testing
**Priority:** Critical | **Effort:** 2 weeks | **Stories:** 40

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

**Status:** ✅ Ready for Implementation  
**Next Action:** Begin Epic 00 (Foundation)
