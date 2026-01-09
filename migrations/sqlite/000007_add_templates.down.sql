DROP INDEX IF EXISTS idx_templates_active;
DROP INDEX IF EXISTS idx_templates_type_default;
DROP INDEX IF EXISTS idx_templates_event_type;

ALTER TABLE templates DROP COLUMN version;
ALTER TABLE templates DROP COLUMN is_active;
ALTER TABLE templates DROP COLUMN description;
ALTER TABLE templates DROP COLUMN event_id;
