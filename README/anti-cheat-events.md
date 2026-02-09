# Anti-Cheat Events Registry

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

# Abuse Detection & Automated Enforcement

The anti-cheat system automatically detects suspicious patterns and applies graduated penalties to protect economy integrity. All enforcement is **automatic** and **immediate** with no manual review gates.

---

## Detection Event Types

| Event Type | Severity | Window | Threshold | Score Delta |
|-----------|----------|--------|-----------|-------------|
| **purchase_burst** | 1 | 10 min | ≥6 star purchases | `(count - 5) × 1.2` |
| **purchase_regular_interval** | 2 | 60 min | ≥6 purchases, mean ≤180s, stddev ≤2.0 | `2.5` |
| **activity_regular_interval** | 1 | 60 min | ≥6 claims, mean ≤240s, stddev ≤3.0 | `2.0` |
| **tick_reaction_burst** | 1 | 30 min | ≥3 purchases within 2s of minute boundary | `count × 0.8` |
| **ip_cluster_activity** | 2 | 10 min | ≥3 distinct players from same IP purchasing | `activePlayers × 0.7` |

---

## Event Definitions

### purchase_burst

**Detects:** Rapid star purchasing (potential spam or panic buying)

**Trigger Conditions:**
- 6+ star purchases within 10 minutes
- Raw count-based detection (no statistical analysis)

**Score Delta:** `(count - 5) × 1.2`
- 6 purchases → 1.2 points
- 10 purchases → 6.0 points

**Rationale:** Legitimate players rarely buy 6+ stars in 10 minutes; signals automated buying or coordinated account activity.

---

### purchase_regular_interval

**Detects:** Bot-like automated purchasing with suspiciously consistent timing

**Trigger Conditions:**
- 6+ star purchases within 60 minutes
- Mean interval between purchases ≤ 180 seconds
- Standard deviation of intervals ≤ 2.0 seconds

**Score Delta:** `2.5` (fixed)

**Rationale:** Human players exhibit irregular timing; bots purchase on fixed schedules. Low stddev (≤2s) is statistically unlikely for humans.

---

### activity_regular_interval

**Detects:** Automated activity faucet claiming

**Trigger Conditions:**
- 6+ activity claims within 60 minutes
- Mean interval ≤ 240 seconds
- Standard deviation ≤ 3.0 seconds

**Score Delta:** `2.0` (fixed)

**Rationale:** Activity faucet has 6-minute cooldown; human players don't claim every 6 minutes precisely. Low stddev indicates automation.

---

### tick_reaction_burst

**Detects:** Purchases aligned with server tick boundaries (sub-second timing patterns)

**Trigger Conditions:**
- 3+ star purchases within 30 minutes
- Purchase timestamps within 2 seconds of minute boundaries (0-2s or 58-60s)

**Score Delta:** `count × 0.8`
- 3 aligned purchases → 2.4 points
- 5 aligned purchases → 4.0 points

**Rationale:** Humans don't consistently time purchases to minute boundaries. Bots often trigger on server ticks or scheduled events.

---

### ip_cluster_activity

**Detects:** Multiple accounts from same IP address (account farming, coordinated abuse)

**Trigger Conditions:**
- 3+ distinct players from same IP address
- All made star purchases within 10 minutes
- IP active (any gameplay action) within 24 hours

**Score Delta:** `activePlayers × 0.7`
- 3 players → 2.1 points
- 5 players → 3.5 points

**Penalty Application:** **All players** from that IP receive the penalty (shared enforcement)

**Rationale:** Legitimate households rarely have 3+ players buying stars simultaneously. High correlation with multi-accounting.

---

## Severity System

### Severity Levels

Abuse score determines severity tier:

| Severity | Score Range | Description | Decay Rate (per hour) |
|----------|------------|-------------|----------------------|
| **0** | 0 - 9.99 | No enforcement | 1.0 (fast decay) |
| **1** | 10 - 24.99 | Light throttling | 0.6 (fast decay) |
| **2** | 25 - 44.99 | Moderate penalties | 0.3 (slow decay, 72h persistence) |
| **3** | 45+ | Heavy penalties | 0.15 (very slow decay, 7d persistence) |

### Persistence

**Severity 2+ can "lock" for extended periods:**
- **Severity 2:** 72 hours persistence if multiple signals within 6 hours
- **Severity 3:** 7 days persistence if multiple signals within 6 hours

Once locked, severity cannot drop below the locked level until persistence expires, even if score decays.

### Account-Level Reputation

Severe abuse (severity 2+) builds **account-level reputation** that carries across seasons:

```
accountScore = sum of all signals across all seasons
combinedScore = max(seasonScore, accountScore × 0.75)
combinedSeverity = max(seasonSeverity, accountSeverity, derivedFromCombinedScore)
```

**Effect:** Repeat offenders face progressively harsher penalties even in new seasons.

---

## Automatic Enforcement

### Severity 1 (Light Throttling)

Applied when score reaches 10:

| Penalty Type | Effect |
|--------------|--------|
| **Star Price** | +5% (multiply by 1.05) |
| **Max Bulk Quantity** | Limited to 4 stars per transaction |
| **Earning Rate** | -10% (multiply by 0.9) |
| **Cooldown Jitter** | +10% random delay added to faucet cooldowns |

---

### Severity 2 (Moderate Penalties)

Applied when score reaches 25:

| Penalty Type | Effect |
|--------------|--------|
| **Star Price** | +15% (multiply by 1.15) |
| **Max Bulk Quantity** | Limited to 3 stars per transaction |
| **Earning Rate** | -25% (multiply by 0.75) |
| **Cooldown Jitter** | +25% random delay added to faucet cooldowns |
| **Admin Notification** | High priority notification sent to moderators/admins |

---

### Severity 3 (Heavy Penalties)

Applied when score reaches 45:

| Penalty Type | Effect |
|--------------|--------|
| **Star Price** | +30% (multiply by 1.3) |
| **Max Bulk Quantity** | Limited to 2 stars per transaction |
| **Earning Rate** | -40% (multiply by 0.6) |
| **Cooldown Jitter** | +50% random delay (capped at 5 minutes) |
| **Admin Notification** | **Critical priority** notification sent immediately |

---

## Enforcement Guarantees

### Minimum Values
All multipliers apply gracefully with floors:
- Star price ≥ 1 microcoin
- Earning rewards ≥ 1 microcoin
- Max bulk quantity ≥ 1 star
- Cooldown jitter ≤ 5 minutes

### Transparency
- Abuse score NOT visible to players
- Enforcement effects (higher prices, longer cooldowns) observable indirectly
- No explicit "you are throttled" messaging

---

## Admin Visibility

### `/admin/abuse-events` Endpoint

Returns last 200 abuse events, ordered by creation time (descending).

**Response Format:**
```json
{
  "ok": true,
  "events": [
    {
      "id": 1234,
      "accountId": "uuid",
      "playerId": "uuid",
      "seasonId": "uuid",
      "eventType": "purchase_burst",
      "severity": 1,
      "scoreDelta": 6.0,
      "details": {
        "count": 10,
        "windowMinutes": 10
      },
      "createdAt": "2026-02-09T12:34:56Z"
    }
  ]
}
```

**Details Field (varies by event type):**

```json
// purchase_burst
{"count": 8, "windowMinutes": 10}

// purchase_regular_interval
{"intervalMeanSeconds": 120.5, "intervalStdSeconds": 1.2, "count": 8}

// activity_regular_interval
{"intervalMeanSeconds": 360.0, "intervalStdSeconds": 2.5, "count": 7}

// tick_reaction_burst
{"count": 4, "windowMinutes": 30}

// ip_cluster_activity
{"ip": "192.168.1.1", "activePlayers": 4, "windowMinutes": 10}
```

---

### `/admin/overview` Metrics

Real-time abuse statistics:

```json
{
  "activeThrottles": 12,           // Players with severity ≥ 1
  "activeAbuseFlags": 3,           // Players with severity ≥ 2
  "abuseEventsLastHour": 45,       // Total events in last hour
  "abuseSevereLastHour": 2         // Severity 3 events in last hour
}
```

---

### `/admin/anti-cheat` Status

System configuration and event counts:

```json
{
  "ok": true,
  "toggles": [
    {"key": "ip_enforcement", "enabled": true},
    {"key": "clustering_detection", "enabled": true},
    {"key": "automatic_throttling", "enabled": true}
  ],
  "sensitivity": {
    "clustering": "medium",
    "throttle": "high"
  },
  "eventCounts24h": {
    "ip_cluster_activity": 15,
    "purchase_burst": 32,
    "total": 78
  }
}
```

---

## Configuration

### Environment Variables

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `ABUSE_INCLUDE_BOTS` | boolean | `false` | Apply abuse detection to bots (default exempts them) |
| `ENABLE_IP_THROTTLING` | boolean | `true` | Enable IP-based enforcement and clustering detection |

**Bot Exemption:**
When `ABUSE_INCLUDE_BOTS=false` (default), bots always receive severity 0 (no enforcement), regardless of behavior. This allows testing without triggering penalties.

---

### Runtime Adjustability

**Currently NOT adjustable at runtime:**
- Detection thresholds (hardcoded per event type)
- Severity score boundaries (10, 25, 45)
- Decay rates (varies by severity level)
- Time windows (varies by detector)
- Enforcement multipliers (price, earning, cooldown, bulk)

**Would require code changes:**
All parameters are compile-time constants. No database-backed configuration system for abuse tuning.

---

## Database Schema

### `player_abuse_state`

Per-player, per-season abuse scoring:

```sql
CREATE TABLE IF NOT EXISTS player_abuse_state (
    player_id TEXT NOT NULL,
    season_id TEXT NOT NULL,
    score DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    severity INT NOT NULL DEFAULT 0,
    last_signal_at TIMESTAMPTZ,
    persistent_until TIMESTAMPTZ,
    last_decay_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (player_id, season_id)
);
```

---

### `account_abuse_reputation`

Cross-season account reputation:

```sql
CREATE TABLE IF NOT EXISTS account_abuse_reputation (
    account_id TEXT PRIMARY KEY,
    score DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    severity INT NOT NULL DEFAULT 0,
    last_signal_at TIMESTAMPTZ,
    persistent_until TIMESTAMPTZ,
    last_decay_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

### `abuse_events`

Append-only audit log:

```sql
CREATE TABLE IF NOT EXISTS abuse_events (
    id BIGSERIAL PRIMARY KEY,
    account_id TEXT NOT NULL,
    player_id TEXT NOT NULL,
    season_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    severity INT NOT NULL,
    score_delta DOUBLE PRECISION NOT NULL,
    details JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_abuse_events_created 
    ON abuse_events (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_abuse_events_player 
    ON abuse_events (player_id, created_at DESC);
```

---

## Detection Philosophy

### Design Goals

1. **Fully Automated** — No manual review gates; system applies enforcement immediately
2. **Graduated Response** — Light throttling first; heavy penalties only after repeated signals
3. **Graceful Degradation** — Never lock players out; always allow minimum earning/buying
4. **Statistical Rigor** — Uses mean/stddev analysis for bot detection, not just raw counts
5. **IP Accountability** — Penalizes all accounts from abusive IPs, not just primary offender
6. **Account Persistence** — Repeat offenders face escalating penalties across seasons
7. **Admin Transparency** — All events logged; admins notified of severe cases

### False Positive Mitigation

- **High thresholds** — Most events require 6+ actions to trigger
- **Fast decay** — Severity 1 penalties fade within hours if behavior normalizes
- **Graceful minimums** — Enforcement never blocks earning/buying completely
- **Bot exemption** — Testing accounts can be marked as bots to avoid triggering detection

---

## Notification Integration

**Admin notifications are triggered automatically:**

| Severity | Priority | When Sent |
|----------|----------|-----------|
| 0 | None | No notification |
| 1 | None | No notification |
| 2 | High | On first detection; after 6h if multiple events |
| 3 | Critical | Immediately on every detection |

**Notification Content:**
```json
{
  "category": "abuse",
  "title": "Abuse Detected: Severity 3",
  "message": "Player <username> triggered purchase_burst (score delta: 6.0)",
  "priority": "critical",
  "playerId": "uuid",
  "eventType": "purchase_burst",
  "severity": 3
}
```

---

## Technical Notes

**Timing:**
- Signals evaluated in batches every ~30-60 seconds
- Enforcement applied immediately after signal detection
- Score decay calculated lazily (on next read/write)

**Performance:**
- Event detection scans last N actions per player (varies by window)
- No global scans; per-player analysis only
- IP clustering requires join across players table (more expensive)

**Edge Cases:**
- Bot flag overrides all enforcement (severity forced to 0)
- Severity 0 players with positive scores decay to 0 within hours
- Account reputation applies at 75% weight to avoid over-penalization
- Persistence locks prevent rapid severity fluctuation

---

## See Also

- [Anti-Abuse](anti-abuse.md) — High-level abuse prevention philosophy
- [HTTP API Reference](http-api-reference.md) — `/admin/abuse-events` endpoint spec
- [Admin Tools](admin-tools.md) — Admin notification system integration
