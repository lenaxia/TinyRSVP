-- Revert preference_questions table changes
-- Migration: 000005_update_preference_questions

-- Create old table structure
CREATE TABLE preference_questions_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    question_type TEXT NOT NULL,
    options JSON,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (question_type IN ('text', 'select', 'boolean'))
);

-- Copy data back (converting question types)
INSERT INTO preference_questions_old (id, event_id, question_text, question_type, options, required, display_order, created_at)
SELECT id, event_id, question_text,
    CASE 
        WHEN question_type = 'single_choice' THEN 'select'
        WHEN question_type = 'multiple_choice' THEN 'select'
        ELSE question_type
    END as question_type,
    options, required, display_order, created_at
FROM preference_questions;

-- Drop new table
DROP TABLE preference_questions;

-- Rename old table back
ALTER TABLE preference_questions_old RENAME TO preference_questions;

-- Recreate indexes
CREATE INDEX idx_questions_event_id ON preference_questions(event_id);
CREATE INDEX idx_questions_display_order ON preference_questions(event_id, display_order);
