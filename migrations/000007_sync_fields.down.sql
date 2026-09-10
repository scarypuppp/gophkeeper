ALTER TABLE files
    ADD COLUMN uploaded_at timestamptz NOT NULL DEFAULT now();

UPDATE files SET uploaded_at = created_at;

ALTER TABLE files
    DROP COLUMN created_at,
    DROP COLUMN updated_at;

ALTER TABLE texts
    DROP COLUMN checksum,
    DROP COLUMN created_at,
    DROP COLUMN updated_at;

ALTER TABLE cards
    DROP COLUMN checksum,
    DROP COLUMN created_at,
    DROP COLUMN updated_at;

ALTER TABLE secrets
    DROP COLUMN checksum,
    DROP COLUMN created_at,
    DROP COLUMN updated_at;
