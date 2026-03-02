# Epic: Test Infrastructure Modernization

**Epic ID:** 12
**Priority:** High  
**Status:** Not Started  
**Target Version:** v0  
**Estimated Effort:** 3-4 weeks (20-26 hours)

---

## Overview

Modernize TinyRSVP's test infrastructure by eliminating massive duplication in test code, implementing generated mocks with gomock, and creating centralized test utilities. This epic will improve developer productivity, reduce maintenance burden, and establish clear testing patterns.

**Goal:** Eliminate 92+ manual mock definitions, consolidate 18+ duplicate pointer helpers, and establish a consistent, maintainable testing infrastructure that scales with the codebase.

---

## Success Criteria

- [ ] All 92+ manual mock definitions replaced with generated mocks
- [ ] 18+ duplicate pointer helper functions consolidated to single location
- [ ] ~30+ generated mocks for all major interfaces (repositories, services)
- [ ] All 236 test files migrated to use centralized utilities
- [ ] 87% reduction in test infrastructure code (15k → 2k lines)
- [ ] Comprehensive testing documentation (TESTING.md)
- [ ] Test data builders for complex objects (Event, Invite, User, RSVP)
- [ ] All tests passing after migration
- [ ] Clear guidelines for when to mock vs use real implementations

---

## User Stories

### Phase 1: Foundation (4-6 hours)
- [x] [`12_TEST_INFRASTRUCTURE/12_STORY_01_testutil_package.md`](12_TEST_INFRASTRUCTURE/12_STORY_01_testutil_package.md) - Create testutil package structure ✅
- [x] [`12_TEST_INFRASTRUCTURE/12_STORY_02_pointer_helpers.md`](12_TEST_INFRASTRUCTURE/12_STORY_02_pointer_helpers.md) - Centralize pointer helpers ✅
- [x] [`12_TEST_INFRASTRUCTURE/12_STORY_03_database_helpers.md`](12_TEST_INFRASTRUCTURE/12_STORY_03_database_helpers.md) - Database test helpers ✅
- [x] [`12_TEST_INFRASTRUCTURE/12_STORY_04_context_helpers.md`](12_TEST_INFRASTRUCTURE/12_STORY_04_context_helpers.md) - Auth context helpers ✅

### Phase 2: Mock Generation Setup (2-3 hours)
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_05_mockgen_setup.md`](12_TEST_INFRASTRUCTURE/12_STORY_05_mockgen_setup.md) - Install and configure mockgen
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_06_priority1_mocks.md`](12_TEST_INFRASTRUCTURE/12_STORY_06_priority1_mocks.md) - Generate core repository mocks
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_07_service_mocks.md`](12_TEST_INFRASTRUCTURE/12_STORY_07_service_mocks.md) - Generate service mocks
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_08_remaining_mocks.md`](12_TEST_INFRASTRUCTURE/12_STORY_08_remaining_mocks.md) - Generate utility mocks

### Phase 3: Migration with Validation (8-12 hours)
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_09_validate_pattern.md`](12_TEST_INFRASTRUCTURE/12_STORY_09_validate_pattern.md) - Migrate 3 example files to validate
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_10_reflect_adjust.md`](12_TEST_INFRASTRUCTURE/12_STORY_10_reflect_adjust.md) - Review and adjust approach
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_11_migrate_handlers.md`](12_TEST_INFRASTRUCTURE/12_STORY_11_migrate_handlers.md) - Migrate handler tests (~81 files)
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_12_migrate_services.md`](12_TEST_INFRASTRUCTURE/12_STORY_12_migrate_services.md) - Migrate service tests (~30 files)
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_13_migrate_repos.md`](12_TEST_INFRASTRUCTURE/12_STORY_13_migrate_repos.md) - Migrate repository tests

### Phase 4: Cleanup & Documentation (2-3 hours)
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_14_cleanup_old_mocks.md`](12_TEST_INFRASTRUCTURE/12_STORY_14_cleanup_old_mocks.md) - Remove manual mocks
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_15_testing_docs.md`](12_TEST_INFRASTRUCTURE/12_STORY_15_testing_docs.md) - Create TESTING.md
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_16_testutil_readme.md`](12_TEST_INFRASTRUCTURE/12_STORY_16_testutil_readme.md) - Document testutil package
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_17_update_llm_guide.md`](12_TEST_INFRASTRUCTURE/12_STORY_17_update_llm_guide.md) - Update README-LLM.md

### Phase 5: Advanced Features (4-6 hours)
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_18_test_builders.md`](12_TEST_INFRASTRUCTURE/12_STORY_18_test_builders.md) - Create test data builders
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_19_http_helpers.md`](12_TEST_INFRASTRUCTURE/12_STORY_19_http_helpers.md) - HTTP test helpers
- [ ] [`12_TEST_INFRASTRUCTURE/12_STORY_20_fixture_files.md`](12_TEST_INFRASTRUCTURE/12_STORY_20_fixture_files.md) - Test fixture files

---

## Dependencies

**Depends on:** None (can run in parallel with other work)  
**Blocks:** None (improves developer experience for all future work)

---

## Technical Overview

### Current State Problems

**Mock Proliferation:**
- 92+ manual mock definitions across 52 test files
- InviteService mocked 26+ different times with inconsistent behavior
- EventRepository mocked 14+ times with different implementations
- Each interface change requires updating 26+ mock definitions

**Code Duplication:**
- 18+ duplicate pointer helper functions (stringPtr, intPtr, etc.)
- 23+ duplicate setupTestDB functions
- ~15,000 lines of duplicate test infrastructure code

**Inconsistent Patterns:**
- Some mocks return nil on unimplemented methods
- Some mocks return errors
- Some mocks return typed domain errors
- No standard for mock behavior

### Proposed Architecture

```
internal/testutil/
├── doc.go                    # Package documentation
├── pointers.go               # StringPtr, IntPtr, BoolPtr, TimePtr
├── database.go               # SetupTestDB, CreateTestUser, CreateTestEvent
├── context.go                # CreateAdminContext, CreateTestContext
├── http.go                   # HTTP request/response builders (Phase 5)
├── fixtures.go               # Load fixture files (Phase 5)
├── mocks/                    # Generated mocks (~21 files)
│   ├── mock_database.go
│   ├── mock_event_repository.go
│   ├── mock_invite_repository.go
│   └── ... (30+ interface mocks)
└── builders/                 # Test data builders (Phase 5)
    ├── event_builder.go
    ├── invite_builder.go
    ├── user_builder.go
    └── rsvp_builder.go
```

### Mock Generation Strategy

**Tool:** `mockgen` (official Go mock generation tool)
- Zero runtime dependencies
- Integrates with gomock for call verification
- Generates consistent, complete mocks
- Widely used in Go community

**Generated Mocks:**
1. Database interfaces (5 mocks)
2. Repository interfaces (10 mocks)
3. Service interfaces (5 mocks)
4. Utility interfaces (5 mocks)

**Total:** ~21 generated files covering 30+ interface variations

### Migration Pattern

**Before (Manual Mock):**
```go
type mockInviteService struct {
    getInviteByIDFunc func(ctx context.Context, id int64) (*models.Invite, error)
}

func (m *mockInviteService) GetInviteByID(ctx context.Context, id int64) (*models.Invite, error) {
    if m.getInviteByIDFunc != nil {
        return m.getInviteByIDFunc(ctx, id)
    }
    return nil, errors.New("not implemented")
}
```

**After (Generated Mock with gomock):**
```go
ctrl := gomock.NewController(t)
defer ctrl.Finish()

mockService := mocks.NewMockInviteService(ctrl)
mockService.EXPECT().
    GetInviteByID(gomock.Any(), int64(1)).
    Return(&models.Invite{ID: 1}, nil)
```

**Benefits:**
- Explicit expectations vs implicit function fields
- Call count verification
- Argument matching
- Better error messages
- Consistent behavior

---

## Testing Philosophy

### Test Categories

**Unit Tests:**
- Purpose: Test single component in isolation
- Speed: < 10ms per test
- Dependencies: All mocked
- Pattern: Use generated mocks

**Integration Tests:**
- Purpose: Test components working together
- Speed: < 100ms per test  
- Dependencies: Real database, mocked external services
- Pattern: Real DB + generated mocks for external deps

**E2E Tests:**
- Purpose: Test complete workflows
- Speed: < 1s per test
- Dependencies: Full stack
- Pattern: Real everything (except external services)

### When to Mock

**ALWAYS Mock:**
- External services (SMTP, OIDC)
- File system operations
- Network calls
- Slow operations

**SOMETIMES Mock:**
- Database (mock for unit tests, real for integration)
- Repositories (mock in service tests)
- Services (mock in handler tests)

**NEVER Mock:**
- Simple validators
- Pure functions
- The code under test

---

## Impact Analysis

### Code Reduction

**Before:**
- 92 manual mock definitions
- 18 duplicate pointer helper functions
- 23 duplicate database setup functions
- ~15,000 lines of test infrastructure code

**After:**
- 21 generated mock files
- 1 centralized testutil package
- ~2,000 lines of test utilities
- **87% reduction** in test infrastructure code

### Developer Experience

**Time Savings:**
- Adding new interface method: ~2 hours → ~5 minutes (regenerate mocks)
- Writing new test: ~30 minutes → ~10 minutes (reuse helpers)
- Fixing broken tests: ~20 minutes → ~5 minutes (clearer errors)

**Quality Improvements:**
- Consistent mock behavior
- Explicit test expectations
- Better test readability
- Easier refactoring

---

## Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Breaking existing tests | High | Medium | Migrate incrementally, file-by-file |
| Pattern doesn't work | High | Low | Validate with 3 examples first |
| Team learning curve | Medium | High | Comprehensive docs, examples |
| Time estimate too low | Medium | Medium | Phase 5 optional, can defer |
| Migration takes too long | Medium | Low | Batch migrations, commit frequently |

---

## Definition of Done

- [ ] All user stories complete
- [ ] All 236 test files passing
- [ ] Zero manual mock definitions remain
- [ ] Centralized testutil package with all helpers
- [ ] 21+ generated mocks for all interfaces
- [ ] TESTING.md documentation complete
- [ ] README-LLM.md updated with testing section
- [ ] Test data builders implemented (Phase 5)
- [ ] HTTP helpers implemented (Phase 5)
- [ ] Fixture files created (Phase 5)
- [ ] All tests passing with improved error messages
- [ ] Code review complete
- [ ] Team trained on new patterns

---

## References

- **Go Testing Best Practices**: https://go.dev/doc/tutorial/add-a-test
- **gomock Documentation**: https://github.com/golang/mock
- **mockgen Guide**: https://pkg.go.dev/go.uber.org/mock/mockgen
- **TinyRSVP Testing Analysis**: See exploration session 2026-01-22

---

## Notes

This epic can run in parallel with other feature work. It improves the developer experience for all future development but doesn't block any functionality.

The work is structured with a validation checkpoint after Phase 2 (Story 09-10) - if the pattern doesn't work as expected, we can adjust the approach before committing to full migration.

Phase 5 (Advanced Features) is optional and can be deferred if time is constrained. Phases 1-4 provide the majority of the value (87% code reduction).

---

## Metrics to Track

**Before Implementation:**
- Manual mock definitions: 92
- Duplicate pointer helpers: 18
- Test infrastructure LOC: ~15,000
- Time to add interface method: ~2 hours

**After Implementation:**
- Manual mock definitions: 0
- Duplicate pointer helpers: 0
- Test infrastructure LOC: ~2,000
- Time to add interface method: ~5 minutes

**Success Measurement:**
- All tests passing: ✓/✗
- Code reduction: [percentage]
- Developer feedback: [positive/negative]
- Time to write new test: [before/after comparison]
