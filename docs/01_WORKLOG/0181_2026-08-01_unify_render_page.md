# Worklog: Consolidate Handler Page Rendering (remove dead fallbacks)

**Date:** 2026-08-01  
**Branch:** `refactor/unify-render-page`

## Summary

Cluster 3 from the tech-debt review: consolidated 15 near-identical `renderPage`-style methods into a single `renderHTML` helper, deleting 13 dead `fmt.Fprintf` HTML fallbacks (unreachable in production — every page template is parsed at startup with `os.Exit(1)` on failure) and fixing a latent header-ordering bug.

## Changes

- **`internal/handlers/render.go`** — new `renderHTML(w, tmpl, name, status, data)` helper. Buffers the render and only writes `Content-Type`/`WriteHeader` after a successful execution, so a mid-render failure produces a clean 500 instead of a truncated response with headers already sent (the old pattern wrote `WriteHeader` before executing). Logs template-execution failures via `slog`.
- **15 render methods now delegate to `renderHTML`**: dashboard, admin (×2), metrics, settings, template_editor, events_web (list/form/detail), rsvp (page, confirmation, unsubscribe), rsvp_summary.
- **Deleted 13 dead fallback HTML strings** that duplicated (and in two cases had drifted from) the real `.html` templates — e.g. the user-management fallback showed `len(data.Users)` while the real template uses `data.Total`; the event-form fallback emitted a bare `<form>`.
- **Header-ordering fix**: all render paths now buffer before writing headers (previously only the confirmation page did).

## Tests

- New `render_test.go` helper `testTemplate(t, name)` for tests to exercise the real render path.
- Updated 42 tests that had relied on the nil-template fallback to either set templates via `SetTemplates`/`SetConfirmationTemplates` or (for the two explicit `*NoTemplate*`/`*WithoutTemplates*` tests) assert `500` as the new correct nil-template behavior.
- Full suite: all 39 non-browser packages pass; `go build`/`go vet` clean.

## Notes

This fixes the duplication bug (finding 1.4/1.7) and the header-ordering bug (1.16) from the review. The middleware `http.Error` paths and the `HandleThemePreview`/`notImplementedHTML` inline HTML are out of scope for this pass.

## Status

**Status:** ✅ Complete  
**Test Pass Rate:** All non-browser tests pass  
**Confidence:** HIGH  
**Production Ready:** Yes
