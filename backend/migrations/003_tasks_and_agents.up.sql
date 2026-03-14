-- Create tasks and agents tables together to resolve the circular FK.
-- tasks.agent_id → agents and agents.task_id → tasks.
-- We create both tables first without the cross-FKs, then add them via ALTER TABLE.

CREATE TABLE tasks (
    id          TEXT PRIMARY KEY,               -- ULID, prefix "task_"
    repo_id     TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    title       TEXT NOT NULL CHECK (char_length(title) <= 255),
    description TEXT,
    status      TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'in_progress', 'done', 'cancelled')),
    priority    INT  NOT NULL DEFAULT 3
                    CHECK (priority BETWEEN 1 AND 5),
    agent_id    TEXT,                           -- FK added below after agents table exists
    created_by  TEXT NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX tasks_repo_id_idx    ON tasks (repo_id);
CREATE INDEX tasks_status_idx     ON tasks (repo_id, status);
CREATE INDEX tasks_agent_id_idx   ON tasks (agent_id) WHERE agent_id IS NOT NULL;

CREATE TABLE agents (
    id             TEXT PRIMARY KEY,            -- ULID, prefix "agent_"
    repo_id        TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    task_id        TEXT,                        -- FK added below after tasks table exists
    name           TEXT NOT NULL,
    persona        TEXT CHECK (persona IN ('backend', 'frontend', 'devops', 'qa', 'general')),
    model          TEXT NOT NULL DEFAULT 'claude-opus-4-5',
    system_prompt  TEXT,
    status         TEXT NOT NULL DEFAULT 'starting'
                       CHECK (status IN ('starting', 'running', 'stopped', 'crashed')),
    worktree_path  TEXT,
    branch         TEXT,
    pid            INT,
    started_at     TIMESTAMPTZ,
    stopped_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (repo_id, name)
);

CREATE INDEX agents_repo_id_idx  ON agents (repo_id);
CREATE INDEX agents_status_idx   ON agents (repo_id, status);
CREATE INDEX agents_task_id_idx  ON agents (task_id) WHERE task_id IS NOT NULL;

-- Add cross-FK constraints now that both tables exist.
ALTER TABLE tasks  ADD CONSTRAINT tasks_agent_id_fk
    FOREIGN KEY (agent_id)  REFERENCES agents(id) ON DELETE SET NULL;

ALTER TABLE agents ADD CONSTRAINT agents_task_id_fk
    FOREIGN KEY (task_id) REFERENCES tasks(id)  ON DELETE SET NULL;
