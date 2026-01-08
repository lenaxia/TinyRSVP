-- TinyRSVP Invite Listing Performance Optimization Rollback
-- Migration: 000004_optimize_invite_listing
-- Purpose: Remove composite indexes for invite listing

DROP INDEX IF EXISTS idx_invites_event_status_viewed_at;

DROP INDEX IF EXISTS idx_invites_event_status_sent_at;

DROP INDEX IF EXISTS idx_invites_event_email_invalid;

DROP INDEX IF EXISTS idx_invites_event_unsubscribed;

DROP INDEX IF EXISTS idx_invites_event_viewed_at;

DROP INDEX IF EXISTS idx_invites_event_sent_at;

DROP INDEX IF EXISTS idx_invites_event_status;
