CREATE TABLE texts
(
    id       bigserial NOT NULL PRIMARY KEY,
    owner    bigint    NOT NULL REFERENCES users (id),
    name     VARCHAR   NOT NULL,
    text     TEXT      NOT NULL,
    metadata VARCHAR   NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX idx_texts_owner_name ON texts (owner, name);
