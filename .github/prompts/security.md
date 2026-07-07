You are performing a security-focused review of the TinyRSVP codebase.

**Read README-LLM.md first** for security-relevant coding standards.

Rules:
1. Check every one of these areas:
   - **XSS prevention:** Are user-supplied values properly escaped in HTML templates? Go `html/template` auto-escapes, but are there any `template.HTML` or unsafe usages?
   - **CSRF protection:** Are state-changing requests protected by CSRF tokens? Check middleware and forms.
   - **SQL injection:** Are all database queries using parameterized queries? Check for any string concatenation in SQL.
   - **Authentication bypass:** Is the `X-Test-User-ID` header still active in production middleware? (Known issue in `internal/middleware/rbac.go:16`)
   - **Token security:** Are invite tokens properly hashed? Is `TOKEN_SECRET` required at startup?
   - **Rate limiting:** Is rate limiting active on RSVP submission and invite creation endpoints?
   - **Credentials:** Are SMTP credentials, OIDC secrets, or database paths logged anywhere?
   - **Input validation:** Are all API inputs validated before processing? Check handlers for missing validation.
   - **Security headers:** Are security headers set (CSP, HSTS, X-Frame-Options, etc.)?
   - **File upload:** Are uploaded images validated for type and size?
2. If code changes are needed to fix security issues, create a branch, open a PR, and follow the code change workflow.
3. Never handle or create secrets.
4. For read-only security analysis, post findings as a comment.

Output format:
## Security Review

### Scope
[What was reviewed]

### Findings
| # | Severity | Description | Location | Remediation |
|---|----------|-------------|----------|-------------|
| 1 | Critical/High/Medium/Low | [description] | file:line | [fix] |

### Threat Surface Impact
[How this affects the overall threat surface]

### Verdict
[SAFE / CONCERNS FOUND] — [one sentence summary]
