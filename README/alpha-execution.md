# Alpha Execution

## Scope
- **Type:** System contract
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

# Alpha Execution Plan (First Playable)

This plan defines the minimum goals and metrics for the first alpha test. It aligns with the first‑playable scope and server‑authoritative economy.

## Goals

- Verify inflation pacing across the Alpha season curve (14 days by default, or accelerated test window).
- Verify bulk‑buy deterrence is strong enough late‑season.
- Observe early vs late‑joiner behavior and perceived fairness.
- Identify abuse vectors around faucets, IP limits, and cooldowns.
- Validate server authority and atomicity under light load.

## Success Criteria

- Players can earn coins via daily login and active play.
- Players can buy stars (single + bulk) and prices rise over time.
- Coin scarcity is felt late‑season without total liquidity collapse.
- No client‑side trust or price manipulation is possible.

## Metrics to Track

Economy:
- Coins emitted per hour (from emission pool)
- Coins earned per hour (faucets; drip is post‑alpha and disabled in current build)
- Coins in circulation
- Stars purchased per hour
- Average star price per hour
- Market pressure value over time

Faucets:
- Daily login claims per day
- Activity claims per hour
- Claim rejection rate (cooldown, daily cap, emission exhausted)

Purchases:
- Star purchase attempts vs success rate
- Bulk purchase distribution (qty and total cost)
- Coin burn totals

Abuse/Access:
- IP‑blocked events
- Rate‑limited signup/login counts
- Account cooldown rejections

## Calibration Workflow (Alpha)

1) Collect at least 48–72 hours of telemetry under normal play.
2) Review emission_tick and faucet_claim rates for smooth throttling (no abrupt stalls).
3) Compare star_purchase_attempt vs star_purchase_success rates; flag sustained affordability failures.
4) Check market_pressure_tick trends for stability and bounded change per hour.
5) Adjust calibration inputs only with telemetry evidence; no blind tuning.
6) Re-run a short validation window after any change and compare deltas.

## Append‑Only Economic Logs (Alpha)

The following logs are append‑only in the database:

- coin_earning_log
- star_purchase_log
- abuse_events
- admin_audit_log

Queryability (Alpha):

- Admin UI exposes the star purchase log and admin audit log.
- Abuse events are visible in admin/moderator views.
- Coin earning history is stored in DB but not exposed in the admin UI (post‑alpha view).

## Telemetry Events (Current Build)

- buy_star
- login

Alpha note: emission pool levels/per‑tick emissions, market pressure per‑tick events, faucet claim events, and star purchase attempt/success events are now emitted.

## Telemetry Sufficiency (Alpha)

Required calibration inputs and current sources:

- Emission pacing: `emission_tick` (emitted, pool levels + per‑tick emissions).
- Faucet pacing and caps: `faucet_claim` (emitted, includes remaining cap context).
- Pricing curves: `star_purchase_attempt` + `star_purchase_success` (emitted, includes price snapshots and quantities).
- Market pressure: `market_pressure_tick` (emitted, includes current/target pressure and rate limit).

Conclusion: telemetry is sufficient to calibrate emission, caps, price curves, and pressure without client trust.

## Test Window

- Recommended: 7–14 days
- Consider accelerated season start via SEASON_START_UTC for testing

## Recruitment (Owner Task)

- Recruit 20–50 testers with mixed activity levels
- Collect feedback on late‑season scarcity and price clarity

## Post‑Test Review

- Review telemetry and identify top 3 economy risks
- Produce prioritized fixes before widening access

---

## Alpha UI Copy (Player-Facing)

About this game (Alpha):

Earn Coins, buy Stars, climb the leaderboard. Coins enter the world as people play, and Stars become harder to buy as the season moves forward. The pressure is shared: when everyone earns and spends, prices rise. Scarcity builds toward the end, and every decision matters.

What is missing (intentional in Alpha):

- Trading
- Multiple seasons
- Messaging and forums
- Cosmetics and profiles
- Advanced admin tools
- Daily tasks, comeback rewards, passive drip

These omissions are temporary and intentional so the core economy can be tested without distractions.

What is coming (post-Alpha / Beta):

We expect to explore trading, messaging and forums, persistent cosmetics and history, multiple seasons, and expanded anti-abuse systems. No timelines or guarantees. Alpha may reset and change.
