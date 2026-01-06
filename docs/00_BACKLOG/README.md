# Backlog

## Purpose

This folder contains sprint stories, epics, and user stories for the TinyRSVP project. Stories are organized by priority and implementation order. Tasks are defined within user story files using checklists.

## Rules

1. **Epics only** - No individual tasks at this level
2. **User stories within epics** - Each epic contains multiple user stories
3. **Tasks within stories** - Use checklists `[ ]` within user story files
4. **Update status regularly** - Mark completed items `[x]`
5. **Keep stories small** - User stories should be completable in 1-3 sessions
6. **Reference from worklog** - Link completed work back to stories

## Structure

```
00_BACKLOG/
├── README.md (this file)
├── 00_EPIC_foundation.md
├── 01_EPIC_auth.md
├── 02_EPIC_events.md
├── 03_EPIC_invites.md
├── 04_EPIC_rsvp.md
├── 05_EPIC_email.md
└── 06_EPIC_templates.md
```

## Story Format

### Epic File Format

```markdown
# Epic: [Epic Name]

**Priority:** High | Medium | Low  
**Status:** Not Started | In Progress | Complete  
**Target Version:** v0 | v1 | Future

## Overview

Brief description of the epic and its goals.

## Success Criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## User Stories

### Story 1: [Story Title]

**As a** [role]  
**I want** [goal]  
**So that** [benefit]

**Status:** Not Started | In Progress | Complete  
**Priority:** High | Medium | Low

#### Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2

#### Tasks
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

#### Technical Notes
- Implementation details
- Dependencies
- Considerations

---

### Story 2: [Story Title]
...
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
