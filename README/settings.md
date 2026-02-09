# Global Settings & Environment

## Scope
- **Type:** System contract
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible
- **Alpha Status:** Fully implemented

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

# Global Settings & Configuration Reference

Comprehensive taxonomy of **environment variables**, **feature flags**, **runtime settings**, and **emergency controls**.

---

## Environment Variables

### Core Configuration

| Variable | Type | Default | Purpose | Required |
|----------|------|---------|---------|----------|
| `DATABASE_URL` | string | — | PostgreSQL connection string | ✅ **Required** |
| `PORT` | string | `8080` | HTTP server port | Optional |
| `APP_ENV` | string | `local` | Runtime environment (`local`, `alpha`, `production`) | Optional |
| `PHASE` | string | `alpha` | Game phase (`alpha`, `beta`, `release`) | Optional |

**Example:**
```bash
DATABASE_URL=postgres://user:pass@localhost:5432/toomanycoins
PORT=8080
APP_ENV=alpha
PHASE=alpha
```

---

### Feature Flags

**Defined in:** [feature_flags.go](../feature_flags.go#L5-L21)

| Flag | Type | Default | Purpose |
|------|------|---------|---------|
| `ENABLE_FAUCETS` | bool | `true` | Enable/disable coin faucets (daily login, activity) |
| `ENABLE_SINKS` | bool | `true` | Enable/disable coin burning mechanics |
| `ENABLE_TELEMETRY` | bool | `true` | Enable/disable telemetry logging (earning events, star purchases) |
| `ENABLE_IP_THROTTLING` | bool | `true` | Enable/disable IP-based rate limiting and clustering detection |

**Boolean Parsing:**
- Accepted values: `true`, `1`, `yes` → Enabled
- All other values (including empty) → Use default

**Example:**
```bash
ENABLE_FAUCETS=true
ENABLE_SINKS=true
ENABLE_TELEMETRY=true
ENABLE_IP_THROTTLING=false  # Disable for local development
```

---

### Alpha Season Extension (Alpha Only)

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `ALPHA_SEASON_EXTENSION_DAYS` | int | `0` | Number of days to extend current Alpha season beyond 14-day base |
| `ALPHA_SEASON_EXTENSION_REASON` | string | — | Audit log reason for extension (required if EXTENSION_DAYS is set) |

**Constraints:**
- Maximum total season length: 90 days (alphaSeasonMaxDays)
- Extension only applies to current active season
- Reason required for audit trail

**Example:**
```bash
ALPHA_SEASON_EXTENSION_DAYS=7
ALPHA_SEASON_EXTENSION_REASON="Week 2 playtest: leaderboard dynamics"
```

---

### Development Mode

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `DEV_MODE` | bool | `false` | Enables development-only features (detailed logging, no TLS enforcement) |

**Example:**
```bash
DEV_MODE=true  # ⚠️  Only for local development
```

**Effects:**
- Detailed startup logging
- Bypass HTTPS-only cookie requirements
- Additional audit log verbosity

---

### Notification Retention

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `NOTIFICATION_RETENTION_HOURS` | int | `48` | Default retention for normal-priority notifications (hours) |

**Priority-Specific Retention:**
- Normal: `NOTIFICATION_RETENTION_HOURS` (default 48 hours)
- High: 7 days (168 hours, hardcoded)
- Critical: 14 days (336 hours, hardcoded)

**Example:**
```bash
NOTIFICATION_RETENTION_HOURS=72  # Extend normal retention to 3 days
```

---

## Global Settings (Database-Persisted)

**Table:** `global_settings`  
**Managed via:** `/admin/settings` endpoint

### Available Settings

| Setting | Type | Default | Purpose | Admin-Configurable |
|---------|------|---------|---------|-------------------|
| `drip_enabled` | bool | `false` | Enable passive income drip (disabled in Alpha) | ✅ Yes |
| `bots_enabled` | bool | `true` | Enable bot player actions | ✅ Yes |
| `active_drip_interval_seconds` | int | `60` | Active player passive income interval | ✅ Yes |
| `idle_drip_interval_seconds` | int | `240` | Idle player passive income interval | ✅ Yes |
| `active_drip_amount` | int | `2000` | Active player passive income (microcoins) | ✅ Yes |
| `idle_drip_amount` | int | `1000` | Idle player passive income (microcoins) | ✅ Yes |
| `activity_window_seconds` | int | `120` | Activity timeout window (seconds) | ✅ Yes |
| `bot_min_star_interval_seconds` | int | `90` | Minimum interval between bot star purchases | ✅ Yes |

**Example API Call:**
```json
PUT /admin/settings
{
  "activeDripIntervalSeconds": 90,
  "drip_enabled": false
}
```

---

## Emergency Season Controls

**Table:** `season_controls`  
**Managed via:** `/admin/seasons/:seasonId/controls` endpoint

### Available Controls

| Control | Type | Purpose | Use Case |
|---------|------|---------|----------|
| `pause_purchases` | bool | Disable all star buying | Emergency stop due to exploit detection |
| `reduce_emission` | bool | Lower daily emission target | Counter hyper-inflation |
| `freeze_season` | bool | Stop all economic activity | Critical system maintenance |

**Example:**
```json
POST /admin/seasons/season-uuid/controls
{
  "controlName": "pause_purchases",
  "value": true,
  "expiresAt": "2026-02-09T18:00:00Z",
  "reason": "Emergency stop due to exploit detection"
}
```

**Audit Trail:**
- All control changes logged in `season_control_events` table
- Includes: reason, old/new values, admin account ID, snapshot data (coins in circulation, active players, market pressure)

**TTL:** 60 seconds (cache refreshed every emission tick)

---

## Database Connection Pool

**Configured in:** [main.go](../main.go#L450-L452)

| Setting | Value | Purpose |
|---------|-------|---------|
| `MaxOpenConns` | `5` | Maximum concurrent database connections |
| `MaxIdleConns` | `5` | Maximum idle connections in pool |
| `ConnMaxLifetime` | `30 minutes` | Connection reuse lifetime |

**Rationale:**
- Low concurrency (5 connections) prevents PostgreSQL connection exhaustion
- 30-minute lifetime ensures connection health (auto-closes stale connections)
- Idle=Open prevents thrashing (all connections remain warm)

---

## Phase-Dependent Defaults

### PhaseAlpha (Current)

| Setting | Default | Rationale |
|---------|---------|-----------|
| `drip_enabled` | `false` | Simplify economy for initial testing |
| `season_length` | `14 days` | Short season for rapid iteration |
| `max_season_days` | `90 days` | Safety cap for extensions |
| `bootstrap_sealed` | Auto-seal after first login | Prevents admin proliferation |

**Enforcement:**
- Phase-dependent logic in [phase.go](../phase.go), [season.go](../season.go)
- Alpha-specific constraints validated at startup
- Cannot advance beyond 90-day total (including extensions)

---

### PhaseBeta (Future)

| Setting | Default | Rationale |
|---------|---------|-----------|
| `drip_enabled` | `true` | Passive income enabled for broader testing |
| `season_length` | `30 days` | Longer season for economic stabilization |
| `tradeable_assets` | Enabled | Cinder Sigils (TSA) become active |

---

### PhaseRelease (Future)

| Setting | Default | Rationale |
|---------|---------|-----------|
| `season_length` | `90 days` | Full-length competitive seasons |
| `cross_season_persistence` | Enabled | Account-wide unlocks and achievements |
| `advanced_telemetry` | Enabled | Full economic audit trail |

---

## Configuration Files

### fly.toml (Fly.io Deployment)

**Location:** [fly.toml](../fly.toml)

**Key Sections:**
```toml
[build]
  dockerfile = "Dockerfile"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = false
  auto_start_machines = true
  min_machines_running = 1

[[vm]]
  memory = '1gb'
  cpu_kind = 'shared'
  cpus = 1

[env]
  APP_ENV = "alpha"
  PHASE = "alpha"
```

**Persistent Volumes:** None (database external)

---

### railway.toml (Railway Deployment)

**Location:** [railway.toml](../railway.toml)

**Key Sections:**
```toml
[build]
  builder = "nixpacks"

[deploy]
  startCommand = "./server"

[[services]]
  name = "too-many-coins"
  port = 8080
```

---

### Dockerfile

**Location:** [Dockerfile](../Dockerfile)

**Key Steps:**
1. Multi-stage build (Go 1.23 builder)
2. Compile binary: `go build -o server .`
3. Minimal runtime image (alpine)
4. Expose port 8080

**No secrets embedded:** All configuration via environment variables

---

## Feature Flag Impact Matrix

| Flag | When Disabled | Impact on Players | Impact on Economy |
|------|--------------|-------------------|-------------------|
| `ENABLE_FAUCETS=false` | Daily login & activity faucets disabled | Cannot earn coins via faucets | Zero new coin emission (no inflation) |
| `ENABLE_SINKS=false` | Burn coins endpoint disabled | Cannot destroy coins | No deflationary pressure |
| `ENABLE_TELEMETRY=false` | Earning/purchase logs not created | No impact (player-invisible) | Cannot analyze economy trends |
| `ENABLE_IP_THROTTLING=false` | IP-based rate limits disabled | No auth throttling (dev only) | Vulnerable to sybil attacks (⚠️ never disable in production) |

---

## Schema Migration Notes

### Active Migrations

None currently pending.

### Historical Migrations

| Migration | Applied | Purpose |
|-----------|---------|---------|
| Add `current_star_price_micro` | 2025-Q4 | Replace decimal `current_star_price` with integer microcoins |
| Add `activity_warmup_*` columns | 2025-Q4 | Support UBI warmup system |
| Add `abuse_events` table | 2025-Q4 | Anti-cheat detection and enforcement |

**Migration Strategy:**
- Additive only (no destructive schema changes in Alpha)
- `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` (idempotent)
- No downtime required (columns nullable or default-valued)

---

## Startup Initialization Sequence

**Order of Operations:**

1. **Environment Loading** (`main.go:433`)
   - Parse `APP_ENV`, `PHASE`, `DATABASE_URL`, `PORT`
   - Load feature flags from env vars

2. **Database Connection** (`main.go:445-456`)
   - Open PostgreSQL connection pool
   - Validate ping
   - Configure connection pool limits

3. **Schema Enforcement** (`main.go:458-459`)
   - Run `ensureSchema(db)` (idempotent migrations)
   - Validate phase-environment consistency

4. **Startup Lock** (`main.go:467-481`)
   - Acquire PostgreSQL advisory lock (`pg_try_advisory_lock`)
   - Only one instance becomes "leader"

5. **Leader-Only Initialization** (`main.go:474-476`)
   - `ensureAlphaAdmin(db)` — Bootstrap first admin if needed
   - Load active season state
   - Initialize economy calibration

6. **Background Workers** (`main.go:522-533`)
   - Start emission tick loop (60-second interval)
   - Start notification pruner (hourly cleanup)
   - Start passive drip loop (1-minute interval, Alpha: disabled)

7. **HTTP Server** (`main.go:535-543`)
   - Register all routes
   - Listen on `0.0.0.0:$PORT`

---

## Design Philosophy

### Configuration Hierarchy

1. **Environment Variables** (highest priority)
   - Deployment-specific (credentials, URLs, ports)
   - Feature flags (global on/off switches)

2. **Database Settings** (runtime-configurable)
   - Economy parameters (drip rates, cooldowns)
   - Emergency controls (pause purchases, freeze season)

3. **Phase Defaults** (built-in fallbacks)
   - Alpha → minimal features
   - Beta → expanded features
   - Release → full feature set

### Admin Control

- **Environment vars:** Require deployment restart (immutable at runtime)
- **Global settings:** Admin-configurable via API (immediate effect)
- **Season controls:** Emergency stop switches with audit trail

### Safety

- **Phase validation:** Alpha env cannot run non-alpha phases
- **Season length cap:** 90-day maximum enforced at startup
- **Bootstrap sealing:** One-time admin creation prevents proliferation

---

## Technical Notes

**Performance:**
- Feature flags loaded once at startup (no per-request checks of env vars)
- Global settings cached with mutex (no DB query per request)
- Season controls cached with 60-second TTL (refreshed per emission tick)

**Security:**
- No secrets in Dockerfile or config files
- `DATABASE_URL` parsed but not logged (prevents credential leaks)
- `DEV_MODE` guards against accidental production use

**Edge Cases:**
- Missing `DATABASE_URL` → Fatal startup error (no fallback)
- Invalid `PHASE` value → Defaults to `alpha` (safe mode)
- Feature flag parsing → Empty string treated as default value

---

## See Also

- [Feature Flags](#environment-variables) — Feature flag impact matrix
- [Admin Tools](admin-tools.md) — Admin settings API endpoints
- [Phase 0 Contract](phase0-contract.md) — Alpha phase constraints
- [Persistent State](persistent-state.md) — Database schema and tables
