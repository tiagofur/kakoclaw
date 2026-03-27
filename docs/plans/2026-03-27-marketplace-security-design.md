# Marketplace Security: AI Scanning + Skill Revocation + User Alerts

**Date:** 2026-03-27
**Status:** Approved

---

## Problem

The marketplace currently has a partial security scanner (security_score, SecurityFindings fields exist) but:
- No clear auto-reject thresholds based on score
- No way to disable a skill after it's been approved
- No user notification when an installed skill is later found to be malicious
- Notification would only appear if user visits the Skills page — insufficient

---

## Goals

1. AI scan at submission time with clear score thresholds (auto-reject low scores)
2. Admin can disable an approved skill post-publication
3. When a skill is disabled, security_score is set to 0 (unambiguous signal)
4. Users who have a disabled skill installed see an app-wide banner (like the LLM banner)
5. The disabled skill is visually flagged in the Skills view so the user knows which one to remove
6. Banner disappears once the user uninstalls all disabled skills

---

## Design

### Security Score Thresholds

| Score | Action |
|-------|--------|
| 75–100 | `approved` (or `needs_review` if specific findings exist) |
| 50–74 | `needs_review` — admin must manually approve |
| 0–49 | `rejected` — automatic, never reaches admin queue |
| 0 (forced) | Set by admin when disabling an approved skill |

Score 0 means **actively dangerous** — either caught by AI upfront or discovered post-approval.

---

### Backend Changes

#### 1. `skill_submissions` schema — new `disabled` status + `disabled_reason`

Add to the existing `status` enum values: `'disabled'`
Add column: `disabled_reason TEXT DEFAULT ''`

Migration: `ALTER TABLE skill_submissions ADD COLUMN disabled_reason TEXT DEFAULT ''`

#### 2. Admin disable endpoint

```
POST /api/v1/admin/submissions/{id}/disable
Body: { "reason": "string" }
```

- Sets `status = 'disabled'`
- Sets `security_score = 0`
- Sets `disabled_reason = reason`
- Sets `updated_at = NOW()`

#### 3. Security alerts endpoint

```
GET /api/v1/marketplace/security-alerts
```

Returns list of slugs with `status = 'disabled'`:
```json
{ "disabled_slugs": ["bad-skill", "another-bad-skill"] }
```

No auth required (public endpoint — slugs are not sensitive).
Can be cached with short TTL (e.g. 5 minutes) if needed.

#### 4. `/api/v1/config/status` — add `securityAlert`

Existing endpoint already called on every app load by `configStore`.
Add logic: cross-reference `disabled_slugs` against user's installed skills (filesystem scan).
Add to response:
```json
{
  "configured": true,
  "securityAlert": true
}
```

`securityAlert = true` when user has ≥1 installed skill whose marketplace slug is disabled.

---

### Frontend Changes

#### 1. `configStore.js`

- Consume `securityAlert` from `/config/status` response
- Expose as `configStore.securityAlert` (reactive ref)

#### 2. `SecurityAlertBanner.vue` (new component, based on `DegradedModeBanner`)

- Shown in `MainLayout.vue` when `configStore.securityAlert === true`
- Non-dismissible (security alerts should not be ignorable)
- Message: "⚠ Una o más skills instaladas fueron desactivadas por razones de seguridad. Revisá tus skills y desinstalá las marcadas."
- CTA button: "Revisar Skills" → routes to `/skills`
- Visual: amber/orange color (distinct from the red LLM banner)

#### 3. Skills view — disabled badge

In the installed skills list, skills with `status = 'disabled'` in the marketplace receive:
- Red "Deshabilitada" badge
- Warning message: "Esta skill fue desactivada por razones de seguridad. Se recomienda desinstalarla."
- Uninstall button highlighted/prominent

Detection: after loading installed skills AND security alerts, cross-reference slugs client-side.

---

### Data Flow

```
App startup
  → GET /api/v1/config/status
  → response includes securityAlert: bool
  → if true: SecurityAlertBanner shown in MainLayout (app-wide)

User navigates to /skills
  → installed skills loaded
  → GET /api/v1/marketplace/security-alerts (list of disabled slugs)
  → cross-reference locally
  → disabled skills get warning badge

User uninstalls disabled skill
  → DELETE /api/v1/skills/{name}
  → configStore refreshes /config/status
  → if no more disabled skills: securityAlert = false → banner disappears
```

---

### What Is NOT in Scope

- Email/push notifications (out of scope for now)
- Per-user install tracking table (not needed for the generic approach)
- Showing which specific skill is mentioned in the banner (intentionally generic)
- Forced auto-uninstall (user must take action manually)

---

## Files to Modify

| File | Change |
|------|--------|
| `pkg/storage/central.go` | Add `disabled_reason` field to schema + `SkillSubmission` struct; add `DisableSkill()` method |
| `pkg/web/handlers_marketplace.go` | Add `handleAdminDisableSkill`; update security scan thresholds; add `handleSecurityAlerts` endpoint |
| `pkg/web/server.go` | Register new routes; update `handleConfigStatus` to include `securityAlert` |
| `pkg/web/frontend/src/stores/configStore.js` | Add `securityAlert` field + expose it |
| `pkg/web/frontend/src/components/SecurityAlertBanner.vue` | New component |
| `pkg/web/frontend/src/components/Layout/MainLayout.vue` | Include `SecurityAlertBanner` |
| `pkg/web/frontend/src/views/SkillsView.vue` | Show disabled badge + warning on affected skills |
