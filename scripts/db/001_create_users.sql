
--  Create the users table
CREATE TABLE IF NOT EXISTS users (
    id         TEXT        PRIMARY KEY,
    name       TEXT        NOT NULL,
    email      TEXT        NOT NULL UNIQUE,
    status     TEXT        NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    version    INT         NOT NULL DEFAULT 1
);