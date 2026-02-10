-- ============================================================
-- ALPHA STATE WIPE
-- Preserves: accounts, player profiles (identity/metadata)
-- Resets: stats, balances, logs, sessions, season state
-- Purpose: Clean slate for new Alpha season while keeping users
-- ============================================================

BEGIN;

-- Ensure any deferrable constraints are checked at commit
SET CONSTRAINTS ALL DEFERRED;

-- =============================================================================
-- PHASE 1: Clear all logs and telemetry
-- =============================================================================
TRUNCATE player_telemetry RESTART IDENTITY CASCADE;
TRUNCATE admin_audit_log RESTART IDENTITY CASCADE;
TRUNCATE abuse_events RESTART IDENTITY CASCADE;
TRUNCATE coin_earning_log RESTART IDENTITY CASCADE;
TRUNCATE star_purchase_log RESTART IDENTITY CASCADE;

-- Clear abuse tracking state
TRUNCATE player_abuse_state CASCADE;
TRUNCATE account_abuse_reputation CASCADE;

-- =============================================================================
-- PHASE 2: Clear notifications and settings
-- =============================================================================
TRUNCATE notification_acks CASCADE;
TRUNCATE notification_deletes CASCADE;
TRUNCATE notification_settings CASCADE;
TRUNCATE notification_reads CASCADE;
TRUNCATE notifications RESTART IDENTITY CASCADE;

-- =============================================================================
-- PHASE 3: Clear all sessions and auth artifacts
-- =============================================================================
TRUNCATE password_resets CASCADE;
TRUNCATE auth_rate_limits CASCADE;
TRUNCATE admin_bootstrap_tokens CASCADE;
TRUNCATE refresh_tokens RESTART IDENTITY CASCADE;
TRUNCATE sessions CASCADE;

-- =============================================================================
-- PHASE 4: Clear IP associations and faucet/sink data
-- =============================================================================
TRUNCATE player_ip_associations CASCADE;
TRUNCATE player_faucet_claims CASCADE;
TRUNCATE player_star_variants CASCADE;
TRUNCATE player_boosts CASCADE;

-- =============================================================================
-- PHASE 5: Clear season data
-- =============================================================================
TRUNCATE season_calibration CASCADE;
TRUNCATE season_final_rankings CASCADE;
TRUNCATE season_end_snapshots CASCADE;
TRUNCATE season_economy CASCADE;
TRUNCATE season_controls CASCADE;
TRUNCATE season_control_events RESTART IDENTITY CASCADE;

-- =============================================================================
-- PHASE 6: Reset player stats (PRESERVE profile identity)
-- =============================================================================
UPDATE players SET
    coins = 0,
    stars = 0,
    last_coin_grant_at = NOW(),
    daily_earn_total = 0,
    last_earn_reset_at = NOW(),
    drip_multiplier = 1.0,
    drip_paused = FALSE,
    burned_coins = 0,
    last_active_at = NOW()
WHERE TRUE;

-- =============================================================================
-- PHASE 7: Clear global settings (will be recreated on startup)
-- =============================================================================
TRUNCATE global_settings CASCADE;

COMMIT;

-- =============================================================================
-- IMPORTANT NOTES:
-- =============================================================================
-- After running this script:
--   1. Restart the application to trigger season initialization
--   2. Or manually call: POST /admin/seasons/recovery
--   3. The server will auto-create a new season on startup
--
-- Preserved:
--   - accounts table (username, password_hash, display_name, email, etc.)
--   - players table (player_id, created_at, created_by, is_bot, bot_profile)
--
-- Reset:
--   - All balances (coins, stars)
--   - All activity tracking (daily_earn_total, last grants)
--   - All logs and history
--   - All sessions (users must log in again)
--   - All season state
-- =============================================================================
