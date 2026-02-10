#!/usr/bin/env bash
# ==============================================================================
# ALPHA STATE WIPE SCRIPT
# Preserves player profiles, resets stats, starts new season, reapplies schema
# ==============================================================================

set -euo pipefail

# ==============================================================================
# Safety checks
# ==============================================================================
if [[ "${PHASE:-}" != "alpha" ]]; then
  echo "❌ Refusing to wipe: PHASE must be 'alpha'." >&2
  exit 1
fi

if [[ "${ALPHA_STATE_WIPE_CONFIRM:-}" != "I_UNDERSTAND_THIS_RESETS_ALL_STATS" ]]; then
  echo "❌ Refusing to wipe: set ALPHA_STATE_WIPE_CONFIRM=I_UNDERSTAND_THIS_RESETS_ALL_STATS to proceed." >&2
  exit 1
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "❌ DATABASE_URL is not set." >&2
  exit 1
fi

# ==============================================================================
# Determine script directory
# ==============================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# ==============================================================================
# Display warning and confirmation
# ==============================================================================
echo "⚠️  ALPHA STATE WIPE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "This will:"
echo "  ✓ Preserve accounts (usernames, passwords, profiles)"
echo "  ✓ Preserve player IDs and creation metadata"
echo "  ✗ Reset all coins and stars to 0"
echo "  ✗ Clear all logs, telemetry, and history"
echo "  ✗ Clear all sessions (users must log in again)"
echo "  ✗ Clear all season state"
echo "  ✓ Reapply schema.sql (ensure migrations applied)"
echo "  ✓ Start a new season"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Press ENTER to continue or Ctrl+C to abort..."
read -r

# ==============================================================================
# Step 1: Run alpha-state-wipe.sql
# ==============================================================================
echo ""
echo "📝 Step 1/3: Running alpha-state-wipe.sql..."
if ! psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f "${PROJECT_ROOT}/alpha-state-wipe.sql"; then
  echo "❌ Failed to execute alpha-state-wipe.sql" >&2
  exit 1
fi
echo "✅ State wipe complete"

# ==============================================================================
# Step 2: Reapply schema.sql (idempotent - uses CREATE IF NOT EXISTS)
# ==============================================================================
echo ""
echo "📝 Step 2/3: Reapplying schema.sql..."
if ! psql "${DATABASE_URL}" -v ON_ERROR_STOP=1 -f "${PROJECT_ROOT}/schema.sql"; then
  echo "❌ Failed to reapply schema.sql" >&2
  exit 1
fi
echo "✅ Schema reapplied"

# ==============================================================================
# Step 3: Initialize new season (via server restart or manual trigger)
# ==============================================================================
echo ""
echo "📝 Step 3/3: Season initialization..."
echo ""
echo "ℹ️  The server will auto-create a new season on next startup."
echo "   Alternatively, you can manually trigger:"
echo "   POST /admin/seasons/recovery"
echo "   (requires admin authentication)"
echo ""

# ==============================================================================
# Success summary
# ==============================================================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ ALPHA STATE WIPE COMPLETE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Next steps:"
echo "  1. Restart your application server"
echo "  2. A new season will be auto-created on startup"
echo "  3. All users can log in with their existing credentials"
echo "  4. All stats and balances start fresh at 0"
echo ""
