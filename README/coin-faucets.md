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

**Currency Model:** All faucet rewards use integer microcoins (1 coin = 1000 microcoins). See [coin-emission.md](coin-emission.md) for details.

Players earn coins through a limited set of server-controlled faucets that draw from the global emission pool.

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

### Daily Login

- Available once per 20 hours per season
- Grants a small, fixed amount of coins
- Designed to reward consistency, not grinding
- Cooldown-based, not grinding-friendly

### Activity Faucet

- Claimable periodically during active play sessions
- Short cooldown (5 minutes default)
- Rewards sustained engagement
- Capped per hour and per day

### Daily Tasks (Post-Alpha)

- A small set of simple tasks refreshed at daily reset
- Tasks grant moderate coin rewards
- Completing all tasks does not exceed the player daily earning cap

### Comeback Reward (Post-Alpha)

- Available only to players who have been inactive for a defined period
- Grants a one-time, modest coin boost
- Cannot exceed a fixed percentage of the daily earning cap

## Daily Earning Caps

Each player has a daily coin earning cap per season:

- Cap decreases as the season progresses
- Late-season caps are significantly lower than early-season caps
- Even late-season caps allow some daily earning
- UBI is NOT subject to daily caps (foundation income)
- Other faucets (daily, activity) ARE subject to caps

## Emission Pool Integration

All faucets draw from the global emission pool:

- If the pool is low, faucet rewards are throttled
- If the pool is empty, optional faucet claims are denied
- UBI grants continue even under pool pressure (with throttling)
- Faucet tuning balances emission with trade burn (post-alpha)
- All grants are validated server-side and logged

## Login Playability Safeguard (Alpha)

Alpha-only safety net for new/returning players:

- Tops up players with extremely low balances
- Draws from the global emission pool
- Short cooldown to prevent abuse
- Intended to keep the game playable within minutes
- May bypass daily earning cap (temporary Alpha measure)