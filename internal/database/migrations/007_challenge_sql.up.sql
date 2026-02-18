ALTER TABLE challenges ADD COLUMN sql_enabled BOOLEAN DEFAULT 0;
ALTER TABLE challenges ADD COLUMN sql_dataset_url TEXT;
ALTER TABLE challenges ADD COLUMN sql_schema_hint TEXT;
