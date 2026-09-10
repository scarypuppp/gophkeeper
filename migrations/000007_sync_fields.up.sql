ALTER TABLE secrets
    ADD COLUMN checksum   VARCHAR     NOT NULL DEFAULT '',
    ADD COLUMN created_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

ALTER TABLE cards
    ADD COLUMN checksum   VARCHAR     NOT NULL DEFAULT '',
    ADD COLUMN created_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

ALTER TABLE texts
    ADD COLUMN checksum   VARCHAR     NOT NULL DEFAULT '',
    ADD COLUMN created_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

ALTER TABLE files
    ADD COLUMN created_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

ALTER TABLE files DROP COLUMN uploaded_at;
