-- Add token column to store plain tokens for retrieval
-- Tokens are shared via email anyway, so storing them is acceptable
ALTER TABLE invites ADD COLUMN token TEXT;

-- Create index for token lookups
CREATE INDEX idx_invites_token ON invites(token);

-- Note: Existing invites will have NULL tokens
-- They can be regenerated if needed
