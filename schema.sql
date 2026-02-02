CREATE TABLE IF NOT EXISTS season_economy (
    season_id TEXT PRIMARY KEY,
    global_coin_pool BIGINT NOT NULL,
    emission_remainder DOUBLE PRECISION NOT NULL,
    last_updated TIMESTAMPTZ NOT NULL
);
