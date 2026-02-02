CREATE TABLE IF NOT EXISTS season_economy (
    season_id TEXT PRIMARY KEY,
    global_coin_pool BIGINT NOT NULL,
    emission_remainder DOUBLE PRECISION NOT NULL,
    last_updated TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS players (
    player_id TEXT PRIMARY KEY,
    coins BIGINT NOT NULL,
    stars BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    last_active_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE players
    ADD COLUMN IF NOT EXISTS last_coin_grant_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE TABLE IF NOT EXISTS player_ip_associations (
    player_id TEXT NOT NULL,
    ip TEXT NOT NULL,
    first_seen TIMESTAMPTZ NOT NULL,
    last_seen TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (player_id, ip)
);

CREATE INDEX IF NOT EXISTS idx_player_ip_associations_ip
    ON player_ip_associations (ip);
