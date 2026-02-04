#!/usr/bin/env bash
set -euo pipefail

if [[ "${APP_ENV:-}" != "alpha" ]]; then
  echo "Refusing to reset: APP_ENV must be 'alpha'." >&2
  exit 1
fi

if [[ "${ALPHA_RESET_CONFIRM:-}" != "I_UNDERSTAND" ]]; then
  echo "Refusing to reset: set ALPHA_RESET_CONFIRM=I_UNDERSTAND to proceed." >&2
  exit 1
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is not set." >&2
  exit 1
fi

echo "ALPHA RESET: dropping schema and re-applying schema.sql..."
psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f "$(dirname "$0")/../schema.sql"

echo "ALPHA RESET COMPLETE."
