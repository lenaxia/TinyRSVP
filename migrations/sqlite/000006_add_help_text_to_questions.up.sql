-- Add help_text column to preference_questions table
-- Migration: 000006_add_help_text_to_questions

ALTER TABLE preference_questions ADD COLUMN help_text TEXT;
