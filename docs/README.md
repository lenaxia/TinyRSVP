# Documentation

## Purpose

This folder contains all project documentation including design documents, specifications, backlog stories, and progress worklogs.

## Rules

1. **Read before editing** - Understand existing docs first
2. **Design docs are numbered** - Use `XX_` prefix (00, 01, 02, etc.)
3. **HLD is authoritative** - [`00_INITIAL_HLD.md`](00_INITIAL_HLD.md) is source of truth
4. **Update regularly** - Keep docs in sync with code
5. **Reference from code** - Link to relevant docs in comments (when necessary)

## Structure

```
docs/
├── README.md                    # This file
├── 00_INITIAL_HLD.md           # High-level design (AUTHORITATIVE)
├── 00_BACKLOG/                 # Sprint stories and epics
│   └── README.md
├── 01_WORKLOG/                 # Progress updates and handoffs
│   └── README.md
└── XX_future_design.md         # Future design documents
```

## Document Types

### Design Documents

**Naming:** `XX_DOCUMENT_NAME.md` where XX is creation order

**Purpose:**
- High-level designs
- Architecture decisions
- Technical specifications
- API designs
- Database schemas

**When to Create:**
- Before implementing major features
- When making architectural decisions
- When defining new subsystems
- When documenting complex algorithms

**Current Design Docs:**
- [`00_INITIAL_HLD.md`](00_INITIAL_HLD.md) - Complete project specification

### Backlog

**Location:** [`00_BACKLOG/`](00_BACKLOG/)

**Purpose:**
- Epics and user stories
- Task tracking
- Priority management

**See:** [`00_BACKLOG/README.md`](00_BACKLOG/README.md)

### Worklog

**Location:** [`01_WORKLOG/`](01_WORKLOG/)

**Purpose:**
- Progress updates
- Handoff documents
- Session summaries
- Decision logs

**See:** [`01_WORKLOG/README.md`](01_WORKLOG/README.md)

## Documentation Workflow

### Before Starting Work

1. Read [`00_INITIAL_HLD.md`](00_INITIAL_HLD.md) for authoritative spec
2. Check [`00_BACKLOG/`](00_BACKLOG/) for current stories
3. Review [`01_WORKLOG/`](01_WORKLOG/) for recent progress

### During Work

1. Update relevant design docs if architecture changes
2. Update backlog story checklists as tasks complete
3. Create worklog entries for significant progress

### After Completing Work

1. Ensure all docs are up to date
2. Create handoff document in worklog
3. Update backlog story status
4. Commit documentation changes

## Key Documents

### Must Read First

1. **[`00_INITIAL_HLD.md`](00_INITIAL_HLD.md)** - Complete project specification
   - Product scope
   - Auth model
   - Database schema
   - API routes
   - Deployment model

2. **[`../README-LLM.md`](../README-LLM.md)** - LLM implementation guide
   - Critical guidelines
   - Type safety rules
   - TDD requirements
   - Common commands

### Reference Often

- **[`00_BACKLOG/README.md`](00_BACKLOG/README.md)** - Current work items
- **[`01_WORKLOG/README.md`](01_WORKLOG/README.md)** - Recent progress

## Maintenance

### After Each Session

- [ ] Update worklog with progress
- [ ] Update backlog story checklists
- [ ] Update design docs if architecture changed
- [ ] Commit documentation changes

### Weekly

- [ ] Review and update this README
- [ ] Archive old worklog entries (>90 days)
- [ ] Review backlog priorities
- [ ] Ensure all docs are current

## Documentation Standards

### Markdown Style

- Use ATX-style headers (`#` not `===`)
- Use fenced code blocks with language tags
- Use tables for structured data
- Use checklists for tasks `- [ ]`
- Use relative links for internal references

### Naming Conventions

- Design docs: `XX_snake_case_name.md`
- Worklog entries: `YYYY-MM-DD_description.md`
- Backlog epics: `XX_EPIC_name.md`

### Content Guidelines

- Be specific and factual
- Avoid vague language
- Include examples
- Document decisions with rationale
- Keep updated with code

## Quick Reference

### Finding Information

**Need to understand the project?**
→ Read [`00_INITIAL_HLD.md`](00_INITIAL_HLD.md)

**Need to know what to work on?**
→ Check [`00_BACKLOG/`](00_BACKLOG/)

**Need to know what was done recently?**
→ Check [`01_WORKLOG/`](01_WORKLOG/)

**Need to understand LLM guidelines?**
→ Read [`../README-LLM.md`](../README-LLM.md)

### Creating New Documents

**Design Document:**
1. Determine next number (XX)
2. Create `XX_descriptive_name.md`
3. Follow design doc template
4. Update this README

**Worklog Entry:**
1. Use date: `YYYY-MM-DD_description.md`
2. Follow worklog template
3. Update worklog index

**Backlog Story:**
1. Add to appropriate epic file
2. Follow story format
3. Update backlog index

---

**Remember:** Documentation is code. Keep it current, accurate, and useful.
