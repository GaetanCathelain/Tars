#!/usr/bin/env sh
# seed.sh: insert development seed data into the database
# Requires: DATABASE_URL env var, or individual POSTGRES_* vars
set -e

DATABASE_URL="${DATABASE_URL:-postgres://${POSTGRES_USER:-tars}:${POSTGRES_PASSWORD:-tars}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${POSTGRES_DB:-tars}?sslmode=disable}"

echo "Seeding database at $DATABASE_URL ..."

psql "$DATABASE_URL" <<'SQL'
-- Development seed: create a test user (password hash is bcrypt of "password")
INSERT INTO users (github_id, login, name, avatar_url, email)
VALUES (
  1,
  'dev-user',
  'Dev User',
  'https://avatars.githubusercontent.com/u/1',
  'dev@example.com'
)
ON CONFLICT (github_id) DO NOTHING;
SQL

echo "Seed complete."
