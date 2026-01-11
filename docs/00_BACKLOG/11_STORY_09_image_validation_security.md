# User Story 11.09: Image Validation & Security

**Epic:** [11_EPIC_rsvp_themes.md](11_EPIC_rsvp_themes.md)  
**Priority:** High  
**Status:** Not Started  
**Estimated Effort:** 1-2 days  
**Owner:** Unassigned  
**Phase:** 2

---

## User Story

As a **security engineer**,  
I want **comprehensive image validation and security measures**,  
So that **malicious file uploads cannot compromise the system or user privacy**.

---

## Acceptance Criteria

### Magic Byte Validation
- [ ] Validate file type by magic bytes (not extension)
- [ ] Reject non-image files
- [ ] Reject files with mismatched extension/content
- [ ] Test with renamed executables
- [ ] Test with polyglot files

### Size & Dimension Limits
- [ ] Enforce 5MB file size limit
- [ ] Enforce 4096x4096 dimension limit
- [ ] Reject oversized files
- [ ] Reject oversized dimensions
- [ ] Test boundary conditions

### EXIF Stripping
- [ ] Strip all EXIF metadata
- [ ] Strip GPS coordinates
- [ ] Strip camera information
- [ ] Strip timestamps
- [ ] Preserve image quality
- [ ] Test with various EXIF data

### Content-Type Validation
- [ ] Validate Content-Type header
- [ ] Detect actual content type
- [ ] Reject mismatched types
- [ ] Set correct Content-Type on serve

### Malicious File Detection
- [ ] Reject files with embedded scripts
- [ ] Reject SVG files (XSS vector)
- [ ] Reject files with suspicious patterns
- [ ] Test with known malicious samples

### Testing
- [ ] Unit tests for each validation rule
- [ ] Security tests with malicious files
- [ ] Performance tests with large files
- [ ] Integration tests

---

## Technical Details

See existing `internal/assets/validator.go` - extend as needed.

---

## Tasks

- [ ] Review existing asset validator
- [ ] Add magic byte validation
- [ ] Add dimension validation
- [ ] Implement EXIF stripping
- [ ] Add malicious pattern detection
- [ ] Write comprehensive tests
- [ ] Security audit

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
