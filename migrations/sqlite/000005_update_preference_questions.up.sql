-- Update preference_questions table to match story requirements
-- Migration: 000005_update_preference_questions

-- Add updated_at column
ALTER TABLE preference_questions ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Note: SQLite doesn't support modifying CHECK constraints directly
-- We need to recreate the table with the new constraint

-- Create new table with correct schema
CREATE TABLE preference_questions_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    question_type TEXT NOT NULL,
    options JSON,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (question_type IN ('text', 'single_choice', 'multiple_choice'))
);

-- Copy data from old table (if any exists)
INSERT INTO preference_questions_new (id, event_id, question_text, question_type, options, required, display_order, created_at, updated_at)
SELECT id, event_id, question_text, 
    CASE 
        WHEN question_type = 'select' THEN 'single_choice'
        WHEN question_type = 'boolean' THEN 'single_choice'
        ELSE question_type
    END as question_type,
    options, required, display_order, created_at, CURRENT_TIMESTAMP
FROM preference_questions;

-- Drop old table
DROP TABLE preference_questions;

-- Rename new table
ALTER TABLE preference_questions_new RENAME TO preference_questions;

-- Recreate indexes
CREATE INDEX idx_questions_event_id ON preference_questions(event_id);
CREATE INDEX idx_questions_display_order ON preference_questions(event_id, display_order);
