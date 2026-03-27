# Marketplace Security: AI Scan Thresholds + Skill Revocation + User Alerts

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add auto-reject thresholds to the AI security scanner, let admins disable approved skills post-publication, and surface a non-dismissible app-wide banner when a user has a disabled skill installed.

**Architecture:** Backend adds a `disabled` status + `disabled_reason` field to `skill_submissions`, exposes a `/marketplace/security-alerts` endpoint returning disabled slugs, and enriches `/config/status` with a `securityAlert` bool. Frontend reads `securityAlert` from configStore (already called on app load), shows a new `SecurityAlertBanner` in MainLayout, and marks disabled skills in SkillsView.

**Tech Stack:** Go 1.26 (SQLite migrations via ALTER TABLE), Vue 3 + Pinia, Tailwind CSS.

---

## Task 1: Add `disabled_reason` to DB schema + struct + `DisableSkill()` method

**Files:**
- Modify: `pkg/storage/central.go`

**Step 1: Add the migration**

In `central.go`, find the `migrations` slice (around line 222–226 where the last ALTER TABLE entries are). Add after the last entry:

```go
`ALTER TABLE skill_submissions ADD COLUMN disabled_reason TEXT NOT NULL DEFAULT '';`,
```

**Step 2: Add field to `SkillSubmission` struct**

In `central.go` around line 832, add after the `Dependencies` field:

```go
DisabledReason string `json:"disabled_reason,omitempty"`
```

**Step 3: Update the scan query in `GetSkillSubmissions` / `GetSkillSubmissionBySlug` to include the new field**

Find every `rows.Scan(` call that reads `skill_submissions` rows. Add `&sub.DisabledReason` at the end of each Scan (after `&sub.Dependencies` or the last field). Search for `skill_slug` in central.go to find all query locations.

**Step 4: Add `DisableSkill()` method**

After `IncrementSkillUsageCount` (around line 1117), add:

```go
// DisableSkill sets a skill's status to 'disabled', security_score to 0,
// and records the reason. Used by admins when a post-publication vulnerability is found.
func (cs *CentralStorage) DisableSkill(id int64, reviewerID int64, reason string) error {
	_, err := cs.db.Exec(`
		UPDATE skill_submissions
		SET status = 'disabled',
		    security_score = 0,
		    disabled_reason = ?,
		    reviewer_id = ?,
		    reviewed_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`, reason, reviewerID, id)
	return err
}
```

**Step 5: Add `GetDisabledSkillSlugs()` method**

```go
// GetDisabledSkillSlugs returns the skill_slug of every disabled marketplace skill.
// Used by the security-alerts endpoint and config/status.
func (cs *CentralStorage) GetDisabledSkillSlugs() ([]string, error) {
	rows, err := cs.db.Query(`
		SELECT skill_slug FROM skill_submissions WHERE status = 'disabled'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var slugs []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return nil, err
		}
		slugs = append(slugs, slug)
	}
	return slugs, rows.Err()
}
```

**Step 6: Build to verify no compile errors**

```bash
go build ./pkg/storage/...
```
Expected: no output (success).

**Step 7: Commit**

```bash
git add pkg/storage/central.go
git commit -m "feat(storage): add disabled_reason field + DisableSkill + GetDisabledSkillSlugs"
```

---

## Task 2: Update security scan thresholds in `handleMarketplaceSubmit`

**Files:**
- Modify: `pkg/web/handlers_marketplace.go`

**Context:** The current scanner sets a `securityScore` and `status`. Find the block that assigns the status after scanning (search for `securityScore` in `handlers_marketplace.go`).

**Step 1: Replace the threshold logic**

Find the block that currently decides `status` based on `securityScore`. Replace it with:

```go
// Determine status based on security score thresholds.
//   score < 50  → auto-rejected (never shown in marketplace)
//   score 50-74 → needs_review (admin must approve manually)
//   score >= 75 → approved (or needs_review if high-severity findings exist)
var submissionStatus string
switch {
case securityScore < 50:
    submissionStatus = "rejected"
case securityScore < 75:
    submissionStatus = "needs_review"
default:
    // High score but specific findings may still require human review
    if hasHighSeverityFindings(securityFindings) {
        submissionStatus = "needs_review"
    } else {
        submissionStatus = "approved"
    }
}

// Private skills skip the public review queue regardless of score,
// but still enforce the hard floor.
if body.Visibility == "private" && submissionStatus != "rejected" {
    submissionStatus = "approved"
}
```

**Step 2: Add `hasHighSeverityFindings` helper** (place near the bottom of the file, before the last closing brace):

```go
// hasHighSeverityFindings returns true when any security finding has severity "high" or "critical".
func hasHighSeverityFindings(findingsJSON string) bool {
    var findings []map[string]interface{}
    if err := json.Unmarshal([]byte(findingsJSON), &findings); err != nil {
        return false
    }
    for _, f := range findings {
        sev, _ := f["severity"].(string)
        if sev == "high" || sev == "critical" {
            return true
        }
    }
    return false
}
```

**Step 3: Build**

```bash
go build ./pkg/web/...
```
Expected: no output.

**Step 4: Commit**

```bash
git add pkg/web/handlers_marketplace.go
git commit -m "feat(marketplace): enforce security score thresholds — auto-reject < 50, review 50-74"
```

---

## Task 3: Add admin `disable` action to `handleAdminSubmissionAction`

**Files:**
- Modify: `pkg/web/handlers_marketplace.go`

**Context:** `handleAdminSubmissionAction` is the handler for `/api/v1/admin/submissions/`. It already handles `/approve` and `/reject` suffix. Find the `handleAdminRejectSubmission` function and add a `disable` case alongside it.

**Step 1: Add the route suffix check**

Inside `handleAdminSubmissionAction`, after the block that routes to `handleAdminApproveSubmission` and `handleAdminRejectSubmission`, add:

```go
if strings.HasSuffix(path, "/disable") {
    s.handleAdminDisableSkill(w, r)
    return
}
```

**Step 2: Add `handleAdminDisableSkill` handler**

Add after `handleAdminRejectSubmission`:

```go
// handleAdminDisableSkill disables a previously-approved skill.
// Sets status='disabled', security_score=0, records the reason.
// POST /api/v1/admin/submissions/{id}/disable
// Body: {"reason": "string"}
func (s *Server) handleAdminDisableSkill(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    if !s.isAdmin(r) {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/submissions/")
    idStr = strings.TrimSuffix(idStr, "/disable")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid submission ID", http.StatusBadRequest)
        return
    }

    var body struct {
        Reason string `json:"reason"`
    }
    _ = json.NewDecoder(r.Body).Decode(&body)
    if strings.TrimSpace(body.Reason) == "" {
        body.Reason = "Disabled by administrator"
    }

    _, userUUID, _ := s.getUserStorage(r)
    reviewerID, _ := s.getUserIDFromUUID(userUUID)

    if err := s.centralStore.DisableSkill(id, reviewerID, body.Reason); err != nil {
        logger.ErrorCF("web", "Failed to disable skill", map[string]interface{}{
            "id": id, "error": err.Error(),
        })
        http.Error(w, "failed to disable skill", http.StatusInternalServerError)
        return
    }

    logger.InfoCF("web", "Skill disabled by admin", map[string]interface{}{
        "submission_id": id, "reviewer": userUUID, "reason": body.Reason,
    })
    writeJSONResponse(w, map[string]string{"status": "disabled"})
}
```

**Step 3: Build**

```bash
go build ./pkg/web/...
```
Expected: no output.

**Step 4: Commit**

```bash
git add pkg/web/handlers_marketplace.go
git commit -m "feat(marketplace): add POST /admin/submissions/{id}/disable endpoint"
```

---

## Task 4: Add `GET /api/v1/marketplace/security-alerts` endpoint

**Files:**
- Modify: `pkg/web/handlers_marketplace.go`
- Modify: `pkg/web/server.go`

**Step 1: Add the handler in `handlers_marketplace.go`**

Add after `handleMarketplaceCategories`:

```go
// handleMarketplaceSecurityAlerts returns the list of disabled skill slugs.
// Public endpoint — no auth required (slugs are not sensitive).
// GET /api/v1/marketplace/security-alerts
func (s *Server) handleMarketplaceSecurityAlerts(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method != http.MethodGet {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    if s.centralStore == nil {
        writeJSONResponse(w, map[string]interface{}{"disabled_slugs": []string{}})
        return
    }

    slugs, err := s.centralStore.GetDisabledSkillSlugs()
    if err != nil {
        logger.ErrorCF("web", "Failed to fetch disabled skill slugs", map[string]interface{}{"error": err.Error()})
        writeJSONResponse(w, map[string]interface{}{"disabled_slugs": []string{}})
        return
    }

    if slugs == nil {
        slugs = []string{}
    }
    writeJSONResponse(w, map[string]interface{}{"disabled_slugs": slugs})
}
```

**Step 2: Register the route in `server.go`**

In `server.go`, find line ~374 where marketplace routes are registered. Add after `handleMarketplaceCategories`:

```go
mux.HandleFunc("/api/v1/marketplace/security-alerts", s.handleMarketplaceSecurityAlerts) // Disabled skills list
```

**Step 3: Build**

```bash
go build ./pkg/web/...
```
Expected: no output.

**Step 4: Commit**

```bash
git add pkg/web/handlers_marketplace.go pkg/web/server.go
git commit -m "feat(marketplace): add GET /marketplace/security-alerts endpoint"
```

---

## Task 5: Add `securityAlert` to `handleConfigStatus`

**Files:**
- Modify: `pkg/web/server.go`

**Context:** `handleConfigStatus` is around line 5044. It already builds the merged effective config and reads the user's installed skills via `getUserStorage`. We need to cross-reference installed skills against disabled slugs.

**Step 1: Add `securityAlert` logic**

Inside `handleConfigStatus`, after the `activeProviders` block (around line 5071), add:

```go
// Check if any of the user's installed skills have been disabled.
securityAlert := false
if s.centralStore != nil {
    if _, userUUID, ok := s.getUserStorage(r); ok && userUUID != "" {
        if disabledSlugs, err := s.centralStore.GetDisabledSkillSlugs(); err == nil && len(disabledSlugs) > 0 {
            // Build a set of disabled slugs for O(1) lookup
            disabledSet := make(map[string]bool, len(disabledSlugs))
            for _, slug := range disabledSlugs {
                disabledSet[slug] = true
            }
            // Scan the user's installed skills
            if userWorkspace, err := config.EnsureUserWorkspace(userUUID); err == nil {
                userLoader := s.newUserSkillsLoader(userUUID, userWorkspace)
                for _, skill := range userLoader.ListUserSkills() {
                    if disabledSet[skill.Slug] {
                        securityAlert = true
                        break
                    }
                }
            }
        }
    }
}
```

**Step 2: Add `securityAlert` to the response map**

Find where `status` map is built (around line 5073):

```go
status := map[string]interface{}{
    "configured":     configured,
    "degradedMode":   degradedMode,
    "securityAlert":  securityAlert,  // <-- add this line
}
```

**Step 3: Build**

```bash
go build ./pkg/web/...
```
Expected: no output.

**Step 4: Commit**

```bash
git add pkg/web/server.go
git commit -m "feat(api): add securityAlert to /config/status response"
```

---

## Task 6: Add `securityAlert` to `configStore.js`

**Files:**
- Modify: `pkg/web/frontend/src/stores/configStore.js`

**Step 1: Add the reactive ref and populate it**

```javascript
// Add after `const statusChecked = ref(false)`:
const securityAlert = ref(false)
```

Inside `checkStatus()`, after `activeProviders.value = response.data.activeProviders || []`:

```javascript
securityAlert.value = response.data.securityAlert === true
```

In the `return` block, add `securityAlert` alongside the other refs.

**Full updated file:**

```javascript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../services/api'

export const useConfigStore = defineStore('config', () => {
  const configured = ref(false)
  const degradedMode = ref(false)
  const reason = ref('')
  const activeProviders = ref([])
  const loading = ref(false)
  const checkError = ref(null)
  const statusChecked = ref(false)
  const securityAlert = ref(false)

  async function checkStatus() {
    loading.value = true
    checkError.value = null
    try {
      const response = await api.get('/config/status')
      configured.value = response.data.configured
      degradedMode.value = response.data.degradedMode
      reason.value = response.data.reason || ''
      activeProviders.value = response.data.activeProviders || []
      securityAlert.value = response.data.securityAlert === true
      statusChecked.value = true
      return response.data
    } catch (error) {
      console.error('Failed to check config status:', error)
      checkError.value = error.message
      degradedMode.value = true
      configured.value = false
      throw error
    } finally {
      loading.value = false
    }
  }

  async function updateProvider(providerData) {
    try {
      const response = await api.post('/config/provider', providerData)
      await checkStatus()
      return response.data
    } catch (error) {
      console.error('Failed to update provider:', error)
      throw error
    }
  }

  async function validateProvider(providerData) {
    try {
      const response = await api.post('/config/validate', providerData)
      return response.data
    } catch (error) {
      console.error('Failed to validate provider:', error)
      throw error
    }
  }

  function needsConfiguration() {
    return degradedMode.value || !configured.value
  }

  return {
    configured,
    degradedMode,
    reason,
    activeProviders,
    loading,
    checkError,
    statusChecked,
    securityAlert,
    checkStatus,
    updateProvider,
    validateProvider,
    needsConfiguration
  }
})
```

Note: also fixed a pre-existing bug on the original line 21 (`configured.value = response.data.degradedMode` was wrong — it should be `degradedMode.value`).

**Step 2: Commit**

```bash
git add pkg/web/frontend/src/stores/configStore.js
git commit -m "feat(store): add securityAlert to configStore + fix degradedMode assignment bug"
```

---

## Task 7: Create `SecurityAlertBanner.vue`

**Files:**
- Create: `pkg/web/frontend/src/components/SecurityAlertBanner.vue`

**Step 1: Create the component**

```vue
<template>
  <div
    v-if="configStore.securityAlert"
    class="bg-amber-500/10 border-b-2 border-amber-500 animate-slideDown sticky top-0 z-banner backdrop-blur-sm px-6 py-4 shadow-sm"
  >
    <div class="flex items-center gap-4 max-w-[1200px] mx-auto flex-wrap md:flex-nowrap">
      <div class="text-[32px] md:text-2xl shrink-0">
        🔒
      </div>
      <div class="flex-1">
        <h3 class="text-makoclaw-text font-semibold text-lg md:text-base mb-1">
          Alerta de Seguridad
        </h3>
        <p class="text-makoclaw-text-secondary text-sm">
          Una o más skills instaladas fueron desactivadas por razones de seguridad.
          Revisá tus skills y desinstalá las marcadas.
        </p>
      </div>
      <button
        class="bg-amber-500 text-white hover:bg-amber-400 rounded-lg px-5 py-2.5 font-semibold transition-all active:scale-[0.97] whitespace-nowrap w-full md:w-auto"
        @click="router.push('/skills')"
      >
        Revisar Skills
      </button>
    </div>
  </div>
</template>

<script setup>
import { useConfigStore } from '../stores/configStore'
import { useRouter } from 'vue-router'

const configStore = useConfigStore()
const router = useRouter()
</script>
```

**Step 2: Commit**

```bash
git add pkg/web/frontend/src/components/SecurityAlertBanner.vue
git commit -m "feat(ui): add SecurityAlertBanner component"
```

---

## Task 8: Add `SecurityAlertBanner` to `MainLayout.vue`

**Files:**
- Modify: `pkg/web/frontend/src/components/Layout/MainLayout.vue`

**Step 1: Add import in `<script setup>`**

In `MainLayout.vue`, find the `<script setup>` section (around line 55). Add the import:

```javascript
import SecurityAlertBanner from '../SecurityAlertBanner.vue'
```

**Step 2: Add banner to template**

Find line 36 where `<DegradedModeBanner />` is rendered. Add the security banner right after it:

```vue
<!-- Degraded Mode Banner -->
<DegradedModeBanner />

<!-- Security Alert Banner -->
<SecurityAlertBanner />
```

**Step 3: Commit**

```bash
git add pkg/web/frontend/src/components/Layout/MainLayout.vue
git commit -m "feat(layout): include SecurityAlertBanner in MainLayout"
```

---

## Task 9: Show disabled badge in `SkillsView.vue`

**Files:**
- Modify: `pkg/web/frontend/src/views/SkillsView.vue`
- Modify: `pkg/web/frontend/src/services/advancedService.js`

**Goal:** After loading installed skills, fetch disabled slugs and mark matching skills with a warning badge.

**Step 1: Add `fetchSecurityAlerts` to `advancedService.js`**

In `advancedService.js`, add:

```javascript
fetchSecurityAlerts: async () => {
  const response = await client.get('/marketplace/security-alerts')
  return response.data
},
```

**Step 2: Add reactive state in `SkillsView.vue`**

In the `<script setup>` section, add near the other refs:

```javascript
const disabledSlugs = ref(new Set())
```

**Step 3: Fetch disabled slugs when loading skills**

In `loadSkills()` (find it — it calls `advancedService.fetchSkills()`), add after loading skills:

```javascript
const alertData = await advancedService.fetchSecurityAlerts()
disabledSlugs.value = new Set(alertData.disabled_slugs || [])
```

Use try/catch so a failure doesn't break the skills load:

```javascript
try {
  const alertData = await advancedService.fetchSecurityAlerts()
  disabledSlugs.value = new Set(alertData.disabled_slugs || [])
} catch {
  // non-critical — silently ignore
}
```

**Step 4: Add `isDisabled` computed helper**

```javascript
const isDisabled = (skill) => disabledSlugs.value.has(skill.slug || skill.name)
```

**Step 5: Add the disabled badge in the installed skills list template**

Find the skill card/row in the "Installed" tab of the template (look for `skill.source === 'user'` or the skills list loop). After the skill name/description, add:

```vue
<div
  v-if="isDisabled(skill)"
  class="mt-2 flex items-center gap-2 bg-red-500/10 border border-red-500/20 rounded-lg px-3 py-2"
>
  <span class="text-red-400 text-xs font-black uppercase tracking-widest">⚠ Deshabilitada</span>
  <span class="text-red-300/70 text-xs">Esta skill fue desactivada por razones de seguridad. Se recomienda desinstalarla.</span>
</div>
```

**Step 6: Refresh config status after uninstalling a skill**

When the user uninstalls a skill (find the uninstall handler), add after the uninstall call:

```javascript
await configStore.checkStatus() // refresh securityAlert in banner
```

Import `useConfigStore` at the top if not already imported:
```javascript
import { useConfigStore } from '../stores/configStore'
const configStore = useConfigStore()
```

**Step 7: Build frontend to check for errors**

```bash
cd pkg/web/frontend
npm run build 2>&1 | tail -20
```
Expected: `dist/` built successfully, no errors.

**Step 8: Commit**

```bash
git add pkg/web/frontend/src/views/SkillsView.vue pkg/web/frontend/src/services/advancedService.js
git commit -m "feat(skills): show disabled badge on revoked skills + refresh banner on uninstall"
```

---

## Manual Testing Checklist

After all tasks are complete and the app is running:

1. **Security scan thresholds**: Submit a skill. Check the resulting `status` in the DB matches the score buckets (score <50 → rejected, 50-74 → needs_review, ≥75 → approved).

2. **Admin disable**: Log in as admin. Call `POST /api/v1/admin/submissions/{id}/disable` with `{"reason": "Test disable"}`. Verify `status='disabled'` and `security_score=0` in the DB.

3. **Security alerts endpoint**: Call `GET /api/v1/marketplace/security-alerts`. Should return `{"disabled_slugs": ["the-disabled-skill"]}`.

4. **Config status**: Call `GET /api/v1/config/status` as a user who has the disabled skill installed. Should return `"securityAlert": true`.

5. **App-wide banner**: As that user, open the app. The amber "Alerta de Seguridad" banner should appear at the top.

6. **Skills view badge**: Navigate to /skills → Installed. The disabled skill should have a red "Deshabilitada" badge.

7. **Banner disappears**: Uninstall the disabled skill. The banner should disappear (config/status is refreshed).

8. **No banner for clean users**: As a user with no disabled skills installed, `securityAlert` should be `false` and no banner shown.
