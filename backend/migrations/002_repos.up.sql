CREATE TABLE repos (
    id             TEXT PRIMARY KEY,            -- ULID, prefix "repo_"
    owner_id       TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name           TEXT NOT NULL,
    path           TEXT NOT NULL,
    github_url     TEXT NOT NULL,
    default_branch TEXT NOT NULL DEFAULT 'main',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_id, name)
);

CREATE INDEX repos_owner_id_idx ON repos (owner_id);
