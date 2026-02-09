# Market Pressure

## Scope
- **Type:** System contract
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

Market pressure represents recent demand for stars and influences star prices.

Market pressure is calculated using rolling averages over time and is rate-limited.

Market pressure inputs:

Total stars purchased in the last 24 hours

Total stars purchased in the last 7 days

Total stars transferred via brokered trades in the last 24 hours (post‑alpha)

Total stars transferred via brokered trades in the last 7 days (post‑alpha)

TSA trading (post‑alpha, Beta‑only):

TSA trades (player‑negotiated; any quantity metric defined for the TSA) contribute to market pressure when enabled.

Market pressure calculation:

Pressure increases when short-term demand exceeds long-term average demand.

Pressure decreases gradually when demand slows.

Market pressure smoothing rules:

Market pressure changes are capped per server tick.

Market pressure may not increase or decrease by more than a small percentage per hour.

Sudden spikes in demand are absorbed over time rather than applied instantly.

Market pressure is derived server-side and stored as a season-level value.
Clients only receive the current pressure value and never compute it.

Market pressure is applied multiplicatively to star prices.

Brokered trading always contributes to market pressure, never relieves it.

Alpha note: trading is disabled; market pressure is derived from star purchases only.
Alpha verification: no trade inputs are wired in Alpha; only star_purchase_log is used.
Alpha verification: market pressure is included in SSE snapshots and the UI binds to it; per-tick rate limiting enforces stability under bursts.

Market pressure must be resistant to day-one coordinated activity and bot-driven bursts.