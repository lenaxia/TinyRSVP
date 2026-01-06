-- TinyRSVP Additional Indexes
-- Migration: 000002_add_indexes
-- Note: Core indexes are already created in 000001_initial_schema
-- This migration is reserved for future performance optimization indexes

-- Placeholder for future additional indexes
-- Examples of indexes that might be added later:
-- CREATE INDEX idx_events_rsvp_deadline ON events(rsvp_deadline) WHERE rsvp_deadline IS NOT NULL;
-- CREATE INDEX idx_invites_event_status ON invites(event_id, status);
-- CREATE INDEX idx_rsvps_created_at ON rsvps(created_at);
