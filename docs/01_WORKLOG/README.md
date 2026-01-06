# Worklog

## Purpose

This folder contains progress updates, handoff documents, and session summaries for the TinyRSVP project. Each worklog entry documents what was accomplished, decisions made, blockers encountered, and context for future sessions.

## Rules

1. **Create a worklog entry after significant work** - Don't wait until end of day
2. **Use date-based naming** - `YYYY-MM-DD_description.md`
3. **Be specific and factual** - What was done, not what was attempted
4. **Document decisions** - Why choices were made
5. **Note blockers** - What's preventing progress
6. **Include next steps** - What should happen next
7. **Update regularly** - Better to have many small entries than one large one

## Naming Convention

```
YYYY-MM-DD_short-description.md
```

**Examples:**
- `2026-01-06_initial-setup.md`
- `2026-01-07_auth-middleware-implementation.md`
- `2026-01-08_database-schema-migration.md`

## Entry Template

```markdown
# Worklog: [Short Description]

**Date:** YYYY-MM-DD  
**Session ID:** [Optional - for tracking multiple sessions per day]  
**Duration:** [Optional - approximate time spent]

## Summary

Brief 1-2 sentence summary of what was accomplished.

## Work Completed

- [ ] Task 1 - Description of what was done
- [ ] Task 2 - Description of what was done
- [ ] Task 3 - Description of what was done

## Decisions Made

### Decision 1: [Title]
**Context:** Why this decision was needed  
**Options Considered:** 
- Option A - pros/cons
- Option B - pros/cons

**Decision:** What was chosen and why

### Decision 2: [Title]
...

## Blockers

- **Blocker 1:** Description and impact
- **Blocker 2:** Description and impact

## Next Steps

1. Next immediate task
2. Following task
3. Future consideration

## Files Changed

- `path/to/file1.go` - What changed
- `path/to/file2.go` - What changed
- `docs/something.md` - What changed

## Tests

- [ ] All tests passing
- [ ] New tests added for: [feature]
- [ ] Coverage: XX%

## Notes

Any additional context, learnings, or observations.

## References

- Related backlog story: `docs/00_BACKLOG/XX_story.md`
- Related design doc: `docs/XX_design.md`
- External references: [links]
```

## Index

Worklog entries in chronological order:

| Date | Description | Key Changes |
|------|-------------|-------------|
| 2026-01-06 | Initial setup | Created project structure, README-LLM.md |
| 2026-01-06 | HLD design review | Comprehensive adversarial review, identified 50+ gaps |
| 2026-01-06 | HLD revision | Complete HLD rewrite addressing all gaps, ready for implementation |

## Maintenance

- **Update index** when creating new entries
- **Archive old entries** (>90 days) to `archive/` subfolder if needed
- **Reference from backlog** stories when completing work
