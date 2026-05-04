#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ ! -f .env ]]; then
  echo "missing .env in deploy directory" >&2
  exit 1
fi

docker compose up -d --build --remove-orphans
docker compose ps
docker image prune -f >/dev/null
