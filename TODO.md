# TODO — Canonical Execution Plan (Single Source of Truth)

This document supersedes any prior task list. It is ordered by dependency and reflects current code + README canon as of February 7, 2026.

## Scope & Change Protocol

- **Scope:** Execution plan only (requirements live in README/*.md and README.md).
- **Change protocol:** When a task modifies canon, update the relevant README/*.md section in the same logical unit.
- **Invariant safety:** If a task impacts Layer 1 invariants (README.md), treat it as a new primary logical unit.

Legend: each task is explicitly marked with a status tag.

Status Tags:
- [DONE]
- [ALPHA REQUIRED]
- [ALPHA EXECUTION]
- [POST-ALPHA]
- [POST-BETA]

---

## Alpha Exit Criteria (Non‑Negotiable)
- [ ] [ALPHA REQUIRED] Players can sign up, log in, and enter the active season without manual admin steps (single season is acceptable for Alpha).
- [ ] [ALPHA REQUIRED] Universal Basic Income (UBI) provides minimum 0.001 coin per tick to all players (economy foundation).
- [ ] [ALPHA REQUIRED] Economy runs continuously without admin intervention (tick loop, emission, UBI, star pricing, market pressure updates).
- [ ] [ALPHA REQUIRED] Daily play loop works: UBI + daily login faucet + active play faucet + decreasing daily cap + emission pool enforcement.
- [ ] [ALPHA REQUIRED] Login playability safeguard keeps new/returning players playable within minutes (alpha‑only, emission‑pool backed).
- [ ] [ALPHA REQUIRED] Players can observe the economy clearly in UI/SSE (time remaining, current star price, coins in circulation, market pressure, emission cadence).
- [ ] [ALPHA REQUIRED] Star purchase flow works end‑to‑end (single + bulk), shows warnings, and is atomic.
- [ ] [ALPHA REQUIRED] Season end freezes all economy actions and writes end‑of‑season snapshots.
- [ ] [ALPHA REQUIRED] Known missing systems are explicitly labeled in UI/docs (trading, multi‑season, cosmetics, communication, rare currencies, etc.).

Allowed to be rough or missing in Alpha:
- Brokered trading (Coins ↔ Stars) and trading UI
- Player‑to‑player trading (TSAs, rare currencies)
- Multi‑season runtime and season lobby
- Cosmetics, badges, titles, and long‑term progression (persistent meta currency, profile collections)
- Direct messaging and forum systems
- Rare currencies (Beta/Release)
- Settings and accessibility polish
- Admin analytics polish beyond core monitoring

---

## Phase 0 — Canon & Reality Reconciliation (Must Stay Honest)
- [x] [DONE] 0.1 Audit repository vs canon sources (README set + SPEC + alpha‑execution + first‑playable)
- [x] [DONE] 0.2 Resolve passive drip contradiction (code runs drip; canon says post‑alpha)
  - [x] [DONE] Decide Alpha stance: disable drip via global_settings.drip_enabled default false OR update canon to allow drip
  - [x] [DONE] Document operational default for alpha (global_settings + runtime log confirms drip disabled/enabled)
- [x] [DONE] 0.3 Resolve admin‑tools overreach (docs list views not present in admin endpoints/UI)
  - [x] [DONE] Audit Alpha admin UI/endpoints vs docs (coin budget remaining, coin earning history, trust/throttle status, IP clustering detail)
  - [x] [DONE] Mark missing views as post‑alpha or add explicit implementation tasks
- [x] [DONE] 0.4 Resolve anti‑abuse doc vs code (CAPTCHA/verification claimed but not implemented)
  - [x] [DONE] Docs already label CAPTCHA/verification as post‑alpha
- [x] [DONE] 0.5 Resolve persistent‑state doc vs schema (seasons/player‑season tables not present)
  - [x] [DONE] persistent‑state.md already marks schema expansion as post‑alpha unless implemented
- [x] [DONE] 0.6 Document login playability safeguard in canon docs
- [x] [DONE] 0.7 Resolve telemetry event naming mismatch (alpha‑execution says join_season; client emits login)
  - [x] [DONE] Decide canonical event names for Alpha and update telemetry contract + TODO list
- [x] [DONE] 0.8 Reconcile TODO.md with Game Bible (February 2026 audit; this document completed)

---

## Phase 0.9 — Currency Normalization (Canonical Authority)
- [x] [DONE] 0.9 Normalize economy to integer microcoins as sole authoritative currency
  - [x] [DONE] 0.9a Schema: Add current_star_price_micro BIGINT column (additive, Alpha-safe)
  - [x] [DONE] 0.9b Economy math: Replace `price + 0.9999` rounding with math.Ceil for prices (conservative charging) and math.Floor for rewards (conservative earning)
  - [x] [DONE] 0.9c API: All coin/price fields exposed as integer microcoins (no decimal types)
  - [x] [DONE] 0.9d Frontend: Confirm formatMicrocoinsToCoins formatter correctly converts to 3-decimal display
  - [x] [DONE] 0.9e Documentation: Update coin-emission.md, coin-faucets.md, star-purchases.md with microcoin canonical model
  - [x] [DONE] 0.9f Confirm: All coin math is integer-only; no floating-point coin comparisons remain

---

## Phase 1 — Backend Foundations (Authoritative Core)
- [x] [DONE] 1.1 Go + net/http backend
- [x] [DONE] 1.2 PostgreSQL persistence
- [x] [DONE] 1.3 Base schema tables
  - [x] [DONE] players, accounts, sessions
  - [x] [DONE] season_economy persistence
  - [x] [DONE] star_purchase_log, notifications
- [x] [DONE] 1.4 Verify schema is aligned with canonical entities (coin_earning_log, abuse_events, telemetry)

---

## Phase 2 — Time System & Season Core
- [x] [DONE] 2.1 Fixed phase‑bound season clock
- [x] [DONE] 2.2 60s tick loop for emission + pressure
- [x] [DONE] 2.3 Season end snapshot on tick
- [x] [DONE] 2.4 Validate time semantics (season day index, reset boundaries, end‑state gating)
- [x] [DONE] 2.4a Expose server day index + total days to UI (no client hardcoding)
- [x] [DONE] 2.4b Alpha season persistence + auto‑advance (single‑season, restart‑safe)
- [ ] [POST-ALPHA] 2.5 Multi‑season runtime model (seasons table, staggered starts, per‑season tick scheduling)

## Phase Transition Tasks (Explicit)
- [ ] [ALPHA REQUIRED] Alpha → Beta: introduce phase config (`PHASE`) and verify Beta season length (28 days) with 2–3 staggered/overlapping seasons (runtime model remains post‑alpha until 2.5 is implemented).
- [ ] [POST-ALPHA] Beta → Release: remove Alpha-only safeguards (single-season lock, alpha extension gates) after multi‑season runtime is stable and verified.

---

## Alpha → Beta Plan (Build-Based, Detailed)

This section defines the Alpha → Beta transition in terms of scope, implementation, affected systems, and player-facing changes. It is the authoritative execution plan for the transition.

### Goals (What changes are going to be made)
- Move from a single-season Alpha runtime to a Beta model with longer seasons (28 days) and early support for multiple overlapping seasons.
- Introduce Beta-only competitive assets (TSAs) and persistent, non-economic meta currency (cosmetics/identity only).
- Enable brokered trading (Coins ↔ Stars) with strict eligibility and tightening rules.
- Establish season-end permanent score conversion (Stars → permanent profile statistic) and prepare for reward grants.
- Expand player-facing UI to show trading desks, profile progression, and season lobby.

### Implementation Outline (How changes will be implemented)
- Phase config switch (`PHASE=beta`) gates Beta-only systems without changing Alpha defaults.
- Add database tables for multi-season runtime, TSA inventory/logs, brokered trades, meta currency, and season history.
- Extend tick loop and economy pipelines to handle per-season scheduling, TSA effects, and brokered trade pressure.
- Add endpoints and SSE payload extensions for new trading, profile, and season lobby views.
- Update admin observability for new logs (trades, TSA minting/burn, meta currency grants).

### Systems Created or Affected
- **Season Runtime:** multi-season scheduling, season lobby listing, season history persistence.
- **Economy:** brokered trading, TSA mint/burn, market pressure coupling, additional logs.
- **Profile/Progression:** permanent star score, meta currency, rewards/badges framework.
- **Anti-Abuse:** new signals for trading abuse and TSA exploitation.
- **Frontend:** trading desks, profile/collection, season lobby, expanded notifications.

### Player Experience (What the player sees and how play changes)
- Players can join active seasons from a lobby and see overlapping season timelines.
- Stars remain non-tradable but convert into a permanent profile score at season end.
- Players can sacrifice Stars to mint TSAs (rank drops immediately).
- Brokered trading offers a system-priced Coins ↔ Stars desk with visible restrictions and costs.
- Profile shows permanent score, cosmetics/identity meta currency, and season history.

### Build Plan (What changes and when)

#### Build B1 — Beta Runtime Foundation
- **Backend:** Add `PHASE` config; extend season model for 28-day seasons; create seasons table and scheduler (Phase 2.5).
- **Schema:** Add seasons, season_history, season_runtime_state.
- **API/SSE:** Season lobby list + current season metadata; expose phase config.
- **Frontend:** Add season lobby UI; show season length and overlap indicators.
- **Systems Affected:** Time system, season core, admin season advance.

#### Build B2 — Persistent Score + Profile Shell
- **Backend:** Implement star conversion at season end (Phase 12.2); store permanent score.
- **Schema:** Add profile_stats (permanent_score, season_count, last_season_id).
- **API/SSE:** Add profile summary endpoint + SSE payload fields for score updates.
- **Frontend:** Add profile screen with permanent score and season history shell.
- **Systems Affected:** Season end, player profile, notifications.

#### Build B3 — Meta Currency (Cosmetic Only)
- **Backend:** Implement meta currency grant logic (non-economic); deny any conversion paths.
- **Schema:** Add meta_currency_wallet and grant_log.
- **API/SSE:** Expose meta currency balances; notifications for grants.
- **Frontend:** Add cosmetics placeholder and meta currency display.
- **Systems Affected:** Progression, rewards, admin visibility.

#### Build B4 — Brokered Trading (Coins ↔ Stars)
- **Backend:** Implement brokered trade engine, eligibility gates, tightening rules, burn mechanics.
- **Schema:** Add brokered_trades + brokered_trade_events logs.
- **API/SSE:** Trading desk endpoints, price/premium display, eligibility feedback.
- **Frontend:** Brokered trade desk UI with warnings, eligibility feedback, and burn breakdown.
- **Systems Affected:** Economy, market pressure, anti-abuse, notifications.

#### Build B5 — TSAs (Competitive Assets)
- **Backend:** Implement Star Sacrifice → TSA minting; TSA inventory; trading stubs.
- **Schema:** Add tsa_inventory, tsa_mint_log, tsa_trade_log (trade API can remain disabled until B6).
- **API/SSE:** TSA inventory endpoints; mint events; admin visibility.
- **Frontend:** TSA inventory UI and mint flow with rank-loss confirmation.
- **Systems Affected:** Economy pressure, season end wipe behavior.

#### Build B6 — TSA Trading + Anti-Abuse Expansion
- **Backend:** Enable TSA trading; add abuse detection for reciprocal trades and IP clustering.
- **Schema:** Add trade_offers and trade_confirmations (if negotiation is added).
- **API/SSE:** Trade negotiation endpoints; trade status notifications.
- **Frontend:** Player-to-player trading UI; moderation hooks.
- **Systems Affected:** Anti-abuse, admin observability, notifications.

#### Build B7 — Beta UX Polish
- **Backend:** Finalize notifications for trades, TSA events, and season lobby.
- **Frontend:** Upgrade economy dashboards with brokered trade metrics and TSA status; refine copy.
- **Systems Affected:** UI/UX, onboarding, documentation alignment.

### Beta Exit Criteria (Draft)
- Multi-season runtime stable with at least 2 overlapping seasons running safely.
- Brokered trading live with eligibility tightening and observable burn.
- TSA minting and trading functional with abuse detection.
- Permanent score conversion at season end and profile view accessible.
- Meta currency exists with no conversion paths into Coins or Stars.

---

## Build Breakdown — Concrete Tasks (Week-Based Schedule)

All dates show target completion (end of week). Current date: Feb 10, 2026. Timeline: B1–B7 spans 7 weeks (Feb 10 – Mar 30).

### B1 — Beta Runtime Foundation (Week 1: Feb 10–16)

**Task B1.1:** Add PHASE config and phase gates to codebase
- **Owner:** Backend  
- **Due:** Feb 14  
- **Description:** Add `PHASE` environment variable (default "alpha"; accepts "alpha", "beta", "release"). Implement phase gates throughout code to gate Beta/Release features.  
- **Acceptance:** `PHASE=beta` compiles without errors; all phase-gated features are disabled unless Phase >= Beta.

**Task B1.2:** Extend season model to support 28-day Beta seasons
- **Owner:** Backend + DB  
- **Due:** Feb 14  
- **Description:** Extend schema `seasons` table with season_length_days (default 14 for Alpha; configurable to 28 for Beta). Update time system to respect season_length_days.  
- **Acceptance:** Schema migrates cleanly; season tick loop respects configurable length.

**Task B1.3:** Implement multi-season scheduler (Phase 2.5)
- **Owner:** Backend  
- **Due:** Feb 16  
- **Description:** Implement basic scheduler that manages multiple active seasons in Beta. Seasons start staggered (e.g., day 1, day 15). Ensure tick loop can drive multiple seasons independently.  
- **Acceptance:** Tick loop runs correctly for 2+ concurrent seasons; admin can list all active seasons.

**Task B1.4:** Add season lobby listing API
- **Owner:** Backend + Frontend  
- **Due:** Feb 16  
- **Description:** Add `/seasons` endpoint returning all active seasons with metadata (start_date, end_date, current_day, player_count). Add frontend UI to display season lobby.  
- **Acceptance:** GET /seasons returns array of active seasons; UI shows season timelines and player counts.

**Task B1.5:** Expose phase config in API and SSE
- **Owner:** Backend + Frontend  
- **Due:** Feb 16  
- **Description:** Include PHASE value in /config response and SSE startup payload so frontend can conditionally render Beta/Release UI.  
- **Acceptance:** Frontend can read PHASE from config and conditionally show/hide Beta features.

---

### B2 — Persistent Score + Profile Shell (Week 2: Feb 17–23)

**Task B2.1:** Implement star conversion at season end
- **Owner:** Backend + DB  
- **Due:** Feb 19  
- **Description:** At season end (tick boundary), convert player stars to permanent_score entry in profile_stats. Store season_id and final_star_count. Never allow mutation.  
- **Acceptance:** Season ends; player stars are locked into permanent_score; UI shows new score in profile.

**Task B2.2:** Add profile_stats table
- **Owner:** DB  
- **Due:** Feb 19  
- **Description:** Create profile_stats table (player_id, total_permanent_score, season_count, last_season_id, badges_earned). Add non-spendable permanent_score column.  
- **Acceptance:** Table exists; migration runs cleanly; can query permanent scores per player.

**Task B2.3:** Implement profile summary API endpoint
- **Owner:** Backend  
- **Due:** Feb 21  
- **Description:** Add GET /players/{player_id}/profile summarizing permanent_score, season_count, last_season_id, and list of best seasons (top 5 by stars).  
- **Acceptance:** Endpoint returns profile summary with all required fields; endpoint is fast (<100ms).

**Task B2.4:** Add profile view UI (shell)
- **Owner:** Frontend  
- **Due:** Feb 23  
- **Description:** Add profile screen showing permanent_score, season history shell placeholder, and cosmetics section (empty for now). Clickable from player card or leaderboard.  
- **Acceptance:** Profile screen loads; shows permanent_score and season count; layout is responsive.

**Task B2.5:** Emit profile update notifications
- **Owner:** Backend  
- **Due:** Feb 23  
- **Description:** When season ends, emit notification to player: "Season ended! You earned +X Stars. Total permanent score: Y." Include link to profile.  
- **Acceptance:** Notification system captures permanent_score event; client receives and displays notification.

---

### B3 — Meta Currency (Cosmetic Only) (Week 3: Feb 24 – Mar 2)

**Task B3.1:** Design meta currency rules doc
- **Owner:** Design  
- **Due:** Feb 25  
- **Description:** Write rules: meta currency is seasonal-persistent, non-tradable, non-convertible to Coins/Stars, used only for cosmetics. Cannot affect economy. Publish in canon.  
- **Acceptance:** Rules doc published in README/; no enforcement gaps identified in review.

**Task B3.2:** Add meta_currency_wallet table
- **Owner:** DB  
- **Due:** Feb 25  
- **Description:** Create meta_currency_wallet (player_id, season_id, balance) and meta_currency_grant_log (player_id, season_id, grant_amount, grant_reason, created_at, admin_note).  
- **Acceptance:** Tables exist; migration runs cleanly; can query balance and log per player/season.

**Task B3.3:** Implement meta currency grant logic
- **Owner:** Backend  
- **Due:** Feb 27  
- **Description:** Add internal grants system: season-end reward grants, event grants, return-player bonuses. All grants logged immutably. Enforce: no conversion paths, no spending (cosmetic-only placeholder).  
- **Acceptance:** Grant system persists to DB; cannot grant Coins/Stars; grants are read-only to player (cannot spend).

**Task B3.4:** Add meta currency to API and SSE
- **Owner:** Backend  
- **Due:** Feb 28  
- **Description:** Include meta_currency_balance in /seasons response and SSE payloads. Add GET /players/{player_id}/meta-currency. Clearly label as "non-tradable cosmetic currency."  
- **Acceptance:** API returns meta currency balance; frontend displays balance clearly labeled as cosmetic-only.

**Task B3.5:** Add cosmetics placeholder to frontend profile
- **Owner:** Frontend  
- **Due:** Mar 2  
- **Description:** Display meta currency balance on profile; add placeholder "Cosmetics Shop" button (links to empty shop for now). Clarify in UI: "Meta currency is cosmetic-only and cannot be traded or spent on game power."  
- **Acceptance:** Profile shows meta currency; cosmetics shop placeholder is visible; warning text is clear.

---

### B4 — Brokered Trading (Coins ↔ Stars) (Week 4: Mar 3–9)

**Task B4.1:** Design brokered trading rules and eligibility gates
- **Owner:** Design + Backend  
- **Due:** Mar 4  
- **Description:** Document eligibility gates (active players, recent spend, star/coin ratios, inflation exposure), tightening curve over season, burn rates, and price premiums. Publish in canon (README/trading or SPEC.md).  
- **Acceptance:** Rules doc complete; no edge cases left unspecified; implementation checklist prepared.

**Task B4.2:** Add brokered_trades table and event logging
- **Owner:** DB  
- **Due:** Mar 4  
- **Description:** Create brokered_trades (trade_id, seller_id, buyer_id, season_id, coin_amount, star_amount, burn_amount, premium_snapshot, created_at) and brokered_trade_events (event_id, trade_id, event_type, snapshot_data, created_at).  
- **Acceptance:** Tables exist and are append-only; migration runs cleanly.

**Task B4.3:** Implement brokered trade engine (pricing + eligibility)
- **Owner:** Backend  
- **Due:** Mar 6  
- **Description:** Implement trade eligibility check (active players, spend history, ratio bands, inflation exposure). Compute trade premium and burn rate based on season day. Price Coins ↔ Stars using market pressure + premium + burn.  
- **Acceptance:** Engine computes correct prices; eligibility gates work; burn amounts are logged; tightening curve is correct.

**Task B4.4:** Add brokered trade endpoints
- **Owner:** Backend  
- **Due:** Mar 7  
- **Description:** Add POST /trades/preview (check eligibility + show quote), POST /trades/execute (execute if eligible), GET /trades/history (list past trades). All endpoints return rich feedback (reasons for ineligibility, burn breakdown, premium explanation).  
- **Acceptance:** Endpoints return correct data; ineligible trades are rejected with clear reasons; eligible trades execute atomically.

**Task B4.5:** Add brokered trading desk UI
- **Owner:** Frontend  
- **Due:** Mar 9  
- **Description:** Add trading desk screen showing Coins ↔ Stars desk. Display eligibility status (green/yellow/red with reasons), preview quote with burn breakdown, premium explanation, and warnings ("Coins are burned"; "You may not be eligible later in season").  
- **Acceptance:** UI shows eligibility feedback; preview shows burn and premium clearly; confirmations have explicit warnings.

---

### B5 — TSAs (Competitive Assets) (Week 5: Mar 10–16)

**Task B5.1:** Design TSA mint and inventory rules
- **Owner:** Design + Backend  
- **Due:** Mar 11  
- **Description:** Document TSA minting via Star Sacrifice (irreversible, immediate rank drop, logged). Define TSA utility and market pressure contribution. Publish in canon.  
- **Acceptance:** Rules doc complete; no minting loopholes identified; rank drop calculation verified.

**Task B5.2:** Add TSA inventory tables and logs
- **Owner:** DB  
- **Due:** Mar 11  
- **Description:** Create tsa_inventory (player_id, season_id, tsa_type, quantity), tsa_mint_log (mint_id, player_id, season_id, tsa_type, quantity, stars_destroyed, active_players_snapshot, created_at), and tsa_event_log for tracking all TSA events.  
- **Acceptance:** Tables exist; migration runs cleanly; can query inventory per player/season.

**Task B5.3:** Implement Star Sacrifice → TSA minting
- **Owner:** Backend  
- **Due:** Mar 13  
- **Description:** Implement minting flow: player sacrifices X Stars, receives Y TSAs (formula: stars_destroyed = TSA_type_cost), rank drops proportionally (stars -= X), mint is logged immutably, market pressure increases. Ensure no refunds.  
- **Acceptance:** Mint succeeds; stars are deducted and logged as destroyed; TSAs appear in inventory; rank drops correctly; pressure increases.

**Task B5.4:** Add TSA inventory endpoint and SSE fields
- **Owner:** Backend  
- **Due:** Mar 14  
- **Description:** Add GET /players/{player_id}/tsa-inventory and include tsa_inventory in /seasons response. Expose TSA types and quantities in SSE on economy updates.  
- **Acceptance:** API returns TSA inventory; SSE includes TSA data; frontend can display inventory.

**Task B5.5:** Add TSA mint flow and inventory to frontend
- **Owner:** Frontend  
- **Due:** Mar 16  
- **Description:** Add TSA Sacrifice UI (confirm stars destruction, show resulting TSA quantity, show rank drop). Display inventory in profile or dedicated TSA screen. Warn: "Sacrifice is irreversible; your rank will drop immediately."  
- **Acceptance:** Mint flow works end-to-end; inventory displays correctly; warnings are clear and prominent.

---

### B6 — TSA Trading + Anti-Abuse Expansion (Week 6: Mar 17–23)

**Task B6.1:** Implement player-to-player TSA trading
- **Owner:** Backend + DB  
- **Due:** Mar 18  
- **Description:** Add trade_offers table (offer_id, seller_id, buyer_id, offer_items, consideration, status, created_at, expires_at). Implement offer negotiation endpoints (POST /trades/offer, PUT /trades/offer/{offer_id}/accept, DELETE /trades/offer/{offer_id}/reject). Enforce atomicity and logging.  
- **Acceptance:** Offers can be created, accepted, rejected; trades are atomic; all trades are logged immutably.

**Task B6.2:** Add trade abuse detection (reciprocal trades, IP clustering)
- **Owner:** Backend  
- **Due:** Mar 19  
- **Description:** Implement abuse signals for TSA trades: reciprocal trades (A sells to B who immediately sells back), IP clustering (multiple accounts on same IP trading heavily), suspicious trade ratios. Log signals to abuse_events; increase throttle/pressure on suspicious accounts.  
- **Acceptance:** Signals fire correctly; throttling applies; admin can see flags in observability.

**Task B6.3:** Add trade endpoints and notifications
- **Owner:** Backend  
- **Due:** Mar 21  
- **Description:** Add GET /trades/offers/received, GET /trades/offers/sent, GET /trades/history. Emit notifications: "New trade offer received", "Offer accepted/rejected", "Trade completed". Include offer details and links.  
- **Acceptance:** Endpoints return correct data; notifications fire on all transition; client displays notifications with links.

**Task B6.4:** Add player-to-player trading UI
- **Owner:** Frontend  
- **Due:** Mar 23  
- **Description:** Add trading interface: view received/sent offers, create new offer, accept/reject/counter. Display trade history. Show warnings for any detected abuse flags (if visible to player).  
- **Acceptance:** Trading UI is functional; offers can be negotiated; history is visible; warnings display (if applicable).

---

### B7 — Beta UX Polish (Week 7: Mar 24–30)

**Task B7.1:** Audit and finalize all notifications
- **Owner:** Backend + Frontend  
- **Due:** Mar 26  
- **Description:** Audit all notification types (trades, TSA events, season lobby, permanent score). Ensure clear copy, links to relevant screens, and consistent presentation. Remove any duplicate or unclear notifications.  
- **Acceptance:** All notifications are canonical; no duplicates or unclear messaging; all link to correct screens.

**Task B7.2:** Upgrade economy dashboard with Beta metrics
- **Owner:** Frontend  
- **Due:** Mar 28  
- **Description:** Add brokered trade metrics to economy dashboard (trade volume, burn rate, premium trend). Add TSA metrics (mint volume, trading volume). Update copy to explain new systems.  
- **Acceptance:** Dashboard shows all Beta metrics; explanatory text is clear; layout is responsive.

**Task B7.3:** Cross-reference canon and code for alignment
- **Owner:** Design + Backend  
- **Due:** Mar 29  
- **Description:** Audit README canon against B1–B6 implementations. Update README sections for persistent-state.md, SPEC.md, and README.md if any drift identified. Document all changes.  
- **Acceptance:** Canon and code aligned; all divergences documented; no hidden assumptions remain.

---

## Canon Alignment (Task-to-README Mapping)

This section maps each build task to the canonical requirements it implements. All tasks must preserve the invariants in README.md.

### Invariants (Immutable, Non-Negotiable)
- **README.md, Layer 1:** No currency may ever convert into Coins or Stars, directly or indirectly.
- **README.md, Layer 1:** Stars are the only direct leaderboard unit.
- **README.md, Layer 1:** Past seasons are immutable and never rewritten.
- **README.md, Layer 1:** Economy logic is server-authoritative.

### Build → Canon Mapping

| Task | README Section | Canon Requirement | Invariant Impact |
|------|------|------|------|
| B1.1 | SPEC.md § Seasons | Phase-bound seasons with staggered/overlapping runtime (Beta: 28 days, 2–3 seasons) | None (phase gating only) |
| B1.2 | SPEC.md § Seasons | Season length configurable; Alpha 14 days, Beta 28 days | None (phased config) |
| B1.3 | Phase 2.5 | Multi-season scheduler with independent tick loops | None (extends Phase 2) |
| B1.4 | SPEC.md § Join rules | Players join active seasons anytime; lobby lists all active seasons | None (informational) |
| B1.5 | SPEC.md § Spec intro | Phase config exposed to client for feature gating | None (frontend gating) |
| B2.1 | Phase 12.2a | At season end, Stars convert to permanent, non-spendable profile statistic | **Invariant**: Stars remain non-tradable and immutable once converted |
| B2.2 | persistent-state.md § profile_stats | Persistent profile entities (total_permanent_score, season_count, last_season_id) | None (new tables) |
| B2.3 | README.md § Between Seasons | Players observe season history and permanent progression through profile | None (read-only) |
| B2.4 | README.md § UI Philosophy | Profile screen shows permanent identity without revealing internal economy formulas | None (UI only) |
| B2.5 | notifications.md | Emit permanent score milestone notifications | None (audit trail) |
| B3.1 | persistent-state.md § Currency model | Persistent meta currency (cosmetic/identity only; non-tradable; non-convertible) | **Invariant**: Meta currency cannot convert into Coins/Stars |
| B3.2 | persistent-state.md § persistent entities | Add meta_currency_wallet and grant_log tables | None (new tables) |
| B3.3 | phase 12 (post-alpha) | Reward grants (badges, titles, cosmetic bonuses) | **Invariant**: Grants cannot mint Coins/Stars; cosmetic-only |
| B3.4 | README.md § System Authority | Meta currency exposed in API; clearly labeled non-economic | None (informational) |
| B3.5 | README.md § UI Philosophy | Cosmetics shop visible but non-functional in Beta; clear labeling on currency purpose | None (UI placeholder) |
| B4.1 | SPEC.md § Trading | Brokered Coins ↔ Stars trading (system-priced, premium, burn, eligibility gates) | **Invariant**: Brokered trades never create Coins or Stars; preserve liquidity |
| B4.2 | persistent-state.md § Trades log | Append-only brokered_trades and brokered_trade_events tables | None (audit trail) |
| B4.3 | SPEC.md § Trading Rules | Eligibility gates, tightening curve, burn mechanics, premium scaling | **Invariant**: Trades always contribute to market pressure (never relieve) |
| B4.4 | http-api-reference.md (future) | REST endpoints for trade preview, execute, history | None (API surface) |
| B4.5 | README.md § UI Philosophy | Trading desk with clear warnings (burns, ineligibility, premium) | None (UI only) |
| B5.1 | SPEC.md § Hard TSA invariants | TSAs are seasonal, system-minted only, never convert into Coins/Stars | **Invariant**: Stars sacrificed are permanently destroyed; TSA supply is observable |
| B5.2 | persistent-state.md § TSA Mint Log | Append-only tsa_inventory and tsa_mint_log tables | None (audit trail) |
| B5.3 | phase 12 (post-alpha) | Star Sacrifice → TSA minting (irreversible, rank drop, logged) | **Invariant**: Stars destroyed; no refunds; minting is immutable |
| B5.4 | SPEC.md § TSA trading (Beta-only) | TSA inventory exposed in API/SSE | None (informational) |
| B5.5 | README.md § UI Philosophy | Mint flow with clear warnings (irreversible, rank drop) | None (UI only, but warnings are mandatory) |
| B6.1 | SPEC.md § TSA trading (Beta-only) | Player-to-player TSA trading (negotiated; server enforces legality, caps, logging) | **Invariant**: Trades are immutable; supply is preserved; no Coin/Star generation |
| B6.2 | anti-abuse.md | New abuse signals for trading (reciprocal trades, IP clustering, suspicious ratios) | None (detection only) |
| B6.3 | notifications.md | Trade offer and completion notifications | None (audit trail) |
| B6.4 | README.md § UI Philosophy | Trading interface with abuse warnings (if visible to player) | None (UI only) |
| B7.1 | notifications.md | Audit all notification types for clarity and correctness | None (quality assurance) |
| B7.2 | README.md § Economy Dashboard | Add Beta metrics (trade volume, burn, premium, TSA activity) | None (informational) |
| B7.3 | README.md § Game Bible | Cross-audit canon and code; document and resolve all drift | None (process) |
| B7.4 | README.md § Bible Requirements | E2E test suite for all B1–B7 features and invariants | None (validation) |

---

## Acceptance Tests (Per Build)

Each build defines explicit acceptance criteria. This section details comprehensive test coverage.

### B1 — Beta Runtime Foundation

**Test B1.1: PHASE config gates Beta features**
- When `PHASE=alpha`, all Beta routes return 404 or error.
- When `PHASE=beta`, Beta routes respond normally.
- When `PHASE=release`, all routes respond normally.
- **Expected:** Feature gating works correctly; no cross-contamination.

**Test B1.2: Season length respects config**
- Alpha season runs 14 days; Beta season runs 28 days.
- Advancing time confirms duration is respected.
- Season end fires at correct tick.
- **Expected:** Season length is configurable and respected; no early/late ends.

**Test B1.3: Multi-season scheduler runs independent tick loops**
- Create season A (day 1) and season B (day 15, overlapping).
- Advance time 1 tick; both seasons tick independently.
- Economy state diverges correctly per season.
- **Expected:** Ticks are independent; economy state is per-season; no leakage.

**Test B1.4: Season lobby lists all active seasons**
- Create 2 active seasons and 1 ended season.
- GET /seasons returns 2 seasons (excludes ended).
- Metadata includes start_date, end_date, current_day, player_count.
- **Expected:** Lobby is correct and complete; ended seasons excluded.

**Test B1.5: Frontend reads PHASE config**
- Frontend requests /config; receives `{"phase":"beta"}`.
- Frontend conditionally renders Beta-only UI (e.g., TSA menu).
- Compare rendered state to phase value.
- **Expected:** UI matches phase; no feature leakage.

---

### B2 — Persistent Score + Profile Shell

**Test B2.1: Stars convert to permanent score at season end**
- Player earns 100 stars mid-season.
- Season ends (tick boundary).
- permanent_score entry created with season_id and final_star_count.
- Player's profile shows permanent_score (100).
- **Expected:** Conversion is atomic and immutable; no star loss.

**Test B2.2: Permanent score is read-only**
- Player attempts to spend permanent_score (via API manipulation).
- Request rejected; permanent_score unchanged.
- Only through Star Sacrifice (B5) can permanent_score be affected indirectly (rank drop).
- **Expected:** Permanent score is immutable; no direct edits possible.

**Test B2.3: Profile endpoint returns all required fields**
- GET /players/{player_id}/profile returns permanent_score, season_count, last_season_id, top_5_seasons.
- Response includes all fields; no missing data.
- Endpoint latency < 100ms.
- **Expected:** Profile data is complete and fast.

**Test B2.4: Profile UI renders without errors**
- Load profile page for player with permanent_score.
- All fields render correctly; layout is responsive (mobile, tablet, desktop).
- Clicking season in history does not error.
- **Expected:** UI is robust and accessible.

**Test B2.5: Permanent score notification fires**
- Season ends; player receives notification within 1 second.
- Notification includes "+X Stars earned" and "Total permanent score: Y".
- Clicking notification navigates to profile.
- **Expected:** Notification is timely and actionable.

---

### B3 — Meta Currency (Cosmetic Only)

**Test B3.1: Meta currency cannot convert to Coins or Stars**
- Player has 100 meta currency.
- Attempts to trade metal currency for Coins (API endpoint manipulation).
- Request rejected; balance unchanged.
- Attempt to spend meta currency on star purchase.
- Request rejected.
- **Expected:** No conversion paths exist; invariant is enforced.

**Test B3.2: Meta currency grant is logged immutably**
- Admin grants 50 meta currency to player.
- grant_log entry created (player_id, season_id, 50, "admin_grant", timestamp, admin_note).
- Player cannot delete or modify log entry.
- **Expected:** Log is append-only; cannot be tampered with.

**Test B3.3: Cosmetics shop is non-functional in Beta**
- Player navigates to cosmetics shop.
- Shop displays but all purchase buttons are disabled or hidden.
- UI explains: "Cosmetics coming soon."
- **Expected:** Shop is placeholder; no unintended purchases.

**Test B3.4: Meta currency appears in API with correct label**
- GET /seasons returns `{"meta_currency_balance":50,"meta_currency_label":"Cosmetic Currency (Non-Tradable)"}`.
- Frontend displays label clearly.
- **Expected:** Frontend accurately represents currency purpose.

---

### B4 — Brokered Trading

**Test B4.1: Eligibility gates work correctly**
- Player A (recent spender, active) can trade.
- Player B (new player, no spend history) cannot trade.
- Player C (throttled due to abuse) cannot trade.
- **Expected:** Gates are enforced; ineligibility is clear.

**Test B4.2: Trade preview shows correct quote and burn**
- Player requests trade preview: sell 10 stars for coins.
- Response includes: coins_offered, coin_burned, premium_percentage, reason_for_ineligibility (if applicable).
- Burn amount is > 0 and < coins_offered.
- **Expected:** Quote is transparent and accurate.

**Test B4.3: Tightening curve works over season**
- Mid-season (day 7/28): trade burn 10%, premium 5%.
- Late season (day 25/28): trade burn 40%, premium 25%.
- Verify burn% and premium% increase monotonically.
- **Expected:** Curve is progressive and predictable.

**Test B4.4: Trade execution is atomic**
- Player executes trade (10 stars for 500 coins, 50 burned, 475 received).
- Transaction fails midway (e.g., DB error).
- State rollback verified; player has original 10 stars, trade_id not created.
- Retry succeeds; trade_id unique, state updated correctly.
- **Expected:** Trades are all-or-nothing; no partial states.

**Test B4.5: Brokered testing contributes to market pressure**
- Before trade: market_pressure = 50.
- Player executes trade; pressure increases to 51 (+1 for trade contribution).
- Multiple trades: pressure increases monotonically.
- **Expected:** Pressure always increases after trades; never decreases.

**Test B4.6: Trading desk UI shows eligibility feedback**
- Eligible player sees: green checkmark; quote shown.
- Ineligible player sees: red X; reason for ineligibility (e.g., "Insufficient spend history").
- Late-season player sees: yellow warning; "Burn rate is high; confirm?".
- **Expected:** Feedback is clear and actionable.

---

### B5 — TSAs (Competitive Assets)

**Test B5.1: Star Sacrifice → TSA minting works**
- Player with 100 stars sacrifices 20 stars.
- 20 stars deducted from star_balance.
- 20 stars logged as destroyed in tsa_mint_log.
- TSA inventory updated (+20 TSAs or appropriate quantity based on 1:1 or other formula).
- **Expected:** Sacrifice is irreversible; no refunds; TSA created.

**Test B5.2: Rank drops after sacrifice**
- Player rank (leaderboard position) before: 10.
- Sacrifice stars; rank after: > 10 (lower rank = further down leaderboard).
- Rank drop magnitude = stars_sacrificed / total_season_stars (approximately).
- **Expected:** Rank drop is proportional and immediate.

**Test B5.3: TSA inventory endpoint returns correct data**
- GET /players/{player_id}/tsa-inventory returns `{"TSA_TYPE_A": 20}`.
- SSE on economy tick includes TSA inventory state.
- Frontend can display inventory without errors.
- **Expected:** Inventory is readable and complete.

**Test B5.4: Mint UI shows warnings clearly**
- Player clicks "Sacrifice Stars for TSAs".
- Modal warns: "Sacrifice is IRREVERSIBLE. Your rank will drop immediately. Do you want to continue?"
- After confirmation, sacrifice completes; warning acknowledged.
- **Expected:** Player makes informed decision; no accidental sacrifices.

**Test B5.5: Season end wipes TSA inventory**
- Player ends season with 20 TSAs.
- Season transitions to ended.
- TSA inventory is zeroed or archived (not carried to next season).
- Next season start: player has 0 TSAs.
- **Expected:** TSAs do not carry between seasons; each season starts fresh.

---

### B6 — TSA Trading + Anti-Abuse

**Test B6.1: Player-to-player TSA trade offer created**
- Player A offers: 5 TSAs to Player B.
- Player B offers: 100 coins back.
- Trade offer created (offer_id, seller_A, buyer_B, items, consideration, status=pending).
- Both players notified.
- **Expected:** Offer is recorded; notifications sent.

**Test B6.2: Trade offer accepted and executed**
- Player B receives offer; views details.
- Clicks "Accept".
- Trade executes atomically: A loses 5 TSAs, B gains 5 TSAs; B loses 100 coins, A gains 100 coins.
- Both players notified: "Trade completed."
- trade_offers.status = accepted; trade_id created in tsa_trade_log.
- **Expected:** Trade is atomic; notifications reflect final state.

**Test B6.3: Reciprocal trade detection works**
- Player A sells 5 TSAs to B for 100 coins.
- Player B immediately offers to sell 5 TSAs back to A for 100 coins (or similar).
- Abuse signal fired for "reciprocal_trade".
- Both players' abuse_events incremented; throttle applied if threshold met.
- **Expected:** Pattern detected; dampening applied.

**Test B6.4: IP clustering detection works**
- 3 accounts on same IP trade heavily (e.g., 10+ trades/hour).
- Abuse signal fired for "ip_clustering_trading".
- Accounts are throttled; trade costs increase or limits tighten.
- **Expected:** Abuse is detected; dampening applied.

**Test B6.5: Trade history endpoint returns complete logs**
- GET /trades/history returns all trades for player (sent offers, received offers, completed trades).
- Data includes timestamps, counterparty IDs, items, and status.
- Admin can query /admin/trades/history?player_id=X for audit.
- **Expected:** History is complete and queryable; admin visibility intact.

---

### B7 — Beta UX Polish

**Test B7.1: All notification types are present and clear**
- Trigger each notification type: trade offer, trade completion, TSA mint, permanent score milestone, season end.
- Verify each notification has clear copy, relevant link, and timestamp.
- No duplicate notifications in same second.
- **Expected:** Notifications are canonical and non-duplicative.

**Test B7.2: Economy dashboard includes Beta metrics**
- Load economy dashboard.
- Verify display of: trade volume (today, week), burn rate (trend), premium (current, trend), TSA minting volume.
- All metrics update in real-time (SSE).
- Explanatory tooltips explain each metric.
- **Expected:** Dashboard is informative and up-to-date.

**Test B7.3: Canon and code aligned (no drift)**
- Audit README sections (SPEC.md, persistent-state.md, SPEC.md) vs. B1–B6 implementations.
- Document any discrepancies.
- Update README or code to resolve drift.
- Final verification: canon and code agree byte-for-byte on core rules (no cherry-picking).
- **Expected:** Drift is zero; future audits will be faster.

**Test B7.4: E2E test suite passes 100%**
- Run full test suite covering B1–B7.
- All tests pass.
- Coverage includes happy path, edge cases, error cases, and anti-patterns.
- Suite completes in < 5 minutes.
- **Expected:** Beta is production-ready; regression tests are fast and reliable.

---

---

## Phase 3 — Economy Emission & Pools
- [x] [DONE] 3.1 Emission pool and daily budget
- [x] [DONE] 3.2 Emission time‑sliced per tick
- [x] [DONE] 3.3 Emission throttling via pool availability
- [x] [DONE] 3.3a Align emission curve to runtime season length (Alpha 14 days / extension-aware)
- [x] [DONE] 3.4 Validate emission pacing vs coin‑emission.md (daily budget, smooth throttle, no abrupt stops)
- [x] [DONE] 3.5 Validate emission floor (prevents pool starvation while respecting scarcity)

---

## Phase 4 — Faucets & Daily Earnings
- [x] [DONE] 4.0 Universal Basic Income (UBI) Implementation
  - [x] [DONE] 4.0a Implement minimum 0.001 coin per tick payout to all active players
  - [x] [DONE] 4.0b Verify UBI draws from emission pool and respects pool throttling
  - [x] [DONE] 4.0c Confirm UBI is foundation (all other faucets are additive)
  - [x] [DONE] 4.0d Document UBI in coin-faucets.md if not already present
  - [x] [DONE] 4.0e Ensure star pricing tuning accounts for UBI + inflation interaction
- [x] [DONE] 4.1 Daily login faucet (cooldown + log + emission cap)
- [x] [DONE] 4.2 Activity faucet (cooldown + log + emission cap)
- [x] [DONE] 4.3 Per‑player daily earn cap with seasonal decay
- [x] [DONE] 4.4 Append‑only coin earning log (source_type + amount)
- [x] [DONE] 4.5 Login playability safeguard (alpha‑only, emission‑pool backed, short cooldown)
- [x] [DONE] 4.6 Verify login safeguard behavior (min balance target, cooldown, no daily‑cap dead‑locks)
- [x] [DONE] 4.7 Confirm faucet priorities and pool gating match canon (no player‑created coins)
- [x] [DONE] 4.8 Resolve passive drip status (enabled vs disabled for Alpha)
- [ ] [POST-ALPHA] 4.9 Daily tasks faucet
- [ ] [POST-ALPHA] 4.10 Comeback reward faucet

---

## Phase 5 — Pricing & Purchases
- [x] [DONE] 5.1 Server‑authoritative star pricing
  - [x] [DONE] Time pressure + late‑season spike
  - [x] [DONE] Quantity scaling for bulk purchases
  - [x] [DONE] Market pressure multiplier + caps
  - [x] [DONE] Affordability guardrail
  - [x] [DONE] Star price persistence (current_star_price in season_economy)
  - [x] [DONE] 5.1a Enforce season-authoritative (player-divergent-free) pricing — star price computed once per tick, shared identically across all players, uses only season-level inputs (no active player metrics)
- [x] [DONE] 5.2 Atomic star purchases (single + bulk)
- [x] [DONE] 5.2a Align pricing time progression to runtime season length (Alpha 14 days / extension-aware)
- [x] [DONE] 5.3 Validate pricing curves vs coin emission (affordability and late‑season scarcity)
- [x] [DONE] 5.4 Validate bulk purchase warnings and re‑check at confirmation
- [x] [DONE] 5.5 Price tick locking for star purchases (client price_tick must match server tick)

---

## Phase 6 — Market Pressure
- [x] [DONE] 6.1 Market pressure computed server‑side
- [x] [DONE] 6.2 Rate‑limited adjustments per tick
- [x] [DONE] 6.3 Validate pressure inputs vs canon (no trade inputs until trading exists)
- [x] [DONE] 6.4 Validate pressure appears in SSE + UI and is stable under bursts

---

## Phase 7 — Telemetry, Calibration, and Economic History (Data‑Driven Only)
- [x] [DONE] 7.1 Telemetry capture + admin telemetry endpoints
- [x] [DONE] 7.2 Season calibration persistence (season_calibration)
- [x] [DONE] 7.3 Ensure telemetry is sufficient to calibrate live values (emission, caps, price curves, pressure)
- [x] [DONE] 7.4 Define and verify telemetry events:
  - [x] [DONE] Faucet claims (daily, activity, passive if enabled, login safeguard)
  - [x] [DONE] Star purchase attempts + successes
  - [x] [DONE] Emission pool levels and per‑tick emissions
  - [x] [DONE] Market pressure value per tick
- [x] [DONE] 7.5 Validate append‑only economic logs are complete and queryable
- [x] [DONE] 7.6 Establish calibration workflow using telemetry history (no blind tuning)
- [x] [DONE] 7.7 Reconcile telemetry taxonomy with alpha‑execution.md (current client emits login + buy_star; join_season is not emitted)

---

## Phase 8 — Anti‑Abuse, Trust, and Access Control (Soft Enforcement Philosophy)

### Anti-Cheat Philosophy: Sentinels, Not Punishers

Anti-cheat is **gradual, invisible, and corrective**, not punitive.

**What anti-cheat NEVER does:**
- Bans automatically
- Suspends accounts automatically
- Zeroes wallets
- Hard-blocks players
- Exposes enforcement actions publicly

**What anti-cheat DOES:**
- Gradually reduces earning rates for suspicious behavior
- Increases star prices for suspicious accounts
- Adds cooldowns and jitter to sensitive actions
- Throttles activity without blocking it

**Goal: Make abuse economically ineffective, not publicly punishing.**

### Implementation Status

- [x] [DONE] 8.1 IP capture + association tracking
- [x] [DONE] 8.2 Enforce one active player per IP per season (soft enforcement: throttles, not hard blocks)
- [x] [DONE] 8.3 Soft IP dampening (delay + multipliers)
- [x] [DONE] 8.4 Rate limiting for signup/login
- [x] [DONE] 8.5 AbuseEvents table + signal aggregation
- [x] [DONE] 8.6 Audit anti‑abuse coverage post‑whitelist removal
- [x] [DONE] 8.7 Update anti‑abuse docs to match Alpha (CAPTCHA/verification is post‑alpha)
- [ ] [ALPHA REQUIRED] 8.8 Verify soft enforcement scaling: minor suspicious activity → minor throttles; extreme abuse → heavy dampening
- [ ] [ALPHA REQUIRED] 8.9 Confirm admin ban capability exists ONLY for extreme cases flagged by anti-cheat
- [ ] [POST-ALPHA] 8.10 CAPTCHA + verification
- [ ] [POST-ALPHA] 8.11 Additional abuse signals + admin visualization improvements
- [ ] [POST-ALPHA] 8.12 Trade-specific abuse detection (reciprocal trades, IP clustering, volume spikes)

---

## Phase 9 — Admin & Observability (Read‑Only in Alpha; Sentinels, Not Gods)

### Admin Role Philosophy: Sentinels, Not Gods

**Admins are sentinels, not gods.** The economy must self-regulate. Admins provide oversight and emergency safeguards, not active management.

**What admins MAY do:**
- Emergency pause 1 or all seasons (temporary freeze)
- Ban extreme abuse cases (only after anti-cheat recommendation)
- Monitor telemetry and economy health (read-only observability)
- Advance seasons manually (recovery only, not normal flow)

**What admins MUST NOT do:**
- Micromanage the economy
- Manually adjust player balances
- Override anti-cheat without justification
- Edit past season data
- Interfere with normal economic flow

**The economy is designed to self-regulate. Admin intervention must be rare, deliberate, and auditable.**

### Implementation Status

- [x] [DONE] 9.1 Alpha admin bootstrap (ENV‑seeded, one‑shot, DB‑sealed)
- [x] [DONE] 9.2 Moderator role support
- [x] [DONE] 9.3 Economy monitoring endpoints (read-only)
- [x] [DONE] 9.4 Notifications system
- [x] [DONE] 9.5 Add notification observability logging
- [x] [DONE] 9.6 Update admin‑tools docs to reflect read‑only Alpha reality
- [ ] [ALPHA REQUIRED] 9.6a Verify notification emission, persistence, and client rendering end‑to‑end
- [ ] [ALPHA REQUIRED] 9.7 Verify admin UI clearly communicates "read-only" status in Alpha
- [ ] [ALPHA REQUIRED] 9.8 Confirm admin manual season advance exists for recovery (POST /admin/seasons/advance)
- [ ] [POST-ALPHA] 9.9 Admin safety tools (pause purchases, adjust emission, freeze season)
- [ ] [POST-ALPHA] 9.9a Admin-triggered broadcast notifications
- [ ] [POST-ALPHA] 9.9b Targeted notifications to individual players
- [ ] [POST-ALPHA] 9.10 Trading visibility (premium, burn rate, eligibility tightness, trade logs)
- [ ] [POST-ALPHA] 9.11 Enhanced player inspection (coin earning history, throttle status detail, IP clustering views)

Alpha admin bootstrap finalized: ENV‑seeded, one‑shot, DB‑sealed. No gate keys.

---

## Phase 10 — Frontend MVP (Alpha)
- [x] [DONE] 10.1 Landing + Auth + Main dashboard + Leaderboard
- [x] [DONE] 10.2 Bulk purchase UI + warnings
- [x] [DONE] 10.3 Admin console entry point
- [x] [DONE] 10.4 Verify UI shows required economy values (price, time, pressure, coins in circulation, next emission)
- [x] [DONE] 10.5 Label missing systems in UI (trading, multi‑season, cosmetics)
- [x] [DONE] 10.5a Season end UI consistency (Ended only; no buy/earn; frozen metrics)
- [ ] [ALPHA REQUIRED] 10.6 Update UI labels to include missing communication systems (direct messaging, forum) and rare currencies
- [ ] [POST-ALPHA] 10.7 Season lobby + brokered trading desk (Coins ↔ Stars)
- [ ] [POST-ALPHA] 10.8 Player profile + collections + settings/accessibility
- [ ] [POST-BETA] 10.9 Player-to-player trading desk (TSAs, rare currencies)
- [ ] [POST-BETA] 10.10 Direct messaging UI
- [ ] [POST-BETA] 10.11 Forum UI

---

## Phase 11 — Game Flow & Playability (Explicit and Tested)
- [x] [DONE] 11.1 Map new‑player journey mid‑season (signup → first earn → first star purchase)
- [x] [DONE] 11.2 Map late‑season joiner viability (can play, not necessarily compete)
- [x] [DONE] 11.3 Define always‑available actions vs tightening actions over time
- [x] [DONE] 11.4 Define daily loop steps (login → faucet(s) → purchase) and failure modes
- [x] [DONE] 11.5 Identify any broken/unclear step and add fix tasks before Alpha test
  - [x] [DONE] Add market pressure + next emission to /seasons response to prevent UI gaps before SSE connects

---

## Phase 12 — Season End, Star Conversion, and Between‑Season Progression
- [x] [DONE] 12.1 End‑of‑season snapshot and economy freeze
- [x] [DONE] 12.1a Expose a single terminal season state to clients (Ended only; "Ending" internal)
- [x] [DONE] 12.1b Season lifecycle integrity: Alpha length guardrails + ended invariants + final snapshot fields
- [ ] [ALPHA REQUIRED] 12.2 Star Conversion to Permanent Profile Statistic
  - [ ] [ALPHA REQUIRED] 12.2a At season end, Stars convert to permanent, non-spendable profile statistic (NOT a currency)
  - [ ] [ALPHA REQUIRED] 12.2b Stars are NOT tradable (before or after conversion)
  - [ ] [ALPHA REQUIRED] 12.2c Stars are NOT spendable (except via Star Sacrifice for TSAs during season in Beta+)
  - [ ] [ALPHA REQUIRED] 12.2d Star value scales with season population (larger/more competitive seasons carry more weight)
  - [ ] [ALPHA REQUIRED] 12.2e Document: "Stars are the permanent score of seasonal performance"
- [ ] [POST-ALPHA] 12.3 Reward granting (badges + titles + recognition)
- [ ] [POST-ALPHA] 12.4 Persistent meta currency grant (cosmetic/identity only; never tradable, never converts to Coins/Stars)
- [ ] [POST-ALPHA] 12.5 Persistent progression + season history + return incentives
- [ ] [POST-ALPHA] 12.6 TSA wipe at season end (all holdings and pending trades; TSAs never carry over)

---

## Post‑Alpha / Beta — Persistent Meta Currency (Canon Only)
- [ ] [POST-ALPHA] Introduce persistent meta currency (Beta) for cosmetics/identity only; non‑tradable, non‑competitive, season‑persistent.
- [ ] [POST-ALPHA] Implement reward grant logic for persistent meta currency (non‑economic, cosmetic only).
- [ ] [POST-ALPHA] Expose persistent meta currency in UI (cosmetics, titles, badges, collections).
- [ ] [POST-ALPHA] Enforce and document: no conversion paths into Coins or Stars, direct or indirect.
- [ ] [POST-ALPHA] (Post‑Release optional) Influence/reputation metric: non‑spendable, eligibility/visibility‑only, never convertible.

## Post‑Alpha / Beta — Tradable Seasonal Assets (TSAs)

_TSAs are seasonal, player‑owned competitive assets (not currencies) introduced in Beta._

- [ ] [POST-ALPHA] Define TSA canon constraints in code (Beta‑only, seasonal competitive asset, system‑minted only, observable supply, no conversion into Coins/Stars, no minting Coins/Stars).
- [ ] [POST-ALPHA] Implement Star Sacrifice → TSA minting (Stars destroyed, immediate rank drop, irreversible).
- [ ] [POST-ALPHA] Implement player‑to‑player TSA trading (negotiated; server enforces legality, caps, and burn; logging).
- [ ] [POST-ALPHA] Add append‑only TSA logs (mint w/ stars_destroyed + source, trade w/ consideration + friction, activation) with admin visibility.
- [ ] [POST-ALPHA] Add TSA season‑end wipe behavior and snapshot/telemetry integration.
- [ ] [POST-ALPHA] Ensure TSA trading contributes to market pressure when enabled.

## Post‑Alpha / Beta — Brokered Trading (Coins ↔ Stars)

_Brokered trading is system‑priced, asymmetric, and costly. It is post‑alpha and currently disabled._

- [ ] [POST-ALPHA] Implement brokered Coins ↔ Stars trading (system-priced, with coin burn and time-based premium).
- [ ] [POST-ALPHA] Enforce eligibility gates:
  - Both players active and time-normalized
  - Both have recent coin spending activity (no pure hoarders)
  - Relative Star holdings within tightening ratio band
  - Coin liquidity within tightening band
  - Inflation exposure difference within tightening band
- [ ] [POST-ALPHA] Implement time-based tightening:
  - Eligibility gates tighten over time
  - Trade burn percentage rises
  - Maximum Stars per trade drops
  - Daily trade limits decrease
- [ ] [POST-ALPHA] Ensure trades always contribute to market pressure (never relieve it).
- [ ] [POST-ALPHA] Add brokered trade logging (burn amounts, eligibility deltas, market pressure contribution).
- [ ] [POST-ALPHA] Add admin visibility for brokered trading (premium, burn rate, eligibility tightness, trade logs).

## Post‑Beta — Rare Currencies (Canon Only)

_Rare currencies are special, limited‑drop currencies that enable pay‑to‑win features. Introduced in Beta (1–2 types), expanded in Release (3–5 types)._

- [ ] [POST-BETA] Define rare currency drop mechanics (random via normal gameplay; slightly influenced by anti-cheat).
- [ ] [POST-BETA] Implement rare currency storage (seasonal; reset at season end).
- [ ] [POST-BETA] Implement rare currency spending (stronger purchases, exclusive advantages; NEVER affects drop rates).
- [ ] [POST-BETA] Implement player‑to‑player rare currency trading (negotiated; server enforces legality and logging).
- [ ] [POST-BETA] Add append‑only rare currency logs (drops, spends, trades) with admin visibility.
- [ ] [POST-BETA] Document rare currency rarity scaling (rarer currency → stronger benefit).
- [ ] [POST-BETA] Ensure rare currencies feel "found, not farmed" (Diablo 2 rune market feel).
- [ ] [POST-BETA] Verify no conversion paths into Coins or Stars exist (direct or indirect).

## Post‑Beta — Player‑to‑Player Trading (TSAs + Rare Currencies)

_Player‑to‑player trading is negotiated and flexible. Introduced in Beta._

Valid trades:
- Coins ↔ Rare Currencies
- Rare Currencies ↔ Rare Currencies
- TSAs ↔ Coins
- TSAs ↔ Rare Currencies
- TSAs ↔ TSAs

Invalid trades:
- Stars (never tradable in any context)
- Meta currency (never tradable)
- Influence/reputation (never tradable)

- [ ] [POST-BETA] Implement player‑to‑player trading interface (negotiation, offers, confirmations).
- [ ] [POST-BETA] Enforce valid trade types and reject invalid combinations.
- [ ] [POST-BETA] Apply friction where appropriate (Coin burn, fees, caps).
- [ ] [POST-BETA] Log all player‑to‑player trades (feeds economic telemetry).
- [ ] [POST-BETA] Add admin visibility for player‑to‑player trades.
- [ ] [POST-BETA] Integrate trades with notifications (offers, confirmations, completions).

## Post‑Beta — Communication Systems (Direct Messaging + Forum)

_Communication systems support trading, social dynamics, and community engagement._

### Direct Messaging (Post-Alpha/Beta)

- [ ] [POST-ALPHA] Implement direct messaging between players.
- [ ] [POST-ALPHA] Integrate messaging with notifications (alerts for new messages).
- [ ] [POST-ALPHA] Integrate messaging with trades (offer negotiation, confirmation).
- [ ] [POST-ALPHA] Add messaging links from player profiles.
- [ ] [POST-ALPHA] Add messaging moderation tools (admin/moderator view, abuse reporting).

### Forum (Post-Beta)

- [ ] [POST-BETA] Implement game‑integrated forum.
- [ ] [POST-BETA] Add forum roles (admins → full moderation; moderators → limited moderation; players → standard posting).
- [ ] [POST-BETA] Support public trade negotiation, strategy discussion, alliances, and rivalries.
- [ ] [POST-BETA] Integrate forum with notifications (replies, mentions, trade offers).
- [ ] [POST-BETA] Integrate forum with player profiles (reputation, badges, contact links).
- [ ] [POST-BETA] Add forum moderation tools (post flags, bans, thread locks).

### Communication Philosophy

Communication systems must evoke:
- **Diablo 2 rune market**: Player-driven, social, economic
- **Competitive tension**: Rivalries, alliances, betrayals
- **Community identity**: Long-term relationships, reputation, history

## Bug Reporting & Feedback System (Alpha → Post‑Release)

_Bug reporting intake is always available from Alpha onward and persists after release._

- [x] [DONE] Implement in-game bug report intake UI (footer/help entry).
- [x] [DONE] Persist bug reports as append-only, immutable records (player_id optional, season_id, timestamp, client version if available).
- [x] [DONE] Add read-only admin visibility for bug reports (view only; no edit/delete/respond).
- [x] [DONE] Confirm Alpha has no attachments and no player feedback loop.
- [ ] [POST-ALPHA] Add admin/moderator bug triage interface.
- [ ] [POST-ALPHA] Integrate bug reports with notifications (alerts for admins/moderators).
- [ ] [POST-ALPHA] Support screenshots/logs attachment for bug reports.
- [ ] [POST-ALPHA] Track bug report status (open, in-progress, resolved, closed).
- [ ] [POST-ALPHA] Define admin/moderator response workflows (if ever; not in Alpha).

---

## Phase 13 — Testing & Validation
- [x] [DONE] 13.1 Simulation engine for pricing + pressure
- [ ] [ALPHA REQUIRED] 13.2 Validate simulation outputs vs live calibration parameters
- [ ] [ALPHA EXECUTION] 13.3 Alpha execution cycle
  - [x] [DONE] Define goals + metrics (README/alpha‑execution.md)
  - [ ] [ALPHA EXECUTION] Recruit testers
  - [ ] [ALPHA EXECUTION] Run 1–2 week test
  - [ ] [ALPHA EXECUTION] Analyze telemetry + prioritize fixes
- [ ] [POST-ALPHA] 13.4 Beta readiness (multi‑season + trading + expanded abuse controls)

---

## Phase 14 — Deployment & Live Ops
- [x] [DONE] 14.1 Fly.io deployment config + migrations
- [x] [DONE] 14.2 Basic monitoring + alerting
- [ ] [POST-ALPHA] 14.3 Backup + restore procedures

---

## Phase 15 — Documentation Alignment (Continuous)
- [x] [DONE] 15.1 Canon README set present
- [ ] [ALPHA REQUIRED] 15.2 Keep docs aligned with code reality after every change
- [x] [DONE] 15.3 Document safe patching & schema evolution rules
- [x] [DONE] 15.4 Formalize Codex governance constraints in docs
- [x] [DONE] 15.5 Document multi-season telemetry & historical integrity guarantees

---

## Final Verification Checklist (Maintain Continuously)
- [ ] [ALPHA REQUIRED] No task depends on a later task
- [ ] [ALPHA REQUIRED] All README requirements represented (or explicitly deferred)
- [ ] [ALPHA REQUIRED] Code‑existing features verified (not just implemented)
- [ ] [ALPHA REQUIRED] Path from today → Alpha → Post‑Alpha is continuous

---

## Phase 16 — Population‑Invariant Economy Validation
- [x] [DONE] Audit all population‑coupled inputs (coins in circulation, active coins, market pressure, affordability guardrails) and document effects at 1/5/500 players.
- [x] [DONE] Verify Universal Basic Income (UBI) provides stable minimum income at all population levels (0.001 coin/tick).
- [x] [DONE] Verify active circulation window (24h) produces stable emission at low population; adjust if faucet starvation occurs.
- [x] [DONE] Add telemetry for activeCoinsInCirculation + activePlayers and confirm visibility in admin telemetry.
- [x] [DONE] Run stress tests: solo season, small group (5‑10), large group (500+) using simulation + live tick metrics.
- [x] [DONE] Validate faucet pacing (UBI + daily/activity/passive if enabled) against emission pool under low‑population conditions.
- [x] [DONE] Confirm star pricing remains purchasable for a solo player over full season without trivializing scarcity.
- [x] [DONE] Document UBI + emission + pricing interaction in coin-emission.md and coin-faucets.md.

**Validation Summary (Feb 9, 2026):**
- Extended validation tests completed at 1, 5, and 500 player populations (see artifacts/validation/)
- UBI consistently delivers 1 microcoin (0.001 coins) per tick across all populations
- Emission pool remains healthy with 0% exhaustion rate across all scenarios
- Affordability: Solo (UBI-only) cannot progress (expected); 5-10 players: 100%; 500 players: 99.8%
- Telemetry enhanced to track activeCoinsInCirculation, activePlayers, and totalCoinsInCirculation
- All population-coupled inputs audited and functioning correctly
