ALTER TABLE jobs ADD COLUMN metrics JSONB DEFAULT '{}'::jsonb;
