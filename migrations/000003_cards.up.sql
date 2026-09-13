CREATE TABLE cards
(
    id         bigserial NOT NULL PRIMARY KEY,
    owner      bigint    NOT NULL REFERENCES users (id),
    name       VARCHAR   NOT NULL,
    number     VARCHAR   NOT NULL,
    holder     VARCHAR   NOT NULL,
    expires_at VARCHAR   NOT NULL,
    cvv        VARCHAR   NOT NULL,
    metadata   VARCHAR   NOT NULL DEFAULT ''
);
CREATE UNIQUE INDEX idx_cards_owner_name ON cards (owner, name);
