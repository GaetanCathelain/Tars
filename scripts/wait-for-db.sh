#!/usr/bin/env sh
# wait-for-db.sh: block until PostgreSQL is accepting connections
set -e

HOST="${DB_HOST:-db}"
PORT="${DB_PORT:-5432}"
USER="${POSTGRES_USER:-tars}"
DB="${POSTGRES_DB:-tars}"
MAX_ATTEMPTS="${MAX_ATTEMPTS:-30}"
SLEEP_SECONDS=2

attempt=1
until pg_isready -h "$HOST" -p "$PORT" -U "$USER" -d "$DB" -q; do
  if [ "$attempt" -ge "$MAX_ATTEMPTS" ]; then
    echo "ERROR: PostgreSQL at $HOST:$PORT did not become ready after $MAX_ATTEMPTS attempts." >&2
    exit 1
  fi
  echo "Waiting for PostgreSQL at $HOST:$PORT ... (attempt $attempt/$MAX_ATTEMPTS)"
  attempt=$((attempt + 1))
  sleep "$SLEEP_SECONDS"
done

echo "PostgreSQL is ready."
exec "$@"
