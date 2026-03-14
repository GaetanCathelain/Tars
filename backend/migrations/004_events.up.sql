CREATE TABLE events (
    id         TEXT PRIMARY KEY,                -- ULID, prefix "evt_"
    repo_id    TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,                   -- e.g. "agent.spawned", "task.created"
    actor_type TEXT NOT NULL
                   CHECK (actor_type IN ('user', 'agent', 'system')),
    actor_id   TEXT,
    agent_id   TEXT REFERENCES agents(id) ON DELETE SET NULL,
    task_id    TEXT REFERENCES tasks(id)  ON DELETE SET NULL,
    payload    JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Primary access pattern: timeline for a repo, newest first.
CREATE INDEX events_repo_timeline_idx ON events (repo_id, created_at DESC);

-- Secondary filters.
CREATE INDEX events_type_idx     ON events (repo_id, type);
CREATE INDEX events_agent_id_idx ON events (agent_id) WHERE agent_id IS NOT NULL;
CREATE INDEX events_task_id_idx  ON events (task_id)  WHERE task_id  IS NOT NULL;
