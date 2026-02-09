# Rate Limits & Throttling

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

# Rate Limiting & Throttling Reference

Comprehensive taxonomy of **global rate limits**, **IP throttling**, **anti-spam timeouts**, and **abuse enforcement multipliers**.

---

## Authentication Rate Limits

**Defined in:** [auth_protection.go](../auth_protection.go#L12-L24)

| Action | Limit | Window | Scope | Error Code |
|--------|-------|--------|-------|------------|
| **Signup** | 5 requests | 600 seconds (10 min) | Per IP | `RATE_LIMIT_SIGNUP` |
| **Login** | 12 requests | 600 seconds (10 min) | Per IP | `RATE_LIMIT_LOGIN` |
| **General Auth** | 10 requests | 600 seconds (10 min) | Per IP | `RATE_LIMIT_AUTH` |

**Constants:**
```go
const (
    SIGNUP_RATE_LIMIT = 5   // per IP per 600 seconds
    LOGIN_RATE_LIMIT  = 12  // per IP per 600 seconds
    AUTH_RATE_LIMIT   = 10  // per IP per 600 seconds
)
```

**Implementation:**
- Stored in `auth_rate_limits` table (ip, action, window_start, attempt_count)
- Resets after 600-second window expires
- 429 status returned when limit exceeded

**Endpoints Affected:**
- `POST /auth/signup`
- `POST /auth/login`
- `POST /auth/request-reset`
- `POST /auth/reset-password`

---

## Faucet Cooldowns

### Daily Login Faucet
**Defined in:** [calibration.go](../calibration.go#L105)

- **Base Cooldown:** 20 hours (72000 seconds)
- **Endpoint:** `POST /claim-daily`
- **Scaling:** Subject to seasonal scaling (0.5x–1.7x multiplier)

**Effective Range:** 10–34 hours (depending on season phase)

---

### Activity Faucet
**Defined in:** [calibration.go](../calibration.go#L104), [handlers.go](../handlers.go#L1960)

- **Base Cooldown:** 360 seconds (6 minutes)
- **Valid Range:** 300–720 seconds (5–12 minutes)
- **Endpoint:** `POST /claim-activity`

**Cooldown Modifiers:**
1. **Seasonal Scaling:** 0.5x–1.7x (early vs late season)
2. **Account Age Multiplier:** 1.0x–1.6x (old vs new accounts)
3. **Abuse Enforcement Jitter:** +0–50% (severity-dependent randomization)
4. **Market Pressure:** Cooldowns extended during high pressure (≥1.7)

**Example Calculation:**
```
base = 360 seconds
seasonal_multiplier = 1.2 (mid-season)
account_age_multiplier = 1.3 (new account)
abuse_jitter = 0.25 (severity 2)

effective_cooldown = 360 × 1.2 × 1.3 = 561 seconds
jitter_added = rand(0, 561 × 0.25) = rand(0, 140) seconds
final_cooldown = 561 + jitter = 561–701 seconds
```

---

## Activity System Timeouts

### Activity Window
**Defined in:** [settings.go](../settings.go#L31)

- **Default:** 120 seconds (2 minutes)
- **Purpose:** Defines how long a player is considered "active" after last activity ping
- **Used For:** UBI warmup calculations, passive income tier determination

**Endpoint:** `POST /activity` (extends activity window by 120 seconds)

---

### UBI Warmup Constants
**Defined in:** [faucet.go](../faucet.go#L244-L252)

| Constant | Value | Purpose |
|----------|-------|---------|
| **warmupDurationSeconds** | 1800 (30 min) | Time to reach maximum warmup (10x) |
| **maxWarmupMultiplier** | 10.0x | Maximum UBI income multiplier |
| **warmupDecayBaseRate** | 0.002 per tick | Warmup decay when idle (60-second ticks) |
| **baseUBIPerTick** | 1 microcoin | Minimum baseline UBI income |

**Effective Warmup Scaling:**
- 0 minutes → 1.0x multiplier
- 15 minutes → ~5.5x multiplier
- 30 minutes → 10.0x multiplier (max)

---

## Bot Rate Limiting

**Defined in:** [settings.go](../settings.go#L32), [bot_runner.go](../cmd/bot-runner/main.go)

| Limit | Value | Purpose |
|-------|-------|---------|
| **BotMinStarIntervalSeconds** | 90 seconds | Minimum interval between bot star purchases |
| **BOT_RATE_LIMIT_MIN_MS** | 3000 ms (3 seconds) | Minimum jitter between bot actions |
| **BOT_RATE_LIMIT_MAX_MS** | 12000 ms (12 seconds) | Maximum jitter between bot actions |

**Jitter Formula:**
```go
jitter = rand(BOT_RATE_LIMIT_MIN_MS, BOT_RATE_LIMIT_MAX_MS)
sleep(jitter)
```

**Purpose:** Prevents synchronized bot requests from creating artificial pressure spikes.

---

## Abuse Enforcement Multipliers

**Defined in:** [abuse.go](../abuse.go#L41-L115)

### Severity Tier Thresholds

| Severity | Score Range | Enforcement Level |
|----------|-------------|-------------------|
| **0** | 0–9.99 | No enforcement |
| **1** | 10–24.99 | Light throttling |
| **2** | 25–44.99 | Moderate penalties |
| **3** | 45+ | Heavy penalties |

---

### Enforcement Multipliers by Severity

| Severity | Price Multiplier | Max Bulk Qty | Earn Multiplier | Cooldown Jitter |
|----------|------------------|--------------|-----------------|-----------------|
| **1** | 1.05x (+5%) | 4 stars | 0.9x (-10%) | 0.1 (10%) |
| **2** | 1.15x (+15%) | 3 stars | 0.75x (-25%) | 0.25 (25%) |
| **3** | 1.3x (+30%) | 2 stars | 0.6x (-40%) | 0.5 (50%) |

---

### Cooldown Jitter Calculation
**Defined in:** [abuse.go](../abuse.go#L118-L135)

```go
func abuseJitterCooldown(baseCooldown time.Duration, jitterFactor float64) time.Duration {
    maxJitter := 5 * time.Minute  // 300 seconds cap
    jitter := time.Duration(float64(baseCooldown) * jitterFactor)
    
    if jitter > maxJitter {
        jitter = maxJitter
    }
    
    return baseCooldown + time.Duration(rand.Int63n(int64(jitter)))
}
```

**Example:**
```
baseCooldown = 360 seconds (activity faucet)
severity = 2 (jitterFactor = 0.25)

jitter = rand(0, 360 × 0.25) = rand(0, 90) seconds
finalCooldown = 360 + jitter = 360–450 seconds
```

**Maximum Jitter Cap:** 5 minutes (300 seconds)  
- Prevents cooldowns from becoming unreasonably long even at severity 3

---

### Abuse Score Decay Rates
**Defined in:** [abuse.go](../abuse.go#L58-L67)

| Severity | Decay Rate | Time to Clear (from min score) |
|----------|------------|-------------------------------|
| **0** | 1.0 per hour | ~10 hours (from 10) |
| **1** | 0.6 per hour | ~25 hours (from 25) |
| **2** | 0.3 per hour | ~83 hours (from 45) |
| **3** | 0.15 per hour | ~300 hours (from score 100+) |

**Persistence Lockouts:**
- **Severity 2:** 72 hours (3 days) if multiple signals within 6 hours
- **Severity 3:** 168 hours (7 days) if multiple signals within 6 hours

---

## System Tick Intervals

**Defined in:** [tick.go](../tick.go#L11-L12)

| Tick Type | Interval | Purpose |
|-----------|----------|---------|
| **emissionTickInterval** | 60 seconds | Economy emission calculations, UBI distribution, market pressure updates |
| **seasonControlsCacheTTL** | 60 seconds | Cache refresh for emergency season controls |

**Emission Tick Actions:**
1. Distribute UBI to active players
2. Update market pressure
3. Adjust star prices
4. Check daily emission cap
5. Trigger abuse decay

---

## Passive Income Intervals

**Defined in:** [calibration.go](../calibration.go#L87-L90), [settings.go](../settings.go#L25-L31)

| State | Interval | Amount | Condition |
|-------|----------|--------|-----------|
| **Active Drip** | 60 seconds | 2 coins (2000 microcoins) | Player active within 120 seconds |
| **Idle Drip** | 240 seconds (4 min) | 1 coin (1000 microcoins) | Player inactive for >120 seconds |

**Scaling Factors:**
- **Seasonal Multiplier:** 0.6x–1.6x (reward scaling)
- **Player Drip Multiplier:** 0.5x–2.0x (per-player modifier, admin-adjustable)
- **IP Dampening:** Reward reduction for multi-accounting
- **Abuse Enforcement:** 0.6x–1.0x (severity-dependent)

**Note:** Passive drip **disabled by default** in Alpha (Phase 0)

---

## Login Safeguard Cooldown

**Defined in:** [earnings.go](../earnings.go#L13)

- **Duration:** 2 minutes (120 seconds)
- **Purpose:** Prevents double-earning on rapid login/logout cycles
- **Applied to:** Initial login earning grant

---

## Owner Bootstrap Claim Window

**Defined in:** [admin_bootstrap.go](../admin_bootstrap.go#L18)

- **Duration:** 5 minutes (300 seconds)
- **Purpose:** Claim code expires after 5 minutes
- **One-time use:** Token invalidated after successful claim

---

## IP Throttling

**Defined in:** [feature_flags.go](../feature_flags.go#L19), [README.md](../README.md#L151)

- **Feature Flag:** `ENABLE_IP_THROTTLING` (default `true`)
- **Disable for Development:** `ENABLE_IP_THROTTLING=false`

**IP-Based Rate Limiting Effects:**
1. **Auth Rate Limits:** Separate counters per IP
2. **IP Clustering Detection:** Multiple accounts from same IP flagged
3. **Reward Dampening:** Reduced earnings for multi-accounting (see [anti-abuse.md](anti-abuse.md))

---

## Purchase Bulk Limits

**Defined in:** [handlers.go](../handlers.go), [abuse.go](../abuse.go)

**Default:** 10 stars per purchase (configurable)

**Abuse-Adjusted Limits:**
| Severity | Max Bulk Qty |
|----------|--------------|
| 0 | 10 stars |
| 1 | 4 stars |
| 2 | 3 stars |
| 3 | 2 stars |

**Account Age Multiplier:**
- New accounts: 0.2x–1.0x (severely restricted)
- Old accounts: 1.0x (full access)

**Example:**
```
defaultBulkMax = 10 stars
accountAgeMultiplier = 0.5 (new account)
abuseMaxQty = 3 (severity 2)

effectiveMax = min(10 × 0.5, 3) = min(5, 3) = 3 stars
```

---

## Error Codes

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `RATE_LIMIT_SIGNUP` | 429 | Signup rate limit exceeded (5/10min per IP) |
| `RATE_LIMIT_LOGIN` | 429 | Login rate limit exceeded (12/10min per IP) |
| `RATE_LIMIT_AUTH` | 429 | General auth rate limit exceeded (10/10min per IP) |
| `COOLDOWN_ACTIVE` | 409 | Faucet cooldown not expired |
| `DAILY_CAP_REACHED` | 429 | Daily earning cap reached |

---

## Design Philosophy

### Abuse Mitigation
- **Graduated Enforcement:** Severity tiers allow proportional response
- **Cooldown Jitter:** Prevents synchronized bot behavior
- **Score Decay:** Allows recovery from false positives over time

### User Experience
- **Seasonal Scaling:** Early-season generosity, late-season scarcity
- **Account Age Protection:** New accounts restricted to prevent throwaway abuse
- **Transparent Limits:** Error responses include retry-after headers

### System Stability
- **IP Throttling:** Prevents single-IP flooding
- **Tick Intervals:** Predictable economy updates (60-second heartbeat)
- **Rate Limit Tables:** Isolated state per IP, no global locks

---

## Technical Notes

**Database Tables:**
- `auth_rate_limits` — IP-based auth rate limiting
- `players` — Per-player cooldown state (last_coin_grant_at, last_earn_reset_at)
- `abuse_events` — Abuse score calculation and enforcement

**Performance:**
- Auth rate limits: Single table scan per IP+action (indexed)
- Cooldown checks: Single player row read (primary key lookup)
- Jitter randomization: Cryptographically secure RNG (not predictable)

**Edge Cases:**
- Cooldown jitter max: 5 minutes (300 seconds) even at severity 3
- Zero-earning players: Still subject to cooldowns (prevents timing attacks)
- Expired auth windows: Reset on next attempt (no leftover state)

---

## See Also

- [Anti-Abuse](anti-abuse.md) — Abuse detection event types and severity scoring
- [Anti-Cheat Events](anti-cheat-events.md) — Event taxonomy and enforcement multipliers
- [Activity System](activity-system.md) — Activity window and warmup mechanics
- [Coin Faucets](coin-faucets.md) — Faucet cooldown formulas and seasonal scaling
- [HTTP API Reference](http-api-reference.md) — Error codes and HTTP status mapping
