You are implementing a feature or user story for the TinyRSVP repository.

**Read README-LLM.md first** — it contains the project guidelines (TDD, type safety, zero technical debt, etc.).

Rules:
1. Read README-LLM.md before making any changes — it contains hard rules for TDD, type safety, architecture, and conventions.
2. Read the relevant design document — `docs/02_DESIGN/02_REVISED_HLD.md` (the authoritative spec), or the relevant `docs/00_BACKLOG/` epic/user-story — before starting.
3. Follow TDD: write tests FIRST — they must fail, then implement, then pass. Multiple happy-path + unhappy-path + edge cases.
4. No `map[string]interface{}` for structured data. Use strongly-typed structs. Never use `interface{}` when the type is known.
5. Never perform destructive git operations.
6. Run full-repo validation before pushing — zero failures required:
   ```bash
   go build ./...
   go test -timeout 30s ./...
   go test -timeout 30s -race ./...
   go vet ./...
   ```
7. Create a work log in `docs/01_WORKLOG/`.
8. Leave the codebase in zero-error state — fix any pre-existing errors you encounter.
