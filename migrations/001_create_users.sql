-- migrations/001_create_users.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    name          VARCHAR(100) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'user',
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_email      ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_is_active  ON users(is_active) WHERE deleted_at IS NULL;
