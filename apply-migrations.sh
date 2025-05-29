#!/usr/bin/env sh
set -e

until pg_isready -h "${DB_HOST:-db}" -p "${DB_PORT:-5432}" -U "${DB_USER:-user}"; do
  echo "Waiting for postgres..."
  sleep 1
done

goose -dir ./internal/storage/postgres/migrations postgres "${DB_DSN}" up
echo "Migrations applied."
