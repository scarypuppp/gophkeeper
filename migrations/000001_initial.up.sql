CREATE TABLE users
(
    id       bigserial NOT NULL PRIMARY KEY,
    login    VARCHAR   NOT NULL,
    password VARCHAR   NOT NULL
);
CREATE UNIQUE INDEX idx_users_login ON users (login);

CREATE TABLE files
(
    id                bigserial NOT NULL PRIMARY KEY,
    owner             bigint    NOT NULL REFERENCES users (id),
    file_name         VARCHAR   NOT NULL,
    file_hash         VARCHAR   NOT NULL,
    storage_file_path VARCHAR   NOT NULL
);
CREATE UNIQUE INDEX idx_files_owner_hash ON files (owner, file_hash);
