# Admin Audit Log & Notification System Research

**Research Date:** February 2026  
**Scope:** Admin audit log (`admin_audit_log` table) and notification system (`notifications` table)

---

## I. ADMIN AUDIT LOG SYSTEM

### A. Table Schema

```sql
CREATE TABLE admin_audit_log (
    id BIGSERIAL PRIMARY KEY,
    admin_account_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    scope_type TEXT NOT NULL,
    scope_id TEXT NOT NULL,
    reason TEXT,  -- NULLABLE; NULLIF($5, '') in code
    details TEXT/JSONB,  -- NULLABLE JSON-encoded details
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Status:** Immutable append-only audit log. All inserts via `logAdminAction()` helper.

---

### B. Complete Action Type Taxonomy

All `action_type` values logged in code (immutable from code analysis):

| Action Type | Scope Type | Trigger | Context Details | Audit Record |
|---|---|---|---|---|
| `auto_admin_bootstrap` | `account` | Auto-creation of alpha-admin account at startup | `username`, `displayName` | Bootstrap initialization event |
| `admin_bootstrap_claim` | `account` | Owner claims bootstrap via secret code window | `username`, `ip` | Initial admin password set |
| `role_update` | `account` | Admin grants/revokes admin/moderator role via `/admin/role` | `username`, `role` | Role change action |
| `season_control_set` | `season_control` | Admin sets season control value via season controls endpoint | Season control ID (format: `{seasonID}:{controlName}`), full control payload with old/new values, reason, snapshot data | Economy state snapshot at time of change |
| `season_advance` | `season` | Admin manually advances season (Alpha-only, recovery) | `previousSeasonId` (if exists), `startedAt` (RFC3339) | Season progression override |
| `season_recovery` | `season` | Admin creates recovery season (Alpha emergency, recovery-only) | Season creation initiation | Season recovery action |
| `notification_create` | `notification` | Admin creates custom notification via `/admin/notifications` | `category`, `type`, `priority`, `message`, `seasonId`, `accountId` | Admin-created notification audit |
| `bot_toggle` | `player` | Toggle bot status on player (deprecated/admin-only) | `wasBot`, `nowBot`, `botProfile` (if applicable) | Bot status change |
| `bot_create` | `player` | Admin creates new bot player | Bot account details | Bot account creation |
| `bot_delete` | `player` | Admin deletes bot player | None typical | Bot account deletion |
| `profile_freeze` | `account` | Admin freezes account (prevents login/activity) | `username`, `playerId` | Account frozen state |
| `profile_unfreeze` | `account` | Admin unfreezes account | `username`, `playerId` | Account unfrozen state |
| `profile_delete` | `account` | Admin permanently deletes account and cascading data | `username`, `playerId`, `previousRole` | Permanent account deletion |

**Total Action Types:** 13

---

### C. Complete Scope Type Taxonomy

| Scope Type | Description | Examples |
|---|---|---|
| `account` | Player or admin account | Account ID or username |
| `season` | Season-level action | Season UUID |
| `season_control` | Season economy control value | Format: `{seasonID}:{controlName}` |
| `notification` | Notification creation | Recipient role (`player`, `moderator`, `admin`, `all`) |
| `player` | Player/bot operations | Player ID |

**Total Scope Types:** 5

---

### D. Retention & Indexing

**Retention Policy:** No explicit retention policy; logs are append-only, immutable, no automatic purge.

**Indexes:** None specified in code; searches via full-table scan of `admin_audit_log` with optional filters.

---

### E. Search & Filter Capabilities

**Endpoint:** `GET /admin/audit-log`

**Query Parameters:**
- `q` (optional) - Full-text search across:
  - `adminUsername` (from joined `accounts` table)
  - `action_type` (ILIKE case-insensitive)
  - `scope_id` (ILIKE)
  - `reason` (ILIKE)
- `limit` (default: 50, max: 200)

**Response Structure:**
```json
{
  "ok": true,
  "items": [
    {
      "id": 123,
      "adminAccount": "account-uuid",
      "adminUsername": "admin1",
      "actionType": "role_update",
      "scopeType": "account",
      "scopeId": "player-uuid",
      "reason": "",
      "details": { "username": "player1", "role": "admin" },
      "createdAt": "2026-02-09T10:30:00Z"
    }
  ],
  "total": 500,
  "limit": 50,
  "query": "role_update"
}
```

---

### F. Admin Action Logging Helper

**Function Signature:**
```go
func logAdminAction(
  db *sql.DB,
  adminAccountID string,      // Account performing action
  actionType string,            // Action type enum
  scopeType string,             // Scope type enum
  scopeID string,               // Resource ID/identifier
  reason string,                // Optional context (e.g., "override", "recovery-only")
  details map[string]interface{} // Context details as JSON
) error
```

**Behavior:**
- Validates non-empty:  `adminAccountID`, `actionType`, `scopeType`, `scopeId`
- Returns `nil` if any validation fails (silent fail)
- Encodes `details` as JSON string if provided
- Inserts with `NULLIF($5, '')` to exclude empty `reason` strings
- Is **fire-and-forget** in most callsites (errors ignored with `_`)

---

### G. Admin Access Control

**Access Level:** Admin-only endpoint. Uses `requireAdmin()` check.

**No Role Hierarchy:** Admins see all audit records. Moderators cannot access audit log.

---

## II. NOTIFICATION SYSTEM

### A. Database Schema

```sql
-- Main notifications table
CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    -- Legacy fields (for backward compat):
    target_role TEXT NOT NULL,
    account_id TEXT,
    message TEXT NOT NULL,
    level TEXT NOT NULL DEFAULT 'info',
    link TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    
    -- Modern fields (Phase 0+):
    recipient_role TEXT,                 -- Replaces target_role
    recipient_account_id TEXT,           -- Replaces account_id
    season_id TEXT,
    category TEXT,                       -- 'economy', 'player_action', 'market', 'abuse', 'system', 'admin'
    type TEXT,                           -- Free-form event type string
    priority TEXT DEFAULT 'normal',      -- 'normal', 'high', 'critical'
    payload JSONB,                       -- Event details
    ack_required BOOLEAN DEFAULT FALSE,  -- Force acknowledgment
    acknowledged_at TIMESTAMPTZ,
    dedupe_key TEXT                      -- Deduplication key
);

-- Read state (optional, implicit read on first fetch)
CREATE TABLE notification_reads (
    notification_id BIGINT PRIMARY KEY,
    account_id TEXT NOT NULL,
    read_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (notification_id, account_id)
);

-- Acknowledgment state (critical notifications)
CREATE TABLE notification_acks (
    notification_id BIGINT PRIMARY KEY,
    account_id TEXT NOT NULL,
    acknowledged_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (notification_id, account_id)
);

-- Soft delete state (player can hide notifications)
CREATE TABLE notification_deletes (
    notification_id BIGINT PRIMARY KEY,
    account_id TEXT NOT NULL,
    deleted_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (notification_id, account_id)
);

-- User notification preferences
CREATE TABLE notification_settings (
    account_id TEXT PRIMARY KEY,
    category TEXT NOT NULL,              -- Notification category
    enabled BOOLEAN DEFAULT TRUE,        -- Show in-app?
    push_enabled BOOLEAN DEFAULT FALSE,  -- Allow browser push?
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (account_id, category)
);
```

**Indexes:**
- `idx_notifications_created_at` on `created_at`
- `idx_notifications_dedupe` on `(dedupe_key, created_at)`

---

### B. Notification Categories (Complete Taxonomy)

**Defined Constants in `notifications.go`:**

```go
const (
    NotificationCategoryEconomy      = "economy"       // Economy state changes
    NotificationCategoryPlayerAction = "player_action" // Player-triggered events
    NotificationCategoryMarket       = "market"        // Market pressure events
    NotificationCategoryAbuse        = "abuse"         // Abuse detection (admins)
    NotificationCategorySystem       = "system"        // System events (season end, etc.)
    NotificationCategoryAdmin        = "admin"         // Admin-only events
)
```

**Total Categories:** 6

---

### C. Notification Types by Category

**ECONOMY Category:**
| Type | Trigger | Priority | Recipients | Example Payload |
|---|---|---|---|---|
| `bulk_star_purchase` | Player buys ≥3 stars in one purchase | `normal` (or `high` if ≥maxQty) | Admin | `{ accountId, playerId, quantity, totalCoinsSpent, finalStarPrice }` |
| `season_emission_threshold_exceeded` | Daily emission exceeds user-configured threshold | `high` | Admin | Economy event details |
| `price_update` | Star price changes significantly | `normal` | Admin | Price delta snapshot |
| `economy_invariant_violation` | Checksum/invariant fail detected | `critical` | Admin | Full economy snapshot |

**PLAYER_ACTION Category:**
| Type | Trigger | Priority | Recipients | Example Payload |
|---|---|---|---|---|
| `purchase_completed` | Star purchase completes successfully | `normal` | Admin (bulk only) | Purchase details |

**MARKET Category:**
| Type | Trigger | Priority | Recipients | Example Payload |
|---|---|---|---|---|
| `market_pressure_spike` | Market pressure crosses threshold (≥1.5; critical ≥1.7) | `high` or `critical` | Admin | `{ last24h, last7d, ratio, desired, marketPressure, pressureClamp }` |
| `market_pressure_recovery` | Market pressure drops below threshold (≤0.8) | `high` | Admin | Recovery event data |

**ABUSE Category:**
| Type | Trigger | Priority | Recipients | Example Payload |
|---|---|---|---|---|
| `severity_2` | Abuse event with severity 2 (first detection or >6h gap) | `high` | Admin | Abuse event details |
| `severity_3` | Abuse event with severity 3 (immediate) | `critical` | Admin | Abuse event details |

**SYSTEM Category:**
| Type | Trigger | Priority | Recipients | Example Payload |
|---|---|---|---|---|
| `season_started` | New season begins | `high` | All (implicit) | Season details |
| `season_ending_soon` | Season in final phase (T-Δ) | `normal` | All | Season end countdown |
| `season_ended` | Season transition complete | `high` | All | Final season stats |
| `season_ended_leaderboard_snapshot` | Leaderboard finalized after season end | `normal` | All | Top rankings |
| `emission_paused_low_liquid` | Faucet emission paused (low liquidity) | `high` | Admin | Liquidity state |
| `emission_paused_user_threshold` | Faucet emission paused (user-set threshold) | `normal` | Admin | Threshold details |
| `tick_abort` | Server tick halted (error state) | `critical` | Admin | Error details |

**ADMIN Category:**
| Type | Trigger | Priority | Recipients | Example Payload |
|---|---|---|---|---|
| `role_updated` | Admin/moderator role granted or revoked | `normal` | Admin | `{ username, role }` |
| `profile_frozen` | Account frozen by admin | `normal` | Admin | Account details |
| `profile_unfrozen` | Account unfrozen by admin | `normal` | Admin | Account details |
| `profile_deleted` | Account permanently deleted | `normal` | Admin | Deletion details |

---

### D. Priority Levels

**Defined Constants:**

```go
const (
    NotificationPriorityNormal   = "normal"    // Default
    NotificationPriorityHigh     = "high"      // Warn-level importance
    NotificationPriorityCritical = "critical"  // Urgent, forces ack
)
```

**Mapping to UI Level:**
- `normal` → UI level `info`
- `high` → UI level `warn`
- `critical` → UI level `urgent`

**Auto-Ack (forced):** All `critical` notifications automatically require acknowledgment (`ack_required=true`).

---

### E. Retention & Expiry

**Default Retention Window:** 48 hours (env var `NOTIFICATION_RETENTION_HOURS`, default 48)

**Expiry by Priority:**
- `critical` → 14 days
- `high` → 7 days
- `normal` → 48 hours (default retention window)

**Pruning:** Background goroutine runs every 30 minutes, deletes:
- Notifications where `expires_at < NOW()`
- Notifications where `expires_at IS NULL` and `created_at < cutoff`
- Cascades deletes to `notification_reads`, `notification_acks`, `notification_deletes`

---

### F. Recipient & Access Control

**Recipient Roles:**
```go
const (
    NotificationRolePlayer    = "player"     // Regular players
    NotificationRoleModerator = "moderator"  // Moderators
    NotificationRoleAdmin     = "admin"      // Admins
)
```

**Access Logic (in `notificationAccessSQL`):**

Notifications are visible if:
1. **Modern targeting** (`recipient_role` NOT NULL):
   - `recipient_role` matches account role (single value visibility)
   - Admins see all role notifications except those targeting specific accounts
   
2. **Legacy targeting** (`target_role` NOT NULL):
   - `target_role = 'all'` → visible to all roles
   - `target_role = 'admin'` → visible to admins
   - `target_role = 'moderator'` → visible to admins + moderators
   - `target_role = 'user'/'player'` → visible to everyone

3. **Account-specific:**
   - Always visible to recipient (`recipient_account_id = $1`)
   - Can also target by account (`account_id = $1`)

---

### G. Deduplication

**Mechanism:** 
- Optional `dedupe_key` + `dedupe_window` (duration)
- On insert, avoids duplicate notifications within time window
- Implementation: Check existing notification with same `dedupe_key` and `created_at > (NOW() - window)`

**Examples in Code:**
- `bulk_star_purchase:{accountId}:{quantity}` with 10-minute window
- `market_pressure_spike` with 60-minute window (or 10-minute for clamps)
- `economy_invariant_violation` with 10-minute window

---

### H. Delivery Guarantees

**Best-Effort:**
- Notifications are emitted asynchronously via `emitNotification()` (fire-and-forget goroutine)
- No guaranteed delivery
- Missed notifications available on next login via `/notifications` endpoint

**Channels:**
1. **In-App Notification Feed** (polling)
   - GET `/notifications` - fetch list
   - Paginated with `after` parameter
   
2. **SSE Streaming** (push)
   - GET `/notifications/stream` - Server-Sent Events
   - Long-lived connection with 3-second poll interval
   - 25-second heartbeat ping
   - Automatic reconnect support via `Last-Event-ID` header

**No Email/SMS:** Email infrastructure optional; Alpha uses in-app only.

---

### I. Ack/Delete Semantics

**Read State:**
- Implicit on first fetch (no dedicated read-mark action)
- Stored in `notification_reads` for tracking

**Acknowledgment State:**
- Explicit via POST `/notifications/ack` with array of IDs
- Required for:
  - Notifications with `ack_required=true`
  - All `critical` priority notifications
- Stored in `notification_acks` with timestamp

**Delete State:**
- Soft delete via POST `/notifications/delete` with array of IDs
- Does not remove notification; marks in `notification_deletes`
- Prevents re-appear in UI (filtered out in fetch queries)
- Original notification remains for audit

**Cascading Deletes:**
- When account is frozen: `delete from notification_reads where account_id = $1`
- When account is unfrozen: full notification history cleared

---

### J. Settings & Preferences

**Endpoint:** `GET/POST /notifications/settings`

**Per-Category Toggle:**
- `enabled` (boolean) - Show in-app?
- `push_enabled` (boolean) - Allow browser push notifications?

**Frontend Implementation:**
```javascript
notificationSettings = {
  "economy": { enabled: true, pushEnabled: false },
  "system": { enabled: true, pushEnabled: true },
  "admin": { enabled: true, pushEnabled: false },
  // ... one per category
}
```

**Browser Push Support:**
- Calls `Notification.requestPermission()` if not granted
- Checks `Notification.permission === 'granted'`
- Suppresses push if user is active (`!document.hidden && notificationPanelOpen`)
- Uses `new Notification()` API (Web Notifications)

**Persistence:**
- Debounced auto-save to `/notifications/settings` (500ms delay)
- Per-account in `notification_settings` table

---

### K. Admin Notification Creation

**Endpoint:** `POST /admin/notifications`

**Request Schema:**
```json
{
  "recipientRole": "player|moderator|admin|all",
  "targetRole": "...",           // Legacy fallback
  "recipientAccountId": "uuid",  // Single-player target (optional)
  "accountId": "uuid",           // Legacy fallback
  "seasonId": "uuid",            // Optional context
  "category": "system",          // Required, normalized to valid category
  "type": "custom_event",        // Free-form event type
  "priority": "high",            // normal|high|critical
  "level": "urgent",             // Legacy; mapped to priority
  "message": "Notification text", // Required
  "link": "/admin.html",         // Optional redirect
  "expiresAt": "2026-02-10T...", // RFC3339, optional
  "ackRequired": true,           // Force ack?
  "payload": {}                  // Custom data
}
```

**Behavior:**
- If `recipientRole` empty, uses `targetRole` (legacy compat)
- If `category` empty, defaults to `NotificationCategoryAdmin`
- If `recipientAccountId` empty, uses `accountId`
- If `role = 'all'`, broadcasts to all three roles (player, moderator, admin)
- Logs audit action: `notification_create` scope `notification`

**Response:**
```json
{ "ok": true }
```

---

### L. SSE Stream Implementation

**Endpoint:** `GET /notifications/stream`

**Features:**
- Long-lived HTTP connection (no timeout)
- Requires authentication
- HTTP/1.1 compatible
- 3-second polling interval for new notifications
- 25-second heartbeat pings (`:ping`)

**Protocol Details:**
```
id: 42\n
event: notification\n
data: { "id": 42, "message": "...", "category": "...", ... }\n
\n
```

**Headers:**
- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`
- `X-Accel-Buffering: no` (disable nginx buffering)

**Client Support:**
- Supports `Last-Event-ID` header for reconnection
- Supports `/notifications/stream?after=123` for catchup
- EventSource API on client side

**Fetch Strategy:**
- After each event, updates `lastNotificationId`
- Next poll fetches notifications with `id > lastID`
- Ascending order in stream

---

### M. UI Integration (Frontend)

**Notification Categories List:**
```javascript
const notificationCategories = [
  "economy", "player_action", "market", "abuse", "system", "admin"
];
```

**UI Panels:**
1. **Notification List Panel**
   - Max 60 notifications default (configurable)
   - Sorted newest → oldest
   - Shows `message`, `category`, `priority` badge

2. **Settings Panel**
   - Toggle in-app per category
   - Toggle push per category
   - Request browser permission on push enable

3. **Notification Details**
   - Full payload rendered if present
   - Ack button for `ack_required` or `critical`
   - Delete button for soft-delete

---

## III. SYSTEM MECHANICS SUMMARY

### Notification Emission Flow

```
Code calls emitNotification(db, NotificationInput)
  ↓
Async goroutine runs insertNotification(db, input)
  ↓
Normalize role, category, priority
  ↓
Set expiry based on priority + retention window
  ↓
Check/apply deduplication (optional)
  ↓
Insert into notifications table
  ↓
Broadcast to SSE streams (if active)
  ↓
New fetch from polling clients
  ↓
UI renders with appropriate styling/behavior
```

### Admin Audit Log Flow

```
Admin action executed (e.g., role_update)
  ↓
Code calls logAdminAction(db, adminID, actionType, scopeType, scopeID, reason, details)
  ↓
Validation: all required fields non-empty
  ↓
JSON-encode details map
  ↓
INSERT INTO admin_audit_log (...)
  ↓
Immutable record stored forever
  ↓
Searchable via /admin/audit-log endpoint
```

---

## IV. MISSING / AMBIGUOUS AREAS

1. **Dedup Implementation:** Exact query logic for `DISTINCT ON` or `WHERE NOT EXISTS` not shown in code search
2. **Telemetry:** Notification emissions sent to telemetry if feature flag enabled
3. **Retention Enforcement:** Manual pruning via `pruneNotifications()` every 30 min; no automated cleanup
4. **Push Notification Backend:** Only browser `Notification` API used; no service worker or backend push service
5. **Admin Audit Indexing:** No explicit indexes on `admin_audit_log`; searches are full table scan
6. **Notification Ordering in Stream:** Ascending by ID; doesn't use timestamp
7. **Cascade on Delete:** Account deletion cascades to `notification_reads` but not `notifications` themselves

---

## V. QUICK REFERENCE

### Action Types (13 total)
`auto_admin_bootstrap`, `admin_bootstrap_claim`, `role_update`, `season_control_set`, `season_advance`, `season_recovery`, `notification_create`, `bot_toggle`, `bot_create`, `bot_delete`, `profile_freeze`, `profile_unfreeze`, `profile_delete`

### Scope Types (5 total)
`account`, `season`, `season_control`, `notification`, `player`

### Notification Categories (6 total)
`economy`, `player_action`, `market`, `abuse`, `system`, `admin`

### Priority Levels (3 total)
`normal`, `high`, `critical`

### Recipient Roles (3 total)
`player`, `moderator`, `admin`

---

**End of Research Document**
