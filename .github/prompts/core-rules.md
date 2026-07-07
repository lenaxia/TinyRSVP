## Core Rules

These rules apply to every response. They are non-negotiable. They are summarized here for the AI workflow; the authoritative source is README-LLM.md (read it in full before making changes).

### 1. Test-Driven Development (TDD) — README-LLM.md Rule 0

Write tests BEFORE writing functional code. Always.

1. Write test
2. Run test (must fail)
3. Write minimal code to pass
4. Run test (must pass)
5. Refactor if needed

Every code change must include: multiple happy-path tests, multiple unhappy-path tests, and edge cases.

**Full-repo validation is mandatory at the end of every task:**
```bash
go build ./...                              # ALL packages must build
go test -timeout 30s ./...                   # ALL tests must pass (zero tolerance for failures, including pre-existing)
go test -timeout 30s -race ./...             # race detector
go vet ./...                                 # static analysis
```

### 2. Type Safety First — README-LLM.md Rule 1

- Define strongly-typed structs for ALL data structures
- NEVER use `map[string]interface{}` for structured data
- NEVER use `interface{}` when you know the type
- NEVER use type assertions when you can use generics
- NEVER pass untyped data between functions

Maps are ONLY acceptable for parsing external JSON/YAML with unknown structure (convert to struct immediately).

### 3. Idiomatic Go — README-LLM.md Rule 2

- Follow Go conventions, not Perl patterns
- Use Go's multiple return values (value, error) pattern
- Avoid global state and exceptions
- Create custom error types for domain-specific errors
- Prefer minimal or no concurrency when possible

### 4. Zero Technical Debt — README-LLM.md Rule 8

- No TODOs, FIXMEs, or commented-out code
- No adapters for backwards compatibility — implement the final solution
- Never hack tests to pass — fix the root cause
- Pre-existing errors are not acceptable — fix them when encountered

### 5. No Unverified Claims

Never state that something exists without showing it (file path + source text). Never state that something does not exist without proving its absence (show the grep command and its empty output).

### 6. No Comments in Code — README-LLM.md Rule 7

Code should be self-documenting through clear naming. Exceptions: package-level doc comments are required on every package. Do NOT add inline comments unless absolutely necessary and timeless.

### 7. Communication Tone — README-LLM.md Rule 6

Always be neutral, factual, objective. Do NOT be sensational, overly agreeable, or a sycophant. Don't be a cheerleader, be a critical collaborator. Never agree with something just because the user stated it.

### 8. Work Logs Are Mandatory

Every significant task MUST create a work log at `docs/01_WORKLOG/NNNN_YYYY-MM-DD_description.md`. A task is NOT complete without a work log.

### 9. No Destructive Git Operations

Multiple agents may work in this repository simultaneously. NEVER run `git checkout .`, `git reset --hard`, or `git clean -fd`. Revert files one at a time with explicit confirmation. Always check `git status` first.

### 10. Status Documentation When Complete

When marking stories/epics complete:
- Run all tests
- Document test pass rate
- Document known issues
- Document confidence level
- Document production readiness
