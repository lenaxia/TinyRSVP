-- Remove help_text column from preference_questions table
-- Migration: 000006_add_help_text_to_questions (rollback)

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
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

INSERT INTO preference_questions_new (id, event_id, question_text, question_type, options, required, display_order, created_at, updated_at)
SELECT id, event_id, question_text, question_type, options, required, display_order, created_at, updated_at
FROM preference_questions;

DROP TABLE preference_questions;

ALTER TABLE preference_questions_new RENAME TO preference_questions;

CREATE INDEX idx_questions_event_id ON preference_questions(event_id);
CREATE INDEX idx_questions_display_order ON preference_questions(event_id, display_order);
