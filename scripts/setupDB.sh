#! /bin/bash
set -euo pipefail

if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

DB_URL="${DB_URL:-}"
if [ -z "$DB_URL" ]; then
  echo "DB_URL is not set in .env" >&2
  exit 1
fi

psql "$DB_URL" -v ON_ERROR_STOP=1 -c "DROP TABLE IF EXISTS chirps;"
psql "$DB_URL" -v ON_ERROR_STOP=1 -c "DROP TABLE IF EXISTS users;"
psql "$DB_URL" -v ON_ERROR_STOP=1 -f sql/schema/001_users.sql
psql "$DB_URL" -v ON_ERROR_STOP=1 -f sql/schema/002_chirps.sql
psql "$DB_URL" -c "\\dt"
