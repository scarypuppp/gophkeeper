ALTER TABLE files
    ADD COLUMN checksum    VARCHAR     NOT NULL DEFAULT '',
    ADD COLUMN size        bigint      NOT NULL DEFAULT 0,
    ADD COLUMN uploaded_at timestamptz NOT NULL DEFAULT now();
