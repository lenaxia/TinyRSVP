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
├── 00_EPIC_foundation.md          # Epic overview
├── 00_STORY_go_module_setup.md    # Detailed user story
├── 00_STORY_config_loader.md      # Detailed user story
│
├── 01_EPIC_auth.md                # Epic overview
├── 01_STORY_oidc_integration.md   # Detailed user story
├── 01_STORY_forward_auth.md       # Detailed user story
│
├── 02_EPIC_events.md              # Epic overview
├── 02_STORY_event_model.md        # Detailed user story
├── 02_STORY_event_crud.md         # Detailed user story
│
└── ... (more epics and stories)
```

## File Formats

### Epic File Format

Epic files provide high-level overview and list related user stories.

**Filename:** `XX_EPIC_name.md`

```markdown
# Epic: [Epic Name]

**Priority:** High | Medium | Low
**Status:** Not Started | In Progress | Complete
**Target Version:** v0 | v1 | Future

## Overview

Brief description of the epic and its goals. What problem does this epic solve?

## Success Criteria

- [ ] Criterion 1 - Measurable outcome
- [ ] Criterion 2 - Measurable outcome
- [ ] Criterion 3 - Measurable outcome

## User Stories

List of user stories in this epic (each has its own file):

- [ ] [`XX_STORY_story_name.md`](XX_STORY_story_name.md) - Brief description
- [ ] [`XX_STORY_another_story.md`](XX_STORY_another_story.md) - Brief description

## Dependencies

- Depends on: Epic YY (if applicable)
- Blocks: Epic ZZ (if applicable)

## Technical Overview

High-level technical approach for this epic.

## References

- Related design docs: [`docs/XX_design.md`](../XX_design.md)
- Related HLD sections: Section X.Y in [`docs/00_INITIAL_HLD.md`](../00_INITIAL_HLD.md)
```

### User Story File Format

User story files are **comprehensive and detailed**, written for an entry-level engineer who is only marginally familiar with the codebase.

**Filename:** `XX_STORY_descriptive_name.md`

```markdown
# User Story: [Story Title]

**Epic:** [`XX_EPIC_name.md`](XX_EPIC_name.md)
**Priority:** High | Medium | Low
**Status:** Not Started | In Progress | Complete
**Estimated Effort:** Small (1 session) | Medium (2-3 sessions) | Large (4+ sessions)

## Story

**As a** [role]
**I want** [goal]
**So that** [benefit]

## Context

### Why This Matters

Explain the business/technical value. Why is this important?

### Current State

What exists now? What's the problem?

### Desired State

What should exist after this story is complete?

## Acceptance Criteria

- [ ] Criterion 1 - Specific, testable condition
- [ ] Criterion 2 - Specific, testable condition
- [ ] Criterion 3 - Specific, testable condition

## Technical Approach

### Architecture

Describe where this fits in the system architecture. Include ASCII diagram if helpful:

```
┌─────────────┐
│  Component  │
│   A         │──────> New Component
└─────────────┘
```

### Files to Create/Modify

List specific files with their purpose:

- **Create:** `internal/auth/oidc.go` - OIDC client implementation
- **Create:** `internal/auth/oidc_test.go` - OIDC client tests
- **Modify:** `internal/config/config.go` - Add OIDC configuration fields
- **Modify:** `cmd/server/main.go` - Initialize OIDC client

### Key Functions/Types

Define the main types and functions needed:

```go
// OIDCConfig holds OIDC provider configuration
type OIDCConfig struct {
    IssuerURL    string
    ClientID     string
    ClientSecret string
    RedirectURL  string
}

// NewOIDCClient creates a new OIDC client
func NewOIDCClient(ctx context.Context, cfg *OIDCConfig) (*OIDCClient, error) {
    // Implementation details
}
```

### Dependencies

**Go Packages:**
- `github.com/coreos/go-oidc/v3/oidc` - OIDC client library
- `golang.org/x/oauth2` - OAuth2 helper

**Internal Packages:**
- `internal/config` - Configuration loading
- `internal/models` - User model

**External Services:**
- OIDC provider (Authentik/Keycloak/etc.)

### Data Flow

Describe the flow of data through the system:

1. User clicks "Login" → `/login` handler
2. Handler redirects to OIDC provider authorization URL
3. Provider redirects back to `/oidc/callback` with code
4. Callback handler exchanges code for tokens
5. Handler validates ID token, extracts claims
6. Handler creates/updates user in database
7. Handler sets session cookie
8. Handler redirects to dashboard

## Implementation Guide

### Step-by-Step Instructions

Detailed steps for an entry-level engineer:

#### Step 1: Set Up Dependencies

```bash
# Add required Go modules
go get github.com/coreos/go-oidc/v3/oidc
go get golang.org/x/oauth2
```

#### Step 2: Define Configuration Struct

In `internal/config/config.go`, add:

```go
type OIDCConfig struct {
    Enabled      bool   `env:"OIDC_ENABLED" default:"false"`
    IssuerURL    string `env:"OIDC_ISSUER_URL"`
    ClientID     string `env:"OIDC_CLIENT_ID"`
    ClientSecret string `env:"OIDC_CLIENT_SECRET"`
    RedirectURL  string `env:"OIDC_REDIRECT_URL"`
}
```

**Why:** This allows configuration via environment variables.

#### Step 3: Write Tests First (TDD)

In `internal/auth/oidc_test.go`:

```go
func TestNewOIDCClient(t *testing.T) {
    tests := []struct {
        name    string
        config  *OIDCConfig
        wantErr bool
    }{
        {
            name: "valid config",
            config: &OIDCConfig{
                IssuerURL:    "https://example.com",
                ClientID:     "test-client",
                ClientSecret: "test-secret",
                RedirectURL:  "http://localhost/callback",
            },
            wantErr: false,
        },
        {
            name: "missing issuer",
            config: &OIDCConfig{
                ClientID:     "test-client",
                ClientSecret: "test-secret",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            client, err := NewOIDCClient(context.Background(), tt.config)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewOIDCClient() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !tt.wantErr && client == nil {
                t.Error("Expected non-nil client")
            }
        })
    }
}
```

#### Step 4: Implement the Function

In `internal/auth/oidc.go`:

```go
package auth

import (
    "context"
    "fmt"
    
    "github.com/coreos/go-oidc/v3/oidc"
    "golang.org/x/oauth2"
)

type OIDCClient struct {
    provider *oidc.Provider
    verifier *oidc.IDTokenVerifier
    config   oauth2.Config
}

func NewOIDCClient(ctx context.Context, cfg *OIDCConfig) (*OIDCClient, error) {
    if cfg.IssuerURL == "" {
        return nil, fmt.Errorf("issuer URL is required")
    }
    
    provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
    if err != nil {
        return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
    }
    
    // Implementation continues...
}
```

**Why:** This follows TDD - tests first, then implementation.

#### Step 5: Run Tests

```bash
go test -timeout 30s ./internal/auth/...
```

**Expected:** Tests should pass.

### Tasks (if needed)

Only include tasks for complex stories that need breakdown:

- [ ] Task 1 - Specific subtask
- [ ] Task 2 - Specific subtask

## Examples

### Similar Code in Codebase

Reference existing patterns:

- See `internal/email/smtp.go` for similar client initialization pattern
- See `internal/config/config.go` for configuration struct examples

### External Examples

- [go-oidc example](https://github.com/coreos/go-oidc/blob/v3/example/idtoken/app.go)
- [OAuth2 flow example](https://pkg.go.dev/golang.org/x/oauth2#example-Config)

## Testing Strategy

### Unit Tests

- Test client initialization with valid/invalid configs
- Test token validation with mock tokens
- Test error handling

### Integration Tests

- Test full OIDC flow with test provider
- Test callback handling
- Test session creation

### Manual Testing

1. Start local OIDC provider (e.g., Keycloak in Docker)
2. Configure app with provider details
3. Navigate to `/login`
4. Complete OIDC flow
5. Verify session created
6. Verify user in database

## Edge Cases

- What if OIDC provider is unreachable?
- What if token validation fails?
- What if user's email changes?
- What if provider doesn't return email claim?

## Rollback Plan

If this needs to be reverted:

1. Remove OIDC routes from router
2. Remove OIDC config from environment
3. Revert to forward auth only
4. No database changes needed (user table unchanged)

## References

- **HLD Section:** Section 3.1 in [`docs/00_INITIAL_HLD.md`](../00_INITIAL_HLD.md)
- **Design Doc:** (if applicable)
- **Related Stories:**
  - [`01_STORY_forward_auth.md`](01_STORY_forward_auth.md)
  - [`01_STORY_session_management.md`](01_STORY_session_management.md)

## Questions to Ask

If you're unsure while implementing:

- Which OIDC claims should we extract?
- Should we auto-create users on first login?
- What should happen if email claim is missing?
- Should we support refresh tokens?

**Ask the user rather than guessing.**

## Success Checklist

Before marking this story complete:

- [ ] All tests passing with timeout
- [ ] Code follows type safety guidelines (no `map[string]interface{}`)
- [ ] Documentation updated (README, architecture diagram if needed)
- [ ] Worklog entry created
- [ ] Changes committed with reference to this story
- [ ] Manual testing completed
- [ ] Edge cases handled
```

## Current Epics

| Epic | Priority | Status | Stories | Completion |
|------|----------|--------|---------|------------|
| 00_EPIC_foundation | High | Not Started | 0 | 0% |
| 01_EPIC_auth | High | Not Started | 0 | 0% |
| 02_EPIC_events | High | Not Started | 0 | 0% |
| 03_EPIC_invites | High | Not Started | 0 | 0% |
| 04_EPIC_rsvp | High | Not Started | 0 | 0% |
| 05_EPIC_email | High | Not Started | 0 | 0% |
| 06_EPIC_templates | Medium | Not Started | 0 | 0% |

## Priority Definitions

**High:** Required for v0 release, blocks other work  
**Medium:** Important but not blocking  
**Low:** Nice to have, can be deferred

## Status Definitions

**Not Started:** No work begun  
**In Progress:** Active development  
**Complete:** All acceptance criteria met, tests passing

## Workflow

1. **Select Story** - Choose highest priority "Not Started" story
2. **Update Status** - Mark as "In Progress"
3. **Work on Tasks** - Complete tasks, mark with `[x]`
4. **Write Tests** - TDD approach, tests first
5. **Implement** - Write code to pass tests
6. **Update Status** - Mark story as "Complete"
7. **Create Worklog** - Document in `docs/01_WORKLOG/`
8. **Commit** - Commit changes with reference to story

## Maintenance

**After Each Session:**
- Update task checklists `[ ]` → `[x]`
- Update story status if complete
- Update epic completion percentage
- Update this README's epic table

**Weekly:**
- Review priorities
- Reorder stories if needed
- Add new stories as discovered
- Archive completed epics (move to `archive/` if needed)

## References

- **Authoritative Spec:** [`docs/00_INITIAL_HLD.md`](../00_INITIAL_HLD.md)
- **Implementation Guide:** [`README-LLM.md`](../../README-LLM.md)
- **Worklog:** [`docs/01_WORKLOG/`](../01_WORKLOG/)
