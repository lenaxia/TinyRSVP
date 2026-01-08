# RSVP Package

## Purpose

This package handles RSVP submission logic, including validation, deadline enforcement, and answer processing. It orchestrates the creation of RSVPs and their associated answers in a transactional manner.

## Rules

- All RSVP submissions must go through the service layer
- Token validation is delegated to the invite service
- Deadline enforcement is strict (no grace period)
- Plus ones are automatically set to 0 for "no" responses
- All database operations must be transactional
- Answer validation must match question types

## Structure

- `service.go` - RSVP service implementation
- `service_test.go` - Service unit tests
- `validator.go` - RSVP and answer validation logic
- `validator_test.go` - Validator unit tests

## Key Interfaces

### Service
Handles RSVP submission with full validation and transaction management.

### Validator
Validates RSVP requests, plus ones, answers, and deadline enforcement.

## Dependencies

- `internal/db/repositories` - Data access layer
- `internal/models` - Domain models
- `internal/invites` - Token validation
- `internal/events` - Event and question services
