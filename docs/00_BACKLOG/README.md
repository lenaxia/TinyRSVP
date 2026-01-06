# Backlog

## Purpose

This folder contains sprint stories, epics, and user stories for the TinyRSVP project. Stories are organized by priority and implementation order. Each epic has its own file, and each user story has its own detailed file.

## Rules

1. **Separate files for epics and stories** - Epics overview, user stories are detailed
2. **User stories are comprehensive** - Written for entry-level engineers unfamiliar with codebase
3. **Include examples and references** - Code snippets, file paths, related functions
4. **Tasks only when necessary** - Use checklists `[ ]` in user story files for complex work
5. **Update status regularly** - Mark completed items `[x]`
6. **Keep stories focused** - User stories should be completable in 1-3 sessions
7. **Reference from worklog** - Link completed work back to stories

## Structure

```
00_BACKLOG/
├── README.md (this file)
│
├── 00_EPIC_foundation.md          # Epic: Foundation & Project Setup
├── 01_EPIC_auth.md                # Epic: Authentication & Authorization
├── 02_EPIC_events.md              # Epic: Event Management
├── 03_EPIC_invites.md             # Epic: Invite & Token Management
├── 04_EPIC_rsvp.md                # Epic: RSVP & Guest Experience
├── 05_EPIC_email.md               # Epic: Email System & Calendar Integration
├── 06_EPIC_templates.md           # Epic: Templates & Asset Management
├── 07_EPIC_frontend.md            # Epic: Frontend & User Experience
├── 08_EPIC_api.md                 # Epic: API & HTTP Layer
│
└── (User story files will be created as epics are started)
```

## Current Epics

| # | Epic | Priority | Status | Stories | Effort | Dependencies |
|---|------|----------|--------|---------|--------|--------------|
| 00 | [Foundation & Project Setup](00_EPIC_foundation.md) | High | Not Started | 7 | 1 week | None |
| 01 | [Authentication & Authorization](01_EPIC_auth.md) | High | Not Started | 8 | 1 week | Epic 00 |
| 02 | [Event Management](02_EPIC_events.md) | High | Not Started | 11 | 2 weeks | Epic 00, 01 |
| 03 | [Invite & Token Management](03_EPIC_invites.md) | High | Not Started | 11 | 1.5 weeks | Epic 00, 01, 02 |
| 04 | [RSVP & Guest Experience](04_EPIC_rsvp.md) | High | Not Started | 11 | 1 week | Epic 00, 02, 03 |
| 05 | [Email System & Calendar](05_EPIC_email.md) | High | Not Started | 15 | 1.5 weeks | Epic 00, 02, 03, 06 |
| 06 | [Templates & Asset Management](06_EPIC_templates.md) | Medium | Not Started | 13 | 1 week | Epic 00, 01 |
| 07 | [Frontend & User Experience](07_EPIC_frontend.md) | High | Not Started | 21 | 1 week | Epic 08 |
| 08 | [API & HTTP Layer](08_EPIC_api.md) | High | Not Started | 18 | 1.5 weeks | All |

**Total Stories:** 115  
**Total Effort:** ~10 weeks  
**v0 Target:** All epics complete

---

## Implementation Order

### Phase 1: Foundation (Week 1)
**Goal:** Establish infrastructure
```
Epic 00: Foundation & Project Setup
  ├─ Database connection
  ├─ Configuration management
  ├─ Migrations
  └─ Repository pattern
```

### Phase 2: Authentication (Week 2)
**Goal:** Secure the application
```
Epic 01: Authentication & Authorization
  ├─ OIDC integration
  ├─ Forward auth
  ├─ Session management
  └─ RBAC middleware
```

### Phase 3: Core Business Logic (Weeks 3-5)
**Goal:** Implement core features
```
Epic 02: Event Management (Week 3-4)
  ├─ Event CRUD
  ├─ Lifecycle states
  ├─ Timezone handling
  └─ Preference questions

Epic 03: Invite & Token Management (Week 4-5)
  ├─ Token generation/validation
  ├─ Invite creation
  ├─ CSV import
  └─ Token lifecycle

Epic 04: RSVP & Guest Experience (Week 5)
  ├─ RSVP submission
  ├─ Plus ones
  ├─ Question answering
  └─ Deadline enforcement
```

### Phase 4: Supporting Systems (Weeks 6-7)
**Goal:** Templates and email
```
Epic 06: Templates & Asset Management (Week 6)
  ├─ Template system
  ├─ Image uploads
  ├─ Storage provider
  └─ XSS prevention

Epic 05: Email System & Calendar (Week 7)
  ├─ SMTP integration
  ├─ Email queue
  ├─ ICS generation
  └─ Retry logic
```

### Phase 5: Integration (Weeks 8-10)
**Goal:** Complete application
```
Epic 08: API & HTTP Layer (Week 8-9)
  ├─ All routes
  ├─ Middleware chain
  ├─ Error handling
  └─ Security headers

Epic 07: Frontend & User Experience (Week 9-10)
  ├─ Mobile-responsive UI
  ├─ Admin dashboard
  ├─ Guest RSVP pages
  └─ Accessibility
```

---

## Epic Dependency Graph

```
                    Epic 00 (Foundation)
                           │
                           ▼
                    Epic 01 (Auth)
                           │
                ┌──────────┴──────────┐
                ▼                     ▼
         Epic 02 (Events)      Epic 06 (Templates)
                │                     │
                ▼                     │
         Epic 03 (Invites)            │
                │                     │
         ┌──────┴──────┐              │
         ▼             ▼              │
  Epic 04 (RSVP)  Epic 05 (Email)◄───┘
         │             │
         └──────┬──────┘
                ▼
         Epic 08 (API)
                │
                ▼
         Epic 07 (Frontend)
```

---

## Priority Definitions

**High:** Required for v0 release, blocks other work  
**Medium:** Important but not blocking  
**Low:** Nice to have, can be deferred

---

## Status Definitions

**Not Started:** No work begun  
**In Progress:** Active development  
**Blocked:** Waiting on dependencies  
**Complete:** All acceptance criteria met, tests passing

---

## Epic Completion Tracking

### Overall Progress
- **Epics Complete:** 0/9 (0%)
- **Stories Complete:** 0/115 (0%)
- **Estimated Remaining:** 10 weeks

### By Priority
- **High Priority:** 0/8 complete
- **Medium Priority:** 0/1 complete

---

## Workflow

### Starting an Epic
1. **Review Dependencies** - Ensure prerequisite epics complete
2. **Read Epic File** - Understand goals and success criteria
3. **Update Status** - Mark epic as "In Progress"
4. **Create User Stories** - Create detailed story files as needed
5. **Work Stories Sequentially** - Follow story order in epic

### Working a User Story
1. **Select Story** - Choose next story in epic
2. **Update Status** - Mark as "In Progress"
3. **Follow TDD** - Write tests first, then implementation
4. **Update Tasks** - Mark checklist items `[x]` as complete
5. **Run Tests** - Ensure all tests pass with timeout
6. **Update Status** - Mark story as "Complete"
7. **Create Worklog** - Document in `docs/01_WORKLOG/`
8. **Commit** - Reference story in commit message

### Completing an Epic
1. **Verify All Stories** - All stories marked complete
2. **Verify Success Criteria** - All criteria met
3. **Run Full Test Suite** - All tests passing
4. **Update Epic Status** - Mark as "Complete"
5. **Update This README** - Update completion tracking
6. **Create Handoff** - Document in worklog

---

## Story Estimation

### Small (1 session, 2-4 hours)
- Simple CRUD operations
- Basic validation
- Straightforward tests

### Medium (2-3 sessions, 1-2 days)
- Complex business logic
- Multiple integrations
- Comprehensive testing

### Large (4+ sessions, 3-5 days)
- Major features
- Multiple components
- Extensive testing
- UI work

---

## Maintenance

### After Each Session
- [ ] Update story task checklists
- [ ] Update story status if complete
- [ ] Update epic completion percentage
- [ ] Update this README's tracking table
- [ ] Commit changes with story reference

### After Each Epic
- [ ] Update epic status to "Complete"
- [ ] Update overall progress metrics
- [ ] Create epic completion worklog
- [ ] Review and adjust remaining epic priorities

### Weekly Review
- [ ] Review priorities
- [ ] Reorder stories if needed
- [ ] Add newly discovered stories
- [ ] Update effort estimates
- [ ] Review blockers

---

## Quick Reference

### Current Sprint Focus
**Sprint 1:** Epic 00 (Foundation)  
**Next:** Epic 01 (Authentication)

### Blocked Stories
None currently

### High Priority Incomplete
All epics (project just starting)

---

## References

- **Authoritative Spec:** [`docs/02_REVISED_HLD.md`](../02_REVISED_HLD.md)
- **LLD Index:** [`docs/04_LLD_INDEX.md`](../04_LLD_INDEX.md)
- **Implementation Guide:** [`README-LLM.md`](../../README-LLM.md)
- **Worklog:** [`docs/01_WORKLOG/`](../01_WORKLOG/)

---

## Epic Quick Links

- [00: Foundation](00_EPIC_foundation.md) - Database, config, migrations
- [01: Authentication](01_EPIC_auth.md) - OIDC, sessions, RBAC
- [02: Events](02_EPIC_events.md) - Event lifecycle, questions
- [03: Invites](03_EPIC_invites.md) - Tokens, CSV import
- [04: RSVP](04_EPIC_rsvp.md) - Guest responses, plus ones
- [05: Email](05_EPIC_email.md) - SMTP, queue, ICS files
- [06: Templates](06_EPIC_templates.md) - Customization, assets
- [07: Frontend](07_EPIC_frontend.md) - UI, mobile-first, accessibility
- [08: API](08_EPIC_api.md) - Routes, middleware, integration

---

**Last Updated:** 2026-01-06  
**Next Review:** After Epic 00 completion
