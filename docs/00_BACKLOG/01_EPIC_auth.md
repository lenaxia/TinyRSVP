# Epic: Authentication & Authorization

**Priority:** High  
**Status:** Not Started  
**Target Version:** v0  
**Estimated Effort:** 1 week

---

## Overview

Implement secure authentication and authorization for admin and event manager users. Support both OIDC (OpenID Connect) and forward auth modes. Establish role-based access control (RBAC) and session management.

**Goal:** Secure the application with proper authentication, create the first admin user automatically, and enforce permissions throughout the system.

---

## Success Criteria

- [ ] OIDC authentication flow working end-to-end
- [ ] Forward auth integration functional
- [ ] First user automatically becomes admin
- [ ] Session management with secure cookies
- [ ] RBAC middleware enforces permissions
- [ ] Users can log in and access dashboard
- [ ] Sessions expire after 7 days
- [ ] Logout invalidates session immediately

---

## User Stories

### Phase 1: Core Authentication
- [ ] [`01_STORY_01_oidc_integration.md`](01_STORY_01_oidc_integration.md) - OIDC authentication flow
- [ ] [`01_STORY_02_forward_auth.md`](01_STORY_02_forward_auth.md) - Forward auth header validation
- [ ] [`01_STORY_03_session_management.md`](01_STORY_03_session_management.md) - Database-backed sessions

### Phase 2: User Management
- [ ] [`01_STORY_04_user_model.md`](01_STORY_04_user_model.md) - User model and service
- [ ] [`01_STORY_05_bootstrap_admin.md`](01_STORY_05_bootstrap_admin.md) - First user becomes admin
- [ ] [`01_STORY_06_user_crud.md`](01_STORY_06_user_crud.md) - User management endpoints

### Phase 3: Authorization
- [ ] [`01_STORY_07_rbac_middleware.md`](01_STORY_07_rbac_middleware.md) - Role-based access control
- [ ] [`01_STORY_08_permission_checks.md`](01_STORY_08_permission_checks.md) - Permission checking service

---

## Dependencies

**Depends on:** Epic 00 (Foundation)  
**Blocks:** Epic 02 (Events), Epic 03 (Invites)

---

## Technical Overview

### Authentication Modes

**Mode 1: OIDC**
```
User → /login → OIDC Provider → /auth/callback → Session Created
```

**Mode 2: Forward Auth**
```
Reverse Proxy → Headers → App validates → Session Created
```

### Session Flow

```
┌──────────────┐
│   Browser    │
└──────┬───────┘
       │ Cookie: tinyrsvp_session=<id>
       ▼
┌──────────────┐
│  Middleware  │
└──────┬───────┘
       │ Validate session
       ▼
┌──────────────┐
│   Database   │
│   sessions   │
└──────────────┘
```

### Role Hierarchy

```
Admin (full control)
  └─> Event Manager (own events only)
      └─> Guest (token-based, no account)
```

---

## Technical Decisions

### Session Storage: Database-backed
- Survives application restarts
- Can be queried for active sessions
- Supports session invalidation
- Enables audit logging

### Cookie Security
- HttpOnly: Prevents JavaScript access
- Secure: HTTPS only
- SameSite=Lax: CSRF protection
- 7-day expiration (non-sliding)

### Bootstrap Admin
- First authenticated user becomes admin
- Subsequent users are event managers
- Admins can promote users to admin

---

## Security Considerations

### OIDC Security
- ID token signature validation
- Nonce validation (CSRF protection)
- State parameter validation
- Token expiration checking

### Session Security
- Cryptographically random session IDs (32 bytes)
- Constant-time session lookup
- Automatic cleanup of expired sessions
- IP and user agent tracking

### Forward Auth Security
- Headers must come from trusted proxy
- Validate proxy IP address
- Headers cannot be spoofed by clients
- Email format validation

---

## References

- **HLD:** Section 4 (Authentication & Authorization), Section 3 (User Roles)
- **LLD:** [`lld/01_AUTH_LLD.md`](../lld/01_AUTH_LLD.md)
- **Libraries:** 
  - `github.com/coreos/go-oidc/v3/oidc`
  - `golang.org/x/oauth2`

---

## Testing Strategy

### Unit Tests
- OIDC token validation
- Session creation/validation
- Permission checking logic
- User role assignment

### Integration Tests
- Full OIDC flow with test provider
- Forward auth with mock headers
- Session persistence across requests
- Permission enforcement on endpoints

### Manual Tests
1. Set up Keycloak/Authentik locally
2. Configure OIDC settings
3. Test login flow
4. Verify first user is admin
5. Test logout
6. Test session expiration

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| OIDC provider unavailable | High | Clear error message, retry mechanism |
| Session hijacking | High | Secure cookies, HTTPS required, IP tracking |
| First user not admin | High | Clear documentation, admin promotion tool |
| Token validation failure | Medium | Detailed error logging, fallback to login |

---

## Definition of Done

- [ ] All user stories complete
- [ ] Both OIDC and forward auth modes working
- [ ] First user becomes admin automatically
- [ ] Sessions persist across restarts
- [ ] RBAC middleware enforces all permissions
- [ ] All tests passing
- [ ] Security review completed
- [ ] Documentation updated
