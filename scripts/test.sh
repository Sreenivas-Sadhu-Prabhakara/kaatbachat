#!/usr/bin/env bash
# Test locally against a throwaway Postgres, then stop the service.
# Spins up a disposable Postgres container on a random host port, runs the Go
# tests (including the /history integration test), and always tears the
# container down on exit — nothing is left running after the tests.
set -euo pipefail

IMAGE="postgres:16"
CID=$(docker run -d -e POSTGRES_PASSWORD=test -e POSTGRES_DB=app -P "$IMAGE")
cleanup() { docker rm -f "$CID" >/dev/null 2>&1 || true; }
trap cleanup EXIT

PORT=$(docker port "$CID" 5432/tcp | head -1 | sed 's/.*://')
export DATABASE_URL="postgres://postgres:test@localhost:${PORT}/app?sslmode=disable"

# Wait for Postgres to accept connections.
for _ in $(seq 1 30); do
  if docker exec "$CID" pg_isready -U postgres -d app >/dev/null 2>&1; then break; fi
  sleep 1
done

go test ./backend/... "$@"
