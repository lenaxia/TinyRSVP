# LLM Workflows

## Purpose

This folder contains standardized workflow templates, prompts, and sequences for common LLM tasks. These workflows ensure consistency, quality, and efficiency when performing repetitive development activities.

## Rules

1. **Use workflows for standardized tasks** - Don't reinvent the wheel
2. **Follow workflow steps exactly** - They're designed for quality
3. **Update workflows when improved** - Capture learnings
4. **Create new workflows** - When patterns emerge
5. **Reference from worklog** - Note which workflow was used

## When to Use Workflows

**ALWAYS use a workflow when:**
- Writing design documents
- Creating comprehensive test suites
- Performing code reviews
- Refactoring existing code
- Updating documentation
- Creating new features
- Debugging complex issues

**DON'T use a workflow when:**
- Making trivial changes
- Fixing typos
- Quick experiments
- One-off tasks

## Available Workflows

### Design & Planning

| Workflow | Purpose | When to Use | File |
|----------|---------|-------------|------|
| Design Document | Create technical design docs | Before implementing major features | `design_document.md` |
| Architecture Review | Review system architecture | When making architectural changes | `architecture_review.md` |
| API Design | Design REST/GraphQL APIs | Before implementing new endpoints | `api_design.md` |

### Development

| Workflow | Purpose | When to Use | File |
|----------|---------|-------------|------|
| TDD Implementation | Test-driven development | Implementing any new functionality | `tdd_implementation.md` |
| Feature Implementation | Complete feature workflow | Building new features end-to-end | `feature_implementation.md` |
| Refactoring | Safe code refactoring | Improving existing code | `refactoring.md` |

### Testing

| Workflow | Purpose | When to Use | File |
|----------|---------|-------------|------|
| Test Suite Creation | Comprehensive test writing | Creating tests for new code | `test_suite_creation.md` |
| Test Review | Review existing tests | Ensuring test quality | `test_review.md` |
| Integration Testing | E2E test scenarios | Testing full workflows | `integration_testing.md` |

### Documentation

| Workflow | Purpose | When to Use | File |
|----------|---------|-------------|------|
| README Update | Update README files | After structural changes | `readme_update.md` |
| Code Documentation | Document complex code | For non-obvious implementations | `code_documentation.md` |
| API Documentation | Document API endpoints | After API changes | `api_documentation.md` |

### Maintenance

| Workflow | Purpose | When to Use | File |
|----------|---------|-------------|------|
| Code Review | Review code changes | Before merging branches | `code_review.md` |
| Debugging | Systematic debugging | Investigating issues | `debugging.md` |
| Performance Analysis | Analyze performance | When performance issues arise | `performance_analysis.md` |

## Workflow Template

Each workflow file should follow this structure:

```markdown
# Workflow: [Name]

## Purpose
Brief description of what this workflow accomplishes.

## When to Use
- Scenario 1
- Scenario 2

## When NOT to Use
- Scenario 1
- Scenario 2

## Prerequisites
- [ ] Prerequisite 1
- [ ] Prerequisite 2

## Steps

### Step 1: [Name]
**Goal:** What this step accomplishes

**Actions:**
1. Action 1
2. Action 2

**Output:** What should be produced

**Validation:**
- [ ] Check 1
- [ ] Check 2

---

### Step 2: [Name]
...

## Completion Checklist
- [ ] All steps completed
- [ ] All validations passed
- [ ] Documentation updated
- [ ] Tests passing
- [ ] Changes committed

## Common Pitfalls
- Pitfall 1 and how to avoid
- Pitfall 2 and how to avoid

## Examples
Link to example PRs or commits that used this workflow.

## Related Workflows
- Related workflow 1
- Related workflow 2
```

## Creating New Workflows

**When to create a new workflow:**
1. You've done the same task 3+ times
2. The task has clear, repeatable steps
3. Quality/consistency matters
4. Others will benefit

**How to create:**
1. Document the steps as you work
2. Refine after 2-3 iterations
3. Add to this index
4. Reference in worklog

## Workflow Best Practices

### Before Starting
1. Read the entire workflow
2. Ensure prerequisites are met
3. Understand the goal
4. Gather required inputs

### During Execution
1. Follow steps in order
2. Complete all validations
3. Document deviations
4. Note improvements

### After Completion
1. Verify completion checklist
2. Update workflow if needed
3. Reference in worklog
4. Share learnings

## Maintenance

**Monthly Review:**
- Review workflow usage
- Update based on feedback
- Archive unused workflows
- Create new workflows as needed

**After Each Use:**
- Note any issues
- Suggest improvements
- Update examples

## Quick Reference

### Most Common Workflows

1. **TDD Implementation** - Use for ALL new code
2. **Test Suite Creation** - Use when adding features
3. **README Update** - Use after structural changes
4. **Code Review** - Use before merging

### Workflow Selection Guide

```
Need to implement feature?
  └─> Use: Feature Implementation → TDD Implementation

Need to fix bug?
  └─> Use: Debugging → TDD Implementation

Need to refactor?
  └─> Use: Refactoring → Test Review

Need to document?
  └─> Use: README Update or Code Documentation

Need to design?
  └─> Use: Design Document → Architecture Review
```

## Index Maintenance

**Update this index when:**
- Adding new workflows
- Removing workflows
- Changing workflow purposes
- Reorganizing structure

---

**Note:** Workflows are living documents. Improve them as you learn.
