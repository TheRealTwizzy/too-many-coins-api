# UI-1 Frontend MVP

The frontend will launch with a minimal set of pages focused on core gameplay.

Required Pages:

Landing Page:

Explains the game in a few sentences.

Provides login and signup entry points.

Authentication Page:

Signup and login.

Account verification and CAPTCHA are post‑alpha.

Season Lobby is post‑alpha (alpha has a single active season only).

Season Dashboard:

Primary gameplay page.

Displays coin balance, star balance, current star price, time remaining, coins in circulation, market pressure, and next emission cadence.

Provides actions to earn coins and buy stars.

Shows basic feedback for price changes and throttles.

Shows trading eligibility status and current trade premium/burn (post‑alpha).

Season-ended behavior (Alpha):

- The season snapshot includes `season_status: active | ended`.
- When `season_status = ended`:
	- Hide or disable Buy Stars and all earn actions.
	- Render a read-only Season Ended state.
	- Do not show live emission/inflation cadence or market pressure as active rates.
	- Coins in circulation show a frozen/final snapshot marker.

Bulk Purchase Modal:

Allows selecting star quantity.

Displays full calculated cost and scaling effects.

Requires explicit confirmation.

Leaderboard Page:

Shows player ranking and tiers.

Displays the player’s current rank.

Profile Page is post‑alpha (badges, titles, cosmetics, season history).

Settings Page is post‑alpha (accessibility options, account preferences).


All pages must be simple, fast, and functional.
No non-essential pages are built before economy testing.

Trading Desk (when enabled):

Allows requesting a system-quoted trade.

Shows eligibility gates, burn cost, and premium.

Requires explicit confirmation and warns that trades are expensive and irreversible.

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
