# User Story: Test Fixture Files

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** Low
**Status:** Not Started
**Estimated Effort:** 1 hour
**Phase:** 5 - Advanced Features

---

## User Story

As a **developer**, I want **test fixture files** so that **I can load consistent test data across multiple tests**.

---

## Acceptance Criteria

- [ ] `testdata/` directory created
- [ ] Fixture files for events, invites, users, RSVPs
- [ ] `internal/testutil/fixtures.go` with load functions
- [ ] Documentation with examples

---

## Implementation

### Directory Structure

```
testdata/
├── events.json          # Sample events
├── invites.json         # Sample invites
├── users.json           # Sample users
└── rsvps.json           # Sample RSVPs
```

### Fixture Files

**testdata/events.json:**
```json
{
  "draft_event": {
    "title": "Draft Event",
    "slug": "draft-event",
    "status": "draft",
    "description": "A draft event for testing",
    "start_time": "2026-12-01T18:00:00Z"
  },
  "published_event": {
    "title": "Published Event",
    "slug": "published-event",
    "status": "published",
    "start_time": "2026-06-15T19:00:00Z"
  }
}
```

### Load Functions

```go
package testutil

func LoadEventFixture(t *testing.T, key string) *models.Event {
    // Load from testdata/events.json
}

func LoadInviteFixture(t *testing.T, key string) *models.Invite {
    // Load from testdata/invites.json
}

func LoadUserFixture(t *testing.T, key string) *models.User {
    // Load from testdata/users.json
}
```

---

## Usage Example

```go
func TestWithFixtures(t *testing.T) {
    event := testutil.LoadEventFixture(t, "published_event")
    invite := testutil.LoadInviteFixture(t, "standard_invite")
    // Use in test...
}
```

---

## Tasks

- [ ] Create testdata/ directory
- [ ] Create fixture JSON files
- [ ] Implement load functions
- [ ] Write tests for loaders
- [ ] Document fixture structure

---

## Dependencies

**Depends on:** Story 17

---

## Benefits

- Consistent test data across tests
- Easy to maintain (edit JSON files)
- Reusable across multiple test files
- Can version control test scenarios
