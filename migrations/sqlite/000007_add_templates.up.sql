ALTER TABLE templates ADD COLUMN event_id INTEGER REFERENCES events(id) ON DELETE CASCADE;
ALTER TABLE templates ADD COLUMN description TEXT;
ALTER TABLE templates ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT 1;
ALTER TABLE templates ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

CREATE INDEX idx_templates_event_type ON templates(event_id, type);
CREATE INDEX idx_templates_type_default ON templates(type, is_default);
CREATE INDEX idx_templates_active ON templates(is_active);
