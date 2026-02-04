The server is the sole authority over time.

The system uses three time layers:

Tick Time:

A fixed server tick runs every 60 seconds.

Tick time is used for coin emission, throttling adjustments, and market pressure smoothing.

Tick execution is idempotent and safe to retry.

Daily Time:

Each season has a daily boundary based on UTC.

Daily resets occur once per day per season.

Daily resets include:

Resetting per-player daily earning totals

Refreshing daily tasks (post‑alpha)

Daily reset logic must be guarded so it runs exactly once per season-day.

Season Time:

Each season has a fixed start_time and end_time, with length defined by server phase:

- Alpha: 14 days by default, extendable up to 21 days only with explicit telemetry-gap configuration.
- Beta: 28 days.
- Release: 28 days.

Season day index is derived from the UTC date difference between the season start day and the current day.

Pricing curves, coin budgets, and caps are based on the derived season day.
Trade eligibility bands, premiums, and burn rates are also based on the derived season day.

Season end transitions the season to ended state and freezes economy actions.

Clients receive all time-related values from the server and never calculate time-sensitive logic locally.