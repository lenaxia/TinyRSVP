# TinyRSVP - Low-Level Design Index

**Version:** 1.0  
**Date:** 2026-01-06  
**Status:** Active

---

## Overview

This document serves as the index and navigation guide for all Low-Level Design (LLD) documents in the TinyRSVP project. Each LLD document provides implementation-ready specifications for a specific domain, including Go package structure, interfaces, types, and function signatures.

**Parent Document:** [High-Level Design (Revised)](02_REVISED_HLD.md)

---

## LLD Documents

### Foundation Layer

#### Domain 7: Database & Persistence
**File:** [`lld/07_DATABASE_LLD.md`](lld/07_DATABASE_LLD.md)
**Status:** ✅ Complete
**HLD Reference:** Section 13 (Database Schema)

**Scope:**
- Database connection management (SQLite, future PostgreSQL)
- Repository pattern implementation
- Transaction management and error handling
- Migration strategy using golang-migrate
- Query optimization and indexing
- Audit logging implementation

**Key Interfaces:**
- `Database` - Database connection interface
- `Repository` - Base repository interface
- `Transactor` - Transaction management
- `Migrator` - Migration execution

**Dependencies:** None (foundation layer)

---

### Core Business Logic Layer

#### Domain 1: Authentication & Authorization
**File:** [`lld/01_AUTH_LLD.md`](lld/01_AUTH_LLD.md)
**Status:** ✅ Complete
**HLD Reference:** Section 4 (Authentication & Authorization)

**Scope:**
- OIDC authentication flow
- Forward auth integration
- Database-backed session management
- User management and role assignment
- RBAC middleware
- Bootstrap admin creation

**Key Interfaces:**
- `Authenticator` - Authentication provider
- `SessionStore` - Session storage
- `AuthorizationChecker` - Permission checking

**Dependencies:**
- Domain 7 (Database)

---

#### Domain 2: Event Management
**File:** [`lld/02_EVENT_LLD.md`](lld/02_EVENT_LLD.md)
**Status:** ✅ Complete
**HLD Reference:** Section 5 (Event Model)

**Scope:**
- Event CRUD operations
- Event lifecycle state machine (draft → published → cancelled → archived)
- Timezone handling (IANA format)
- Optimistic locking for concurrent updates
- Event validation and business rules

**Key Interfaces:**
- `EventService` - Event business logic
- `EventRepository` - Event persistence
- `EventValidator` - Event validation

**Dependencies:**
- Domain 1 (Auth) - for permission checks
- Domain 7 (Database) - for persistence

---

#### Domain 3: Invite & Token Management
**File:** [`lld/03_INVITE_LLD.md`](lld/03_INVITE_LLD.md)
**Status:** ✅ Complete
**HLD Reference:** Section 6 (Invite & Guest Access Model)

**Scope:**
- Invite creation (individual, bulk CSV, manual)
- Token generation (256-bit cryptographically secure)
- Token hashing (HMAC-SHA256)
- Token validation (constant-time comparison)
- Token lifecycle (expiration, revocation, regeneration)
- Invite status tracking

**Key Interfaces:**
- `InviteService` - Invite business logic
- `InviteRepository` - Invite persistence
- `TokenGenerator` - Token generation
- `TokenValidator` - Token validation

**Dependencies:**
- Domain 2 (Event) - invites belong to events
- Domain 5 (Email) - for sending invites
- Domain 7 (Database) - for persistence

---

#### Domain 4: RSVP & Preference Questions
**File:** [`lld/04_RSVP_LLD.md`](lld/04_RSVP_LLD.md)
**Status:** ✅ Complete
**HLD Reference:** Section 7 (RSVP Model), Section 8 (Preference Questions)

**Scope:**
- RSVP submission and updates
- RSVP state transitions (yes/no/maybe)
- Plus ones validation
- RSVP deadline enforcement
- Preference question management (text, select, boolean)
- Answer validation and storage

**Key Interfaces:**
- `RSVPService` - RSVP business logic
- `RSVPRepository` - RSVP persistence
- `QuestionService` - Question management
- `QuestionRepository` - Question persistence

**Dependencies:**
- Domain 3 (Invite) - RSVPs belong to invites
- Domain 2 (Event) - questions belong to events
- Domain 5 (Email) - for confirmations
- Domain 7 (Database) - for persistence

---

### Supporting Systems Layer

#### Domain 6: Template & Asset Management
**File:** [`lld/06_TEMPLATE_LLD.md`](lld/06_TEMPLATE_LLD.md)
**Status:** ✅ Complete
**HLD Reference:** Section 11 (Templates), Section 12 (Asset Storage)

**Scope:**
- Template management (invite_email, rsvp_page, confirmation_page)
- Template validation and security (XSS prevention)
- Go html/template integration
- Image upload and validation
- Storage provider abstraction (local FS, future S3)
- Asset access control and deletion

**Key Interfaces:**
- `TemplateService` - Template management
- `TemplateRenderer` - Template rendering
- `StorageProvider` - Storage abstraction
- `AssetValidator` - Asset validation

**Dependencies:**
- Domain 1 (Auth) - for access control
- Domain 7 (Database) - for template storage

---

#### Domain 5: Email System
**File:** [`lld/05_EMAIL_LLD.md`](lld/05_EMAIL_LLD.md)
**Status:** ✅ Complete
**HLD Reference:** Section 9 (Email System), Section 10 (Calendar Integration)

**Scope:**
- SMTP configuration and validation
- Email queue management (database-backed)
- Hybrid send strategy (immediate + background retry)
- Retry policy with exponential backoff
- Rate limiting (50/minute configurable)
- ICS calendar file generation (RFC 5545)
- Bounce handling and unsubscribe

**Key Interfaces:**
- `EmailSender` - Email sending
- `EmailQueue` - Queue management
- `ICSGenerator` - Calendar file generation
- `TemplateRenderer` - Email template rendering

**Dependencies:**
- Domain 2 (Event) - for event details
- Domain 3 (Invite) - for recipient info
- Domain 6 (Template) - for email templates
- Domain 7 (Database) - for queue storage

---

### Orchestration Layer

#### Domain 8: API & HTTP Handlers
**File:** [`lld/08_API_LLD.md`](lld/08_API_LLD.md)
**Status:** ✅ Complete
**HLD Reference:** Section 18 (API Routes), Section 19 (Request Flow)

**Scope:**
- HTTP router configuration
- Request/response handling
- Input validation and sanitization
- Error response formatting
- CSRF protection
- Security headers (CSP, HSTS, etc.)
- Rate limiting
- Health check and metrics endpoints
- All API routes (50+ endpoints)

**Key Interfaces:**
- `Handler` - HTTP handler
- `Middleware` - HTTP middleware
- `Validator` - Input validation
- `ErrorFormatter` - Error response formatting

**Dependencies:**
- All domains (orchestration layer)

---

## Document Relationships

### Dependency Graph

```
                    ┌─────────────────┐
                    │   Domain 8      │
                    │   API & HTTP    │
                    │ (Orchestration) │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   Domain 1   │    │   Domain 2   │    │   Domain 5   │
│     Auth     │───▶│    Events    │◀───│    Email     │
└──────┬───────┘    └──────┬───────┘    └──────┬───────┘
       │                   │                    │
       │            ┌──────┴──────┐             │
       │            ▼             ▼             │
       │    ┌──────────────┐ ┌──────────────┐  │
       │    │   Domain 3   │ │   Domain 4   │  │
       │    │   Invites    │ │     RSVP     │  │
       │    └──────┬───────┘ └──────┬───────┘  │
       │           │                 │          │
       │           └────────┬────────┘          │
       │                    │                   │
       │            ┌───────┴────────┐          │
       │            ▼                ▼          │
       │    ┌──────────────┐ ┌──────────────┐  │
       └───▶│   Domain 6   │ │   Domain 7   │◀─┘
            │   Template   │ │   Database   │
            └──────────────┘ └──────────────┘
                                (Foundation)
```

### Cross-Reference Map

| From Domain | To Domain | Reason |
|-------------|-----------|--------|
| 1 (Auth) | 7 (Database) | User and session storage |
| 2 (Event) | 1 (Auth) | Permission checks |
| 2 (Event) | 7 (Database) | Event persistence |
| 3 (Invite) | 2 (Event) | Invites belong to events |
| 3 (Invite) | 5 (Email) | Sending invites |
| 3 (Invite) | 7 (Database) | Invite persistence |
| 4 (RSVP) | 3 (Invite) | RSVPs belong to invites |
| 4 (RSVP) | 2 (Event) | Questions belong to events |
| 4 (RSVP) | 5 (Email) | Confirmation emails |
| 4 (RSVP) | 7 (Database) | RSVP persistence |
| 5 (Email) | 2 (Event) | Event details for emails |
| 5 (Email) | 3 (Invite) | Recipient information |
| 5 (Email) | 6 (Template) | Email templates |
| 5 (Email) | 7 (Database) | Queue storage |
| 6 (Template) | 1 (Auth) | Access control |
| 6 (Template) | 7 (Database) | Template storage |
| 8 (API) | All | Orchestrates all domains |

---

## Implementation Order

### Phase 1: Foundation (Priority 1)
**Goal:** Establish data layer and authentication

1. **Domain 7: Database & Persistence** - Must be first (foundation)
2. **Domain 1: Authentication & Authorization** - Required for all protected endpoints

**Estimated Effort:** 1 week  
**Deliverables:** Database layer, migrations, auth middleware

---

### Phase 2: Core Entities (Priority 2)
**Goal:** Implement core business logic

3. **Domain 2: Event Management** - Core entity
4. **Domain 3: Invite & Token Management** - Depends on events
5. **Domain 4: RSVP & Preference Questions** - Depends on invites

**Estimated Effort:** 2 weeks  
**Deliverables:** Event CRUD, invite system, RSVP handling

---

### Phase 3: Supporting Systems (Priority 3)
**Goal:** Implement email and templates

6. **Domain 6: Template & Asset Management** - Required for email
7. **Domain 5: Email System** - Depends on templates

**Estimated Effort:** 1 week  
**Deliverables:** Template engine, email queue, ICS generation

---

### Phase 4: Integration (Priority 4)
**Goal:** Wire everything together

8. **Domain 8: API & HTTP Handlers** - Orchestrates all domains

**Estimated Effort:** 1 week  
**Deliverables:** Complete HTTP API, all routes working

---

## Reading Guide

### For Implementers

**Start Here:**
1. Read [High-Level Design](02_REVISED_HLD.md) for overall architecture
2. Read [LLD Plan](03_LLD_PLAN.md) for document organization
3. Read this index to understand domain relationships
4. Read LLD documents in implementation order (7 → 1 → 2 → 3 → 4 → 6 → 5 → 8)

**For Specific Features:**
- Authentication: Domain 1
- Event creation: Domain 2
- Sending invites: Domain 3 + Domain 5
- Guest RSVP: Domain 4
- Email templates: Domain 6
- Database queries: Domain 7
- API endpoints: Domain 8

### For Reviewers

**Review Order:**
1. Domain 7 (Database) - Foundation correctness
2. Domain 1 (Auth) - Security review
3. Domains 2-4 (Business Logic) - Business rules correctness
4. Domain 5 (Email) - Reliability review
5. Domain 6 (Template) - Security review (XSS)
6. Domain 8 (API) - Integration review

---

## Document Status

| Domain | Document | Status | Lines | Interfaces | Models | Tests |
|--------|----------|--------|-------|------------|--------|-------|
| 1 | Auth | ✅ Complete | ~1000 | 4 | 3 | TDD |
| 2 | Event | ✅ Complete | ~400 | 2 | 0 | TDD |
| 3 | Invite | ✅ Complete | ~600 | 3 | 3 | TDD |
| 4 | RSVP | ✅ Complete | ~500 | 2 | 2 | TDD |
| 5 | Email | ✅ Complete | ~700 | 3 | 2 | TDD |
| 6 | Template | ✅ Complete | ~500 | 3 | 1 | TDD |
| 7 | Database | ✅ Complete | ~900 | 12 | 11 | TDD |
| 8 | API | ✅ Complete | ~600 | 2 | 2 | TDD |

**Legend:**
- 📝 Planned - Not yet written
- 🚧 In Progress - Currently being written
- ✅ Complete - Written and reviewed
- 🔄 Revision - Under revision

**Summary:**
- **Total Documents:** 8
- **Complete:** 8 (100%)
- **Total Lines:** ~5,200
- **Total Interfaces:** 34
- **Total Models:** 24
- **Testing Approach:** Test-Driven Development (TDD)

---

## Key Concepts

### Domain-Driven Design

Each domain represents a **bounded context** with:
- Clear responsibilities
- Well-defined interfaces
- Minimal coupling to other domains
- High cohesion within domain

### Repository Pattern

All data access goes through repositories:
- Abstracts database implementation
- Enables testing with mocks
- Centralizes query logic
- Enforces data access patterns

### Service Layer

Business logic lives in service layer:
- Orchestrates repository calls
- Enforces business rules
- Handles transactions
- Validates inputs

### Interface-Based Design

All dependencies are interfaces:
- Enables dependency injection
- Facilitates testing
- Allows implementation swapping
- Reduces coupling

---

## Common Patterns

### Error Handling Pattern

```go
func (s *Service) DoSomething(ctx context.Context, input Input) (*Output, error) {
    if err := s.validator.Validate(input); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    result, err := s.repository.Save(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("failed to save: %w", err)
    }
    
    return result, nil
}
```

### Transaction Pattern

```go
func (s *Service) DoTransactionalWork(ctx context.Context) error {
    return s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
        if err := s.repo1.DoWork(ctx, tx); err != nil {
            return err
        }
        if err := s.repo2.DoWork(ctx, tx); err != nil {
            return err
        }
        return nil
    })
}
```

### Validation Pattern

```go
type Validator interface {
    Validate(input interface{}) error
}

func (v *EventValidator) Validate(event *Event) error {
    var errs []error
    
    if len(event.Title) < 3 || len(event.Title) > 200 {
        errs = append(errs, errors.New("title must be 3-200 characters"))
    }
    
    if len(errs) > 0 {
        return &ValidationError{Errors: errs}
    }
    
    return nil
}
```

---

## Testing Strategy

### Test Organization

Each domain has:
- Unit tests for business logic (`*_test.go`)
- Integration tests for database operations (`*_integration_test.go`)
- Mock implementations for interfaces (`mock_*.go`)

### Test-Driven Development

**Mandatory workflow:**
1. Write test first
2. Run test (should fail)
3. Write minimal code to pass
4. Run test (should pass)
5. Refactor if needed

**See:** [README-LLM.md](../README-LLM.md#testing-requirements) for detailed TDD guidelines

---

## Type Safety Requirements

### Strongly-Typed Structs

**ALWAYS DO ✅:**
- Define structs for all data structures
- Export structs for reuse across packages
- Use explicit types for all fields

**NEVER DO ❌:**
- Use `map[string]interface{}` for structured data
- Use `interface{}` when type is known
- Pass untyped data between functions

**See:** [README-LLM.md](../README-LLM.md#1-type-safety-first) for complete type safety guidelines

---

## Security Considerations

### Per-Domain Security

| Domain | Security Focus |
|--------|----------------|
| 1 (Auth) | Session security, OIDC validation, CSRF tokens |
| 2 (Event) | Authorization checks, input validation |
| 3 (Invite) | Token security, HMAC validation, timing attacks |
| 4 (RSVP) | Input validation, deadline enforcement |
| 5 (Email) | SMTP security, rate limiting, bounce handling |
| 6 (Template) | XSS prevention, template injection, file upload validation |
| 7 (Database) | SQL injection prevention, parameterized queries |
| 8 (API) | Input sanitization, security headers, rate limiting |

**See:** [HLD Section 16](02_REVISED_HLD.md#16-security) for complete security requirements

---

## Performance Considerations

### Per-Domain Performance

| Domain | Performance Focus |
|--------|-------------------|
| 1 (Auth) | Session lookup optimization, cache sessions |
| 2 (Event) | Index on created_by, status, start_time |
| 3 (Invite) | Index on token_hash, event_id, email |
| 4 (RSVP) | Index on invite_id, response |
| 5 (Email) | Background processing, batch sending, rate limiting |
| 6 (Template) | Template caching, compiled templates |
| 7 (Database) | Connection pooling, prepared statements, indexes |
| 8 (API) | Middleware ordering, response caching |

---

## Glossary

| Term | Definition | Domain |
|------|------------|--------|
| Authenticator | Interface for authentication providers | 1 |
| Repository | Data access layer interface | 7 |
| Service | Business logic layer | All |
| Validator | Input validation interface | All |
| Token | Cryptographically secure random value | 3 |
| HMAC | Hash-based Message Authentication Code | 3 |
| ICS | iCalendar file format | 5 |
| CSP | Content Security Policy | 8 |
| CSRF | Cross-Site Request Forgery | 8 |
| RBAC | Role-Based Access Control | 1 |

---

## References

**Project Documentation:**
- [README.md](../README.md) - User-facing documentation
- [README-LLM.md](../README-LLM.md) - LLM implementation guide
- [High-Level Design](02_REVISED_HLD.md) - Authoritative specification
- [Design Review](01_HLD_DESIGN_REVIEW.md) - Design review findings
- [LLD Plan](03_LLD_PLAN.md) - LLD planning document

**Standards:**
- RFC 5545: iCalendar specification
- RFC 6749: OAuth 2.0 Authorization Framework
- OpenID Connect Core 1.0
- CAN-SPAM Act compliance

**Go Libraries:**
- github.com/coreos/go-oidc - OIDC client
- golang.org/x/oauth2 - OAuth2 helper
- github.com/mattn/go-sqlite3 - SQLite driver
- github.com/golang-migrate/migrate - Database migrations

---

## Maintenance

### Updating This Index

**When to Update:**
- New LLD document created
- LLD document status changes
- Domain dependencies change
- New cross-references identified

**Update Checklist:**
- [ ] Update document status table
- [ ] Update dependency graph if needed
- [ ] Add cross-references
- [ ] Update statistics (lines, interfaces, models)

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01-06 | Initial index created |

---

**Last Updated:** 2026-01-06  
**Next Review:** After first LLD document completion
