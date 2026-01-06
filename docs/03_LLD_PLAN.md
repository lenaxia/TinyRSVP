# TinyRSVP - Low-Level Design (LLD) Document Plan

**Date:** 2026-01-06  
**Status:** Planning Complete  
**Purpose:** Define structure and organization of LLD documents

---

## Overview

This document outlines the structure and organization of Low-Level Design (LLD) documents for TinyRSVP. Each LLD document focuses on a specific domain and provides implementation-ready specifications including Go package structure, interfaces, types, and function signatures.

---

## LLD Document Structure

### Document Organization

```
docs/
├── 02_REVISED_HLD.md                    # High-Level Design (authoritative)
├── 03_LLD_PLAN.md                       # This document
├── 04_LLD_INDEX.md                      # Index of all LLD documents
└── lld/                                 # Low-Level Design documents
    ├── 01_AUTH_LLD.md                   # Authentication & Authorization
    ├── 02_EVENT_LLD.md                  # Event Management
    ├── 03_INVITE_LLD.md                 # Invite & Token Management
    ├── 04_RSVP_LLD.md                   # RSVP & Preference Questions
    ├── 05_EMAIL_LLD.md                  # Email System
    ├── 06_TEMPLATE_LLD.md               # Template & Asset Management
    ├── 07_DATABASE_LLD.md               # Database & Persistence
    └── 08_API_LLD.md                    # API & HTTP Handlers
```

---

## Domain Breakdown

### Domain 1: Authentication & Authorization
**File:** `docs/lld/01_AUTH_LLD.md`

**Scope:**
- OIDC authentication flow
- Forward auth integration
- Session management (database-backed)
- User management
- Role-based access control (RBAC)
- Bootstrap admin creation
- Middleware for authentication and authorization

**Key Packages:**
- `internal/auth/` - Authentication logic
- `internal/middleware/` - HTTP middleware
- `internal/models/` - User and Session models

**Dependencies:**
- Database layer (Domain 7)
- HTTP handlers (Domain 8)

**Interfaces:**
- `Authenticator` - Authentication provider interface
- `SessionStore` - Session storage interface
- `AuthorizationChecker` - Permission checking interface

---

### Domain 2: Event Management
**File:** `docs/lld/02_EVENT_LLD.md`

**Scope:**
- Event CRUD operations
- Event lifecycle state machine (draft → published → cancelled → archived)
- Event validation
- Timezone handling
- Optimistic locking for concurrent updates
- Event ownership and permissions

**Key Packages:**
- `internal/events/` - Event business logic
- `internal/models/` - Event model

**Dependencies:**
- Authentication (Domain 1)
- Database layer (Domain 7)
- Template system (Domain 6)

**Interfaces:**
- `EventService` - Event business logic interface
- `EventRepository` - Event persistence interface
- `EventValidator` - Event validation interface

---

### Domain 3: Invite & Token Management
**File:** `docs/lld/03_INVITE_LLD.md`

**Scope:**
- Invite creation (individual, bulk CSV, manual)
- Token generation (256-bit cryptographically secure)
- Token hashing (HMAC-SHA256)
- Token validation (constant-time comparison)
- Token lifecycle (expiration, revocation, regeneration)
- Invite status tracking

**Key Packages:**
- `internal/invites/` - Invite business logic
- `pkg/token/` - Token generation and validation (reusable)
- `internal/models/` - Invite model

**Dependencies:**
- Event management (Domain 2)
- Email system (Domain 5)
- Database layer (Domain 7)

**Interfaces:**
- `InviteService` - Invite business logic interface
- `InviteRepository` - Invite persistence interface
- `TokenGenerator` - Token generation interface
- `TokenValidator` - Token validation interface

---

### Domain 4: RSVP & Preference Questions
**File:** `docs/lld/04_RSVP_LLD.md`

**Scope:**
- RSVP submission and updates
- RSVP state transitions (yes/no/maybe)
- Plus ones validation
- RSVP deadline enforcement
- Preference question management (text, select, boolean)
- Answer validation and storage
- Question lifecycle management

**Key Packages:**
- `internal/rsvp/` - RSVP business logic
- `internal/models/` - RSVP and Question models

**Dependencies:**
- Invite management (Domain 3)
- Event management (Domain 2)
- Email system (Domain 5) - for confirmations
- Database layer (Domain 7)

**Interfaces:**
- `RSVPService` - RSVP business logic interface
- `RSVPRepository` - RSVP persistence interface
- `QuestionService` - Question management interface
- `QuestionRepository` - Question persistence interface

---

### Domain 5: Email System
**File:** `docs/lld/05_EMAIL_LLD.md`

**Scope:**
- SMTP configuration and validation
- Email queue management (database-backed)
- Hybrid send strategy (immediate + background retry)
- Retry policy with exponential backoff
- Rate limiting (50/minute configurable)
- Email templates (invite, confirmation, update, cancellation)
- ICS calendar file generation (RFC 5545)
- Bounce handling
- Unsubscribe mechanism

**Key Packages:**
- `internal/email/` - Email business logic
- `pkg/ics/` - ICS calendar generation (reusable)
- `internal/models/` - EmailQueue model

**Dependencies:**
- Event management (Domain 2)
- Invite management (Domain 3)
- Template system (Domain 6)
- Database layer (Domain 7)

**Interfaces:**
- `EmailSender` - Email sending interface
- `EmailQueue` - Queue management interface
- `ICSGenerator` - Calendar file generation interface
- `TemplateRenderer` - Email template rendering interface

---

### Domain 6: Template & Asset Management
**File:** `docs/lld/06_TEMPLATE_LLD.md`

**Scope:**
- Template management (invite_email, rsvp_page, confirmation_page)
- Template validation and security (XSS prevention)
- Template variable interpolation (Go html/template)
- Image upload and validation
- Asset storage abstraction (local FS, future S3)
- Asset access control
- Asset deletion policy

**Key Packages:**
- `internal/templates/` - Template management
- `internal/storage/` - Storage provider abstraction
- `internal/models/` - Template model

**Dependencies:**
- Authentication (Domain 1) - for access control
- Database layer (Domain 7)

**Interfaces:**
- `TemplateService` - Template management interface
- `TemplateRenderer` - Template rendering interface
- `StorageProvider` - Storage abstraction interface
- `AssetValidator` - Asset validation interface

---

### Domain 7: Database & Persistence
**File:** `docs/lld/07_DATABASE_LLD.md`

**Scope:**
- Database connection management (SQLite, future PostgreSQL)
- Repository pattern implementation
- Transaction management
- Migration strategy (golang-migrate)
- Query optimization and indexing
- Connection pooling
- Error handling and retries
- Audit logging

**Key Packages:**
- `internal/db/` - Database connection and utilities
- `internal/models/` - All data models
- `migrations/` - SQL migration files

**Dependencies:**
- None (foundational layer)

**Interfaces:**
- `Database` - Database connection interface
- `Repository` - Base repository interface
- `Transactor` - Transaction management interface
- `Migrator` - Migration execution interface

---

### Domain 8: API & HTTP Handlers
**File:** `docs/lld/08_API_LLD.md`

**Scope:**
- HTTP router configuration
- Request/response handling
- Input validation and sanitization
- Error response formatting
- CSRF protection
- Security headers
- Rate limiting
- Health check endpoint
- Metrics endpoint (Prometheus)
- All API routes (admin, events, invites, RSVP, templates, assets)

**Key Packages:**
- `internal/handlers/` - HTTP handlers
- `internal/middleware/` - HTTP middleware
- `cmd/server/` - Application entrypoint

**Dependencies:**
- All other domains (orchestration layer)

**Interfaces:**
- `Handler` - HTTP handler interface
- `Middleware` - Middleware interface
- `Validator` - Input validation interface
- `ErrorFormatter` - Error response formatting interface

---

## Cross-Domain Dependencies

### Dependency Graph

```
┌─────────────────────────────────────────────────────────────┐
│                    Domain 8: API & HTTP                      │
│                    (Orchestration Layer)                     │
└──────────────────────┬──────────────────────────────────────┘
                       │
       ┌───────────────┼───────────────┬───────────────┐
       ▼               ▼               ▼               ▼
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│ Domain 1 │    │ Domain 2 │    │ Domain 5 │    │ Domain 6 │
│   Auth   │───▶│  Events  │◀───│  Email   │    │ Template │
└────┬─────┘    └────┬─────┘    └────┬─────┘    └────┬─────┘
     │               │               │               │
     │         ┌─────┴─────┐         │               │
     │         ▼           ▼         │               │
     │    ┌──────────┐ ┌──────────┐ │               │
     │    │ Domain 3 │ │ Domain 4 │ │               │
     │    │ Invites  │ │   RSVP   │ │               │
     │    └────┬─────┘ └────┬─────┘ │               │
     │         │             │       │               │
     └─────────┴─────────────┴───────┴───────────────┘
                       │
                       ▼
              ┌─────────────────┐
              │    Domain 7     │
              │    Database     │
              │ (Foundation)    │
              └─────────────────┘
```

### Dependency Rules

1. **Foundation Layer (Domain 7)** - No dependencies, used by all
2. **Core Business Logic (Domains 1-6)** - Depend on foundation, may depend on each other
3. **Orchestration Layer (Domain 8)** - Depends on all other domains

---

## LLD Document Template

Each LLD document will follow this structure:

```markdown
# Domain Name - Low-Level Design

**Version:** 1.0
**Date:** YYYY-MM-DD
**HLD Reference:** Section X.Y

## 1. Overview
- Purpose and scope
- Key responsibilities
- Design principles

## 2. Package Structure
- Go package organization
- File layout
- Import dependencies

## 3. Data Models
- Struct definitions
- Field descriptions
- Validation rules

## 4. Interfaces
- Interface definitions
- Method signatures
- Contract specifications

## 5. Implementation Details
- Core algorithms
- State machines
- Error handling
- Concurrency considerations

## 6. Dependencies
- Internal dependencies
- External libraries
- Cross-domain interactions

## 7. Testing Strategy
- Unit test approach
- Test fixtures
- Mock interfaces
- Edge cases

## 8. Security Considerations
- Input validation
- Authorization checks
- Sensitive data handling

## 9. Performance Considerations
- Optimization strategies
- Caching (if applicable)
- Query optimization

## 10. Error Scenarios
- Error types
- Error handling
- Recovery strategies

## 11. Examples
- Usage examples
- Code snippets
- Integration examples

## 12. Open Questions
- Unresolved design decisions
- Future enhancements
```

---

## Implementation Order

### Phase 1: Foundation (Week 1)
1. **Domain 7: Database & Persistence** - Foundation for all data operations
2. **Domain 1: Authentication & Authorization** - Required for all protected endpoints

### Phase 2: Core Business Logic (Week 2-3)
3. **Domain 2: Event Management** - Core entity
4. **Domain 3: Invite & Token Management** - Depends on events
5. **Domain 4: RSVP & Preference Questions** - Depends on invites

### Phase 3: Supporting Systems (Week 4)
6. **Domain 6: Template & Asset Management** - Required for email
7. **Domain 5: Email System** - Depends on templates

### Phase 4: Integration (Week 5)
8. **Domain 8: API & HTTP Handlers** - Orchestrates all domains

---

## Cross-Reference Strategy

### Reference Format

When referencing other LLD documents:

```markdown
**See:** [Domain Name LLD](./XX_DOMAIN_LLD.md#section-name)
```

### Common Cross-References

1. **Authentication checks** → Reference Domain 1 (Auth)
2. **Database operations** → Reference Domain 7 (Database)
3. **Email sending** → Reference Domain 5 (Email)
4. **Template rendering** → Reference Domain 6 (Template)
5. **Event validation** → Reference Domain 2 (Event)

---

## Documentation Standards

### Code Examples

All code examples must:
- Be valid Go syntax
- Include error handling
- Follow project conventions (from README-LLM.md)
- Use strongly-typed structs (no `map[string]interface{}`)
- Include comments for complex logic

### Interface Definitions

All interfaces must:
- Have clear method signatures
- Document parameters and return values
- Specify error conditions
- Include usage examples

### Type Definitions

All types must:
- Use explicit types (no `interface{}`)
- Include struct tags for JSON/DB mapping
- Document field constraints
- Include validation rules

---

## Review Checklist

Before finalizing each LLD:

- [ ] All interfaces defined with clear contracts
- [ ] All data models specified with validation rules
- [ ] Error handling strategy documented
- [ ] Testing approach defined
- [ ] Security considerations addressed
- [ ] Performance considerations noted
- [ ] Cross-references to other LLDs included
- [ ] Code examples provided
- [ ] Consistent with HLD specifications
- [ ] No `map[string]interface{}` usage
- [ ] TDD approach outlined

---

## Next Steps

1. Create LLD index document (`docs/04_LLD_INDEX.md`)
2. Create `docs/lld/` directory
3. Write LLD documents in implementation order
4. Review each LLD for completeness and consistency
5. Update index as documents are completed

---

**Status:** Planning complete, ready to begin LLD document creation
