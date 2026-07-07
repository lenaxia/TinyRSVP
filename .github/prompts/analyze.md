You are performing a deep analysis of the TinyRSVP codebase. This is a READ-ONLY task — do not make any code changes.

**Read README-LLM.md first** for full architectural context.

Rules:
1. Read README-LLM.md for the project architecture (Chi router, SQLite/PostgreSQL, OIDC + Forward Auth, token-based guest access, async email queue, template system).
2. Read `docs/02_DESIGN/02_REVISED_HLD.md` and relevant `docs/02_DESIGN/lld/` items as needed.
3. Be specific — reference file paths, function names, and type names. Do NOT reference line numbers (they drift).
4. If you find bugs or design flaws, describe them precisely with reproduction steps or code references.
5. Do not create branches, PRs, or make any file changes.
6. If the analysis reveals issues that should be fixed, suggest using `/fix` or `/implement` in your response.

Output format:
## Analysis

### Topic
[What was analyzed]

### Findings
[Detailed findings with code references]

### Recommendations
[Suggested actions, if any — reference appropriate commands like `/fix` or `/implement`]
