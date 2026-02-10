# Admin Tools

## Scope
- **Type:** System contract
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

The system must include a minimal internal admin and observability interface.

Admin access is restricted to authorized accounts only.

---

## Admin Role Philosophy — Sentinels, Not Gods

**Admins are sentinels, not gods.**

The economy must **self-regulate**. Admins provide oversight and emergency safeguards, not active management.

### What Admins MAY Do

Admins may:

- **Emergency pause 1 or all seasons** (temporary freeze)
- **Ban extreme abuse cases** (only after anti-cheat recommendation)
- **Monitor telemetry and economy health** (read-only observability)
- **Advance seasons manually** (recovery only, not normal flow)

### What Admins MUST NOT Do

Admins must NOT:

- Micromanage the economy
- Manually adjust player balances
- Override anti-cheat without justification
- Edit past season data
- Interfere with normal economic flow

**The economy is designed to self-regulate.** Admin intervention must be rare, deliberate, and auditable.

---

## Telemetry Rules — Player-Facing Subset Only

Only telemetry that influences **player decisions** may be shown to players.

### Allowed Player-Facing Telemetry

- **Total Coins in Circulation**: Sum of all player wallet balances (NOT the unreleased Coin Pool)
- **Market pressure trends**: Graphs showing pressure changes over time
- **Star price graphs**: Historical and current pricing
- **Time remaining**: Season countdown and milestones
- **Leaderboard changes**: Rank movements and competitive dynamics

Telemetry may include:

- Graphs
- Timers
- Trend indicators
- Aggregated statistics

### Forbidden Player-Facing Telemetry

Telemetry must NOT expose:

- **Exact internal formulas** (pricing algorithms, emission calculations)
- **Anti-cheat thresholds** (thresholds, multipliers, detection signals)
- **Exploitable signals** (precise emission pool size, exact faucet pacing)
- **Admin control granularity** (exact admin actions beyond high-level pauses)

### Admin-Only Telemetry

Admins have access to:

- Full economy state (emission pool, exact rates, throttle status)
- Anti-cheat signals and abuse event logs
- Player-level detail (balances, earning history, throttle state)
- Trade logs and market activity (when trading is enabled)

Admin telemetry is **never exposed to players**.

---

## Required admin capabilities are split by phase. Alpha is read‑only.

Alpha (read‑only, current build):

Implemented (Alpha):

Season monitoring (single season):

- Active season status (active vs ended).
- Season time remaining (via season snapshot).

Alpha override (admin‑only, recovery):

- Manual season advance when no active season exists or the current season has ended (POST /admin/seasons/advance).
- No parameters are accepted; this is an override, not the normal flow.

Alpha rule:

- “Ending” is internal only; admin UI shows only **Active** or **Ended**.
- When ended, admin economy indicators are read-only and present frozen/final markers (no live emission/inflation rates).
- Admin control strips (pause/freeze/emission controls) are hidden in Alpha.

Economy monitoring (per season):

- Current base star price.
- Current effective star price.
- Current market pressure.
- Daily emission target.
- Daily cap early/late.

Telemetry (current build):

- Event counts per hour by type (from player telemetry stream).
- Notification emit events are logged for observability.
- Emitted event types include: emission_tick, market_pressure_tick, faucet_claim, star_purchase_attempt, star_purchase_success, notification_emitted.
- Admin UI remains read‑only and currently exposes counts, not full raw payloads.

Notification visibility (read‑only):

- Admins can see notification emission events (counts and aggregates only).
- Admins cannot send custom notifications in Alpha.
- Admins cannot target individual players in Alpha.

Player inspection (read‑only):

- Player search by username/account/player ID.
- Trust status and flag count.

Abuse monitoring (read‑only):

- Recent abuse events list.
- Anti‑cheat toggle status (visibility only; not configurable).

Auditability (read‑only):

- Star purchase log.
- Admin audit log.

Bug reports (read‑only):

- View bug report list and details (title, description, category, player_id if present, season_id, timestamp, client version if available).
- No editing, deletion, or player responses.
- No attachments in Alpha; admin view is observational only.

Not yet in Alpha (post‑alpha or pending implementation):

- Global coin budget remaining for the day.
- Coin emission rate and throttling state details.
- Coins emitted per hour, coins earned per hour, and average star price over time (beyond event counts).
- Per‑player coin earning history view.
- Per‑player coin and star balance detail view (beyond search results).
- Throttle status per player.
- IP clustering detail views beyond aggregate signals.

Post‑Alpha (planned):

Trading visibility:

Current trade premium and burn rate.

Current trade eligibility tightness.

Stars transferred via trades per hour.

Coins burned via trades per hour.

View trade eligibility status and recent trades.

Safety tools (admin‑only, auditable):

Temporarily pause star purchases per season if needed.

Temporarily reduce coin emission rates.

Freeze a season in emergency cases.

Temporarily disable trading per season if needed.

All admin actions are logged and auditable.

---

## Admin Action Audit Trail

All administrative actions are logged to an **append-only, immutable audit log** for accountability and recovery.

### Action Type Taxonomy

| Action Type | Scope Type | When Triggered | Details Captured |
|------------|-----------|----------------|-----------------|
| **auto_admin_bootstrap** | `account` | Server startup creates first admin | `{"username": "alpha-admin", "autoCreated": true}` |
| **admin_bootstrap_claim** | `account` | Owner claims bootstrap with code | `{"username": "...", "claimedAt": "..."}` |
| **role_update** | `account` | Admin changes user role | `{"username": "...", "oldRole": "player", "newRole": "admin"}` |
| **profile_freeze** | `account` | Admin freezes account | `{"username": "...", "reason": "..."}` |
| **profile_unfreeze** | `account` | Admin unfreezes account | `{"username": "...", "reason": "..."}` |
| **profile_delete** | `account` | Admin deletes account (reserved) | `{"username": "...", "reason": "..."}` |
| **season_control_set** | `season_control` | Admin sets season control (emergency) | `{"controlName": "pause_purchases", "value": "true", "expiresAt": "..."}` |
| **season_advance** | `season` | Admin manually advances season | `{"oldSeasonId": "...", "newSeasonId": "...", "reason": "..."}` |
| **season_recovery** | `season` | Admin creates recovery season (Alpha) | `{"newSeasonId": "...", "reason": "...", "confirm": "..."}` |
| **notification_create** | `notification` | Admin sends broadcast notification | `{"category": "system", "priority": "high", "targetRole": "all", "message": "..."}` |
| **bot_toggle** | `player` | Admin enables/disables bot | `{"playerId": "...", "isBot": true}` |
| **bot_create** | `player` | Admin creates test bot | `{"username": "bot_1", "profile": "threshold_buyer"}` |
| **bot_delete** | `player` | Admin deletes bot account | `{"playerId": "...", "username": "bot_1"}` |

### Scope Type Taxonomy

| Scope Type | Meaning | Scope ID Format |
|-----------|---------|----------------|
| **account** | Action affects user account | Account UUID or username |
| **season** | Action affects season lifecycle | Season UUID |
| **season_control** | Emergency season control switch | Season UUID |
| **notification** | Notification creation/broadcast | Notification ID or "broadcast" |
| **player** | Action affects player state | Player UUID |

### Database Schema

```sql
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

CREATE INDEX IF NOT EXISTS idx_admin_audit_log_created 
    ON admin_audit_log (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_admin 
    ON admin_audit_log (admin_account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_action 
    ON admin_audit_log (action_type, created_at DESC);
```

### Querying the Audit Log

**Endpoint:** `GET /admin/audit-log`

**Query Parameters:**
- `limit` (default 50, max 200) — Number of entries to return
- `offset` (default 0) — Pagination offset
- `search` (optional) — Full-text search across admin username, action type, scope ID, reason

**Response:**
```json
{
  "ok": true,
  "items": [
    {
      "id": 1234,
      "adminAccountId": "admin-uuid",
      "adminUsername": "alpha-admin",
      "actionType": "role_update",
      "scopeType": "account",
      "scopeId": "player-uuid",
      "reason": "Promoting moderator",
      "details": { /* JSON */ },
      "createdAt": "2026-02-09T12:34:56Z"
    }
  ],
  "total": 1457,
  "limit": 50,
  "offset": 0
}
```

### Retention & Immutability

**Append-Only Guarantee:**
- No UPDATE queries (actions cannot be modified)
- No DELETE queries (actions cannot be removed)
- Monotonic IDs (sequential ID proves chronological order)

**Retention Policy:**
- **Alpha:** Indefinite (no purge, short-lived seasons)
- **Beta:** Indefinite (compliance audit trail)
- **Release:** 2+ years (legal/compliance requirements)

### Security & Access Control

| Role | Access |
|------|--------|
| **Admin** | Full access (all entries) |
| **Moderator** | No access (reserved for post-alpha) |
| **Player** | No access |

**Entry Creation:**
- Only admins can create entries (via admin actions)
- System can create bootstrap entries (`auto_admin_bootstrap`)
- No API for arbitrary entry creation (prevents log forgery)

For complete action type details and edge cases, see the full audit log specification in the codebase.