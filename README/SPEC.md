Game: Too Many Coins! (MMO web)

Seasons (phase‑bound, server‑defined):

- Alpha: 14 days by default; extension up to 21 days only with explicit telemetry‑gap configuration. Single active season only.
- Beta: 28 days, total seasons 2–3, staggered/overlapping.
- Release: 28 days, concurrent seasons up to 4, staggered.

Join rules: players can join any active season anytime

Core loop: earn coins (faucets) -> buy stars (system store) -> optional brokered trading -> leaderboard rank (stars are the only direct ranking unit)

Trading: optional, costly, asymmetric, brokered Coins-for-Stars only; system-priced with premium and coin burn; eligibility gates tighten over time

TSA trading (post‑alpha, Beta‑only): player‑negotiated, freely tradable player‑to‑player. The system enforces legality, caps, and logging; it does not set prices. TSA trades never create Coins or Stars and always contribute to market pressure when enabled.

Bulk buy allowed but price scales so hard it’s almost infeasible late-season

Inflation model: server-emitted coin pool, time-sliced, global daily budget decreases over season; trade burn is modeled and balanced to preserve liquidity

Inflation pressure is monotonically increasing; delay is punished and mistakes are permanent

Late-season: star prices higher, coin supply lower, trading narrower and more expensive

Affordability: star prices stay aligned with per-player coin emission so most active players can buy stars throughout the season

Anti-abuse: 1 active account per IP per season by default; no whitelist, throttles + trust-based enforcement only

Rewards: cosmetics/titles/badges persist between seasons; coins/stars reset per season

Post‑alpha currency model: introduce a persistent meta currency (Beta) for cosmetic/identity use only; it cannot be traded, cannot convert into Coins or Stars, and cannot affect competitive power. An optional influence/reputation metric may exist post‑release; it is non‑spendable, eligibility/visibility‑only, and never convertible.

Post‑alpha seasonal instruments (Beta‑only): Tradable Seasonal Assets (TSAs) are seasonal, player‑owned competitive assets (not currencies). TSAs are system‑minted only (including via Star sacrifice), reset at season end, never convert into Coins or Stars, and never generate Coins or Stars. TSAs may affect rank indirectly through their utility, but never grant Stars directly.

Hard TSA invariants: TSAs cannot mint Coins or Stars; TSAs cannot be converted into Coins or Stars; Stars sacrificed for TSAs are permanently destroyed; TSA supply is observable and auditable.

Hard prohibition: no currency may ever convert into Coins or Stars, directly or indirectly.

Real-time: prices and season stats update via SSE or WebSockets

Server-authoritative economy: all pricing, caps, purchases enforced server-side