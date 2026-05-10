#!/usr/bin/env bash
set -euo pipefail

# Load environment variables
source "$1"

PGPASSWORD="$DB_PASSWORD" psql \
  -h "$DB_HOST" \
  -U "$DB_USER" \
  -d "$DB_NAME" \
  -f "$SEED_FILE"

echo "Database seeded"