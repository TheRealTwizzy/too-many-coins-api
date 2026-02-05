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

-- Alpha bootstrap uses ENV-seeded password; no gate key table.
