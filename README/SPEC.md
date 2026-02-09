# VISION-1 Game Spec

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

---

## Rare and Hard-to-Find Currencies (Post-Alpha)

Rare currencies are special, limited-drop currencies introduced in Beta and expanded in Full Release.

**Beta**: Introduce 1–2 rare currencies  
**Full Release**: Introduce 3–5 rare currencies

### Rare Currency Rules

- Rare currencies drop randomly via normal gameplay
- Drop rates may be slightly influenced by anti-cheat (bad actors receive fewer drops)
- Rare currencies enable **pay-to-win features** (stronger purchases, exclusive advantages)
- Rarity scales value: rarer currency → stronger benefit
- Rare currencies ARE tradable (player-to-player, negotiated trades)
- Purchases NEVER affect drop rates (no pay-to-increase-drops)
- Rare currencies are **seasonal** and reset at season end

### Design Philosophy

Rare currencies introduce asymmetric power for players who:

- Play actively
- Get lucky
- Trade strategically

They create:

- Social trading markets (Diablo 2 rune market feel)
- Risk and regret (spend now or hold for better trade?)
- Meaningful differentiation between players

Rare currencies are NOT required to compete, but they offer **meaningful advantages** for those who obtain them.

Rare currencies must feel **found**, not farmed.

---

## Trading Rules (Comprehensive)

### Brokered Trading — Coins ↔ Stars (Post-Alpha)

Brokered trading is system-priced and controlled:

- Trades are **Coins-for-Stars only** (no Coin-for-Coin, no Star-for-Star)
- The system sets the price (players do not negotiate)
- Every trade **burns Coins** as overhead (Coins are destroyed, never re-enter the economy)
- Trades never create Coins or Stars
- Trades are **asymmetric**: buyer pays more than seller receives
- Trades are priced **at or above current system star price + time-based premium**
- Trades **always contribute to market pressure** (never relieve it)

Eligibility gates (must pass ALL):

- Both players must be active and time-normalized
- Both must have recent coin spending activity (no pure hoarders)
- Relative Star holdings must be within a tightening ratio band
- Coin liquidity must be within a tightening band (not too low, not too high)
- Inflation exposure difference must be within a tightening band

**As the season progresses**:

- Eligibility gates TIGHTEN
- Trade burn percentage RISES
- Maximum Stars per trade DROPS
- Daily trade limits DECREASE

Late-season trading is expensive, dangerous, and narrow — but still rational in specific cases.

### Player-to-Player Trading — Rare Currencies and TSAs (Post-Beta)

Player-to-player trading is negotiated and flexible:

- Valid trades:
  - Coins ↔ Rare Currencies
  - Rare Currencies ↔ Rare Currencies
  - TSAs ↔ Coins
  - TSAs ↔ Rare Currencies
  - TSAs ↔ TSAs

- Invalid trades:
  - Stars (never tradable in any context)
  - Meta currency (never tradable)
  - Influence/reputation (never tradable)

All player-to-player trades:

- Are **player-negotiated** (the system does not set prices)
- Are **enforced by the system** (legality, caps, logging)
- Must be **logged** (all trades feed economic telemetry)
- May include **friction** (Coin burn, fees, caps)
- Must feel **meaningful for both parties**

### Trading Philosophy

Trading is **pressure, not relief**:

- Trades do not save players from inflation
- Trades add scarcity, risk, and social dynamics
- Trades are costly tools for repositioning, not catch-up mechanics

Trading must create:

- Social markets
- Negotiation dynamics
- Risk and regret
- Opportunities for clever plays
- Punishment for mistakes

---

## Communication Systems (Post-Alpha / Beta)

Players need robust communication to support trading, social dynamics, and community engagement.

### Direct Messaging (Post-Alpha)

- Players can send direct messages to other players
- Messages must integrate with:
  - Notifications (alerts for new messages)
  - Trades (offer negotiation, confirmation)
  - Player profiles (message from profile link)

### Forum (Post-Beta)

A game-integrated forum where players can:

- Discuss strategies
- Negotiate trades publicly
- Build community reputation
- Form alliances and rivalries

Forum roles:

- **Admins** → Forum admins (full moderation powers)
- **Moderators** → Forum moderators (limited moderation powers)
- **Players** → Forum users (standard posting privileges)

### Communication Integration

Messaging and forum must integrate with:

- **Notifications**: Real-time alerts for messages, replies, trade offers
- **Trades**: Offer negotiation, confirmation, completion notifications
- **Economy events**: Price spikes, season milestones, leaderboard changes
- **Player profiles**: Contact links, reputation, badges

### Overall Feel

The communication systems must evoke:

- **Diablo 2 rune market**: Player-driven, social, economic
- **Competitive tension**: Rivalries, alliances, betrayals
- **Community identity**: Long-term relationships, reputation, history

Players must feel:

- Connected to the community
- Incentivized to negotiate and trade
- Invested in their reputation and relationships
