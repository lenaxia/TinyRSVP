-- Remove RSVP configuration fields from events table
-- Migration: 000013_add_rsvp_configuration (down)

DROP TRIGGER IF EXISTS check_event_capacity_positive;
DROP TRIGGER IF EXISTS check_event_capacity_positive_update;
DROP TRIGGER IF EXISTS check_adults_count_non_negative;
DROP TRIGGER IF EXISTS check_adults_count_non_negative_update;
DROP TRIGGER IF EXISTS check_kids_count_non_negative;
DROP TRIGGER IF EXISTS check_kids_count_non_negative_update;

ALTER TABLE events DROP COLUMN allow_rsvp_after_deadline;
ALTER TABLE events DROP COLUMN allow_maybe_rsvp;
ALTER TABLE events DROP COLUMN private_guest_list;
ALTER TABLE events DROP COLUMN family_headcount;
ALTER TABLE events DROP COLUMN event_capacity;

ALTER TABLE rsvps DROP COLUMN adults_count;
ALTER TABLE rsvps DROP COLUMN kids_count;
