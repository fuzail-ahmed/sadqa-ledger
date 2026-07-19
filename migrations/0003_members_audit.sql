-- Adds member create/update audit trail (docs/PRD.md §5: every write records
-- the acting admin), mirroring admins.created_by_admin_id (0002).
ALTER TABLE members ADD COLUMN created_by_admin_id INTEGER REFERENCES admins(id);
ALTER TABLE members ADD COLUMN updated_by_admin_id INTEGER REFERENCES admins(id);
