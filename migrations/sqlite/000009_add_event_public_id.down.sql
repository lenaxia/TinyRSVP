-- Rollback: Remove public_id and friendly_name from events table
-- Migration: 000009_add_event_public_id

-- Drop indexes
DROP INDEX IF EXISTS idx_events_friendly_name;
DROP INDEX IF EXISTS idx_events_public_id;

-- Remove columns (SQLite doesn't support DROP COLUMN directly in older versions)
-- We need to recreate the table without these columns
CREATE TABLE events_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    location TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    timezone TEXT NOT NULL,
    rsvp_deadline TIMESTAMP,
    max_plus_ones INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'draft',
    template_id INTEGER REFERENCES templates(id) ON DELETE SET NULL,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL DEFAULT 1,
    ics_sequence INTEGER NOT NULL DEFAULT 0,
    CHECK (status IN ('draft', 'published', 'cancelled', 'archived')),
    CHECK (max_plus_ones >= 0 AND max_plus_ones <= 10),
    CHECK (end_time IS NULL OR end_time > start_time),
    CHECK (rsvp_deadline IS NULL OR rsvp_deadline < start_time)
);

INSERT INTO events_backup SELECT 
    id, title, description, location, start_time, end_time, timezone,
    rsvp_deadline, max_plus_ones, status, template_id, created_by,
    created_at, updated_at, version, ics_sequence
FROM events;

DROP TABLE events;

ALTER TABLE events_backup RENAME TO events;

-- Recreate original indexes
CREATE INDEX idx_events_created_by ON events(created_by);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_start_time ON events(start_time);
