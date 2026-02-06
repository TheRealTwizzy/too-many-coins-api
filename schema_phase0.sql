-- ============================================================
-- Too Many Coins — Phase 0 Minimal Schema (Alpha Reset Only)
--
-- Reset-friendly identity + persistence only.
-- ============================================================

CREATE TABLE accounts (
    account_id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

<<<<<<< HEAD
-- Admin accounts are not players; do not create player rows for admin roles.

=======
>>>>>>> a7f569c (Refactor authentication flow and database schema for Phase 0)
CREATE TABLE sessions (
    session_id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE players (
    player_id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(account_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL,
    last_login_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE player_state (
    player_id TEXT PRIMARY KEY REFERENCES players(player_id) ON DELETE CASCADE,
    state JSONB NOT NULL DEFAULT '{}'::jsonb
);

<<<<<<< HEAD
-- =========================
-- Season admin controls (key-value store)
-- Required by tick.go and admin_handlers.go for runtime control queries
-- =========================
CREATE TABLE IF NOT EXISTS season_controls (
    season_id UUID NOT NULL,
    control_name TEXT NOT NULL,
    value JSONB NOT NULL,
    expires_at TIMESTAMPTZ NULL,
    last_modified_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_modified_by UUID NOT NULL,
    PRIMARY KEY (season_id, control_name)
);

=======
>>>>>>> a7f569c (Refactor authentication flow and database schema for Phase 0)
-- Alpha bootstrap uses ENV-seeded password; no gate key table.
