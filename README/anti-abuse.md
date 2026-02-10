# Anti Abuse

## Scope
- **Type:** System contract
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

The system must actively prevent and mitigate abuse and coordinated manipulation.

---

## Anti-Cheat Philosophy — Soft Enforcement

Anti-cheat is **gradual, invisible, and corrective**, not punitive.

### What Anti-Cheat NEVER Does

Anti-cheat **NEVER**:

- Bans automatically
- Suspends accounts automatically
- Zeroes a wallet
- Hard-blocks players
- Exposes enforcement actions publicly

### What Anti-Cheat DOES

Anti-cheat acts by **adjusting coin faucet flow**:

- Gradually reduces earning rates for suspicious behavior
- Increases star prices for suspicious accounts
- Adds cooldowns and jitter to sensitive actions
- Throttles activity without blocking it

### Enforcement Scaling

**Effects should be**:

- **Gradual**: Small adjustments at first, increasing over time
- **Mostly invisible**: Players feel resistance, not punishment
- **Severe only in extreme abuse**: Heavy throttles reserved for confirmed bad actors

**Behavior → consequence must scale smoothly**:

- Minor suspicious activity → minor throttles
- Moderate abuse patterns → noticeable resistance
- Extreme abuse → heavy economic dampening

### Admin Involvement

Admins may ban ONLY:

- **After anti-cheat recommendation** (system flags extreme cases)
- **In extreme cases** (confirmed, egregious abuse)

Admins **must not**:

- Micromanage economy
- Manually adjust individual player balances
- Override anti-cheat without justification

The goal: **Make abuse economically ineffective, not publicly punishing.**

---

## Access controls:

Only one active player per IP address per season is the default baseline.

Additional accounts from the same IP are not hard-blocked; they are throttled through economic dampening, cooldowns, and trust-based enforcement.

No whitelist or allowlist is used in alpha.

Account protections:

Account creation is rate-limited.

CAPTCHA and verification are post‑alpha.

Account age is a soft signal only. Hard gating first-session play is forbidden.
New accounts may face softer throttles (cooldown multipliers, reward dampening, bulk limits), but never hard blocks.

Throttles:

Per-player star purchase rate limits exist.

Per-IP star purchase limits exist, especially early in the season.

Coin earning and star buying may be dynamically throttled for suspicious activity.
Brokered trading eligibility may be tightened or suspended for suspicious activity.

Detection:

The system monitors for clustering patterns such as many new accounts from related IP ranges acting similarly.

Suspicious activity generates abuse events.

Abuse events may trigger automatic temporary throttles.

Trade-specific detection:

Repeated reciprocal trades between the same accounts

Trading patterns that concentrate Stars across related IP ranges

Unusual trade volume spikes relative to participation

(Trade-specific detection is post‑alpha while trading is disabled.)

Enforcement:

Throttles are gradual and reversible.

The goal is to make abuse economically ineffective, not to punish publicly.

All abuse decisions and throttles are enforced server-side.

---

## Alpha Audit — Post‑Whitelist Removal

Implemented (confirmed in code):

- Whitelisting removed; alpha relies on throttles only.
- Auth rate limits for signup/login (IP‑based windows).
- Account age used as a soft throttling signal (cooldown/reward/bulk multipliers), never a hard gate.
- IP association tracking and dampening (delay + reward/price multipliers when multiple accounts share an IP).
- Abuse scoring with throttles (earn multiplier, price multiplier, bulk max, cooldown jitter) driven by detected signals.
- Abuse signals include purchase bursts, regular purchase cadence, activity cadence, tick‑reaction patterns, and IP clustering.
- Abuse events are logged and emit moderator/admin notifications.
- Bot star‑purchase rate limit enforced via minimum interval.

Gaps / Alpha‑known limitations:

- No explicit hard per‑IP star purchase limit beyond IP dampening and abuse scoring.
- No explicit per‑player star purchase rate limit beyond abuse scoring and bot interval limits.
- CAPTCHA and verification remain post‑alpha.
- Trade‑specific abuse detection is inactive while trading is disabled.

---

## Anti-Cheat Events Registry

The anti-cheat system automatically detects suspicious patterns and applies graduated penalties to protect economy integrity. All enforcement is **automatic** and **immediate** with no manual review gates.

### Detection Event Types

| Event Type | Severity | Window | Threshold | Score Delta |
|-----------|----------|--------|-----------|-------------|
| **purchase_burst** | 1 | 10 min | ≥6 star purchases | `(count - 5) × 1.2` |
| **purchase_regular_interval** | 2 | 60 min | ≥6 purchases, mean ≤180s, stddev ≤2.0 | `2.5` |
| **activity_regular_interval** | 1 | 60 min | ≥6 claims, mean ≤240s, stddev ≤3.0 | `2.0` |
| **tick_reaction_burst** | 1 | 30 min | ≥3 purchases within 2s of minute boundary | `count × 0.8` |
| **ip_cluster_activity** | 2 | 10 min | ≥3 distinct players from same IP purchasing | `activePlayers × 0.7` |

### Severity System

| Severity | Score Range | Decay Rate (per hour) | Price | Max Bulk | Earn | Cooldown Jitter |
|----------|------------|----------------------|-------|----------|------|-----------------|
| **0** | 0-9.99 | 1.0 (fast) | 1.0x | — | 1.0x | 0% |
| **1** | 10-24.99 | 0.6 (fast) | 1.05x | 4 stars | 0.9x | +10% |
| **2** | 25-44.99 | 0.3 (slow, 72h persistence) | 1.15x | 3 stars | 0.75x | +25% |
| **3** | 45+ | 0.15 (very slow, 7d persistence) | 1.3x | 2 stars | 0.6x | +50% |

### Admin Visibility

**Endpoints:**
- `GET /admin/abuse-events` — Last 200 abuse events
- `GET /admin/overview` — Real-time abuse statistics
- `GET /admin/anti-cheat` — System configuration and event counts

**Database:**
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
```

For complete event definitions and enforcement mechanics, see the implementation in abuse.go.