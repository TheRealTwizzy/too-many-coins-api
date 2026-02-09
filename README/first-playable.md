# First Playable

## Scope
- **Type:** System contract
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

The first playable milestone is a stripped-down but fully real version of the game.

Scope of the first playable:

Included:

One active season only

Real season timer (Alpha: 14 days by default; extension up to 21 days only with explicit telemetry‑gap configuration; may be accelerated for testing)

Account creation and login

Coin emission system (simplified but real)

One coin faucet: daily login

Login playability safeguard (alpha) to ensure a minimal playable balance

One coin faucet: capped active play

Star purchasing (single and bulk)

Full pricing logic (time, quantity, market pressure, late-season spike)

Leaderboard

Basic admin monitoring

IP-based season access enforcement

Excluded:

Multiple concurrent seasons

No whitelist system; rely on throttles and abuse detection

Daily tasks

Comeback rewards

Brokered player trading

Cosmetics and meta-progression

Endgame challenges

Season modifiers

Goals of the first playable:

Verify inflation pacing

Verify bulk-buy deterrence

Observe early vs late player behavior

Identify abuse vectors

Validate server authority and atomicity

The first playable is considered successful when:

Players can earn coins

Players can buy stars

Prices visibly rise over time

Coin scarcity is felt

The economy does not collapse under light load

The first playable may be discarded or reset frequently.

---

## Alpha Daily Loop (Explicit)

Primary loop (Alpha):

1) Log in (or sign up)
- Server verifies session and season availability.
- Login playability safeguard may top up very low balances from the emission pool (short cooldown, alpha‑only; may bypass daily cap).

2) Claim faucets
- Daily login faucet (once per 20 hours per season) draws from the emission pool.
- Activity faucet is available when recently active and draws from the emission pool.
- Per‑day earning cap applies; late‑season caps are lower.
- If the emission pool is low, rewards are throttled; if empty, claims are denied.

3) Buy stars
- Single or bulk purchase.
- Server calculates price using time progression, quantity scaling, market pressure, late‑season spike, and affordability guardrail.
- Purchase is atomic and re‑checks balance at confirmation.

4) Observe economy feedback
- UI shows time remaining, current star price, coins in circulation, market pressure, and next emission cadence.
- SSE updates reflect changes in near‑real time.

Common failure modes (Alpha):

- SEASON_ENDED: season is over; earning and purchases are blocked.
- FEATURE_DISABLED: faucets or purchases disabled by phase/flags.
- COOLDOWN: faucet claim denied due to cooldown; wait for next window.
- DAILY_CAP: per‑day earning cap reached; wait for daily reset.
- EMISSION_EXHAUSTED: emission pool empty; faucet claims denied until replenished.
- NOT_ENOUGH_COINS: star purchase denied; earn more coins or wait for price changes.

---

## New‑Player Journey (Mid‑Season)

Goal: ensure a brand‑new player can earn coins and buy their first star within a short session.

1) Signup
- Create account and auto‑join the single active season.
- Server sets initial player state and applies IP throttles if needed.

2) First login
- Login safeguard may top up to a minimum playable balance (alpha‑only, emission‑pool backed, short cooldown).

3) First earn
- Claim daily login faucet (if not on cooldown).
- Claim activity faucet after any active action (if within activity window).
- If the emission pool is low, rewards are throttled; if empty, faucet claims are denied.

4) First star purchase
- Buy 1 star (or a small bulk) using server‑calculated price.
- If not enough coins, return to faucet claims or wait for price changes.

5) Feedback and next action
- UI shows time remaining, price, pressure, coins in circulation, and next emission cadence.
- Player can repeat activity faucet loops until daily cap or cooldowns apply.

Failure checkpoints to watch:

- Daily faucet cooldown not clear to player.
- Emission pool exhaustion makes first earn impossible.
- Daily cap reached before first star purchase.
- Price spikes outpace first‑session earnings.

---

## Late‑Season Joiner Viability (Alpha)

Goal: a late‑season joiner can still perform meaningful actions (earn coins, buy at least some stars), even if they cannot compete for top rank.

Viability expectations:

- Login safeguard prevents zero‑action sessions by topping up very low balances (emission‑pool backed).
- Daily login and activity faucets still provide non‑zero rewards under late‑season caps.
- Star prices are higher, but affordability guardrails keep single‑star purchases possible for active players.

Late‑season failure signals to watch:

- Emission pool depletion blocks all faucet claims for extended periods.
- Daily caps drop so low that a first star purchase becomes impossible.
- Market pressure spikes make prices outpace faucet earnings for days at a time.
- UI fails to surface the next emission cadence, causing confusion about when to act.

---

## Always‑Available vs Tightening Actions (Alpha)

Always‑available actions (subject to cooldowns/flags):

- Log in / view season dashboard.
- Claim daily login faucet (if cooldown elapsed and emission pool has coins).
- Claim activity faucet (if recently active and emission pool has coins).
- Buy a single star when affordable.

Tightening actions over time:

- Bulk star purchases become increasingly inefficient due to quantity scaling and late‑season spike.
- Daily earning caps decline with season progress.
- Emission pool throttling becomes more common as scarcity increases.

Hard‑blocked actions in Alpha:

- Trading (disabled).
- Multi‑season selection (single season only).
- Cosmetics, collections, and post‑alpha progression systems.