-- Add public_id and friendly_name to events table
-- Migration: 000009_add_event_public_id

-- Add public_id column (unique, non-guessable identifier)
ALTER TABLE events ADD COLUMN public_id TEXT;

-- Add friendly_name column (optional, URL-friendly slug)
ALTER TABLE events ADD COLUMN friendly_name TEXT;

-- Create unique index on public_id
CREATE UNIQUE INDEX idx_events_public_id ON events(public_id) WHERE public_id IS NOT NULL;

-- Create unique index on friendly_name
CREATE UNIQUE INDEX idx_events_friendly_name ON events(friendly_name) WHERE friendly_name IS NOT NULL;
