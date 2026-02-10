# Release Plan

Purpose: This document is the internal hub for the full release roadmap. It links the website experience plan, anti-cheat/abuse operations, and subsystem readiness criteria so we can move from Beta to Release with clear, testable gates.

## Document Index (Existing Canon)
- SPEC and phases: README/SPEC.md
- Alpha goals and exit criteria: README/alpha-execution.md
- Frontend MVP contract: README/frontend-mvp.md
- Admin tools contract: README/admin-tools.md
- Anti-abuse philosophy and event registry: README/anti-abuse.md
- API surface: README/http-api-reference.md
- Subsystems: README/notifications.md, README/season-end.md, README/between-seasons.md, README/server-authority.md, README/settings.md, README/communication.md
- Gaps audit: DOCUMENTATION_GAPS.md

## Release Roadmap (Alpha -> Beta -> Release)

### Alpha Exit Criteria (confirm or update)
Owner: TBD (Engineering lead)

Checklist:
- Core gameplay loop stable (daily/activity claims, market actions, season cycle).
- Admin tools operate in read-only mode with recovery-only controls.
- Anti-abuse event registry implemented with server-side enforcement logging.
- Bug reporting and notification paths function end-to-end.
- No whitelist requirement in Alpha; throttles only.

### Beta Goals
Owner: TBD (Design + Engineering)

Checklist:
- Gameplay economy balances within acceptable variance under load testing.
- Anti-cheat operations are actionable: escalation path, admin review panels, audit logging.
- Moderator tooling is defined, limited in scope, and auditable.
- Player experience and onboarding are complete with consistent error handling.
- Notification taxonomy is consistent and documented.

### Release Gates
Owner: TBD (Product)

Checklist:
- All readiness criteria in the Subsystem Readiness Matrix are met.
- Admin and moderator workflows cover the full incident response loop.
- Player experience is complete, accessible, and stable on desktop and mobile.
- Documentation gaps are resolved or explicitly deferred with owners and dates.
- Operational runbooks and on-call handoffs are finalized.

## Website Plan (Information Architecture by Role)

### Player Experience
Owner: TBD (Frontend)
Primary surfaces and required information:
- Landing, signup, login, and profile management.
- Season dashboard with balances, daily and activity claim status, notifications, and recent events.
- Leaderboards and market context summaries.
- Bug report entry point and status feedback.

Player action coverage:
- Claim daily and activity rewards.
- View season timing, rules, and end-of-season outcomes.
- Manage notification preferences.

### Moderator Experience
Owner: TBD (Ops)
Scope and visibility:
- Profile lookup and limited profile edits only (no economy edits).
- Abuse reports triage and tagging for admin escalation.
- Strict audit trail for all moderator actions.

Moderator action coverage:
- Resolve reports with notes.
- Flag accounts for admin review.
- Apply only approved moderation actions (to be defined in the capability matrix).

### Admin Experience
Owner: TBD (Ops)
Core panels and data surfaces:
- Season operations and emergency controls.
- Anti-cheat and abuse monitoring dashboard.
- Player account diagnostics (read-heavy, write-light).
- Audit log and notifications.

Admin action coverage:
- Emergency season controls with audit logging.
- Manual account review tools.
- Anti-cheat enforcement actions with escalation levels and reasons.

## Anti-Cheat / Abuse Operations Plan

### Principles
Owner: TBD (Security)
- Server authority over all state changes.
- Enforcement ladder with escalating responses.
- Full auditability for any intervention.

### Signals
- Event registry from README/anti-abuse.md is the baseline.
- Add anomaly detection for:
  - Rapid claim frequency.
  - Unusual market activity patterns.
  - Multi-account correlation on the same IP range.

### Workflow
- Detection: server emits abuse events with severity and score deltas.
- Triage: admin dashboard groups by account, IP cluster, and season.
- Action: apply throttles, freezes, or bans only for extreme cases.
- Audit: all actions logged with reason and operator.
- Notify: player-facing notices for account freezes or bans only.
- Detection -> Triage -> Action -> Audit -> Player notification.
- All actions are logged with operator, reason, and time.
- Escalation paths are defined for repeat offenders.

## Subsystem Readiness Matrix

Each subsystem must define a beta-ready and release-ready checklist with owners.

- Economy and market pressure
  - Beta: Stable under load and tuned within target variance.
  - Release: Sustained balance across multiple seasons.

- Seasons and between-seasons
  - Beta: Season end flow and rewards are correct.
  - Release: Between-seasons progression is complete and tested.

- Notifications
  - Beta: All critical notifications delivered.
  - Release: Category taxonomy is consistent and documented.

- Admin and moderation
  - Beta: Admin recovery actions are safe and auditable.
  - Release: Moderator capability matrix enforced and logged.
  - Owner: TBD (Ops)

- Anti-cheat / abuse
  - Beta: Core event registry wired to detection panels.
  - Release: Escalation ladder and automation thresholds are validated.

## Inconsistency Resolutions

- Whitelist policy: Whitelist is deprecated in Alpha; UI must show post-alpha placeholder only.
- Admin bootstrap: ENV-seeded bootstrap is canonical; claim-code flow is disabled by default.
- Moderator scope: Capability matrix and audit trail live in admin-tools and release plan docs.
- Documentation gaps: Refresh DOCUMENTATION_GAPS.md and note resolved items.

## Beta -> Release Checklist (Operational)

- Load testing and stability pass.
- Admin and moderator runbooks finalized.
- Anti-cheat escalation ladder validated with dry runs.
- Player experience tested on mobile and desktop.
- Telemetry and audit logs reviewed for completeness.
- Incident response simulation completed.

## Ownership and Next Steps

Assign owners to each subsection and list target dates for:
- Beta readiness review
- Release readiness review
- Post-release monitoring plan
- First public release candidate
