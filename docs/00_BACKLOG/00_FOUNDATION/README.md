# Epic: Foundation & Project Setup

**Priority:** High
**Status:** ✅ Complete
**Target Version:** v0
**Estimated Effort:** 1 week
**Completed:** 2026-01-06

---

## Overview

Establish the foundational infrastructure for TinyRSVP including Go module setup, configuration management, database layer, and migrations. This epic provides the base upon which all other features will be built.

**Goal:** Create a working Go application with database connectivity, configuration loading, and basic project structure.

---

## Success Criteria

- [x] Go module initialized with all required dependencies
- [x] Configuration loaded from environment variables
- [x] Database connection established (SQLite)
- [x] All 9 database tables created via migrations
- [x] Repository pattern implemented for data access
- [x] Health check endpoint returns database status
- [x] Application starts successfully in Docker container

---

## User Stories

**Naming Convention:** `{EPIC_NUMBER}_STORY_{STORY_NUMBER}_{description}.md`

### Phase 1: Project Bootstrap
- [x] [`00_STORY_01_go_module_setup.md`](00_STORY_01_go_module_setup.md) - Initialize Go module and project structure
- [x] [`00_STORY_02_config_management.md`](00_STORY_02_config_management.md) - Environment-based configuration system

### Phase 2: Database Layer
- [x] [`00_STORY_03_database_connection.md`](00_STORY_03_database_connection.md) - Database connection management and pooling
- [x] [`00_STORY_04_database_migrations.md`](00_STORY_04_database_migrations.md) - Migration system using golang-migrate
- [x] [`00_STORY_05_repository_pattern.md`](00_STORY_05_repository_pattern.md) - Base repository interfaces and implementations

### Phase 3: Infrastructure
- [x] [`00_STORY_06_health_checks.md`](00_STORY_06_health_checks.md) - Health check and readiness endpoints
- [x] [`00_STORY_07_docker_setup.md`](00_STORY_07_docker_setup.md) - Dockerfile and docker-compose configuration

---

## Dependencies

**Depends on:** None (foundation layer)  
**Blocks:** All other epics

---

## Technical Overview

### Architecture

```
┌─────────────────────────────────────┐
│         Application Entry           │
│        (cmd/server/main.go)         │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      Configuration Loader           │
│     (internal/config/config.go)     │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      Database Connection            │
│        (internal/db/db.go)          │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│         Migrations                  │
│   (migrations/sqlite/*.sql)         │
└─────────────────────────────────────┘
```

### Key Components

1. **Configuration Management**
   - Environment variable parsing
   - Validation and defaults
   - Type-safe configuration structs

2. **Database Layer**
   - SQLite connection with proper settings
   - Connection pooling
   - Transaction support
   - Migration execution

3. **Repository Pattern**
   - Base repository interface
   - Transaction management
   - Error handling patterns

---

## Technical Decisions

### Database Choice: SQLite (v0)
- Zero configuration
- Single file storage
- Perfect for homelab deployment
- PostgreSQL support deferred to v1+

### Migration Tool: golang-migrate
- Industry standard
- Supports up/down migrations
- Works with SQLite and PostgreSQL
- Embedded migrations in binary

### Configuration: Environment Variables
- 12-factor app methodology
- Docker-friendly
- No config files to manage
- Type-safe parsing with validation

---

## References

- **HLD:** Section 13 (Database Schema), Section 20 (Deployment)
- **LLD:** [`lld/07_DATABASE_LLD.md`](../lld/07_DATABASE_LLD.md)
- **README-LLM:** Type Safety Guidelines, TDD Requirements

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| SQLite performance issues | Medium | Use proper indexes, connection pooling |
| Migration failures | High | Test migrations thoroughly, support rollback |
| Configuration errors | High | Validate on startup, fail fast with clear errors |

---

## Definition of Done

- [x] All user stories complete
- [x] All tests passing with timeout
- [x] Application starts successfully
- [x] Database migrations run automatically
- [x] Health check endpoint functional
- [x] Docker image builds successfully
- [x] Documentation updated
