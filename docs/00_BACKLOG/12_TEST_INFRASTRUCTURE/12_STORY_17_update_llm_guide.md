# User Story: Update README-LLM with Testing Guidelines

**Epic:** [12_EPIC_test_infrastructure.md](../12_EPIC_test_infrastructure.md)
**Priority:** High
**Status:** Not Started
**Estimated Effort:** 30 minutes
**Phase:** 4 - Cleanup & Documentation

---

## User Story

As an **AI assistant implementing TinyRSVP**, I want **testing guidelines in README-LLM.md** so that **I follow consistent testing patterns**.

---

## Acceptance Criteria

- [ ] Testing section added to README-LLM.md
- [ ] TDD workflow documented
- [ ] Mock usage guidelines
- [ ] Reference to TESTING.md

---

## Content to Add

```markdown
## Testing Guidelines

### Test-Driven Development (TDD)
**HARD RULE**: Always write tests BEFORE implementation.

Workflow:
1. Write test (should fail)
2. Write minimal code to pass
3. Refactor if needed
4. Repeat

### Using Mocks
All major interfaces have generated mocks in `internal/testutil/mocks/`.

Setup:
[Example with gomock]

Expectations:
[Example]

Regenerate mocks after interface changes:
\`\`\`bash
./scripts/generate_mocks.sh
\`\`\`

### Test Utilities
Use `internal/testutil` for common operations:
- testutil.StringPtr()
- testutil.SetupTestDB()
- testutil.CreateTestUser()
- testutil.CreateAdminContext()

### When to Mock vs Real
**Unit Tests:** Mock all dependencies
**Integration Tests:** Real DB, mock external services
**E2E Tests:** Real everything (except external services)

See `docs/TESTING.md` for complete guidelines.
```

---

## Dependencies

**Depends on:** Story 15 (TESTING.md exists to reference)
