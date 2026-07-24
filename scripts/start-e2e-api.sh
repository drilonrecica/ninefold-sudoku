#!/usr/bin/env bash
set -euo pipefail

e2e_directory=".e2e"
database_path="$(pwd)/${e2e_directory}/ninefold.db"
mkdir -p "${e2e_directory}"
rm -f "${database_path}" "${database_path}-shm" "${database_path}-wal"

export NINEFOLD_ENVIRONMENT="test"
export NINEFOLD_PUBLIC_URL="http://127.0.0.1:4173"
export NINEFOLD_HTTP_ADDRESS="127.0.0.1:8080"
export NINEFOLD_DATABASE_PATH="${database_path}"
export NINEFOLD_ALLOWED_ORIGINS="http://127.0.0.1:4173"
export NINEFOLD_COOKIE_SECRET="AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
export NINEFOLD_REPLAY_SIGNING_KEY="MC4CAQAwBQYDK2VwBCIEIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
export NINEFOLD_REPLAY_SIGNING_KEY_ID="e2e-1"
export NINEFOLD_ADMIN_PROXY_HEADER="X-Ninefold-Admin"
export NINEFOLD_LOG_LEVEL="error"
export NINEFOLD_REPLAY_RETENTION="168h"
export NINEFOLD_MATCH_TOMBSTONE_RETENTION="720h"
export NINEFOLD_COMMAND_RECEIPT_RETENTION="24h"
export NINEFOLD_SHUTDOWN_TIMEOUT="10s"

cd apps/server
go run ./cmd/maintenance seed-e2e internal/puzzle/catalog/catalog.jsonl
exec go run -ldflags="-X main.buildVersion=e2e" ./cmd/api
