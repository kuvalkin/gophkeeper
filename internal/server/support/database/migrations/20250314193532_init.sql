-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(250) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    version int8 NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS entries (
    user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    key TEXT NOT NULL,
    name TEXT NOT NULL,
    notes TEXT DEFAULT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    version int8 NOT NULL DEFAULT 1,
    user_version int8 NOT NULL,
    content_name TEXT NOT NULL,

    PRIMARY KEY (user_id, key)
);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS entries;
