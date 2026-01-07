# Invites Package

## Purpose

The invites package provides a service layer for managing guest invitations with secure token-based access. It integrates the token generation and hashing system from [`pkg/token`](../../pkg/token/) with the invite data model and repository.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    InviteService                         │
│  - Orchestrates invite creation with token generation   │
│  - Manages invite lifecycle (create, retrieve, revoke)  │
│  - Integrates token.Generator with InviteRepository     │
└─────────────────────────────────────────────────────────┘
                    │                │
                    ▼                ▼
        ┌──────────────────┐  ┌──────────────────┐
        │  token.Generator │  │ InviteRepository │
        │  - Generate()    │  │  - Create()      │
        │  - Hash()        │  │  - GetByToken()  │
        └──────────────────┘  │  - Update()      │
                              └──────────────────┘
```

## Key Components

### InviteService Interface

The service provides the following operations:

- **CreateInvite**: Generate token → Hash token → Store invite → Return invite + plaintext token
- **GetInviteByToken**: Hash token → Retrieve invite by hash
- **GetInviteByID**: Retrieve invite by database ID
- **RevokeInvite**: Update invite status to revoked (with validation)
- **ListInvitesByEventID**: List invites for an event with filtering

### Token Integration

The service integrates with [`pkg/token`](../../pkg/token/) for secure token operations:

1. **Token Generation**: Uses `token.Generator.Generate()` to create cryptographically secure 43-character tokens
2. **Token Hashing**: Uses `token.Generator.Hash()` to create 43-character HMAC-SHA256 hashes
3. **Token Storage**: Only hashes are stored in the database, never plaintext tokens
4. **Token Validation**: Hashes incoming tokens and looks up by hash

### Security Properties

- Tokens have 256 bits of entropy (2^256 possible values)
- HMAC-SHA256 prevents token forgery
- Constant-time comparison prevents timing attacks
- Plaintext tokens never stored in database
- Tokens returned only once during creation

## Usage Examples

### Creating an Invite

```go
import (
    "context"
    "time"
    
    "github.com/lenaxia/tinyrsvp/internal/invites"
    "github.com/lenaxia/tinyrsvp/internal/db/repositories"
    "github.com/lenaxia/tinyrsvp/pkg/token"
)

// Setup
secret := []byte("your-secret-key-32-bytes-long!")
generator := token.NewGenerator(secret)
repo := repositories.NewInviteRepository(database)
service := invites.NewInviteService(generator, repo)

// Create invite
ctx := context.Background()
name := "John Doe"
email := "john@example.com"
expiresAt := time.Now().Add(30 * 24 * time.Hour)

invite, plainToken, err := service.CreateInvite(
    ctx,
    eventID,
    &name,
    &email,
    2, // max plus ones
    expiresAt,
)
if err != nil {
    return err
}

// plainToken should be sent to the guest via email
// invite.TokenHash is stored in the database
```

### Retrieving an Invite by Token

```go
// When guest clicks invite link with token
plainToken := "token-from-url-parameter"

invite, err := service.GetInviteByToken(ctx, plainToken)
if err != nil {
    return err
}

// Check if invite is valid
if invite.Status == models.InviteStatusRevoked {
    return errors.New("invite has been revoked")
}

if time.Now().After(invite.ExpiresAt) {
    return errors.New("invite has expired")
}
```

### Revoking an Invite

```go
// Revoke an invite (prevents further use)
err := service.RevokeInvite(ctx, inviteID)
if err != nil {
    return err
}

// Status transitions are validated:
// - draft → revoked ✓
// - sent → revoked ✓
// - viewed → revoked ✓
// - responded → revoked ✗ (terminal state)
// - revoked → * ✗ (terminal state)
```

### Listing Invites

```go
import "github.com/lenaxia/tinyrsvp/internal/db/repositories"

// List all invites for an event
filters := repositories.InviteFilters{
    Limit: 100,
    Offset: 0,
}

invites, err := service.ListInvitesByEventID(ctx, eventID, filters)
if err != nil {
    return err
}

// Filter by status
status := models.InviteStatusSent
filters.Status = &status

sentInvites, err := service.ListInvitesByEventID(ctx, eventID, filters)
```

## Validation Rules

The service enforces the following validation rules:

### Email Validation
- Must be valid email format (uses `net/mail.ParseAddress`)
- Maximum 255 characters
- Required for invites with status "sent"

### Token Validation
- Token hash must be exactly 43 characters
- Token hash must be unique across all invites
- Generated using cryptographically secure random source

### Business Rules
- Event ID must be positive and reference existing event
- Max plus ones must be between 0 and 10
- Expiration time must be in the future
- Status transitions must follow valid state machine

## Error Handling

The service returns wrapped errors with context:

```go
// Token generation errors
"failed to generate token: %w"

// Token hashing errors
"failed to hash token: %w"

// Repository errors
"failed to create invite: %w"
"failed to get invite: %w"
"failed to update invite: %w"

// Validation errors
models.ValidationError with specific field and message
```

## Testing

### Unit Tests
- [`service_test.go`](service_test.go) - Service layer tests with mocks
- Tests token generation failures
- Tests hashing failures
- Tests repository failures
- Tests validation errors

### Integration Tests
- [`integration_test.go`](integration_test.go) - End-to-end workflow tests
- Tests full invite creation workflow
- Tests token hashing consistency
- Tests multiple invite scenarios
- Tests email validation integration

Run tests:
```bash
go test -timeout 30s ./internal/invites/...
```

## Dependencies

### Required Packages
- [`pkg/token`](../../pkg/token/) - Token generation and hashing
- [`internal/models`](../models/) - Invite data model
- [`internal/db/repositories`](../db/repositories/) - Invite repository

### External Dependencies
- `net/mail` - Email validation
- `database/sql` - Database operations (via repository)

## Performance Considerations

1. **Token Generation**: Uses crypto/rand (system entropy)
2. **Token Hashing**: HMAC-SHA256 is fast (~1μs per operation)
3. **Database Lookups**: Unique index on token_hash enables fast retrieval
4. **Batch Operations**: Not currently supported (use repository directly if needed)

## Security Notes

1. **Secret Key Management**
   - Secret key must be at least 32 bytes
   - Must be randomly generated using crypto/rand
   - Must be kept confidential
   - Should be stored in environment variables or secrets manager

2. **Token Handling**
   - Plaintext tokens returned only during creation
   - Tokens should be transmitted over HTTPS only
   - Tokens should be included in email links
   - Never log plaintext tokens

3. **Hash Storage**
   - Only hashes stored in database
   - Hashes are deterministic (same token → same hash)
   - Different secret keys produce different hashes

## Future Enhancements

Potential additions for future stories:

- Batch invite creation with token generation
- Token regeneration for existing invites
- Invite expiration management
- Invite statistics and analytics
- Email sending integration

## Related Documentation

- [Token Package](../../pkg/token/README.md) - Token generation and hashing
- [Invite Model](../models/invite.go) - Data model and validation
- [Invite Repository](../db/repositories/invite_repository.go) - Database operations
- [Story 03](../../docs/00_BACKLOG/03_STORY_03_invite_model.md) - Original requirements
