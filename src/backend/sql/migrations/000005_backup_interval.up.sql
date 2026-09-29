ALTER TABLE backup_targets ADD COLUMN interval_hours INTEGER NOT NULL DEFAULT 24;
ALTER TABLE backup_targets ADD COLUMN last_full_at TEXT;
