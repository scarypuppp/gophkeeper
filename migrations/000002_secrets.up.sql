CREATE TABLE secrets
(
    id       bigserial NOT NULL PRIMARY KEY,
    owner    bigint    NOT NULL REFERENCES users (id),
    name     VARCHAR   NOT NULL,
    login    VARCHAR   NOT NULL,
    password VARCHAR   NOT NULL,
    metadata VARCHAR   NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX idx_secrets_owner_name ON secrets (owner, name);
