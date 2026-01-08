-- TinyRSVP Invite Listing Performance Optimization
-- Migration: 000004_optimize_invite_listing
-- Purpose: Add composite indexes for common invite listing query patterns

CREATE INDEX IF NOT EXISTS idx_invites_event_status ON invites(event_id, status);

CREATE INDEX IF NOT EXISTS idx_invites_event_sent_at ON invites(event_id, sent_at);

CREATE INDEX IF NOT EXISTS idx_invites_event_viewed_at ON invites(event_id, viewed_at);

CREATE INDEX IF NOT EXISTS idx_invites_event_unsubscribed ON invites(event_id, unsubscribed);

CREATE INDEX IF NOT EXISTS idx_invites_event_email_invalid ON invites(event_id, email_invalid);

CREATE INDEX IF NOT EXISTS idx_invites_event_status_sent_at ON invites(event_id, status, sent_at);

CREATE INDEX IF NOT EXISTS idx_invites_event_status_viewed_at ON invites(event_id, status, viewed_at);
