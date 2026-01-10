# Security Notice for Test Environment

## ⚠️ TEST ENVIRONMENT ONLY

All secrets, passwords, and cryptographic keys in this directory are **INTENTIONALLY PUBLIC** and **SAFE TO COMMIT** to version control.

### Why These Secrets Are Safe

1. **Test Environment Only**: These credentials only work in the local Docker test environment
2. **No Production Access**: They provide zero access to any production systems
3. **Publicly Documented**: All passwords are documented in README files
4. **Isolated**: The test environment runs in isolated Docker containers
5. **No Real Data**: No real user data or sensitive information exists in test environment

### Test Credentials

The following credentials are intentionally public and safe:

#### User Passwords
- `admin` / `admin123`
- `testuser` / `test123`
- `guest` / `guest123`

#### Configuration Secrets
- Session secret: `insecure_session_secret_for_testing_only`
- Encryption key: `insecure_encryption_key_for_testing_only_must_be_32_chars!!`
- HMAC secret: `insecure_hmac_secret_for_testing_only_must_be_very_long_string`
- OIDC client secret: `insecure_secret_for_testing_only`

#### Cryptographic Material
- Argon2 password hashes (for test passwords above)
- RSA private key (generated for test OIDC provider)

### Secret Scanning Configuration

This repository includes `.gitleaks.toml` which explicitly allows these test secrets:
- Prevents false positives in security scans
- Documents which secrets are intentionally public
- Helps distinguish test vs production secrets

### Production Deployment

**CRITICAL**: When deploying to production:

1. ✅ **DO** use a real OIDC provider (Authentik, Keycloak, Okta, etc.)
2. ✅ **DO** generate new, strong secrets for all configuration
3. ✅ **DO** use environment variables or secret management systems
4. ✅ **DO** enable TLS/HTTPS with valid certificates
5. ✅ **DO** use strong, unique passwords for all accounts
6. ❌ **DO NOT** use any credentials from this test directory
7. ❌ **DO NOT** commit production secrets to version control
8. ❌ **DO NOT** use simple passwords like "admin123"

### Questions?

If you're unsure whether a secret is safe to commit:
- If it's in the `test/` directory → Probably safe (but verify it's documented here)
- If it's anywhere else → **DO NOT COMMIT** without review

### References

- [OWASP Secrets Management](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [GitHub Secret Scanning](https://docs.github.com/en/code-security/secret-scanning/about-secret-scanning)
- [GitLeaks Documentation](https://github.com/gitleaks/gitleaks)
