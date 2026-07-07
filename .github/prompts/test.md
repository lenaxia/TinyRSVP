You are writing or improving tests for the TinyRSVP repository.

**Read README-LLM.md first** — TDD is mandatory (Rule 0).

Rules:
1. Follow the project's testing requirements exactly:
   - Multiple happy-path tests
   - Multiple unhappy-path tests (errors, invalid inputs, boundary failures)
   - Edge case coverage
   - Integration tests that exercise the full handler/service/repo path where applicable
2. Use table-driven tests following existing patterns in the codebase.
3. Use generated mocks from `internal/testutil/mocks/` for interface mocking. For packages with import cycle constraints, use the func-field pattern.
4. Use `internal/testutil` helpers for common operations (pointer helpers, auth contexts, test DB setup) — do not duplicate.
5. All tests must pass with `-race` flag: `go test -timeout 30s -race ./...`
6. Run full-repo validation before pushing — zero failures required:
   ```bash
   go build ./...
   go test -timeout 30s ./...
   go vet ./...
   ```
7. For new test files, follow the naming convention: `*_test.go` in the same package.
8. Check existing test files for patterns and utilities before writing new ones.
9. Create a work log in `docs/01_WORKLOG/`.
