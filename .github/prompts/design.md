You are iterating on a **design document** for the TinyRSVP repository — the step that comes *before* `/implement` or `/fix`. The goal is a reviewed, approved design, not code.

Output target: a design document under `docs/02_DESIGN/` (for cross-cutting/architectural work) or `docs/00_BACKLOG/EPIC-XX/` (for epic- or story-scoped work), following the repository's existing conventions.

Rules:
1. Read README-LLM.md first — especially the architecture section, the project rules, and the technology stack. Read any existing doc that touches the same area before writing.
2. Decide where the design lives:
   - Cross-cutting / architectural → a new file in `docs/02_DESIGN/` named descriptively, or an in-place edit to `docs/02_DESIGN/02_REVISED_HLD.md`.
   - Story- or epic-scoped → the relevant `docs/00_BACKLOG/` directory.
   - Updating an existing design → edit it in place; do not silently duplicate.
3. Scope the design to the request text from the collaborator. If the request is ambiguous, state the ambiguity explicitly and pick the narrowest reasonable scope.
4. A design doc must cover at minimum: problem statement, goals/non-goals, proposed design, alternatives considered, data-flow / component interactions, failure-mode analysis, and open questions.
5. State assumptions up front (with confidence levels) and validate each one against source/tests before relying on it.
6. Workflow — follow the Code Change Workflow: feature branch (`design/` or `docs/` prefix), open a PR, iterate through the automated review until it posts APPROVE.
7. **MERGE HOLD — this command never auto-merges.** After the automated review posts APPROVE, STOP. Do not merge. Post a comment on the PR summarising the design and stating it is approved and awaiting an explicit `/merge`.
8. Do not write production code in this step — only the design document and supporting diagrams/tables.
