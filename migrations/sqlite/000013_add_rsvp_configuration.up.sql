-- Add RSVP configuration fields to events table
-- Migration: 000013_add_rsvp_configuration

ALTER TABLE events ADD COLUMN allow_rsvp_after_deadline BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE events ADD COLUMN allow_maybe_rsvp BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE events ADD COLUMN private_guest_list BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE events ADD COLUMN family_headcount BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE events ADD COLUMN event_capacity INTEGER;

-- Add check constraint for event_capacity (must be positive if set)
CREATE TRIGGER check_event_capacity_positive
BEFORE INSERT ON events
WHEN NEW.event_capacity IS NOT NULL AND NEW.event_capacity <= 0
BEGIN
    SELECT RAISE(ABORT, 'event_capacity must be positive');
END;

CREATE TRIGGER check_event_capacity_positive_update
BEFORE UPDATE ON events
WHEN NEW.event_capacity IS NOT NULL AND NEW.event_capacity <= 0
BEGIN
    SELECT RAISE(ABORT, 'event_capacity must be positive');
END;

-- Add family headcount fields to rsvps table
ALTER TABLE rsvps ADD COLUMN adults_count INTEGER;
ALTER TABLE rsvps ADD COLUMN kids_count INTEGER;

-- Add check constraints for headcount fields (must be non-negative if set)
CREATE TRIGGER check_adults_count_non_negative
BEFORE INSERT ON rsvps
WHEN NEW.adults_count IS NOT NULL AND NEW.adults_count < 0
BEGIN
    SELECT RAISE(ABORT, 'adults_count must be non-negative');
END;

CREATE TRIGGER check_adults_count_non_negative_update
BEFORE UPDATE ON rsvps
WHEN NEW.adults_count IS NOT NULL AND NEW.adults_count < 0
BEGIN
    SELECT RAISE(ABORT, 'adults_count must be non-negative');
END;

CREATE TRIGGER check_kids_count_non_negative
BEFORE INSERT ON rsvps
WHEN NEW.kids_count IS NOT NULL AND NEW.kids_count < 0
BEGIN
    SELECT RAISE(ABORT, 'kids_count must be non-negative');
END;

CREATE TRIGGER check_kids_count_non_negative_update
BEFORE UPDATE ON rsvps
WHEN NEW.kids_count IS NOT NULL AND NEW.kids_count < 0
BEGIN
    SELECT RAISE(ABORT, 'kids_count must be non-negative');
END;
