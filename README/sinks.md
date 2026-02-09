# Coin Sinks

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

# Coin Sinks & Voluntary Destruction

Coin sinks remove coins from circulation to counter inflation. The primary sinks are star purchases and coin burning.

---

## Star Purchases (Primary Sink)

See [star-purchases.md](star-purchases.md) for complete star purchasing mechanics.

**Summary:**
- Star purchases are the **primary competitive sink**
- Coins spent on stars are permanently removed from circulation
- Price increases over time and with demand (market pressure)
- Bulk purchases penalized through quantity scaling

---

## Coin Burning (Voluntary Destruction)

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
- Value in microcoins (1 coin = 1000 microcoins)

**Response:**
```json
{
  "ok": true,
  "coinsBurned": 100,
  "playerCoins": 4900,
  "burnedTotal": 250
}
```

**Fields:**
- `coinsBurned`: Amount destroyed in this transaction
- `playerCoins`: Updated coin balance after burn
- `burnedTotal`: Lifetime cumulative coins burned by this player

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

Atomic increment ensures accurate lifetime burn total.

---

## Design Intent

### Why Burn Coins?

Burning provides **no immediate gameplay benefit**, but serves several design purposes:

1. **Counter-Inflationary Tool**
   - Removes coins from active circulation
   - Reduces total supply without admin intervention
   - Player-driven deflation mechanism

2. **Prestige Signaling**
   - Cumulative burn total visible on profile (if implemented)
   - Display of wealth/confidence ("I have so many coins I can afford to waste them")
   - Flex mechanic for top earners

3. **Future Unlocks** (Post-Alpha)
   - May gate cosmetics, titles, badges
   - Possible achievement system tied to burn milestones
   - Potential rare currency conversion (e.g., 1000 burned coins → 1 prestige token)

4. **Strategic Denial**
   - Prevent accidental spending on stars near season end
   - Reduce coins-in-circulation metric to slow star price inflation
   - Competitive tactic: burn to starve market pressure

---

## Economy Impact

### Tracked Metrics

**Per-Player:**
- `burned_coins` field in `players` table (lifetime cumulative)
- Returned in `/player` API response as `burnedTotal`

**Global:**
- ⚠️ **NOT TRACKED** — No aggregate "total coins burned" metric in `season_economy`
- No telemetry event emitted (unlike star purchases)
- Not recorded in `coin_earning_log` or `star_purchase_log`

### Circulation Impact

Burned coins are **permanently destroyed:**
- Reduces `player.coins` balance
- Does NOT reduce `season_economy.coins_in_circulation` directly
- Circulation metric updates on next emission/faucet cycle when burned coins are no longer counted

**Example:**
```
Before burn:
  player.coins = 5000
  coins_in_circulation = 50000 (aggregate of all players)

Player burns 100 coins:
  player.coins = 4900 (immediate)
  coins_in_circulation = 50000 (unchanged until next recalculation)

Next emission tick:
  coins_in_circulation = 49900 (recomputed aggregate reflects burn)
```

### Strategic Implications

**For Players:**
- Pure coin sink with no return
- May slow star price inflation if many players burn
- Future utility uncertain (post-alpha features TBD)

**For Economy:**
- Voluntary deflation tool (complements mandatory star purchases)
- Low adoption expected unless incentivized (cosmetics, achievements)
- Burn totals may serve as wealth indicator or gating mechanism

---

## Gating & Requirements

| Requirement | Check | Error Code |
|------------|-------|-----------|
| **Authentication** | Valid session | `UNAUTHORIZED` |
| **Season Active** | `!isSeasonEnded()` | `SEASON_ENDED` |
| **Feature Enabled** | `ENABLE_SINKS=true` | `FEATURE_DISABLED` |
| **Valid Amount** | 1 ≤ amount ≤ 1000 | `INVALID_AMOUNT` |
| **Sufficient Balance** | `player.coins >= amount` | `NOT_ENOUGH_COINS` |

---

## Technical Notes

**Implementation:**
- Single UPDATE query (atomic)
- No telemetry emission (silent operation)
- No admin notification (low-priority action)
- No abuse detection (burning coins harms player, not economy)

**Performance:**
- O(1) operation (primary key lookup + update)
- No side effects or cascading updates
- Safe to call repeatedly

**Edge Cases:**
- Burning to exactly 0 coins: allowed
- Burning more than balance: fails with `NOT_ENOUGH_COINS`
- Burning during season end: fails with `SEASON_ENDED`
- Multiple rapid burns: subject to auth rate limits only (no burn-specific cooldown)

---

## Future Enhancements (Post-Alpha)

Potential burn mechanics under consideration:

1. **Burn Milestones:**
   - 100 coins burned → Bronze badge
   - 1000 coins burned → Silver badge
   - 10,000 coins burned → Gold badge

2. **Rare Currency Conversion:**
   - 1000 burned coins → 1 prestige token
   - Prestige tokens unlock exclusive cosmetics/titles

3. **Profile Display:**
   - Lifetime burn total visible on profile
   - Leaderboard for "most coins burned" (flex competition)

4. **Strategic Burn Incentives:**
   - Burn during final week → bonus star discount
   - Burn threshold gates "comeback reward" eligibility

None of these are implemented in Alpha. Current burn mechanic is **pure destruction with no reward**.

---

## See Also

- [Star Purchases](star-purchases.md) — Primary competitive coin sink
- [Coin Emission](coin-emission.md) — Coin supply and circulation mechanics
- [HTTP API Reference](http-api-reference.md) — `/burn-coins` endpoint spec
