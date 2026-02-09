The database stores only authoritative facts required to reconstruct the game state.

Alpha note: the current schema is minimal and does not yet include all entities below. The list is the target canonical model and is post‑alpha unless explicitly implemented.

Currency model (post‑alpha canon, not implemented in Alpha):

Coins and Stars remain seasonal and reset each season.
Post‑alpha introduces a persistent meta currency for cosmetic / identity use only; it cannot be traded, cannot convert into Coins or Stars, and cannot affect competitive power.
An optional influence / reputation metric may exist post‑release; it is non‑spendable, eligibility/visibility‑only, and never convertible.
No currency may ever convert into Coins or Stars, directly or indirectly.

Post‑alpha seasonal instruments (Beta‑only):

Tradable Seasonal Assets (TSAs) are seasonal, player‑owned competitive assets (not currencies). TSAs are system‑minted only (including via Star sacrifice), reset at season end, and never convert into Coins or Stars or generate Coins or Stars. TSA trading is player‑negotiated; the system enforces legality, caps, and logging. Trading remains disabled in Alpha.

Persistent entities include:

Players:

player_id

account_id

created_at

last_login_at

trust_status (normal, throttled, flagged; admin-only internal flag)

Seasons:

season_id

start_time

end_time

current_day

status (active, ended)

PlayerSeasonState:

player_id

season_id

coin_balance

star_balance

daily_earn_total

last_earn_reset_at

last_action_at

EconomyState (per season):

season_id

current_base_price

market_pressure

global_coin_budget_remaining

last_emission_tick_at

trade_premium

trade_burn_rate

trade_eligibility_tightness

Purchases (append-only log):

purchase_id

player_id

season_id

star_quantity

total_coin_cost

price_snapshot

quantity_multiplier_snapshot

market_pressure_snapshot

created_at

Trades (append-only log):

trade_id

seller_player_id

buyer_player_id

season_id

star_quantity

coin_price

coin_burned

trade_premium_snapshot

eligibility_snapshot

created_at

TSA Mint Log (append-only, post‑alpha/Beta-only):

tsa_mint_id
tsa_type
season_id
player_id
minted_quantity
mint_source (star_sacrifice, system_drop)
stars_destroyed
active_players_snapshot
day_index
created_at

TSA Trade Log (append-only, post‑alpha/Beta-only):

tsa_trade_id
tsa_type
season_id
seller_player_id
buyer_player_id
consideration_summary
coin_burned
star_burned
friction_snapshot
destroyed
created_at

TSA Activation Log (append-only, post‑alpha/Beta-only):

tsa_activation_id
tsa_type
season_id
player_id
activation_effect
created_at

CoinEarnings (append-only log):

earning_id

player_id

season_id

source_type (login, task, activity, comeback)

amount

created_at

AbuseEvents:

event_id

player_id

season_id

event_type

severity

created_at

Derived values such as star prices, caps, throttles, and rankings are computed server-side and are not trusted if provided by clients.

All coin and star balance changes must occur inside database transactions.