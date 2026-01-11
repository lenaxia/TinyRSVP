ALTER TABLE templates ADD COLUMN category TEXT NOT NULL DEFAULT 'plain';
ALTER TABLE templates ADD COLUMN thumbnail_url TEXT;
ALTER TABLE templates ADD COLUMN image_url TEXT;
ALTER TABLE templates ADD COLUMN tags TEXT NOT NULL DEFAULT '[]';
ALTER TABLE templates ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

CREATE INDEX idx_templates_category ON templates(category);
CREATE INDEX idx_templates_sort_order ON templates(sort_order);

UPDATE templates SET category = 'plain' WHERE category IS NULL OR category = '';
