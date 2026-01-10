-- Remove token column
DROP INDEX IF EXISTS idx_invites_token;
ALTER TABLE invites DROP COLUMN token;
