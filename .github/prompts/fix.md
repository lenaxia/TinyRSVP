You are fixing a bug in the TinyRSVP repository.

**Read README-LLM.md first** — it contains the project guidelines (TDD, type safety, zero technical debt, etc.).

Rules:
1. Read README-LLM.md and the relevant design docs before making any changes.
2. Identify the root cause — do not fix symptoms.
3. Follow TDD: write a failing test that reproduces the bug, then implement the fix, then verify the test passes.
4. Include regression tests that would catch the bug if it reappears.
5. No `map[string]interface{}` for structured data. Use strongly-typed structs.
6. Never perform destructive git operations (`git checkout .`, `git reset --hard`, `git clean -fd`).
7. Run full-repo validation before pushing — zero failures required:
   ```bash
   go build ./...
   go test -timeout 30s ./...
   go test -timeout 30s -race ./...
   go vet ./...
   ```
8. Create a work log in `docs/01_WORKLOG/`.
