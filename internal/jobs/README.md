# Jobs Package

## Purpose
Scheduled background jobs for system maintenance and automation.

## Rules
- All jobs must be idempotent (safe to run multiple times)
- All jobs must handle errors gracefully
- All jobs must log their activity
- All jobs must support context cancellation
- Jobs should not require authentication context

## Structure
- `archiver.go` - Event archiving job
- `archiver_test.go` - Event archiving tests

## Key Files
- `archiver.go` - Archives events older than configured threshold
