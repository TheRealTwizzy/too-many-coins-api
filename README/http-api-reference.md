# HTTP API Reference

## Scope
- **Type:** System contract  
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible
- **Alpha Status:** Complete (47 endpoints)

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

# HTTP API Specification

Complete reference for all public, player, admin, and moderator endpoints.

---

## Response Format Convention

All endpoints return JSON with a consistent structure:

**Success Response:**
```json
{
  "ok": true,
  "data": { /* endpoint-specific fields */ }
}
```

**Error Response:**
```json
{
  "ok": false,
  "error": "ERROR_CODE",
  /* optional context fields, e.g., "nextAvailableInSeconds" */
}
```

**HTTP Status Codes:**
- `200 OK` — Most endpoints (success or application-level error)
- `400 Bad Request` — Malformed JSON, invalid type
- `401 Unauthorized` — Missing or invalid session
- `403 Forbidden` — Valid session but insufficient permissions
- `404 Not Found` — Endpoint doesn't exist
- `405 Method Not Allowed` — Wrong HTTP method
- `503 Service Unavailable` — Server initializing, timeout

---

## Error Code Taxonomy

| ERROR CODE | HTTP Status | Context | Meaning |
|-----------|-----------|---------|---------|
| `INVALID_CREDENTIALS` | 200 | Login | Username/password mismatch; account not found |
| `ACCOUNT_FROZEN` | 200 | Login | Account suspended; cannot log in |
| `UNAUTHORIZED` | 401 | General | Session missing or expired; request new login |
| `FORBIDDEN` | 403 | Auth | Valid session but lacks permission (non-admin accessing admin endpoint) |
| `ADMIN_BOOTSTRAP_REQUIRED` | 403 | Admin | Owner bootstrap not yet claimed; use `/admin/bootstrap/claim` first |
| `INVALID_REQUEST` | 400 | General | Malformed JSON or missing required fields |
| `INVALID_PLAYER_ID` | 200 | Player | PlayerID not valid UUID format |
| `INVALID_AMOUNT` | 200 | Economy | Coin/star amount invalid (≤0 or out of range) |
| `PLAYER_NOT_FOUND` | 200 | Admin | Admin search or control on non-existent player |
| `PLAYER_NOT_REGISTERED` | 200 | Earn | Player exists but state not initialized |
| `SEASON_NOT_FOUND` | 200 | Season | Season ID not found in database |
| `SEASON_ENDED` | 200 | Earn/Buy | Active season has ended; no earning or purchases allowed |
| `FEATURE_DISABLED` | 200 | Earn/Buy | Feature flag disabled (faucets, sinks, telemetry, etc.) |
| `COOLDOWN` | 200 | Earn | Faucet claim still in cooldown; includes `nextAvailableInSeconds` |
| `DAILY_CAP` | 200 | Earn | Daily coin earning limit reached for this player |
| `EMISSION_EXHAUSTED` | 200 | Earn | Global emission pool depleted; no coins available |
| `NOT_ENOUGH_COINS` | 200 | Buy | Insufficient coin balance for purchase or burn |
| `NOT_ENOUGH_STARS` | 200 | Buy | Insufficient star balance for star purchase |
| `METHOD_NOT_ALLOWED` | 405 | General | Wrong HTTP method for endpoint (e.g., GET on POST-only) |
| `INTERNAL_ERROR` | 500 | General | Unhandled server error; contact admin |
| `SERVICE_UNAVAILABLE` | 503 | Startup | Server still initializing; retry in a few seconds |

---

## Authentication & Sessions

All player endpoints requiring authentication use **HTTP-only session cookies**.

**Session Flow:**
```
1. POST /auth/signup   →  Create account
2. POST /auth/login    →  Authenticate, receive session cookie
3. Subsequent requests include cookie automatically (browser)
4. POST /auth/logout   →  Destroy session
```

**Cookie Details:**
- Name: `session` (or server-configured name)
- HttpOnly: true (JavaScript cannot access)
- Secure: true (HTTPS only in production)
- SameSite: Lax
- Path: `/`

---

## Endpoint Catalog

### PUBLIC ENDPOINTS

#### `GET /`
Serves the main HTML UI (index.html).

| Property | Value |
|----------|-------|
| Authentication | None |
| Response | HTML document |

---

#### `GET /health`

Health check endpoint. Returns server status.

| Property | Value |
|----------|-------|
| Authentication | None |
| Request | (none) |
| Response | `{ "ok": true, "status": "healthy" }` |
| HTTP Status | 200 or 503 |

**When unhealthy (503):**
- Database connection failed
- Season state not initialized yet

---

#### `GET /leaderboard`

Player rankings by star count for current season.

| Property | Value |
|----------|-------|
| Authentication | None |
| Query Parameters | `?limit=100&offset=0` |
| Request | (none) |
| Response | `{ "ok": true, "leaderboard": [ { "rank": 1, "username": "...", "displayName": "...", "stars": 500, ... } ] }` |

---

#### `POST /telemetry`

Client-side event tracking (e.g., button clicks, engagements).

| Property | Value |
|----------|-------|
| Authentication | Optional |
| Request | `{ "eventType": "buy_star", "payload": { ... } }` |
| Response | `{ "ok": true }` (always 204 No Content on success) |

**Event Types (Client):**
- `login` — User logged in
- `buy_star` — User initiated star purchase
- (others as per telemetry schema)

---

### PLAYER ENDPOINTS

#### `GET /player`

Fetch current player's profile, economy state, and activity warmup.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | (none) |
| Response | `{ "ok": true, "player": { "playerId": "uuid", "coins": 1000, "stars": 50, "displayName": "Player1", "activityWarmup": 0.75, "ubiMultiplier": 7.75, "currentUBIPerTick": 7 } }` |
| Error Codes | `INVALID_PLAYER_ID`, `PLAYER_NOT_REGISTERED`, `INTERNAL_ERROR` |

---

#### `GET /auth/me`

Get current authenticated user's account info.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | (none) |
| Response | `{ "ok": true, "account": { "username": "player1", "email": "...", "playerId": "uuid", "isAdmin": false, "isModerator": false } }` |
| Error Codes | `UNAUTHORIZED` |

---

#### `GET /profile` or `POST /profile`

Get or update player profile.

| Property | Value |
|----------|-------|
| Authentication | Required |
| GET Request | (none) |
| GET Response | `{ "ok": true, "profile": { "displayName": "Player1", "bio": "...", "pronouns": "...", "location": "...", "website": "...", "avatarUrl": "..." } }` |
| POST Request | `POST /profile` with JSON body: `{ "displayName": "NewName", "bio": "...", ... }` |
| POST Response | `{ "ok": true, "profile": { /* updated fields */ } }` |
| Error Codes | `INVALID_REQUEST`, `UNAUTHORIZED` |

---

#### `GET /seasons`

Fetch current active season data (price, coins in circulation, time remaining, etc.).

| Property | Value |
|----------|-------|
| Authentication | None (public but can be authenticated) |
| Request | (none) |
| Response | `{ "ok": true, "season": { "seasonId": "uuid", "status": "active", "dayIndex": 5, "totalDays": 14, "secondsRemaining": 777600, "currentStarPrice": 450, "coinsInCirculation": 50000, ... } }` |
| HTTP Status | 200 (normal) or 503 (initializing) |

**Status Field Values:**
- `"active"` — Season ongoing
- `"ended"` — Season finished; no new earning/buying allowed

---

#### `POST /activity`

Signal player presence for activity tracking.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | (empty JSON body: `{}`) |
| Response | `{ "ok": true, "timestamp": "2026-02-09T12:34:56Z" }` |
| Error Codes | `UNAUTHORIZED`, `INTERNAL_ERROR` |

**Behavior:** Updates `players.last_active_at` to current time. Called by frontend every ~20 seconds.

---

#### `POST /claim-daily`

Claim daily login faucet reward.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | `{}` |
| Response | `{ "ok": true, "reward": 150, "playerCoins": 5234, "nextAvailableInSeconds": 72000 }` |
| Error Codes | `SEASON_ENDED`, `FEATURE_DISABLED`, `COOLDOWN`, `DAILY_CAP`, `EMISSION_EXHAUSTED` |

**Cooldown:** ~20 hours (parametrizable)

---

#### `POST /claim-activity`

Claim activity faucet reward (frequent payout for active players).

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | `{}` |
| Response | `{ "ok": true, "reward": 12, "playerCoins": 5246, "nextAvailableInSeconds": 342 }` |
| Error Codes | `SEASON_ENDED`, `FEATURE_DISABLED`, `COOLDOWN`, `DAILY_CAP`, `EMISSION_EXHAUSTED` |

**Cooldown:** ~6 minutes  
**Requirement:** Player must be active (within activity window)

---

#### `GET /events`

Server-Sent Events (SSE) stream for real-time updates.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | (none; keep-alive stream) |
| Response Stream | SSE events (e.g., `data: {"type":"price_update","price":450}`) |

---

#### `POST /buy-star`

Purchase a single star with coins.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | `{ "quantity": 1 }` (usually 1) |
| Response | `{ "ok": true, "starsPurchased": 1, "coinsBurned": 450, "playerCoins": 4796, "playerStars": 51, "priceAfter": 455 }` |
| Error Codes | `SEASON_ENDED`, `FEATURE_DISABLED`, `NOT_ENOUGH_COINS`, `INTERNAL_ERROR` |

---

#### `POST /buy-star/quote`

Get price quote for stars without purchasing.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | `{ "quantity": 1 }` |
| Response | `{ "ok": true, "quantity": 1, "pricePerUnit": 450, "totalCost": 450, "playerCoins": 5246, "canAfford": true }` |
| Error Codes | `INVALID_REQUEST`, `INVALID_AMOUNT` |

---

#### `POST /buy-variant-star`

Purchase specialty/variant stars (post-alpha feature).

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | `{ "variantType": "...", "quantity": 1 }` |
| Response | `{ "ok": true, "variant": "...", "purchased": 1, "coinsBurned": 500 }` |
| Error Codes | `SEASON_ENDED`, `FEATURE_DISABLED`, `NOT_ENOUGH_COINS` |

---

#### `POST /buy-boost`

Purchase an activity or other boost.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | `{ "boostType": "activity", "quantity": 1 }` |
| Response | `{ "ok": true, "boostType": "activity", "durationSeconds": 3600, "coinsBurned": 250 }` |
| Error Codes | `SEASON_ENDED`, `FEATURE_DISABLED`, `NOT_ENOUGH_COINS` |

---

#### `POST /burn-coins`

Burn (sink) coins from player account.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | `{ "amount": 100 }` (microcoins) |
| Response | `{ "ok": true, "coinsBurned": 100, "playerCoins": 5146 }` |
| Error Codes | `SEASON_ENDED`, `FEATURE_DISABLED`, `NOT_ENOUGH_COINS`, `INVALID_AMOUNT` |

---

#### `GET/POST /notifications`

Fetch or send notifications.

| Property | Value |
|----------|-------|
| Authentication | Required |
| GET Request | (none) |
| GET Response | `{ "ok": true, "notifications": [ { "id": 1, "category": "system", "title": "...", "message": "...", "createdAt": "..." } ] }` |
| POST Request | `{ "message": "..." }` (set notification) |
| POST Response | `{ "ok": true }` |

---

#### `POST /notifications/ack`

Mark notification as acknowledged.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | `{ "notificationId": 123 }` |
| Response | `{ "ok": true }` |

---

#### `POST /notifications/delete`

Delete notification.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | `{ "notificationId": 123 }` |
| Response | `{ "ok": true }` |

---

#### `GET /notifications/settings` or `POST /notifications/settings`

Get or update notification preferences.

| Property | Value |
|----------|-------|
| Authentication | Required |
| GET Response | `{ "ok": true, "settings": { "category": "system", "enabled": true, "pushEnabled": false } }` |
| POST Request | `{ "category": "system", "enabled": true, "pushEnabled": false }` |
| POST Response | `{ "ok": true }` |

---

#### `GET /notifications/stream`

SSE stream for real-time notifications (alternative to polling).

| Property | Value |
|----------|-------|
| Authentication | Required |
| Response Stream | SSE events with notification payloads |

---

#### `POST /bugs/report`

Submit a bug report.

| Property | Value |
|----------|-------|
| Authentication | Optional |
| Request | `{ "title": "...", "description": "...", "category": "gameplay" }` |
| Response | `{ "ok": true, "bugId": 456, "status": "submitted" }` |
| Error Codes | `INVALID_REQUEST`, `INTERNAL_ERROR` |

---

### AUTHENTICATION ENDPOINTS

#### `POST /auth/signup`

Create a new account.

| Property | Value |
|----------|-------|
| Authentication | None |
| Request | `{ "username": "newplayer", "password": "...", "email": "..." }` |
| Response | `{ "ok": true, "playerId": "uuid", "username": "newplayer" }` |
| Error Codes | `INVALID_REQUEST` |

---

#### `POST /auth/login`

Authenticate and create session.

| Property | Value |
|----------|-------|
| Authentication | None |
| Request | `{ "username": "player1", "password": "..." }` |
| Response | `{ "ok": true, "username": "player1", "playerId": "uuid", "isAdmin": false, "isModerator": false }` |
| Error Codes | `INVALID_CREDENTIALS`, `ACCOUNT_FROZEN` |
| Side Effect | Sets `session` cookie |

---

#### `POST /auth/logout`

Destroy session.

| Property | Value |
|----------|-------|
| Authentication | Required |
| Request | (none) |
| Response | `{ "ok": true }` |
| Side Effect | Clears `session` cookie |

---

#### `POST /auth/request-reset`

Request a password reset (sends email link).

| Property | Value |
|----------|-------|
| Authentication | None |
| Request | `{ "username": "player1" }` or `{ "email": "player@example.com" }` |
| Response | `{ "ok": true, "message": "Reset link sent" }` |

---

#### `POST /auth/reset-password`

Complete password reset using reset token.

| Property | Value |
|----------|-------|
| Authentication | None |
| Request | `{ "token": "...", "newPassword": "..." }` |
| Response | `{ "ok": true, "message": "Password reset" }` |

---

### ADMIN ENDPOINTS

**Authentication:** All require valid admin session.

#### `GET /admin/bootstrap/status`

Check admin bootstrap status.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | (none) |
| Response | `{ "ok": true, "bootstrapSealed": true, "adminExists": true, "passwordChanged": true }` |
| Error Codes | `ADMIN_BOOTSTRAP_REQUIRED` |

---

#### `POST /admin/bootstrap/claim`

Claim admin ownership with rotating claim code.

| Property | Value |
|----------|-------|
| Authentication | None (pre-bootstrap) |
| Request | `{ "claimCode": "XXXXXX" }` |
| Response | `{ "ok": true, "message": "Bootstrap claimed; set password at /admin/initialize" }` |
| Error Codes | `INVALID_REQUEST` |

---

#### `GET /admin/overview`

Real-time overview of economy and player metrics (last hour, last 24h, last 7d).

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | (none) |
| Response | `{ "ok": true, "coinsEmittedLastHour": 500, "starsPurchasedLastHour": 12, "uniquePlayersLast24h": 45, "uniquePlayersLast7d": 180 }` |

---

#### `GET /admin/economy`

Economy calibration and state.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | (none) |
| Response | `{ "ok": true, "dailyEmissionTarget": 10000, "baseStarPrice": 400, "currentStarPrice": 450, "marketPressure": 0.45, "dailyCapEarly": 60, "dailyCapLate": 25 }` |

---

#### `GET /admin/telemetry`

Event telemetry analytics (48 hours, binned by hour).

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | (none) |
| Query | `?eventType=buy_star` (optional filter) |
| Response | `{ "ok": true, "series": [ { "bucket": "2026-02-08T12:00:00Z", "eventType": "buy_star", "count": 45 } ] }` |

---

#### `GET /admin/abuse-events`

Log of detected abuse patterns (last 7 days).

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | (none) |
| Response | `{ "ok": true, "events": [ { "playerId": "uuid", "eventType": "ip_cluster_activity", "severity": "medium", "timestamp": "2026-02-08T12:34:56Z" } ] }` |

---

#### `GET /admin/anti-cheat`

Anti-cheat system status and toggles.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | (none) |
| Response | `{ "ok": true, "toggles": [ { "key": "ip_enforcement", "enabled": true }, { "key": "clustering_detection", "enabled": true } ], "sensitivity": { "clustering": "medium", "throttle": "high" } }` |

---

#### `GET /admin/bugs`

Paginated list of user-submitted bug reports.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Query | `?limit=20&offset=0` |
| Response | `{ "ok": true, "items": [ { "id": 456, "title": "...", "description": "...", "category": "gameplay", "createdAt": "...", "status": "open" } ], "total": 127 }` |

---

#### `GET /admin/audit-log`

Admin action audit log.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Query | `?limit=50&offset=0&search=...` (optional filter) |
| Response | `{ "ok": true, "items": [ { "id": 1, "adminUsername": "admin1", "actionType": "player_control", "scopeType": "player", "scopeId": "uuid", "reason": "...", "timestamp": "..." } ], "total": 500 }` |

---

#### `GET /admin/player-search`

Search for players by username.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Query | `?username=player&limit=10` |
| Response | `{ "ok": true, "players": [ { "playerId": "uuid", "username": "player1", "stars": 50, ... } ] }` |

---

#### `POST /admin/player-controls`

Get or modify player state (coins, stars, drip status, etc.).

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | `{ "playerId": "uuid" }` or `{ "username": "player1" }` |
| Response | `{ "ok": true, "playerId": "uuid", "coins": 5000, "stars": 50, "dripMultiplier": 1.0, "isBot": false }` |

---

#### `GET/POST /admin/settings`

Get or update global game settings.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| GET Response | `{ "ok": true, "settings": { "activeDripIntervalSeconds": 60, "activityWindowSeconds": 120, ... } }` |
| POST Request | `{ "activityWindowSeconds": 150 }` |
| POST Response | `{ "ok": true, "settings": { /* updated */ } }` |

---

#### `GET /admin/star-purchases`

History/log of all star purchases.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Query | `?limit=50&offset=0&playerId=...` (optional filter) |
| Response | `{ "ok": true, "purchases": [ { "id": 1, "playerId": "uuid", "quantity": 5, "costCoins": 2250, "timestamp": "..." } ], "total": 5000 }` |

---

#### `GET /admin/bots`

List of all active bots.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | (none) |
| Response | `{ "ok": true, "bots": [ { "playerId": "uuid", "username": "bot_1", "displayName": "TestBot", "isBot": true, "botProfile": "threshold_buyer" } ] }` |

---

#### `POST /admin/bots/create`

Create a new test bot.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | `{ "username": "bot_2", "profile": "threshold_buyer", "threshold": 500 }` |
| Response | `{ "ok": true, "playerId": "uuid", "username": "bot_2" }` |

---

#### `POST /admin/bots/delete`

Delete a bot.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | `{ "playerId": "uuid" }` |
| Response | `{ "ok": true }` |

---

#### `POST /admin/notifications`

Send admin-created notification to player(s).

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | `{ "title": "Maintenance", "message": "Server down at 3pm UTC", "targetPlayers": ["uuid1", "uuid2"] }` (or empty for broadcast) |
| Response | `{ "ok": true, "notificationsSent": 2 }` |

---

#### `POST /admin/role`

Grant or revoke admin/moderator role.

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Request | `{ "username": "player1", "role": "admin" }` (or "moderator", "player") |
| Response | `{ "ok": true, "username": "player1", "role": "admin" }` |

---

#### `POST /admin/seasons/advance`

Manually advance to next season (recovery only).

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Phase | Alpha only |
| Request | (none) |
| Response | `{ "ok": true, "newSeasonId": "uuid", "message": "Season advanced" }` |
| Error Codes | `INVALID_PHASE` |

---

#### `POST /admin/seasons/recovery`

Create recovery season (Alpha emergency use).

| Property | Value |
|----------|-------|
| Authentication | Admin |
| Phase | Alpha only |
| Request | `{ "confirm": "I understand this is a recovery action", "reason": "..." }` |
| Response | `{ "ok": true, "newSeasonId": "uuid" }` |

---

### MODERATOR ENDPOINTS

#### `GET/POST /moderator/profile`

Moderator-level profile viewing and editing.

| Property | Value |
|----------|-------|
| Authentication | Moderator or Admin |
| GET Query | `?username=player1` |
| GET Response | `{ "ok": true, "profile": { "username": "player1", "displayName": "...", "email": "...", "bio": "...", "pronouns": "...", "location": "...", "website": "...", "avatarUrl": "..." } }` |
| POST Request | `{ "username": "player1", "displayName": "NewName", ... }` |
| POST Response | `{ "ok": true, "profile": { /* updated */ } }` |

---

## Rate Limiting & Throttling

**Auth Endpoints:**
- Signup: 5 attempts per IP per hour
- Login: 12 attempts per IP per hour
- Auth actions: 10 per IP per hour

**Gameplay Endpoints:**
- Daily claim: 1 per player per ~20 hours (cooldown-enforced)
- Activity claim: 1 per player per ~6 minutes (cooldown-enforced)
- Star purchase: No per-endpoint limit; daily cap enforced

**Admin Endpoints:**
- No rate limits (trusted admin);assumes internal use only

---

## Examples

### Typical Game Loop

```
1. Player logs in
   POST /auth/login { "username": "player1", "password": "..." }
   ← receive session cookie

2. Player enters game
   GET /seasons
   ← get current season, price, time remaining
   
   GET /player
   ← get coin/star balance, activity warmup

3. Player plays for 20 seconds
   POST /activity {}
   ← signal presence to server

4. Player clicks "Claim" button
   POST /claim-activity {}
   ← earn 10-15 coins if eligible

5. Player buys a star
   POST /buy-star { "quantity": 1 }
   ← star added, coins deducted, price raised

6. Player logs out
   POST /auth/logout {}
   ← session destroyed
```

### Admin Monitoring

```
1. Admin logs in
   POST /auth/login { "username": "admin", "password": "..." }

2. Admin opens dashboard
   GET /admin/overview
   ← see metrics

   GET /admin/economy
   ← see calibration

   GET /admin/telemetry?eventType=buy_star
   ← see purchase events

3. Admin detects abuse
   GET /admin/abuse-events
   ← see flagged patterns

   GET /admin/player-search?username=suspicious
   ← find suspect player

   POST /admin/player-controls { "username": "suspicious" }
   ← inspect state

4. Admin takes action
   POST /admin/role { "username": "suspicious", "role": "player" }
   ← revoke if moderator, or leave as-is
```

---

## See Also

- [Activity System](activity-system.md) — `/activity` and `/claim-activity` details
- [Coin Faucets](coin-faucets.md) — Earning mechanics background
- [Admin Tools](admin-tools.md) — Admin governance and capabilities
