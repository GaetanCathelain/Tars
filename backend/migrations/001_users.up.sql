CREATE TABLE users (
    id         TEXT PRIMARY KEY,                -- ULID, prefix "usr_"
    github_id  BIGINT UNIQUE NOT NULL,
    login      TEXT NOT NULL,
    name       TEXT,
    avatar_url TEXT,
    email      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
