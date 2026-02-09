# Notifications

## Scope
- **Type:** System contract
- **Status:** Canonical (Game Bible)
- **Owner:** Game Bible
- **Alpha Status:** Fully implemented

## Change Protocol
- Update alongside related systems and TODO entries in the same logical unit.
- Preserve cross-file invariants defined in README.md.
- Prefer additive clarifications; flag any breaking change explicitly.

---

# Notification System

In-app notification system for **economy events**, **player actions**, **market pressure**, **abuse detection**, **system events**, and **admin broadcasts**.

---

## Category Taxonomy

| Category | Purpose | Typical Priority | Example Triggers |
|---------|---------|------------------|------------------|
| **economy** | Economy thresholds, bulk purchases, price milestones | Normal–High | Bulk star purchase (≥3), emission threshold breach, price milestone, economy violation |
| **player_action** | Individual player milestones | Normal | First login, season join, milestone achievement |
| **market** | Market pressure spikes/recovery | High–Critical | Pressure spike (≥1.7 critical), pressure recovery below threshold |
| **abuse** | Abuse detection events | High–Critical | Severity 2 detection (high), severity 3 detection (critical, auto-escalate) |
| **system** | Season lifecycle, emission, tick health | Normal–Critical | Season start/end, emission pause, tick abort/recovery |
| **admin** | Admin actions affecting players | High | Role update, profile freeze/unfreeze, account deletion |

---

## Priority Levels & Auto-Expiry

| Priority | Retention | Auto-Expiry | Forced Ack | Use Case |
|---------|-----------|-------------|-----------|----------|
| **normal** | 48 hours | ✅ Yes | ❌ No | Routine events, minor milestones |
| **high** | 7 days | ✅ Yes | ❌ No | Market pressure, bulk purchases, moderate abuse |
| **critical** | 14 days | ✅ Yes | ✅ **Required** | Tick aborts, severe abuse, season recovery, admin broadcasts |

**Retention Configuration:**
```env
NOTIFICATION_RETENTION_HOURS=48  # normal priority default
```

**Auto-Expiry Behavior:**
- Expired notifications soft-deleted (hidden from queries, preserved in DB)
- Unread critical notifications block certain actions until acknowledged
- Admins can override default retention per notification

---

## Notification Structure

**Database Schema:**
```sql
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    player_id TEXT NOT NULL,
    category TEXT NOT NULL,
    priority TEXT NOT NULL DEFAULT 'normal',
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    details JSONB NOT NULL DEFAULT '{}',
    read BOOLEAN NOT NULL DEFAULT FALSE,
    acknowledged BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notifications_player 
    ON notifications (player_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_unread 
    ON notifications (player_id, read) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_priority 
    ON notifications (player_id, priority) WHERE deleted_at IS NULL;
```

**Example Entry:**
```json
{
  "id": 1234,
  "playerId": "player-uuid",
  "category": "economy",
  "priority": "high",
  "title": "Bulk Star Purchase Detected",
  "message": "You purchased 5 stars in quick succession. This may affect market pricing.",
  "details": {
    "starCount": 5,
    "totalCost": 2450,
    "newBalance": 3200,
    "priceImpact": "+2.3%"
  },
  "read": false,
  "acknowledged": false,
  "deletedAt": null,
  "createdAt": "2026-02-09T12:34:56Z",
  "expiresAt": "2026-02-16T12:34:56Z"
}
```

---

## Category Triggers

### Economy Category
**Trigger Events:**
- **Bulk Star Purchase (≥3 stars)** — Player buys 3+ stars in single transaction
- **Emission Threshold Breach** — Current emission exceeds target by ≥10%
- **Price Milestone** — Price crosses 500, 1000, 2000 thresholds
- **Economy Violation** — Admin manually flags economy anomaly

**Example Notification:**
```json
{
  "category": "economy",
  "priority": "normal",
  "title": "Price Milestone Reached",
  "message": "Star price has crossed 1000 coins.",
  "details": {
    "oldPrice": 987,
    "newPrice": 1023,
    "milestone": 1000
  }
}
```

---

### Player Action Category
**Trigger Events:**
- **First Login** — Player's first authenticated session
- **Season Join** — Player joins new season (first claim/purchase)
- **Milestone Achievement** — Star count milestones (10, 50, 100, 500)

**Example Notification:**
```json
{
  "category": "player_action",
  "priority": "normal",
  "title": "First Season Join",
  "message": "Welcome to Season Alpha! Your activity tracking starts now.",
  "details": {
    "seasonId": "alpha-s1",
    "joinedAt": "2026-02-09T12:34:56Z"
  }
}
```

---

### Market Category
**Trigger Events:**
- **Pressure Spike (≥1.7 critical)** — Market pressure reaches critical threshold
- **Pressure Recovery** — Market pressure falls below critical after spike

**Example Notification:**
```json
{
  "category": "market",
  "priority": "critical",
  "title": "Critical Market Pressure Spike",
  "message": "Market pressure reached 1.85 (critical threshold 1.7). Faucet cooldowns extended.",
  "details": {
    "pressureValue": 1.85,
    "threshold": 1.7,
    "cooldownMultiplier": 1.92
  }
}
```

---

### Abuse Category
**Trigger Events:**
- **Severity 2 Detection (high priority)** — Anti-cheat detects score 25–45
- **Severity 3 Detection (critical priority)** — Anti-cheat detects score ≥45, auto-escalates to critical

**Example Notification:**
```json
{
  "category": "abuse",
  "priority": "critical",
  "title": "Severe Abuse Detection",
  "message": "Your account triggered severe anti-cheat flags (score 48). Your earning rate reduced by 40%.",
  "details": {
    "eventType": "purchase_burst",
    "severity": 3,
    "score": 48,
    "enforcementMultipliers": {
      "earning": 0.6,
      "price": 1.3
    }
  }
}
```

**Acknowledgment Requirement:**
- Severity 3 abuse notifications require explicit acknowledgment before further purchases
- Prevents accidental penalty escalation due to unread warnings

---

### System Category
**Trigger Events:**
- **Season Start/End** — Season lifecycle transitions
- **Emission Pause** — Daily emission target reached, purchases paused
- **Tick Abort/Recovery** — Server tick health events

**Example Notification:**
```json
{
  "category": "system",
  "priority": "high",
  "title": "Season Ended",
  "message": "Season Alpha has concluded. Final leaderboard locked. New season starts in 24 hours.",
  "details": {
    "seasonId": "alpha-s1",
    "endedAt": "2026-02-09T23:59:59Z",
    "nextSeasonStartsAt": "2026-02-10T23:59:59Z"
  }
}
```

---

### Admin Category
**Trigger Events:**
- **Role Update** — Admin changes player role (player → moderator → admin)
- **Profile Freeze/Unfreeze** — Admin suspends or restores account
- **Account Deletion** — Admin deletes player account (reserved)

**Example Notification:**
```json
{
  "category": "admin",
  "priority": "high",
  "title": "Role Updated",
  "message": "Your role was updated from player to moderator by alpha-admin.",
  "details": {
    "oldRole": "player",
    "newRole": "moderator",
    "adminUsername": "alpha-admin",
    "updatedAt": "2026-02-09T12:34:56Z"
  }
}
```

---

## Delivery Mechanisms

### Polling (Primary Method)
**Endpoint:** `GET /notifications`

**Request:**
```
GET /notifications?limit=50&offset=0&unread_only=true&category=abuse
```

**Query Parameters:**
- `limit` (default 50, max 200) — Number of notifications to return
- `offset` (default 0) — Pagination offset
- `unread_only` (default false) — Filter to unread notifications only
- `category` (optional) — Filter by category (economy, player_action, market, abuse, system, admin)

**Response:**
```json
{
  "ok": true,
  "notifications": [
    {
      "id": 1234,
      "category": "abuse",
      "priority": "critical",
      "title": "Severe Abuse Detection",
      "message": "Your account triggered severe anti-cheat flags.",
      "details": { /* ... */ },
      "read": false,
      "acknowledged": false,
      "createdAt": "2026-02-09T12:34:56Z",
      "expiresAt": "2026-02-23T12:34:56Z"
    }
  ],
  "total": 12,
  "unreadCount": 3
}
```

**Automatic Read Tracking:**
- Fetching a notification automatically marks it as `read = true`
- Acknowledgment requires explicit `POST /notifications/:id/ack`

---

### SSE Streaming (Real-Time)
**Endpoint:** `GET /notifications/stream`

**Connection Behavior:**
```
GET /notifications/stream HTTP/1.1
Accept: text/event-stream
```

**Event Stream:**
```
data: {"type":"notification","id":1234,"category":"market","priority":"high",...}

data: {"type":"heartbeat"}

data: {"type":"notification","id":1235,"category":"abuse","priority":"critical",...}
```

**Implementation Details:**
- **Poll Interval:** 3 seconds (checks for new notifications)
- **Heartbeat Interval:** 25 seconds (keeps connection alive)
- **Reconnection:** Client auto-reconnects on disconnect (exponential backoff)

**Event Types:**
- `notification` — New notification payload (full JSON structure)
- `heartbeat` — Keep-alive signal (no payload)

---

## Notification Settings

**Per-Category Toggles:**
```sql
CREATE TABLE IF NOT EXISTS notification_settings (
    player_id TEXT NOT NULL,
    category TEXT NOT NULL,
    in_app_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    push_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (player_id, category)
);
```

**Default Settings:**
| Category | In-App | Push (Browser) |
|---------|--------|---------------|
| All categories | ✅ Enabled | ❌ Disabled |

**Settings Endpoint:**
```
PUT /settings/notifications
Body:
{
  "economy": { "in_app_enabled": true, "push_enabled": false },
  "abuse": { "in_app_enabled": true, "push_enabled": true }
}
```

**Push Notifications (Web Notifications API):**
- Requires user permission (browser prompt)
- Only sent if `push_enabled = true` for category
- Delivered via Service Worker (future: FCM/APNs)

---

## Acknowledgment & Deletion

### `/notifications/:id/ack` (POST)
**Purpose:** Explicitly acknowledge critical notifications

**Request:**
```
POST /notifications/1234/ack
```

**Response:**
```json
{
  "ok": true,
  "acknowledged": true,
  "acknowledgedAt": "2026-02-09T13:00:00Z"
}
```

**Required For:**
- Critical priority notifications (severity 3 abuse, tick aborts, admin actions)
- Blocks certain actions until acknowledged (e.g., purchases after abuse detection)

---

### `/notifications/:id` (DELETE)
**Purpose:** Soft-delete notification (hide from queries)

**Request:**
```
DELETE /notifications/1234
```

**Response:**
```json
{
  "ok": true,
  "deletedAt": "2026-02-09T13:00:00Z"
}
```

**Behavior:**
- Sets `deleted_at = NOW()`
- Notification hidden from queries but preserved in DB
- Cannot delete unacknowledged critical notifications (must ack first)

---

## Admin Creation

### `/admin/notifications` (POST)
**Purpose:** Admin broadcasts custom notification

**Request:**
```
POST /admin/notifications
Body:
{
  "category": "system",
  "priority": "high",
  "title": "Maintenance Window",
  "message": "Server will restart at 3pm UTC for updates.",
  "targetRole": "all",
  "targetPlayers": [],
  "expiresIn": "7d"
}
```

**Parameters:**
- `category` (required) — One of: economy, player_action, market, abuse, system, admin
- `priority` (optional, default "normal") — normal, high, critical
- `title` (required) — Notification title (max 100 chars)
- `message` (required) — Notification body (max 500 chars)
- `targetRole` (optional) — Broadcast to role: "all", "admin", "moderator"
- `targetPlayers` (optional) — Array of player IDs (empty + targetRole="all" → everyone)
- `expiresIn` (optional, default "48h") — Duration string (e.g., "7d", "24h", "30m")

**Response:**
```json
{
  "ok": true,
  "created": 1547,
  "preview": {
    "id": 5678,
    "category": "system",
    "priority": "high",
    "title": "Maintenance Window",
    "expiresAt": "2026-02-16T13:00:00Z"
  }
}
```

**Broadcast Behavior:**
- `targetRole: "all"` + empty `targetPlayers` → All players receive notification
- `targetRole: "admin"` → Only admin role receives
- `targetPlayers: [...]` → Only listed players receive (overrides targetRole)

---

## Priority Override

**Admin Priority Control:**
Admins can force critical priority for any notification, bypassing category defaults:

```json
{
  "category": "economy",
  "priority": "critical",  // Forced critical (normally "normal")
  "title": "Emergency: Star Purchases Paused",
  "message": "All star purchases halted due to exploit investigation."
}
```

**Forced Acknowledgment:**
- Critical priority notifications require explicit acknowledgment
- Player cannot dismiss without acknowledging
- Blocks purchases/claims until acknowledged

---

## Technical Notes

### Performance
- **Indexes:** `player_id + created_at DESC`, `player_id + read`, `player_id + priority`
- **Query Optimization:** `WHERE deleted_at IS NULL` filter on all queries (excludes soft-deleted)
- **Expiry Cleanup:** Hourly cron job soft-deletes expired notifications (sets `deleted_at`)

### Storage
- **Average Notification:** ~300 bytes (title + message + details)
- **Alpha Season (14 days, 100 players, 5 notifications/player):** ~150 KB
- **Retention:** 48h–14d (priority-dependent), soft-deleted after expiry

### Edge Cases
- **Unread Critical Notifications:** Block purchases until acknowledged (prevents penalty escalation)
- **Notification Flood:** Max 10 notifications per player per category per hour (spam prevention)
- **Expired Notifications:** Soft-deleted but preserved in DB for audit trail
- **Deleted Player Notifications:** Cascade deleted on account deletion

---

## Design Philosophy

### User Awareness
- Critical events (market pressure, abuse, season end) delivered with high priority
- Auto-expiry ensures notification inbox doesn't grow indefinitely
- Forced acknowledgment prevents accidental penalty escalation

### Admin Control
- Admins can broadcast system-wide notifications (maintenance, events)
- Priority override for emergency communications
- Audit trail via admin_audit_log (action_type: notification_create)

### Privacy
- Notifications stored per-player (no global broadcast table)
- Soft-deletion preserves audit trail
- Settings per-category (granular control)

---

## See Also

- [Anti-Cheat Events](anti-cheat-events.md) — Abuse detection triggers for notifications
- [Market Pressure](market-pressure.md) — Market pressure spike thresholds
- [Admin Tools](admin-tools.md) — Admin notification creation capabilities
- [HTTP API Reference](http-api-reference.md) — `/notifications` endpoint spec
