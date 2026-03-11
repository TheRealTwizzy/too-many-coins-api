# Too Many Coins — API & Game Design Reference

**Too Many Coins** is a deterministic, season-based economic competition built around constrained scarcity, transparent rules, and strategic sacrifice. This repository contains the authoritative game bible and design specification for the server-side API.

---

## Game Overview

Players join time-bounded seasons, earn **Coins** through Universal Basic Income (UBI), convert those Coins into **Seasonal Stars**, and decide when to lock in their value or spend it for tactical advantage. The game's central tension is the trade-off between **Rank**, **Power**, and **Safety** — every season forces a choice.

The server is the single source of truth. All economy math, action ordering, rankings, and outcomes are authoritative at the server level. The client is never trusted for legality or state.

---

## Season Structure

- Seasons last **28 days** and new seasons begin every **7 days**, creating four overlapping active seasons at any time.
- A player may **spectate** any season, but may **economically participate** in only one active season at a time.
- Re-entry into a season is allowed but resets all season-bound resources (Coins, Seasonal Stars, Sigils) to zero.

### Season Lifecycle Phases

| Phase      | Description                                                                  |
|------------|------------------------------------------------------------------------------|
| Scheduled  | Season is upcoming; spectating allowed                                       |
| Active     | Players may join, earn Coins, buy Stars, trade, and Lock In                  |
| Blackout   | Final 72 hours before expiration; Lock-In is disabled, join/re-entry allowed |
| Expired    | Season has ended; finalization and award distribution complete                |

---

## Core Resources

| Resource         | Scope           | Description                                                                       |
|------------------|-----------------|-----------------------------------------------------------------------------------|
| **Coins**        | Season-bound    | The only live seasonal currency. Minted exclusively by UBI. Destroyed on exit.   |
| **Seasonal Stars** | Season-bound  | Rank on the seasonal leaderboard and spendable budget for Sigil Vault purchases. |
| **Global Stars** | Yearly          | Earned from Seasonal Stars on exit. Determines global yearly leaderboard rank.   |
| **Sigils**       | Season-bound    | Temporary strategic items acquired via drops or Vault purchase. Destroyed on exit.|
| **Boosts**       | Season-bound    | Temporary power effects activated by spending Sigils.                             |
| **Cosmetics**    | Permanent       | Profile and interface enhancements purchased with Global Stars. No gameplay effect.|

---

## Core Mechanics

### Universal Basic Income (UBI)
Coins enter the economy only through UBI. Rate is dynamically adjusted based on:
- Player **activity state** (`Active` or `Idle` — offline is not a gameplay state)
- **Inflation dampening** based on total seasonal Coin supply
- **Hoarding suppression** based on recent spend behavior

### Lock-In
Players may voluntarily exit a season early by **Locking In**:
- Converts current Seasonal Stars to Global Stars at a **1:1 ratio**
- Destroys all remaining Coins, Sigils, and Boosts
- Removes the player from the live leaderboard
- Not available during Blackout

### Natural Season End
Players who remain until the season expires are eligible for:
- **Participation Bonus** — based on total Active ticks (max 56 Global Stars)
- **Placement Bonus** — 100, 60, and 40 Global Stars for 1st, 2nd, and 3rd place

### Trading
- Players may trade **Coins** and/or **Sigils** (both parties must contribute positive value)
- Seasonal Stars and Global Stars are **never tradeable**
- Trades are escrowed at initiation and expire after **3,600 seconds**
- Fees are symmetric and burned from both parties

---

## Progression and Prestige

Progression is deliberately minimal. There are no levels, XP systems, stat growth, or tenure bonuses.

| Prestige Type        | Description                                                                     |
|----------------------|---------------------------------------------------------------------------------|
| **Seasonal**         | Top-3 placement badges awarded at natural season end                           |
| **Yearly**           | Permanent top-10 badges awarded at yearly reset (every 12 season expirations)  |
| **Cosmetic**         | Permanent profile enhancements purchased with Global Stars; zero gameplay impact|

**Yearly reset** wipes all Global Stars and yearly standings. Cosmetics and badges are permanent.

---

## Server Architecture Principles

- **Single authoritative tick** — The simulation advances on a global 1-second tick (`global_tick_index` = Unix time in seconds).
- **Determinism** — All math, ordering, RNG, and replay must be fully deterministic from authoritative inputs.
- **Closed action set** — Only explicitly defined actions exist. No undefined or implicit behaviors.
- **Immutable logs** — All events are logged with hash-chained snapshots for auditability and replay.
- **Staff non-intervention** — Staff cannot edit balances, prices, ranks, rewards, timing, or outcomes.
- **Server modes** — `NORMAL`, `MAINTENANCE_LOCKDOWN`, `READ_ONLY_ECONOMY`, `LOCKDOWN_CONNECTIONS`, `RATE_LIMIT_ACTIONS`

### Tick Processing Order (per tick)
1. Classify tick; capture authoritative start-of-tick snapshot
2. Process scheduled system events
3. Run Sigil drop evaluation
4. Resolve player actions
5. Execute UBI accrual and system math
6. Publish next authoritative surfaces (prices, vault costs)
7. Evaluate activity transitions
8. Recalculate leaderboards
9. Persist audit and state records

---

## Documentation Structure

This repository contains the full game design specification (Game Bible):

| File | Contents |
|------|----------|
| `section_00_table_of_contents_v1.txt` | Complete table of contents for all chapters |
| `chapter_01_introduction_v7.txt` | Overview, game concept, vision, target audience |
| `chapter_02_core_game_loops_v15.txt` | Primary loop, seasonal loop, Lock-In decision loop |
| `chapter_03_core_mechanics_and_systems_v16.txt` | Player actions, activity requirements, boosts, constraints |
| `chapter_04_resources_and_currencies_v12.txt` | Coins, Seasonal Stars, Global Stars, Sigils, drop rules |
| `chapter_05_economy_design_v11.txt` | UBI, sinks, scarcity, inflation controls, trading friction |
| `chapter_06_progression_systems_v7.txt` | In-season and cross-season progression, leaderboards |
| `chapter_07_time_seasons_and_pacing_v9.txt` | Season schedule, join/exit constraints, blackout rules |
| `chapter_08_multiplayer_and_social_systems_v16.txt` | Player interaction, chat, social graph, spectatorship |
| `chapter_09_win_loss_and_success_definitions_v19.txt` | Win/loss definitions, valid success states |
| `chapter_10_ui_and_client_state_assumptions_v19.txt` | Client state model, UI contracts, mobile/desktop |
| `chapter_11_edge_cases_and_failure_states_v16.txt` | Disconnects, idle handling, season-end edge cases |
| `chapter_12_admin_moderation_and_governance_v8.txt` | Staff authority, moderation tools, reporting |
| `chapter_13_data_persistence_and_resets_v9.txt` | What persists, what is destroyed, reset triggers |
| `chapter_14_lifecycle_and_release_model_v9.txt` | Alpha, Beta, and Release phase definitions |
| `chapter_15_technical_and_implementation_references_v12.txt` | Event taxonomy, RNG contracts, logging, replay integrity |
| `too_many_coins_descriptive_unified_summary.md` | Unified narrative summary of the full design |

---

## Lifecycle Phases

The service is always in exactly one phase:

| Phase       | Description                                                              |
|-------------|--------------------------------------------------------------------------|
| **Alpha**   | Invite-only; feature gating and wipes allowed; no monetization          |
| **Beta**    | Requires beta entry wipe from Alpha; economy rules enforced              |
| **Release** | Full public availability; determinism and outcome history guaranteed     |

Phase transitions take effect only at tick boundaries and require a defined transition sequence.

---

## Key Design Principles

1. **Structural fairness** — No hidden systems, tenure bonuses, or administrative interventions in competitive outcomes.
2. **Timing over presence** — Advantage comes from strategic timing, not constant attendance.
3. **Transparency** — All seasonal economy data and leaderboards are publicly visible in real time.
4. **Closed rule set** — If a mechanic, social affordance, or admin capability is not explicitly defined, it does not exist.
5. **Auditability** — Immutable hash-chained logs and deterministic replay from authoritative inputs.