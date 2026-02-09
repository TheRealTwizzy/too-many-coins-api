# Admin Audit Log

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

# Admin Action Audit Trail

All administrative actions are logged to an **append-only, immutable audit log** for accountability and recovery.

---

## Action Type Taxonomy

| Action Type | Scope Type | When Triggered | Details Captured |
|------------|-----------|----------------|-----------------|
| **auto_admin_bootstrap** | `account` | Server startup creates first admin | `{"username": "alpha-admin", "autoCreated": true}` |
| **admin_bootstrap_claim** | `account` | Owner claims bootstrap with code | `{"username": "...", "claimedAt": "..."}` |
| **role_update** | `account` | Admin changes user role | `{"username": "...", "oldRole": "player", "newRole": "admin"}` |
| **profile_freeze** | `account` | Admin freezes account | `{"username": "...", "reason": "..."}` |
| **profile_unfreeze** | `account` | Admin unfreezes account | `{"username": "...", "reason": "..."}` |
| **profile_delete** | `account` | Admin deletes account (reserved) | `{"username": "...", "reason": "..."}` |
| **season_control_set** | `season_control` | Admin sets season control (emergency) | `{"controlName": "pause_purchases", "value": "true", "expiresAt": "..."}` |
| **season_advance** | `season` | Admin manually advances season | `{"oldSeasonId": "...", "newSeasonId": "...", "reason": "..."}` |
| **season_recovery** | `season` | Admin creates recovery season (Alpha) | `{"newSeasonId": "...", "reason": "...", "confirm": "..."}` |
| **notification_create** | `notification` | Admin sends broadcast notification | `{"category": "system", "priority": "high", "targetRole": "all", "message": "..."}` |
| **bot_toggle** | `player` | Admin enables/disables bot | `{"playerId": "...", "isBot": true}` |
| **bot_create** | `player` | Admin creates test bot | `{"username": "bot_1", "profile": "threshold_buyer"}` |
| **bot_delete** | `player` | Admin deletes bot account | `{"playerId": "...", "username": "bot_1"}` |

---

## Scope Type Taxonomy

| Scope Type | Meaning | Scope ID Format |
|-----------|---------|----------------|
| **account** | Action affects user account | Account UUID or username |
| **season** | Action affects season lifecycle | Season UUID |
| **season_control** | Emergency season control switch | Season UUID |
| **notification** | Notification creation/broadcast | Notification ID or "broadcast" |
| **player** | Action affects player state | Player UUID |

---

## Database Schema

```sql
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id BIGSERIAL PRIMARY KEY,
    admin_account_id TEXT NOT NULL,
    action_type TEXT NOT NULL,
    scope_type TEXT NOT NULL,
    scope_id TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    details JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_log_created 
    ON admin_audit_log (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_admin 
    ON admin_audit_log (admin_account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_audit_log_action 
    ON admin_audit_log (action_type, created_at DESC);
```

---

## Log Entry Structure

**Example Entry:**
```json
{
  "id": 1234,
  "adminAccountId": "admin-uuid",
  "adminUsername": "alpha-admin",
  "actionType": "role_update",
  "scopeType": "account",
  "scopeId": "player-uuid",
  "reason": "Promoting moderator to admin for testing",
  "details": {
    "username": "player1",
    "oldRole": "moderator",
    "newRole": "admin",
    "effectiveAt": "2026-02-09T12:34:56Z"
  },
  "createdAt": "2026-02-09T12:34:56Z"
}
```

---

## Action Type Details

### Account Management Actions

#### `auto_admin_bootstrap`
**When:** Server startup detects no admin exists and bootstrap not sealed  
**Triggered by:** Automated system process (no human admin)  
**Details:**
```json
{
  "username": "alpha-admin",
  "autoCreated": true,
  "mustChangePassword": true
}
```

**Immutability:** Cannot be deleted or modified; proves ownership chain

---

#### `admin_bootstrap_claim`
**When:** Owner enters valid claim code at `/admin/bootstrap/claim`  
**Triggered by:** Human owner action (pre-bootstrap phase)  
**Details:**
```json
{
  "username": "alpha-admin",
  "claimedAt": "2026-02-09T12:00:00Z",
  "ipAddress": "203.0.113.45"
}
```

**Security:** Claim code NOT logged (sensitive); only timestamp and IP recorded

---

#### `role_update`
**When:** Admin changes user role via `/admin/role`  
**Triggered by:** Admin action  
**Details:**
```json
{
  "username": "player1",
  "oldRole": "player",
  "newRole": "moderator",
  "effectiveAt": "2026-02-09T12:34:56Z"
}
```

**Scope ID:** Target account ID

---

#### `profile_freeze` / `profile_unfreeze`
**When:** Admin suspends or restores account  
**Triggered by:** Admin action (future feature, reserved in schema)  
**Details:**
```json
{
  "username": "player1",
  "reason": "Multiple ToS violations",
  "duration": "7 days",
  "autoUnfreezeAt": "2026-02-16T12:34:56Z"
}
```

---

#### `profile_delete`
**When:** Admin permanently deletes account  
**Triggered by:** Admin action (reserved, not yet implemented)  
**Details:**
```json
{
  "username": "player1",
  "reason": "GDPR deletion request",
  "deletedAt": "2026-02-09T12:34:56Z",
  "confirmedBy": "admin-uuid"
}
```

**Immutability:** Cannot delete audit log entry even after account deletion

---

### Season Control Actions

#### `season_control_set`
**When:** Admin sets emergency season control switch  
**Triggered by:** Admin action (Alpha: read-only, controls not exposed)  
**Details:**
```json
{
  "controlName": "pause_purchases",
  "value": "true",
  "expiresAt": "2026-02-09T18:00:00Z",
  "reason": "Emergency stop due to exploit detection"
}
```

**Available Controls (reserved for post-alpha):**
- `pause_purchases` — Disable all star buying
- `reduce_emission` — Lower daily emission target
- `freeze_season` — Stop all economic activity

---

#### `season_advance`
**When:** Admin manually advances to next season via `/admin/seasons/advance`  
**Triggered by:** Admin action (Alpha recovery only)  
**Details:**
```json
{
  "oldSeasonId": "season-uuid-1",
  "newSeasonId": "season-uuid-2",
  "reason": "Manual advance for testing",
  "autoAdvance": false,
  "economySnapshotTaken": true
}
```

**Scope ID:** Old season UUID

---

#### `season_recovery`
**When:** Admin creates recovery season (Alpha emergency use)  
**Triggered by:** Admin action via `/admin/seasons/recovery`  
**Details:**
```json
{
  "newSeasonId": "season-uuid-3",
  "reason": "Database corruption recovery",
  "confirm": "I understand this is a recovery action",
  "economyReset": true,
  "playerDataPreserved": false
}
```

**Requirements:**
- Phase must be Alpha
- Explicit confirmation text required
- Reason field mandatory

---

### Notification Actions

#### `notification_create`
**When:** Admin creates notification via `/admin/notifications`  
**Triggered by:** Admin action  
**Details:**
```json
{
  "category": "system",
  "priority": "high",
  "title": "Maintenance Window",
  "message": "Server will restart at 3pm UTC for updates",
  "targetRole": "all",
  "targetPlayers": ["uuid1", "uuid2"],
  "expiresAt": "2026-02-09T15:00:00Z"
}
```

**Target Options:**
- `targetRole: "all"` → Broadcast to all players
- `targetRole: "admin"` → Only admins
- `targetPlayers: [...]` → Specific player list

---

### Bot Management Actions

#### `bot_toggle`
**When:** Admin enables/disables bot flag on player  
**Triggered by:** Admin action (via `/admin/player-controls` or global settings)  
**Details:**
```json
{
  "playerId": "bot-uuid",
  "username": "bot_alpha_01",
  "isBot": true,
  "previousValue": false
}
```

---

#### `bot_create`
**When:** Admin creates test bot via `/admin/bots/create`  
**Triggered by:** Admin action  
**Details:**
```json
{
  "playerId": "bot-uuid",
  "username": "bot_alpha_01",
  "profile": "threshold_buyer",
  "threshold": 500,
  "maxStarsPerDay": 50
}
```

---

#### `bot_delete`
**When:** Admin deletes bot account via `/admin/bots/delete`  
**Triggered by:** Admin action  
**Details:**
```json
{
  "playerId": "bot-uuid",
  "username": "bot_alpha_01",
  "deletedAt": "2026-02-09T12:34:56Z"
}
```

---

## Querying the Audit Log

### `/admin/audit-log` Endpoint

**Request:**
```
GET /admin/audit-log?limit=50&offset=0&search=player1
```

**Query Parameters:**
- `limit` (default 50, max 200) — Number of entries to return
- `offset` (default 0) — Pagination offset
- `search` (optional) — Full-text search across admin username, action type, scope ID, reason

**Response:**
```json
{
  "ok": true,
  "items": [
    {
      "id": 1234,
      "adminAccountId": "admin-uuid",
      "adminUsername": "alpha-admin",
      "actionType": "role_update",
      "scopeType": "account",
      "scopeId": "player-uuid",
      "reason": "Promoting moderator",
      "details": { /* JSON */ },
      "createdAt": "2026-02-09T12:34:56Z"
    }
  ],
  "total": 1457,
  "limit": 50,
  "offset": 0
}
```

---

### Search Behavior

**Full-text search matches:**
- Admin username (case-insensitive)
- Action type (exact or partial match)
- Scope ID (exact or partial match)
- Reason field (case-insensitive substring)

**Example Queries:**
```
?search=bootstrap           → All bootstrap-related actions
?search=player1             → All actions affecting player1
?search=season_advance      → All manual season advances
?search=alpha-admin         → All actions by alpha-admin
```

---

## Retention & Immutability

### Append-Only Guarantee

- **No UPDATE queries** — Actions cannot be modified after creation
- **No DELETE queries** — Actions cannot be removed (including account deletions)
- **Monotonic IDs** — Sequential ID proves chronological order

### Retention Policy

| Environment | Policy | Rationale |
|-----------|--------|-----------|
| **Alpha** | Indefinite (no purge) | Short-lived seasons; minimal volume |
| **Beta** | Indefinite (no purge) | Compliance audit trail |
| **Release** | 2+ years | Legal/compliance requirements; DB archival after 2 years |

**Future Archival:**
- Compress old entries to cold storage (S3, etc.)
- Keep last 90 days in hot DB for admin UI
- Restore from archive on-demand for investigations

---

## Security & Access Control

### Who Can View Audit Log

| Role | Access |
|------|--------|
| **Admin** | Full access (all entries) |
| **Moderator** | No access (reserved for post-alpha) |
| **Player** | No access |

### Who Can Create Entries

- **Only admins** can create entries directly (via admin actions)
- **System** can create bootstrap entries (`auto_admin_bootstrap`)
- **No API** for arbitrary entry creation (prevents log forgery)

---

## Design Philosophy

### Accountability
- Every admin action logged with reason and timestamp
- Cannot be hidden or retroactively modified
- Proves ownership chain from bootstrap to current state

### Recovery
- Audit log enables rollback/investigation of admin errors
- Details field captures before/after snapshots
- Season recovery actions preserve economic state snapshots

### Compliance
- GDPR-compliant: logs retained even after account deletion
- Immutable trail for legal/regulatory review
- Export functionality (future: CSV/JSON dump)

---

## Technical Notes

**Performance:**
- Indexes on `created_at`, `admin_account_id`, `action_type` for fast queries
- JSONB details field enables flexible querying without schema changes
- Append-only → no row locking or UPDATE contention

**Storage:**
- ~500 bytes per entry (average)
- 10,000 entries ≈ 5 MB
- Alpha (14-day season, ~500 admin actions) ≈ 250 KB

**Edge Cases:**
- Bootstrap entries have `admin_account_id = system` (no human admin yet)
- Deleted accounts: `admin_account_id` preserved, `adminUsername` shows account_id if lookup fails
- Failed actions NOT logged (only successful operations recorded)

---

## See Also

- [Admin Tools](admin-tools.md) — Admin governance and role philosophy
- [HTTP API Reference](http-api-reference.md) — `/admin/audit-log` endpoint spec
- [Notifications](notifications.md) — Admin notification creation mechanics
