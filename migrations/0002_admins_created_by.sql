-- Adds admin-creation audit trail (docs/SCHEMA.md §3.1): who created whom.
-- NULL for the first/setup-created admin, which has no creator.
ALTER TABLE admins ADD COLUMN created_by_admin_id INTEGER REFERENCES admins(id);
