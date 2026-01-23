# Epic 12: Test Infrastructure Modernization - Stories Index

This directory contains all user stories for Epic 12: Test Infrastructure Modernization.

## Epic Overview

**Goal:** Eliminate 92+ manual mock definitions, consolidate 18+ duplicate pointer helpers, and establish a consistent, maintainable testing infrastructure.

**Estimated Effort:** 20-26 hours (3-4 weeks)

**Status:** Not Started

---

## User Stories by Phase

### Phase 1: Foundation (4-6 hours)

| Story | Title | Effort | Status |
|-------|-------|--------|--------|
| [01](12_STORY_01_testutil_package.md) | Create testutil Package Structure | 30 min | Not Started |
| [02](12_STORY_02_pointer_helpers.md) | Centralize Pointer Helpers | 30 min | Not Started |
| [03](12_STORY_03_database_helpers.md) | Database Test Helpers | 1 hour | Not Started |
| [04](12_STORY_04_context_helpers.md) | Auth Context Helpers | 30 min | Not Started |

**Phase 1 Total: 2.5 hours**

### Phase 2: Mock Generation Setup (2-3 hours)

| Story | Title | Effort | Status |
|-------|-------|--------|--------|
| [05](12_STORY_05_mockgen_setup.md) | Install and Configure mockgen | 30 min | Not Started |
| [06](12_STORY_06_priority1_mocks.md) | Generate Priority 1 Repository Mocks | 1 hour | Not Started |
| [07](12_STORY_07_service_mocks.md) | Generate Service Mocks | 30 min | Not Started |
| [08](12_STORY_08_remaining_mocks.md) | Generate Remaining Mocks | 30 min | Not Started |

**Phase 2 Total: 2.5 hours**

### Phase 3: Migration with Validation (8-12 hours)

| Story | Title | Effort | Status |
|-------|-------|--------|--------|
| [09](12_STORY_09_validate_pattern.md) | Migrate 3 Example Files to Validate | 2 hours | Not Started |
| [10](12_STORY_10_reflect_adjust.md) | Review and Adjust Approach | 1 hour | Not Started |
| [11](12_STORY_11_migrate_handlers.md) | Migrate Handler Tests (~81 files) | 4-5 hours | Not Started |
| [12](12_STORY_12_migrate_services.md) | Migrate Service Tests (~30 files) | 2-3 hours | Not Started |
| [13](12_STORY_13_migrate_repos.md) | Migrate Repository Tests | 1 hour | Not Started |

**Phase 3 Total: 10-12 hours**

### Phase 4: Cleanup & Documentation (2-3 hours)

| Story | Title | Effort | Status |
|-------|-------|--------|--------|
| [14](12_STORY_14_cleanup_old_mocks.md) | Remove Manual Mocks and Duplicates | 1 hour | Not Started |
| [15](12_STORY_15_testing_docs.md) | Create TESTING.md | 1 hour | Not Started |
| [16](12_STORY_16_testutil_readme.md) | Document testutil Package | 30 min | Not Started |
| [17](12_STORY_17_update_llm_guide.md) | Update README-LLM with Testing | 30 min | Not Started |

**Phase 4 Total: 3 hours**

### Phase 5: Advanced Features (4-6 hours)

| Story | Title | Effort | Status |
|-------|-------|--------|--------|
| [18](12_STORY_18_test_builders.md) | Create Test Data Builders | 2-3 hours | Not Started |
| [19](12_STORY_19_http_helpers.md) | HTTP Test Helpers | 1-2 hours | Not Started |
| [20](12_STORY_20_fixture_files.md) | Test Fixture Files | 1 hour | Not Started |

**Phase 5 Total: 4-6 hours**

---

## Total Effort Summary

| Phase | Hours | Percentage |
|-------|-------|------------|
| Phase 1: Foundation | 2.5 | 11% |
| Phase 2: Mock Generation | 2.5 | 11% |
| Phase 3: Migration | 10-12 | 48% |
| Phase 4: Cleanup & Docs | 3 | 13% |
| Phase 5: Advanced | 4-6 | 20% |
| **TOTAL** | **22-26** | **100%** |

---

## Critical Path

The critical path through the epic:

```
Story 01 → Story 02, 03, 04 (parallel) → 
Story 05 → Stories 06, 07, 08 (parallel) →
Story 09 → Story 10 (CHECKPOINT) →
Stories 11, 12, 13 (parallel) →
Story 14 → Stories 15, 16, 17 (parallel) →
Stories 18, 19, 20 (parallel, optional)
```

**Checkpoint at Story 10:** Decision point to PROCEED, ADJUST, or ABORT full migration.

---

## Success Metrics

**Before Implementation:**
- 92 manual mock definitions
- 18 duplicate pointer helpers
- ~15,000 lines of test infrastructure code
- ~2 hours to add new interface method

**After Implementation:**
- 0 manual mock definitions
- 1 centralized testutil package
- ~2,000 lines of test infrastructure code
- ~5 minutes to add new interface method

**Expected Reduction:** 87% in test infrastructure code

---

## Dependencies

**Epic Dependencies:**
- **Depends on:** None (can run in parallel with other work)
- **Blocks:** None (improves developer experience for future work)

**Story Dependencies:**
- Stories 01-04: Foundation (sequential within phase, parallel with each other)
- Stories 05-08: Require Story 01, can run parallel with each other
- Story 09: Requires Stories 06-08 (all mocks generated)
- Story 10: Requires Story 09 (CRITICAL CHECKPOINT)
- Stories 11-13: Require Story 10 decision to PROCEED
- Story 14: Requires Stories 11-13 complete
- Stories 15-17: Require Story 14
- Stories 18-20: Require Story 17, can run parallel

---

## Notes

### Validation Checkpoint (Story 10)

After Story 09, we pause to evaluate:
- Does the pattern work as expected?
- Are tests more readable?
- Do we need additional helpers?
- Should we proceed with full migration?

**Possible Outcomes:**
- ✅ **PROCEED** - Continue with Stories 11-20
- 🔄 **ADJUST** - Refine approach, create helpers, re-validate
- ❌ **ABORT** - Revert and reconsider strategy

### Phase 5 Optional

Phase 5 (Stories 18-20) provides nice-to-have features:
- Test data builders
- HTTP helpers
- Fixture files

These can be deferred if time is constrained. Phases 1-4 provide the core value (87% code reduction).

---

## Quick Start

To begin Epic 12 implementation:

1. Start with **Story 01**: Create testutil package structure
2. Complete **Phase 1** (Stories 01-04): Foundation
3. Complete **Phase 2** (Stories 05-08): Mock generation
4. **CRITICAL**: Complete Story 09-10 and evaluate results
5. If proceeding, complete remaining phases

---

## Questions?

See the main [Epic document](../12_EPIC_test_infrastructure.md) for complete details including:
- Technical overview
- Current state problems
- Proposed architecture
- Testing philosophy
- Risk mitigation
- Definition of done
