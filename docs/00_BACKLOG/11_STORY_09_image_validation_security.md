# User Story 11.09: Image Validation & Security

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
**Priority:** High
**Status:** ✅ Complete
**Estimated Effort:** 1-2 days
**Owner:** LLM (2026-01-11)
**Phase:** 2
**Completed:** 2026-01-11

---

## User Story

As a **security engineer**,  
I want **comprehensive image validation and security measures**,  
So that **malicious file uploads cannot compromise the system or user privacy**.

---

## Acceptance Criteria

### Magic Byte Validation
- [x] Validate file type by magic bytes (not extension)
- [x] Reject non-image files
- [x] Reject files with mismatched extension/content
- [x] Test with renamed executables
- [x] Test with polyglot files

### Size & Dimension Limits
- [x] Enforce 5MB file size limit
- [x] Enforce 4096x4096 dimension limit
- [x] Reject oversized files
- [x] Reject oversized dimensions
- [x] Test boundary conditions

### EXIF Stripping
- [x] Strip all EXIF metadata
- [x] Strip GPS coordinates
- [x] Strip camera information
- [x] Strip timestamps
- [x] Preserve image quality
- [x] Test with various EXIF data

### Content-Type Validation
- [x] Validate Content-Type header
- [x] Detect actual content type
- [x] Reject mismatched types
- [x] Set correct Content-Type on serve

### Malicious File Detection
- [x] Reject files with embedded scripts
- [x] Reject SVG files (XSS vector)
- [x] Reject files with suspicious patterns
- [x] Test with known malicious samples

### Testing
- [x] Unit tests for each validation rule
- [x] Security tests with malicious files
- [x] Performance tests with large files
- [x] Integration tests

---

## Technical Details

See existing `internal/assets/validator.go` - extend as needed.

---

## Tasks

- [x] Review existing asset validator
- [x] Add magic byte validation (already implemented)
- [x] Add dimension validation (already implemented)
- [x] Implement EXIF stripping (already implemented)
- [x] Add malicious pattern detection (enhanced)
- [x] Write comprehensive tests
- [x] Security audit

---

## Dependencies

**Depends on:**
- Story 11.08: Custom Image Upload
- Story 06.08: Storage Provider

**Blocks:**
- Story 11.10: Custom Image Preview

---

## References

- **Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)
- **Asset Validator:** `internal/assets/validator.go`
