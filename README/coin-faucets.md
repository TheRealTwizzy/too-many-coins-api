# Coin Faucets

## Scope
- **Type:** System contract
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

# Coin Faucets

## Currency Model: Integer Microcoins

**Canonical Currency Unit: Microcoins (integer only)**

The economy uses integer microcoins as the sole authoritative currency:

- **1 Coin = 1000 microcoins**
- All faucet rewards are expressed and stored as integer microcoins
- All faucet logic (cooldowns, rate-limiting, caps) uses integer microcoins only
- No floating-point faucet amounts exist at runtime
- Coins are a **display format only** derived from microcoins: `microcoins / 1000` with exactly 3 decimal places

This ensures:
- Perfect precision in reward calculations
- No rounding errors in daily caps or accumulation
- Consistent earning semantics across all faucets
- Clean audit trail for reward verification

---

## Players Earn Coins Through Faucets

Players earn coins through a limited set of server-controlled faucets.

## Universal Basic Income (UBI) — Dynamic Activity-Based Payout

Every player receives coins every game tick (continuous income).

**Base payout per tick: 1 microcoin (0.001 coin)**

**Active players earn up to 10x more through activity warmup:**

- Players start at the base rate (1 microcoin/tick)
- Sustained activity increases the "warmup level" over ~30 minutes
- At maximum warmup, players earn 10 microcoins (0.010 coins) per tick
- When idle, warmup decreases gradually
- Players with more recent activity lose warmup more slowly

**Activity Warmup Mechanics:**

- **Warmup increase**: Linear growth during sustained activity (30 minutes to reach maximum)
- **Warmup decay**: Gradual decrease when idle, slower decay for recently active players
- **Activity threshold**: Based on the active activity window (configurable, default varies by phase)
- **Income scaling**: Linear from 1x (base) to 10x (fully warmed up)

**Coins are displayed with exactly 3 decimal places** (thousandths precision).

This ensures:

- Players are **never unable to play** (minimum 1 microcoin/tick guaranteed)
- Active engagement is rewarded with higher sustained income
- Even late-season or throttled players receive minimal income
- The game remains accessible at all population levels
- Active players have a meaningful income advantage without gatekeeping

**Star costs MUST scale over the season** to prevent UBI from trivializing scarcity.

**Inflation and UBI must be considered together** when tuning the economy.

UBI is the foundation of the economy; all other faucets are additive.

---

## Faucet Types

Daily Login:

Available once per 20 hours per season.

Grants a small, fixed amount of coins (stored as microcoins, displayed with 3 decimals).

Designed to reward consistency, not grinding.

Daily Tasks:

Post‑alpha only.

A small set of simple tasks refreshed at daily reset.

Tasks grant moderate coin rewards (stored as microcoins).

Completing all tasks does not exceed the player daily earning cap.

Active Play:

Coins are granted at a slow, steady rate during active participation.

Active play rewards are capped per hour and per day.

AFK or idle behavior does not generate coins.

Alpha runtime default: passive drip is disabled (`drip_enabled=false`).

Comeback Reward:

Post‑alpha only.

Available only to players who have been inactive for a defined period.

Grants a one-time, modest coin boost.

Cannot exceed a fixed percentage of the daily earning cap.

Caps:

Each player has a daily coin earning cap per season.

The cap decreases as the season progresses.

Late-season caps are significantly lower than early-season caps.

Even late-season caps allow some daily earning; faucets never drop to zero.

All faucet grants draw from the global emission pool.
If the pool is low, faucet rewards are proportionally throttled and may grant a partial amount.
If the pool is empty, the faucet claim is denied.

Faucet tuning is balanced against trade burn to keep the economy liquid enough for daily action.
Coin shortage is possible but rare.

All faucet usage and coin grants are validated server-side and logged.

Login Playability Safeguard (Alpha):

If a player logs in with a balance too low to make near-term progress,
the server may top them up to a minimum playable balance.

This safeguard:

Draws from the global emission pool
Has a short cooldown
Is intended to keep the game playable within minutes
May bypass the per-day earning cap as an alpha-only safety net