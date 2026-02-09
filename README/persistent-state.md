# Persistent State

## Scope
- **Type:** System contract
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

The database stores only authoritative facts required to reconstruct the game state.

Alpha note: the current schema is minimal and does not yet include all entities below. The list is the target canonical model and is post‑alpha unless explicitly implemented.

Currency model (post‑alpha canon, not implemented in Alpha):

Coins and Stars remain seasonal and reset each season.
Post‑alpha introduces a persistent meta currency for cosmetic / identity use only; it cannot be traded, cannot convert into Coins or Stars, and cannot affect competitive power.
An optional influence / reputation metric may exist post‑release; it is non‑spendable, eligibility/visibility‑only, and never convertible.
No currency may ever convert into Coins or Stars, directly or indirectly.

Post‑alpha seasonal instruments (Beta‑only):

Tradable Seasonal Assets (TSAs) are seasonal, player‑owned competitive assets (not currencies). TSAs are system‑minted only (including via Star sacrifice), reset at season end, and never convert into Coins or Stars or generate Coins or Stars. TSA trading is player‑negotiated; the system enforces legality, caps, and logging. Trading remains disabled in Alpha.

Persistent entities include:

Players:

player_id

account_id

created_at

last_login_at

trust_status (normal, throttled, flagged; admin-only internal flag)

Seasons:

season_id

start_time

end_time

current_day

status (active, ended)

PlayerSeasonState:

player_id

season_id

coin_balance

star_balance

daily_earn_total

last_earn_reset_at

last_action_at

EconomyState (per season):

season_id

current_base_price

market_pressure

global_coin_budget_remaining

last_emission_tick_at

trade_premium

trade_burn_rate

trade_eligibility_tightness

Purchases (append-only log):

purchase_id

player_id

season_id

star_quantity

total_coin_cost

price_snapshot

quantity_multiplier_snapshot

market_pressure_snapshot

created_at

Trades (append-only log):

trade_id

seller_player_id

buyer_player_id

season_id

star_quantity

coin_price

coin_burned

trade_premium_snapshot

eligibility_snapshot

created_at

TSA Mint Log (append-only, post‑alpha/Beta-only):

tsa_mint_id
tsa_type
season_id
player_id
minted_quantity
mint_source (star_sacrifice, system_drop)
stars_destroyed
active_players_snapshot
day_index
created_at

TSA Trade Log (append-only, post‑alpha/Beta-only):

tsa_trade_id
tsa_type
season_id
seller_player_id
buyer_player_id
consideration_summary
coin_burned
star_burned
friction_snapshot
destroyed
created_at

TSA Activation Log (append-only, post‑alpha/Beta-only):

tsa_activation_id
tsa_type
season_id
player_id
activation_effect
created_at

CoinEarnings (append-only log):

earning_id

player_id

season_id

source_type (ubi, login, activity, task, comeback, playability_safeguard)

amount

created_at

AbuseEvents:

event_id

player_id

season_id

event_type

severity

created_at

Derived values such as star prices, caps, throttles, and rankings are computed server-side and are not trusted if provided by clients.

All coin and star balance changes must occur inside database transactions.

---

## Database Indexing & Optimization

### Existing Indexes (Alpha Schema)

**Core Indexes:**

```sql
-- Sessions
CREATE INDEX idx_sessions_account_id ON sessions (account_id);

-- Player IP Associations
CREATE INDEX idx_player_ip_associations_ip ON player_ip_associations (ip);

-- Notifications
CREATE INDEX idx_notifications_created_at ON notifications (created_at);
CREATE INDEX idx_notifications_dedupe ON notifications (dedupe_key, created_at);

-- Bug Reports
CREATE INDEX idx_bug_reports_season_id ON bug_reports (season_id);
CREATE INDEX idx_bug_reports_player_id ON bug_reports (player_id);
CREATE INDEX idx_bug_reports_created_at ON bug_reports (created_at DESC);

-- Admin Password Gates (DEPRECATED)
CREATE UNIQUE INDEX idx_admin_password_gates_active 
    ON admin_password_gates (account_id) WHERE used_at IS NULL;
```

---

### Recommended Indexes (High-Value)

**Player Lookups (Not Yet Indexed):**

```sql
-- Fast player lookup by season + account_id
-- Used by: /player API (primary query pattern)
CREATE INDEX idx_players_season_account 
    ON players (season_id, account_id);

-- Fast player lookup by username
-- Used by: Admin search, leaderboard display
CREATE INDEX idx_accounts_username 
    ON accounts (username);

-- Fast player activity queries
-- Used by: Activity tracking, warmup calculations
CREATE INDEX idx_players_last_active 
    ON players (last_active_at DESC);
```

**Leaderboard Queries:**

```sql
-- Leaderboard by stars (descending)
-- Used by: /leaderboard endpoint (primary sort)
CREATE INDEX idx_players_stars_desc 
    ON players (season_id, stars DESC);

-- Leaderboard with tie-breaking (stars DESC, coins DESC)
-- Used by: Final ranking calculations
CREATE INDEX idx_players_ranking 
    ON players (season_id, stars DESC, coins DESC);
```

**Economy Telemetry:**

```sql
-- Star purchase log by player + season
-- Used by: Admin telemetry, player purchase history
CREATE INDEX idx_star_purchase_log_player 
    ON star_purchase_log (player_id, season_id, created_at DESC);

-- Coin earning log by player + season
-- Used by: Admin telemetry, earning audits
CREATE INDEX idx_coin_earning_log_player 
    ON coin_earning_log (player_id, season_id, created_at DESC);

-- Time-series queries for economy analysis
CREATE INDEX idx_star_purchase_log_created 
    ON star_purchase_log (created_at DESC);
CREATE INDEX idx_coin_earning_log_created 
    ON coin_earning_log (created_at DESC);
```

**Abuse & Moderation:**

```sql
-- Abuse events by player
-- Used by: Abuse score calculation, enforcement checks
CREATE INDEX idx_abuse_events_player 
    ON abuse_events (player_id, created_at DESC);

-- Abuse events by severity (high-priority queries)
-- Used by: Admin anti-cheat dashboard
CREATE INDEX idx_abuse_events_severity 
    ON abuse_events (severity DESC, created_at DESC);

-- Admin audit log by admin account
-- Used by: Admin action history, accountability audits
CREATE INDEX idx_admin_audit_log_admin 
    ON admin_audit_log (admin_account_id, created_at DESC);

-- Admin audit log by action type
-- Used by: Compliance queries (e.g., "all role_update actions")
CREATE INDEX idx_admin_audit_log_action 
    ON admin_audit_log (action_type, created_at DESC);
```

---

### Query Pattern Analysis

**Most Frequent Queries:**

| Query Pattern | Frequency | Current Index | Recommended Index |
|---------------|-----------|---------------|-------------------|
| `SELECT * FROM players WHERE account_id = ? AND season_id = ?` | Every /player request | ❌ None (table scan) | ✅ `idx_players_season_account` |
| `SELECT * FROM players WHERE season_id = ? ORDER BY stars DESC LIMIT 50` | Every /leaderboard request | ❌ None (sort on disk) | ✅ `idx_players_stars_desc` |
| `SELECT * FROM star_purchase_log WHERE player_id = ? ORDER BY created_at DESC` | Admin telemetry queries | ❌ None (table scan) | ✅ `idx_star_purchase_log_player` |
| `SELECT * FROM abuse_events WHERE player_id = ? ORDER BY created_at DESC` | Abuse score calculations | ❌ None (table scan) | ✅ `idx_abuse_events_player` |

**Impact Estimate:**
- **Player queries:** 10-100x faster with composite index (season_id + account_id)
- **Leaderboard queries:** 5-50x faster with stars DESC index (avoids full table sort)
- **Telemetry queries:** 20-200x faster with time-series indexes (created_at DESC)

---

### Missing Tables (Schema Gaps)

**Identified During P1 Canon Promotion:**

⚠️ **Warning:** The following tables are used in code but **missing from schema.sql**:

```sql
-- Activity Boosts (used in buy_boost handler)
CREATE TABLE IF NOT EXISTS player_boosts (
    player_id TEXT NOT NULL,
    boost_type TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (player_id, boost_type)
);

CREATE INDEX idx_player_boosts_expires 
    ON player_boosts (expires_at) WHERE expires_at > NOW();

-- Variant Stars (used in buy_variant_star handler)
CREATE TABLE IF NOT EXISTS player_star_variants (
    player_id TEXT NOT NULL,
    variant TEXT NOT NULL,
    quantity BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (player_id, variant)
);

-- Abuse Events (used in abuse detection system)
CREATE TABLE IF NOT EXISTS abuse_events (
    event_id BIGSERIAL PRIMARY KEY,
    player_id TEXT NOT NULL,
    season_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    severity INT NOT NULL,
    score_delta DOUBLE PRECISION NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_abuse_events_player 
    ON abuse_events (player_id, created_at DESC);
CREATE INDEX idx_abuse_events_severity 
    ON abuse_events (severity DESC, created_at DESC);

-- Admin Audit Log (used in admin action tracking)
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id BIGSERIAL PRIMARY KEY,
    admin_account_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    scope_type TEXT NOT NULL,
    scope_id TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    details JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_audit_log_created 
    ON admin_audit_log (created_at DESC);
CREATE INDEX idx_admin_audit_log_admin 
    ON admin_audit_log (admin_account_id, created_at DESC);
CREATE INDEX idx_admin_audit_log_action 
    ON admin_audit_log (action_type, created_at DESC);
```

**Action Required:** Add these tables to `schema.sql` in next schema consolidation pass.

---

### PostgreSQL Maintenance

**Recommended Maintenance Schedule:**

| Operation | Frequency | Purpose | Command |
|-----------|-----------|---------|---------|
| **VACUUM** | Weekly | Reclaim dead row space, prevent bloat | `VACUUM ANALYZE;` |
| **ANALYZE** | Daily | Update query planner statistics | `ANALYZE;` |
| **REINDEX** | Monthly | Rebuild indexes (remove bloat) | `REINDEX TABLE players;` |

**Automatic Maintenance (Alpha):**

PostgreSQL autovacuum enabled by default (recommended):
```sql
-- Check autovacuum status
SELECT name, setting FROM pg_settings WHERE name LIKE 'autovacuum%';

-- Verify last vacuum/analyze
SELECT relname, last_vacuum, last_autovacuum, last_analyze, last_autoanalyze 
FROM pg_stat_user_tables 
ORDER BY last_autovacuum DESC;
```

**Manual Maintenance (Production):**

```bash
# Weekly maintenance script
#!/bin/bash
psql $DATABASE_URL -c "VACUUM ANALYZE;"
psql $DATABASE_URL -c "REINDEX DATABASE toomanycoins;"
```

---

### Performance Monitoring

**Query Performance Tracking:**

```sql
-- Enable query statistics (set in postgresql.conf)
shared_preload_libraries = 'pg_stat_statements'
pg_stat_statements.track = all

-- View slowest queries
SELECT query, calls, total_exec_time, mean_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- View most frequent queries
SELECT query, calls, total_exec_time
FROM pg_stat_statements
ORDER BY calls DESC
LIMIT 10;
```

**Index Usage Statistics:**

```sql
-- Identify unused indexes (candidates for removal)
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY indexname;

-- Identify missing indexes (table scans)
SELECT schemaname, tablename, seq_scan, seq_tup_read, 
       idx_scan, idx_tup_fetch, seq_tup_read / seq_scan AS avg_seq
FROM pg_stat_user_tables
WHERE seq_scan > 0
ORDER BY seq_tup_read DESC;
```

---

### Database Size & Growth

**Current Size Estimates (Alpha, 14-day season):**

| Table | Rows | Size Per Row | Total Size |
|-------|------|--------------|------------|
| `players` | 100 | 500 bytes | 50 KB |
| `accounts` | 100 | 400 bytes | 40 KB |
| `star_purchase_log` | 5,000 | 200 bytes | 1 MB |
| `coin_earning_log` | 50,000 | 150 bytes | 7.5 MB |
| `notifications` | 500 | 300 bytes | 150 KB |
| `abuse_events` | 200 | 250 bytes | 50 KB |

**Total Database:** ~10 MB (Alpha, 100 players, 14-day season)

**Growth Projections:**

| Metric | 100 Players | 1,000 Players | 10,000 Players |
|--------|-------------|---------------|----------------|
| **14-day season** | 10 MB | 100 MB | 1 GB |
| **90-day season** | 65 MB | 650 MB | 6.5 GB |
| **1 year (4 seasons)** | 260 MB | 2.6 GB | 26 GB |

**Index Overhead:** ~30-50% of table size (indexes stored separately)

---

### Connection Pool Configuration

**Current Settings ([main.go](../main.go#L450-L452)):**

```go
db.SetMaxOpenConns(5)      // Maximum concurrent connections
db.SetMaxIdleConns(5)      // Maximum idle connections in pool
db.SetConnMaxLifetime(30 * time.Minute)  // Connection reuse lifetime
```

**Rationale:**
- **Low concurrency (5):** Alpha workload <50 concurrent requests
- **Idle=Open:** All connections remain warm (no connection thrashing)
- **30-minute lifetime:** Automatic connection refresh prevents stale connections

**Scaling Recommendations:**

| Workload | MaxOpenConns | MaxIdleConns | ConnMaxLifetime |
|----------|--------------|--------------|-----------------|
| **Alpha (100 players)** | 5 | 5 | 30 minutes |
| **Beta (1,000 players)** | 20 | 10 | 15 minutes |
| **Release (10,000 players)** | 50 | 25 | 10 minutes |

---

### Transaction Isolation

**Default Level:** `READ COMMITTED` (PostgreSQL default)

**Critical Transactions (require serialization):**

```sql
-- Star purchase (prevent double-spend)
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;
UPDATE players SET coins = coins - ?, stars = stars + ? WHERE player_id = ?;
INSERT INTO star_purchase_log (...) VALUES (...);
COMMIT;

-- Economy emission (prevent race conditions)
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE;
UPDATE season_economy SET global_coin_pool = global_coin_pool + ? WHERE season_id = ?;
UPDATE players SET coins = coins + ? WHERE player_id IN (...);
COMMIT;
```

**Best Practices:**
- Use `SERIALIZABLE` for money-like operations (coins, stars)
- Use `READ COMMITTED` for read-only queries (leaderboard, telemetry)
- Keep transactions short (minimize lock duration)

---

### Backup & Recovery

**Recommended Backup Strategy:**

| Backup Type | Frequency | Retention | Purpose |
|-------------|-----------|-----------|---------|
| **Full Backup** | Daily | 7 days | Disaster recovery |
| **Incremental Backup** | Hourly | 24 hours | Point-in-time recovery |
| **WAL Archiving** | Continuous | 7 days | Transaction-level recovery |

**Backup Commands:**

```bash
# Full backup (pg_dump)
pg_dump $DATABASE_URL --format=custom --file=backup_$(date +%Y%m%d).dump

# Restore from backup
pg_restore --dbname=$DATABASE_URL --clean --if-exists backup_20260209.dump

# Point-in-time recovery (requires WAL archiving)
pg_basebackup -D /backup/base -Ft -z -P
```

**Alpha Backup (Fly.io):**
- **Managed by:** Fly.io Postgres managed service
- **Frequency:** Daily automated snapshots
- **Retention:** 7 days (free tier)

---

### Edge Cases & Constraints

**Uniqueness Constraints:**
- `accounts.username` — UNIQUE (case-sensitive, no collisions)
- `players.player_id` — PRIMARY KEY (UUID, globally unique)
- `sessions.session_id` — PRIMARY KEY (cryptographically random)

**Foreign Key Constraints:**
- **Not enforced in Alpha** (schema uses TEXT references, not FK constraints)
- Reason: Simplifies testing (no cascading deletes required)
- Post-Alpha: Add FK constraints for data integrity

**Soft Deletes:**
- `notifications.deleted_at` — Soft delete (hidden from queries, preserved in DB)
- `notification_deletes` table — Explicit per-player deletion tracking

**Append-Only Tables:**
- `star_purchase_log` — No UPDATE or DELETE queries (immutable audit trail)
- `coin_earning_log` — No UPDATE or DELETE queries (compliance requirement)
- `admin_audit_log` — No UPDATE or DELETE queries (accountability)

---

## See Also

- [HTTP API Reference](http-api-reference.md) — Query patterns and endpoint usage
- [Settings](settings.md) — Database connection pool configuration
- [Admin Audit Log](admin-audit-log.md) — Audit log schema and indexing
- [Telemetry](../telemetry.go) — Telemetry logging and query patterns