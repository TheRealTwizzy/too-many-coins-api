# Star Purchases

## Scope
- **Type:** System contract
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

# Star Purchases

**Currency Model:** All star prices and purchases use integer microcoins (1 coin = 1000 microcoins). See [coin-emission.md](coin-emission.md) for details.

## Star Purchasing System

Stars may only be obtained by purchasing them from the system using coins.

Stars are minted only by the system. Brokered trading may transfer existing Stars between players but never creates new Stars and never bypasses scarcity.

## Star Characteristics

**Stars are NOT tradable in any context** (player-to-player trading, brokered trading, or any other mechanism).

**Stars are NOT spendable** (except via Star Sacrifice for TSAs in Beta+, which permanently destroys them).

**Stars are seasonal competitive units** that:

- Determine leaderboard rank directly
- Reset at season end (do not carry over as currency)
- Convert into a **permanent profile statistic** after season end
- Influence long-term profile rank and identity

**Star value scales with season population**. Stars earned in larger, more competitive seasons carry more weight as a permanent statistic.

**Stars are the permanent score of seasonal performance.**

Star purchases follow these rules:

Players may buy stars one at a time or in bulk.
Bulk purchases are allowed but heavily penalized through scaling.

The total coin cost of a star purchase is determined by:

A base star price that increases over the season

A quantity multiplier that scales non-linearly with purchase size

A market pressure factor based on recent purchase activity

A late-season spike applied during the final week

An affordability guardrail that keeps prices aligned to average coins per player and emission so stars remain purchasable through the season

Quantity scaling:

The quantity multiplier must grow faster than linear.

Large bulk purchases rapidly become inefficient.

Extremely large bulk purchases may include additional hard multipliers.

Bulk purchase interfaces must:

Show the full calculated cost before confirmation

Show how quantity affects price

Warn players when purchases are highly inefficient

Require explicit confirmation

Alpha verification:

- Server recomputes the bulk quote at purchase time; price and balance are re‑checked before commit.
- Bulk warnings are derived from the max bulk multiplier (medium/high/severe thresholds).

Star purchases:

Must be atomic

Must re-check price and balance at confirmation time

Must fail safely if conditions change

Star supply is system-managed and cannot be exhausted.
Scarcity is enforced through pricing, not limited stock.

All star purchases are validated server-side and recorded in an append-only log.

## Star Price Persistence

The current star price is persisted in the season economy table and updated each emission tick (every 60 seconds).

This ensures:
- Star price remains **identically consistent** across all players at any given moment
- Star price remains consistent across server restarts
- Price continuity is maintained during unexpected downtime
- The season-authoritative star price is recoverable from database state

## Price Tick Locking (Authoritative Match)

Every season snapshot includes a `price_tick` identifier alongside `current_star_price`.

Client purchase requests MUST include the `price_tick` observed by the client.

Server validation:
- If `request.price_tick == current_price_tick`: purchase proceeds using the snapshotted price
- If `request.price_tick != current_price_tick`: request fails with `PRICE_CHANGED`

This guarantees:
- UI price and server-enforced price match exactly (or fail explicitly)
- No silent recomputation or price drift between display and enforcement
- `NOT_ENOUGH_COINS` is reserved strictly for true insufficiency at the locked price

### Star Price Computation (Season-Level Authority)

The star price is computed from **season-level inputs only**. All players see the **identical price** at any given moment in the season.

Season-level inputs:
- Time progression within the season
- Total coins in circulation (across all players)
- Stars purchased this season
- Market pressure (aggregate purchase activity)
- Late-season spike (time-based multiplier)
- Affordability guardrail (derived from total coins / expected player base)

**Active player metrics are NOT a direct input** to star price computation. Player activity influences pricing only indirectly through market pressure (purchase activity).

The computation is performed **once per server tick** and stored in the database. The same value is broadcast to all players via SSE and API endpoints.

If the persisted price is NULL on startup (new season or legacy data), it will be populated by the next emission tick.

### Price Displayed vs. Price Paid

The displayed to all players is the season-authoritative price.

Purchase flow:
- Player sees the authoritative star price
- Anti-abuse logic may affect:
  - Purchase allowance (buying limits)
  - Effective price paid (through cooldowns or other mechanisms)
  - But NOT the displayed price
- Star purchase log records both:
  - season_price_snapshot (authoritative)
  - effective_price_paid (actual cost including anti-abuse adjustments)

---

## Variant Stars (Alpha Feature)

Variant stars are specialty stars purchased separately from regular stars. They **do not count toward leaderboard rank** but are tracked per player for cosmetic/collection purposes.

### Available Variants

| Variant | Cost Multiplier | Display Price Formula |
|---------|----------------|----------------------|
| **Ember** | 2.0x | `base_star_price × 2.0` |
| **Void** | 4.0x | `base_star_price × 4.0` |

### Mechanics

**Purchase Endpoint:** `POST /buy-variant-star`

**Request:**
```json
{
  "variantType": "ember",
  "quantity": 1
}
```

**Response:**
```json
{
  "ok": true,
  "variant": "ember",
  "purchased": 1,
  "coinsBurned": 900,
  "playerCoins": 4100,
  "variantCount": 5
}
```

**Pricing:**
- Base price = season-authoritative star price (same as regular stars)
- Display price = base price × variant multiplier
- Effective price = display price × abuse enforcement multipliers (if applicable)

**Leaderboard Impact:** **NONE** — variant stars stored separately, not counted in `player.stars`

**Storage:** `player_star_variants` table (player_id, variant, count)

**Purchase Log:** Recorded in `star_purchase_log` with `purchase_type = "variant"`

### Gating & Requirements

- Same as regular stars (must be authenticated, season active, sinks enabled)
- No additional unlock requirements
- Subject to bot rate limits if player is bot

### Design Intent

- **Cosmetic expression** — Players collect variants for prestige without competitive advantage
- **Coin sink diversification** — Provides alternative spending target beyond leaderboard position
- **Future collectibles** — May unlock badges, titles, or cosmetic rewards post-alpha

---

## Activity Boosts (Alpha Feature)

Activity boosts are temporary purchased buffs that enhance coin earning from activity faucets.

### Available Boosts

| Boost Type | Cost | Duration | Effect |
|-----------|------|----------|--------|
| **Activity** | 25 coins | 30 minutes | +1 coin per activity faucet claim |

### Mechanics

**Purchase Endpoint:** `POST /buy-boost`

**Request:**
```json
{
  "boostType": "activity",
  "quantity": 1
}
```

**Response:**
```json
{
  "ok": true,
  "boostType": "activity",
  "durationSeconds": 1800,
  "expiresAt": "2026-02-09T13:04:56Z",
  "coinsBurned": 25,
  "playerCoins": 5221
}
```

**Cost Calculation:**
- Base cost: 25 coins
- Subject to abuse enforcement price multipliers (if player flagged)
- Subject to IP dampening multipliers (if player throttled)

**Duration:**
- 30 minutes from purchase time
- Multiple purchases **extend** expiration (does not stack quantity)
- Example: Buy boost at 12:00 → expires at 12:30. Buy again at 12:15 → expires at 12:45.

**Effect Application:**
- When `HasActiveBoost(db, playerID, "activity")` returns true during activity claim
- Reward calculation: `base_reward + 1` coin
- Applied **before** abuse enforcement earning multipliers

**Storage:** `player_boosts` table (player_id, boost_type, expires_at)

**Schema (⚠️ Missing from schema.sql):**
```sql
CREATE TABLE IF NOT EXISTS player_boosts (
    player_id TEXT NOT NULL,
    boost_type TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (player_id, boost_type)
);
```

### Gating & Requirements

- Season must be active (`SEASON_ENDED`)
- Sinks must be enabled (`ENABLE_SINKS` feature flag)
- Player must have sufficient coins (`NOT_ENOUGH_COINS`)
- Boost type must be valid (`INVALID_BOOST`)

### Design Intent

- **Time investment strategy** — Players who commit to extended play sessions earn more
- **Coin sink timing** — Encourages spending coins on earning multiplier rather than hoarding for stars
- **Optional optimization** — Not required for competitive play; casual players can skip entirely

### Economics

**Break-even analysis:**
```
Cost: 25 coins
Duration: 30 minutes
Boost: +1 coin per claim
Activity cooldown: 6 minutes

Theoretical max claims in 30 min: 5 claims
Max gain: 5 coins
Net loss: 20 coins
```

**Never profitable directly** — Boost costs more than it earns. Strategic value:
- Sustaining warmup multiplier during intensive play
- Ensuring cap claims don't leave coins on table
- Psychological engagement (active optimization vs passive earning)

---

## Coin Burning (Voluntary Sink)

Players may voluntarily burn (destroy) coins for no immediate gameplay benefit.

### Mechanics

**Endpoint:** `POST /burn-coins`

**Request:**
```json
{
  "amount": 100
}
```
- `amount` must be 1-1000 microcoins per transaction

**Response:**
```json
{
  "ok": true,
  "coinsBurned": 100,
  "playerCoins": 4900,
  "burnedTotal": 250
}
```

### What Happens to Burned Coins

1. **Immediate deduction:** `player.coins` reduced by `amount`
2. **Tracking:** `player.burned_coins` incremented by `amount` (permanent cumulative total)
3. **Circulation removed:** Coins are permanently destroyed, not redistributed
4. **No rewards:** No gameplay benefit, achievement, or currency exchange

### Exchange Rate

**Fixed 1:1 destruction** — No conversion or exchange; purely destructive.

### Constraints

| Constraint | Value | Error Code |
|-----------|-------|-----------|
| **Minimum** | 1 microcoin | `INVALID_AMOUNT` |
| **Maximum** | 1000 microcoins (1 coin) | `INVALID_AMOUNT` |
| **Balance** | Must have sufficient coins | `NOT_ENOUGH_COINS` |
| **Season** | Must be active | `SEASON_ENDED` |
| **Feature Flag** | Sinks must be enabled | `FEATURE_DISABLED` |

### Database Schema

**Field:** `players.burned_coins`

```sql
ALTER TABLE players ADD COLUMN IF NOT EXISTS 
    burned_coins BIGINT NOT NULL DEFAULT 0;
```

**Update Pattern:**
```sql
UPDATE players 
SET coins = coins - $1, 
    burned_coins = burned_coins + $1
WHERE player_id = $2 
  AND coins >= $1;
```

### Design Intent

**Why Burn Coins?**

Burning provides **no immediate gameplay benefit**, but serves several design purposes:

1. **Counter-Inflationary Tool** — Removes coins from active circulation
2. **Prestige Signaling** — Cumulative burn total visible on profile (flex mechanic)
3. **Future Unlocks** — May gate cosmetics, titles, badges (post-alpha)
4. **Strategic Denial** — Prevent accidental spending; reduce market pressure

### Economy Impact

**Per-Player:**
- `burned_coins` field in `players` table (lifetime cumulative)
- Returned in `/player` API response as `burnedTotal`

**Global:**
- ⚠️ **NOT TRACKED** — No aggregate "total coins burned" metric in `season_economy`
- Circulation metric updates on next emission/faucet cycle

**Strategic Implications:**
- Pure coin sink with no return
- May slow star price inflation if many players burn
- Future utility uncertain (post-alpha features TBD)

---

## See Also

- [Coin Emission](coin-emission.md) — Coin supply and circulation mechanics
- [Market Pressure](market-pressure.md) — Price dynamics and demand signals
- [Anti-Abuse](anti-abuse.md) — Purchase restrictions and enforcement
