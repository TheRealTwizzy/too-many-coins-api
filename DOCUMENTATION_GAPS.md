# Documentation Gaps Audit

## Summary
This document identifies implemented features, systems, and mechanics that are either completely undocumented or only partially documented in the Game Bible.

**Total Gaps Found: 24+**

---

## Critical Gaps (Block Understanding of Core Systems)

### 1. Activity System Mechanics (CRITICAL)
**Status:** Implemented in code, **NOT** documented
**Location:** `faucet.go`, `handlers.go`
**Issue:** 
- Activity "claim" handler exists (`/claim-activity`)
- Activity warmup state is returned via `/player` endpoint
- But **what constitutes an activity action is never defined**
- When `recent_activity_seconds` is updated is unclear
- How activity differs from daily login is unexplained
- Activity tracking logic is invisible to documentation readers

**What's Missing:**
- Definition of "activity" (is it any endpoint? specific actions only?)
- Player-visible mechanics for activity logging
- When activity window is checked/reset
- Interaction with warmup decay

---

### 2. Activity Cooldown & Faucet Claim Dynamics
**Status:** Partially implemented, **NOT** centrally documented
**Location:** `calibration.go` (ActivityCooldownSeconds, ActivityReward), `handlers.go` (activityClaimHandler)
**Issue:**
- Activity cooldown in calibration.go (~360 seconds, 4% of daily cap)
- But exact claim rules, remaining cap tracking, and cooldown semantics are missing
- Cooldown application logic undocumented

**What's Missing:**
- Exact formulas for activity reward calculation
- How remaining daily cap affects activity claims
- Claim denial reasons and order of checks
- Interaction with UBI warmup system

---

### 3. HTTP Error Codes & Response Taxonomy
**Status:** Implemented ad-hoc, **NO** centralized documentation
**Location:** All handlers (`handlers.go`, `admin_handlers.go`, etc.)
**Issue:**
- Error codes used throughout: SEASON_ENDED, FEATURE_DISABLED, INVALID_REQUEST, PLAYER_NOT_FOUND, etc.
- No centralized error code spec
- No documentation of HTTP status codes returned
- Clients must infer error semantics from code

**What's Missing:**
- Complete list of all error codes by endpoint
- HTTP status code strategy (e.g., when is 400 vs 403 used?)
- Error code meanings and player visibility
- Retry semantics

---

### 4. Feature Flags (Not in Game Bible)
**Status:** Implemented in `feature_flags.go`, **ZERO** documentation
**Issue:**
- Four feature flags control system behavior:
  - ENABLE_FAUCETS
  - ENABLE_SINKS
  - ENABLE_TELEMETRY
  - ENABLE_IP_THROTTLING
- No documentation of what happens when each is disabled
- No runtime behavior spec
- No deployment guidance

**What's Missing:**
- Purpose and impact of each flag
- Default values and rationale
- When to disable each in production
- Fallback behavior when disabled

---

### 5. Rate Limiting Strategy (Incomplete)
**Status:** Partially documented, incomplete
**Location:** `auth_protection.go` documents auth limits, but game mechanics limits missing
**Issue:**
- Auth actions have documented rate limits (signup, login, auth)
- But gameplay endpoints have undocumented/inconsistent limits
- Faucet claim limits not centrally specified
- Purchase limits not documented
- Star purchase frequency not specified

**What's Missing:**
- Complete list of all rate-limited endpoints
- Per-endpoing rate limit values
- Limit window durations
- Backoff/throttle response strategy

---

### 6. Admin Audit Log Action Types (Taxonomy Documented)
**Status:** Documented in README/admin-tools.md
**Issue:**
- Taxonomy is now listed in admin-tools, but may diverge from implementation as new actions are added.

**What's Missing:**
- Audit log retention policy
- Ongoing process to keep taxonomy synced with code

---

### 7. Anti-Cheat Event Types (Documented)
**Status:** Documented in README/anti-abuse.md
**Issue:**
- Thresholds and response logic are now centralized, but require periodic validation against code.

**What's Missing:**
- False-positive handling policy
- Operator runbook for overrides

---

## Major Gaps (Incomplete System Documentation)

### 8. Boost System (Completely Undocumented)
**Status:** Implemented endoint `/buy-boost`, **NO** documentation
**Issue:**
- Handler exists but feature is invisible to players/designers
- Mechanics completely unknown

**What's Missing:**
- What is a boost?
- Cost/economics
- Effect Duration
- Strategic purpose
- UI/UX contract with client

---

### 9. Variant Stars (Completely Undocumented)
**Status:** Implemented endpoint `/buy-variant-star`, **NO** documentation
**Issue:**
- Handler exists, feature unexplained

**What's Missing:**
- What are variant stars?
- How do they differ from base stars?
- Cost/pricing model
- Leaderboard effect (if any)
- When available/gated

---

### 10. Burn Coins System (Completely Undocumented)
**Status:** Implemented endpoint `/burn-coins`, **NO** documentation
**Issue:**
- Handler exists, purpose unknown
- Game design rationale missing

**What's Missing:**
- Why players might burn (sink mechanism? cosmetic?)
- Exchange rate (fixed? variable?)
- Economics impact on emoji system
- When available

---

### 11. IP Tracking & Dampening Delay
**Status:** Implemented (`auth.go`, `handlers.go`), **partially** documented
**Location:** `anti-abuse.md` mentions IP enforcement, but mechanics not detailed
**Issue:**
- `RecordPlayerIP` and `ApplyIPDampeningDelay` exist but duration/impact not documented
- Why applied not explained

**What's Missing:**
- Exact dampening delay duration
- When delay applies (new IP? every login?)
- Player-visible behavior
- Abuse scenario it prevents

---

### 12. Telemetry Event Taxonomy (Incomplete)
**Status:** Partially implemented and documented
**Location:** `alpha-execution.md` lists some, but not all
**Issue:**
- Documented: buy_star, login
- Alpha note mentions: emission_tick, faucet_claim, star_purchase_attempt, star_purchase_success, market_pressure_tick
- But also: ubi_tick, activity_regular_interval, tick_reaction_burst, etc.
- Complete list not in one place

**What's Missing:**
- Single authoritative telemetry event registry
- All event types and payload schemas
- Sampling strategy (if any)
- Client-facing vs server-only events

---

### 13. Notification System (Partially Documented)
**Status:** Implemented, `notifications.md` exists but incomplete
**Issue:**
- Notification categories exist but not enumerated
- Push notification behavior unclear
- Delivery guarantees not specified
- Settings behavior not explained

**What's Missing:**
- Complete list of notification categories
- When each is triggered
- Push notification opt-in mechanism
- Notification retention policy
- Ack/delete semantics

---

### 14. Moderator Role & Capabilities (Documented)
**Status:** Documented in README/admin-tools.md
**Issue:**
- Capability matrix defined; audit logging for moderator actions still pending implementation.

**What's Missing:**
- Finalized action_type list for moderator audit logs

---

### 15. Player Profile Fields (Not Documented)
**Status:** `/profile` endpoint exists, schema incomplete
**Issue:**
- Profile endpoint returns fields not spec'd
- Biography, pronouns, location, website, avatar stored but game contract missing

**What's Missing:**
- Complete player profile schema
- Moderation status fields
- Cosmetic fields (if supported)
- Privacy model for profile fields

---

### 16. Global Settings Available Values (Not Centralized)
**Status:** Settings stored and queryable, **NO** master list
**Location:** `settings.go`, `calibration.go`
**Issue:**
- Implemented settings: ActiveDripInterval, IdleDripInterval, ActivityWindow, etc.
- But no central enumeration of all available settings
- No documentation of setting ranges/validation
- Calibration-related settings not explained

**What's Missing:**
- Complete enumeration of all global settings
- Allowed ranges and data types
- Calibration impact of each setting
- When changes take effect (immediate? next tick?)

---

### 17. Cooldown & Throttle Systems (Scattered)
**Status:** Multiple systems implemented, **NO** unified model
**Issue:**
- Activity cooldown: ~360 seconds
- Login reward cooldown: hours
- Rate limit windows: seconds to minutes
- IP dampening delay: unspecified
- No unified throttle/cooldown taxonomy

**What's Missing:**
- Unified cooldown model/spec
- Per-endpoint cooldown values table
- Cooldown reset conditions
- User-visible cooldown states

---

## Minor Gaps (Implementation Details)

### 18. CoinEarnings Log Source Types (Incomplete)
**Status:** Documented as (ubi, login, activity, task, comeback, playability_safeguard)
**Issue:**
- Documentation shows these but not all are used in code
- What actually triggers each source_type unclear
- Post-alpha sources may not be implemented

**What's Missing:**
- Confirmation of which source_type values are actually used in Alpha
- Conditions triggering each source_type
- Historical completeness guarantee

---

### 19. Database Schema Indexing Strategy (Not Documented)
**Status:** Schema exists, query performance strategy missing
**Issue:**
- No documentation of which columns are indexed
- No query optimization guidelines
- No schema design rationale

**What's Missing:**
- Index list and justification
- Query performance characteristics
- Archival/retention policy
- Cold query optimization strategy (if any)

---

### 20. Bot Configuration Profile Types (Incomplete)
**Status:** `bot-runner.md` partially documented
**Issue:**
- Bot strategies exist (threshold_buyer, etc.) but not all enumerated
- Bot profile types in code not in docs

**What's Missing:**
- Complete list of bot strategies
- Strategy parameters and ranges
- Bot behavior under each strategy
- Bot detection countermeasures

---

### 21. Season End Snapshot Details (Partially Documented)
**Status:** `/seasons` endpoint, frozen snapshot mentioned
**Issue:**
- Frozen state during season-ended is documented
- But exact field inclusion/values not specified

**What's Missing:**
- All fields included in ended season view
- Final values computation (last tick? last emission?)
- Immutability guarantees
- Archive behavior

---

### 22. Admin Control Strips & Season Controls (Incomplete)
**Status:** Handlers exist, UI shows "disabled" 
**Issue:**
- Code mentions future controls (pause purchases, reduce emission, freeze)
- But not clear what's actually implemented in Alpha

**What's Missing:**
- Which season controls are actually implemented in Alpha vs post-Alpha
- API contracts for implemented controls
- Effect/latency of controls
- Rollback capability

---

### 23. Health Check Behavior (Minimal)
**Status:** `/health` endpoint exists
**Issue:**
- Returns 200 OK but exact contract not documented
- What "healthy" means not defined
- SLA/monitor expectations not set

**What's Missing:**
- Health check response format
- What components are checked
- Unhealthy condition responses
- Monitoring/alerting expectations

---

### 24. Leaderboard Tie-Breaking (Not Documented)
**Status:** Leaderboard handler exists
**Issue:**
- `/leaderboard` endpoint exists
- Tie-breaking logic not documented
- Ranking display rules not specified

**What's Missing:**
- Exact ranking algorithm
- Tie-breaking order (when star counts equal)
- Historical ranking preservation
- Pagination/sorting strategy

---

## Cross-Cutting Issues

### Documentation Organization
- **No centralized HTTP API reference** - endpoints scattered across files/docs
- **No error code taxonomy** - error codes ad-hoc throughout
- **No constant/configuration reference** - settings/limits not enumerated
- **No system boundary definitions** - unclear what's Alpha vs post-Alpha for small features

### System Interdependencies Not Documented
- Activity warmup + UBI interaction: documented but edge cases missing
- Faucet claims + daily claims + UBI: mechanics not clearly separated
- Rate limits + throttling + cooldowns: overlapping systems
- Admin controls + season state: interaction unclear

---

## Recommendations (Priority Order)

### P0: Unblock Players & Developers
1. **[REQ-1]** Create [Activity System](README/activity-system.md):
   - Define activity trigger conditions
   - Document activity vs login faucets
   - Explain recent_activity_seconds lifecycle

2. **[REQ-2]** Create [HTTP API Reference](README/http-api-reference.md):
   - All endpoints with methods, paths, auth requirements
   - Request/response schemas
   - Complete error code taxonomy

3. **[REQ-3]** Document Feature Flags in [README.md](README.md):
   - Add section: "Feature Flags (Deployment Configuration)"
   - Per-flag purpose, impact, default value

### P1: Completeness & Governance
4. **[REQ-4]** Create [Anti-Cheat Events Registry](README/anti-cheat-events.md)
5. **[REQ-5]** Document Boost and Variant Stars in [README/star-purchases.md](README/star-purchases.md)
6. **[REQ-6]** Add Burn Coins mechanics to [README/economy.md](README/economy.md)
7. **[REQ-7]** Create [Admin Audit Log](README/admin-audit-log.md) with action_type taxonomy
8. **[REQ-8]** Create [Notification System Spec](README/notifications.md) (expand existing)

### P2: Developer Productivity
9. Centralize all rate limit values in [README/rate-limits.md](README/rate-limits.md)
10. Centralize all cooldown/throttle mechanics in one spec
11. Document global settings in [README/settings.md](README/settings.md)
12. Add schema indexing notes to [README/persistent-state.md](README/persistent-state.md)

---

## Severity Classification

| Severity | Count | Examples |
|----------|-------|----------|
| **CRITICAL** (Breaks understanding) | 7 | Activity system, error codes, feature flags, admin audit log |
| **MAJOR** (Significant gaps) | 10 | Boosters, variant stars, burn coins, notifications, IP tracking |
| **MINOR** (Implementation details) | 7 | CoinEarnings source types, indexing, bot profiles, leaderboard |

**Immediate Action Needed:** REQ-1, REQ-2, REQ-3 block player onboarding and developer iteration.
