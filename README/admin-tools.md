The system must include a minimal internal admin and observability interface.

Admin access is restricted to authorized accounts only.

Required admin capabilities are split by phase. Alpha is read‑only.

Alpha (read‑only, current build):

Season monitoring:

View the active season and its current day.

View season status (active, ending, ended).

Economy monitoring (per season):

Current base star price.

Current market pressure.

Current effective star price.

Global coin budget remaining for the day.

Coin emission rate and throttling state.

Flow metrics (if available in telemetry):

Coins emitted per hour.

Coins earned per hour.

Stars purchased per hour.

Average star price over time.

Player inspection (read‑only):

View individual player season state.

View coin and star balances.

View purchase history.

View coin earning history.

View trust and throttle status.

Abuse monitoring (read‑only):

View recent abuse events.

View IP-based clustering signals.

View throttles currently applied.

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