# Activity System

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

# Player Activity Tracking & Dynamic Rewards

The Activity System enables player engagement tracking and dynamic reward scaling. Active players earn more coins and build warmup multipliers; inactive players gradually lose both as an incentive to return.

---

## Definition: What is Activity?

**Activity** = Temporal proximity to active gameplay.

A player is considered **active** if the current time minus their last activity timestamp is within the **activity window** (default: 2 minutes, configurable via global settings).

```
last_active_at := timestamp when player last called POST /activity

now - last_active_at ≤ activity_window  →  player is ACTIVE
now - last_active_at > activity_window  →  player is IDLE
```

**Key:** Activity is NOT tied to specific actions. The frontend advances `last_active_at` every 20 seconds while the player is present, regardless of which endpoint they use.

---

## Activity Tracking

### `POST /activity` Endpoint

**Purpose:** Update player's `last_active_at` timestamp to signal ongoing engagement.

**Request:**
```json
POST /activity
Content-Type: application/json
Authentication: Session cookie (required)

{}  // Empty body
```

**Response:**
```json
{
  "ok": true,
  "timestamp": "2026-02-09T12:34:56Z"
}
```

**Behavior:**
- Called by frontend every 20 seconds during active gameplay
- Updates `last_active_at` in players table to current time
- Does NOT grant coins or trigger faucet mechanics
- Does NOT reset cooldowns or daily cap
- Purely a "presence signal" for engagement tracking

**Data Persisted:**
- `players.last_active_at` ← current timestamp

---

## Activity Window

**Default Duration:** 2 minutes (120 seconds)

**Configurable Via:** Global setting `activity_window_seconds`

**Example:**
```
player.last_active_at = 12:00:00 UTC
current_time = 12:01:50 UTC

elapsed = 12:01:50 - 12:00:00 = 1 min 50 sec
activity_window = 2 min

1 min 50 sec < 2 min  →  Player is ACTIVE
```

An idle player who stops sending `/activity` pings will become IDLE after 2+ minutes without any interaction.

---

## Activity Faucet (Frequent Claim)

### `POST /claim-activity` Endpoint

**Purpose:** Active players earn coins frequently by claiming the Activity Faucet.

**Request:**
```json
POST /claim-activity
Content-Type: application/json
Authentication: Session cookie (required)

{}  // Empty body
```

**Response (Success):**
```json
{
  "ok": true,
  "reward": 12,                          // coins granted (microcoins)
  "playerCoins": 5234,                   // new coin balance after grant
  "nextAvailableInSeconds": 342          // when next claim can occur
}
```

**Response (Failure):**
```json
{
  "ok": false,
  "error": "COOLDOWN",                   // or other error code
  "nextAvailableInSeconds": 342          // context-specific
}
```

### Eligibility Requirements (Checked in Order)

1. **Season Active** — Current season has not ended (`SEASON_ENDED`)
2. **Feature Enabled** — Faucets enabled via `ENABLE_FAUCETS` flag (`FEATURE_DISABLED`)
3. **Player Valid** — Player registered and valid (`PLAYER_NOT_REGISTERED`)
4. **Cooldown Satisfied** — Since last activity claim OR daily login claim, cooldown has elapsed (`COOLDOWN`)
5. **Daily Cap Available** — Remaining daily coin limit > 0 (`DAILY_CAP`)
6. **Emission Pool Available** — Global coin pool has coins to emit (`EMISSION_EXHAUSTED`)
7. ****Player is ACTIVE**  —  `now - last_active_at ≤ activity_window` (**critical difference from daily login**)

If any requirement fails, claim is denied with appropriate error code.

### Cooldown Mechanics

**Base Cooldown:** ~360 seconds (6 minutes)

**Calculated as:**
```
cooldown = activity_cooldown_seconds (from calibration)
         × seasonScaling (early vs late season)
         × accountAgeScaling (new accounts may have different cooldown)
         × abuseJitter (anti-pattern detection adds variance)
```

**Interaction with Daily Login Faucet:**
- Both activity and daily claims share the same **daily cap** (total coin limit per day)
- Cooldowns are **independent** (activity has 6-min window, daily has 20-hour window)
- Can claim both if eligible and cap-limited, but together they consume the daily budget

### Reward Amount

**Base Reward:** ~4% of daily early-season emission cap (varies by population)

**Example (at 100 players):**
```
daily_cap_early = ~60 coins
activity_reward = 60 × 0.04 = ~2.4 coins per claim
```

**Scaling:** Reward scales with population (more players → higher emission cap → higher activity reward)

**Bonuses:**
- Activity boosts (+1 coin if active boost purchased)
- Never scaled below 1 microcoin

### Daily Cap Interaction

Activity claims count toward the daily earning cap like all faucet claims.

**Example:**
```
daily_cap = 60 coins
claimed_today = 50 coins (from various faucets)
remaining_cap = 10 coins

claim_activity()
→ request gives 12 coins worth
→ capped to remaining 10 coins
→ player receives 10 coins
→ remaining_cap becomes 0
```

Next activity claim (within 6 min) will fail with `DAILY_CAP` error until date resets.

---

## Comparison: Activity vs Daily Login Faucet

| Mechanic | Activity | Daily Login |
|----------|----------|-------------|
| **Endpoint** | `POST /claim-activity` | `POST /claim-daily` |
| **Cooldown** | ~6 minutes | ~20 hours |
| **Frequency** | Every 6 min (if active) | Once per day |
| **Reward** | ~4% daily cap | ~25% daily cap |
| **Requirement** | Player must be ACTIVE | No activity requirement |
| **Boost** | Supported (BoostActivity) | None |
| **Daily Cap** | Counts toward limit | Counts toward limit |
| **Daily Cap Impact** | Uses 6× or more per day (if claimed frequently) | Uses ~25% of cap |
| **Strategic Use** | Engaged players earn steady passive income | Casual players get larger one-time reward |

---

## Activity Warmup Multiplier

The **Activity Warmup System** amplifies UBI (Universal Basic Income) rewards based on sustained engagement.

### Warmup Level (0.0 to 1.0)

A floating-point value representing the player's engagement streak:

```
warmup_level = 0.0    →  Idle; base UBI (1x multiplier)
warmup_level = 0.5    →  Semi-active; 5.5x UBI multiplier
warmup_level = 1.0    →  Highly active; 10x UBI multiplier
```

### Accumulation (Active Phase)

While player is **active** (within activity window):

```
warmup_duration = 30 minutes
warmup_increase_per_tick = 1 / (30 × 60) minutes
                         = ~0.000556 per tick (1-2 second ticks)

Each tick: warmup_level += 0.000556
           warmup_level = min(warmup_level, 1.0)
```

**Player reaches maximum warmup (1.0) after 30 minutes of continuous activity.**

### Decay (Idle Phase)

While player is **idle** (outside activity window):

```
base_decay_rate = 0.002 per tick

decay_multiplier = 1.0 + (recent_activity_ratio × 2.0)
                = 1.0 to 3.0x depending on engagement history

effective_decay = base_decay_rate / decay_multiplier

Each tick: warmup_level -= effective_decay
           warmup_level = max(warmup_level, 0.0)
```

**Decay is slower after recent activity:** If player was just active, they retain warmup 3× longer than an idle player would.

**Time to lose warmup (from full to zero):**
- Just became idle: ~30-90 minutes
- Been idle for hours: ~15 minutes

### Recent Activity Seconds (Decay Driver)

`recent_activity_seconds` tracks elapsed time while active to control warmup decay rate.

**Accumulation:**
```
while active:
  recent_activity_seconds += 60 seconds per tick
  cap at: 2 × warmup_duration_seconds = 3600 seconds
```

**Decay:**
```
while idle:
  recent_activity_seconds -= 30 seconds per tick (half rate)
  floor at: 0
```

**Used in:**
```
activity_ratio = recent_activity_seconds / 1800  (1800 = 30 min)
decay_multiplier = 1.0 + (activity_ratio × 2.0)
```

Higher `recent_activity_seconds` → slower decay of warmup level → engaged players are more forgiving of brief absences.

### UBI Multiplier Calculation

```
base_ubi = 1 microcoin per tick
multiplier = 1.0 + (warmup_level × 9.0)  (ranges 1x to 10x)
actual_ubi = base_ubi × multiplier

warmup=0.0  →  1 microcoin/tick
warmup=0.5  →  5.5 microcoin/tick
warmup=1.0  →  10 microcoin/tick
```

**Updated:** Every UBI tick (1-2 seconds), automatically.

---

## Update Frequency & Lifecycle

| Data | Updated By | Frequency | Trigger |
|------|-----------|-----------|---------|
| `last_active_at` | `POST /activity` | ~every 20 seconds | Frontend ping |
| `activity_warmup_level` | UBI distribution loop | ~every 1-2 seconds | Server tick |
| `activity_warmup_updated_at` | UBI distribution loop | ~every 1-2 seconds | Server tick |
| `recent_activity_seconds` | UBI distribution loop | ~every 1-2 seconds | Server tick |

### Player-Visible State (API)

```json
GET /player
{
  "activityWarmup": 0.75,                 // 0.0-1.0
  "ubiMultiplier": 7.75,                  // derived: 1.0 + (0.75 × 9.0)
  "currentUBIPerTick": 7,                 // current UBI in microcoins
  "lastActiveAt": "2026-02-09T12:34:56Z"
}
```

---

## Edge Cases & Guarantees

### Player Goes Idle

**Scenario:** Player active for 1 hour, then closes browser.

**Warmup behavior:**
- Reaches warmup_level = 1.0 (after 30 min active)
- Remains at 1.0 for ~30-90 min while decaying slowly (3x slower because recent_activity_seconds is high)
- Drops to 0.5 after ~1-2 hours idle
- Reaches 0.0 after ~2-4 hours idle (depending on decay rate)

### Player Returns After Long Absence

**Scenario:** Player inactive for 24 hours, logs back in.

**Warmup behavior:**
- Starts at warmup_level = 0.0 (warmed up completely)
- Immediately starts accumulating warmup again at first `/activity` ping
- Reaches 1.0 within 30 minutes of resumed play
- Faster "return to form" because no prior engagement to decay from

### Concurrent Activity & Cap Conflict

**Scenario:** Player claims activity faucet (caps to remaining daily limit), then tries to claim daily login faucet and emit baseline UBI.

**Behavior:**
- Activity claim: respects remaining daily cap
- Daily login claim (if eligible): respects remaining daily cap
- UBI emission: UBI is NOT subject to daily cap (always emits, adding to coins)
- Result: Player can exceed daily cap via UBI; daily cap only applies to **faucet claims** (activity + daily login + other active methods)

---

## Configuration & Admin Control

### Settable Parameters

Via `/admin/settings`:

- `activity_window_seconds` (default 120) — how long player is considered active after last ping
- `activity_cooldown_seconds` (default 360) — faucet claim cooldown
- `activity_reward` (computed by calibration, not directly settable) — coin amount per claim

### Feature Flags

- `ENABLE_FAUCETS` — if false, `/claim-activity` returns `FEATURE_DISABLED`

### Calibration (Auto-Computed by Population)

Activity reward & cooldown scale based on active player count:

```
activity_cooldown = clamp(6 × 60 sec, 300, 720)     // 5-12 minute range
activity_reward = clamp(0.04 × daily_cap, 1, 6)    // ~4% of daily budget
```

---

## Design Goals

1. **Engagement Incentive:** Active players earn significantly more (frequent claims + UBI multiplier)
2. **Passive Penalty:** Idle players lose both activity claim eligibility and warmup multiplier
3. **Return-Friendly:** Brief absences (< 2-4 hours) don't completely reset progress; "comeback" is faster than initial grind
4. **Tuneable:** Activity window, cooldown, and reward are all adjustable without code changes
5. **Fairness:** Activity is time-based, not action-based; server measures engagement via presence, not button clicks
6. **Anti-Farm:** 6-minute cooldown prevents spamming a single claim; daily cap prevents abuse of multiple alts

---

## Technical Notes

**Database:**
- Schema: `players` table columns `activity_warmup_level`, `activity_warmup_updated_at`, `recent_activity_seconds`, `last_active_at`
- Index: Primary key on `player_id`; activity columns have no special indexes (assume O(1) lookups)

**Timing:**
- All timestamps in UTC
- Server-authoritative; client `last_active_at` not trusted
- Ticks are 1-2 seconds (internal server loop); frontend pings every 20 seconds

**Interaction with Other Systems:**
- UBI: Multiplier calculated from warmup_level (see [coin-faucets.md](coin-faucets.md))
- Emission: Activity claim costs coins from emission pool (daily cap not enforced on UBI)
- Boosts: Activity boost provides +1 coin per claim if active
- Telemetry: Activity claims emit `faucet_claim` event with reward amount

---

## See Also

- [Coin Faucets](coin-faucets.md) — full earning mechanics including activity, daily, and UBI
- [Coin Emission](coin-emission.md) — emission pool and UBI distribution
- [Anti-Abuse](anti-abuse.md) — throttle/cooldown interaction with activity tracking
