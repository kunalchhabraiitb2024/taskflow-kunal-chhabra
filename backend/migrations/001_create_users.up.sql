-- Enable pgcrypto for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,  -- bcrypt hash, never plaintext
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Fast lookup by email (login)
CREATE INDEX idx_users_email ON users(email);
