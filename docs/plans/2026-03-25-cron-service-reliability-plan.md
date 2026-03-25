# Cron Service Reliability Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix cron job management in web mode so agents use the CronTool API instead of write_file, and add mtime-based safety net to prevent external file changes from being overwritten.

**Architecture:** Three changes: (1) wire CronTool into web chat agent loops, (2) add mtime detection + union merge to CronService.saveStoreUnsafe(), (3) InitializeAllUsers in web mode (already done).

**Tech Stack:** Go 1.26, existing CronService + CronTool + MultiUserChannelManager

**Design doc:** `docs/plans/2026-03-25-cron-service-reliability-design.md`

---

### Task 1: Add mtime tracking field to CronService

**Files:**
- Modify: `pkg/cron/service.go:69-77`

**Step 1: Add `lastSavedMtime` field to `CronService` struct**

In `pkg/cron/service.go`, add `lastSavedMtime time.Time` to the struct:

```go
type CronService struct {
	storePath      string
	store          *CronStore
	onJob          JobHandler
	mu             sync.RWMutex
	running        bool
	stopChan       chan struct{}
	gronx          *gronx.Gronx
	lastSavedMtime time.Time // mtime of jobs.json after our last write
}
```

**Step 2: Update `loadStore()` to record mtime after loading**

At the end of `loadStore()`, after successfully reading the file, stat the file and record its mtime:

```go
func (cs *CronService) loadStore() error {
	cs.store = &CronStore{
		Version: 1,
		Jobs:    []CronJob{},
	}

	data, err := os.ReadFile(cs.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, cs.store); err != nil {
		return err
	}

	// Record file mtime so we can detect external changes later
	if info, err := os.Stat(cs.storePath); err == nil {
		cs.lastSavedMtime = info.ModTime()
	}

	return nil
}
```

**Step 3: Run existing tests to verify no regression**

Run: `go test ./pkg/cron/ -v`
Expected: All existing tests PASS

**Step 4: Commit**

```bash
git add pkg/cron/service.go
git commit -m "feat(cron): add mtime tracking field to CronService"
```

---

### Task 2: Add union merge helper to CronService

**Files:**
- Modify: `pkg/cron/service.go` (add new method)
- Test: `pkg/cron/service_test.go`

**Step 1: Write the failing test for mergeExternalChanges**

Add to `pkg/cron/service_test.go`:

```go
func TestMergeExternalChanges_AddsNewJobsFromDisk(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "cron", "store.json")
	cs := NewCronService(storePath, nil)

	// Add a job via API (in-memory + disk)
	_, err := cs.AddJob(1, "api-job", everySchedule(60000), "msg1", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Simulate external edit: write a new job directly to disk
	externalStore := CronStore{
		Version: 1,
		Jobs: []CronJob{
			{ID: "external-1", UserID: 1, Name: "external-job", Enabled: true,
				Schedule: everySchedule(30000),
				Payload:  CronPayload{Kind: "agent_turn", Message: "ext msg"}},
		},
	}
	data, _ := json.MarshalIndent(externalStore, "", "  ")
	os.MkdirAll(filepath.Dir(storePath), 0755)
	os.WriteFile(storePath, data, 0644)

	// Trigger merge
	cs.mu.Lock()
	cs.mergeExternalChanges()
	cs.mu.Unlock()

	// Should have both: the API job (from memory) + the external job (from disk)
	jobs := cs.ListJobs(true)
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs after merge, got %d", len(jobs))
	}

	names := map[string]bool{}
	for _, j := range jobs {
		names[j.Name] = true
	}
	if !names["api-job"] {
		t.Error("api-job should be preserved from memory")
	}
	if !names["external-job"] {
		t.Error("external-job should be added from disk")
	}
}

func TestMergeExternalChanges_MemoryWinsOnConflict(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "cron", "store.json")
	cs := NewCronService(storePath, nil)

	// Add a job via API
	job, err := cs.AddJob(1, "original", everySchedule(60000), "msg1", "task", "telegram", "123", "")
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Update it in memory (simulate state change like last_run)
	cs.mu.Lock()
	for i := range cs.store.Jobs {
		if cs.store.Jobs[i].ID == job.ID {
			cs.store.Jobs[i].Name = "updated-in-memory"
		}
	}
	cs.mu.Unlock()

	// Write a conflicting version to disk with the same ID
	externalStore := CronStore{
		Version: 1,
		Jobs: []CronJob{
			{ID: job.ID, UserID: 1, Name: "updated-on-disk", Enabled: true,
				Schedule: everySchedule(60000),
				Payload:  CronPayload{Kind: "agent_turn", Message: "disk msg"}},
		},
	}
	data, _ := json.MarshalIndent(externalStore, "", "  ")
	os.WriteFile(storePath, data, 0644)

	// Trigger merge
	cs.mu.Lock()
	cs.mergeExternalChanges()
	cs.mu.Unlock()

	// Memory should win
	jobs := cs.ListJobs(true)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job after merge, got %d", len(jobs))
	}
	if jobs[0].Name != "updated-in-memory" {
		t.Errorf("expected memory version 'updated-in-memory', got %q", jobs[0].Name)
	}
}

func TestMergeExternalChanges_NoFileIsNoop(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "cron", "nonexistent.json")
	cs := NewCronService(storePath, nil)

	_, _ = cs.AddJob(1, "memory-only", everySchedule(60000), "msg", "task", "", "", "")

	// Delete the file to simulate it not existing
	os.Remove(storePath)

	// Merge should be a no-op (no crash, no data loss)
	cs.mu.Lock()
	cs.mergeExternalChanges()
	cs.mu.Unlock()

	jobs := cs.ListJobs(true)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job preserved, got %d", len(jobs))
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./pkg/cron/ -run TestMergeExternal -v`
Expected: FAIL — `mergeExternalChanges` not defined

**Step 3: Implement `mergeExternalChanges()`**

Add to `pkg/cron/service.go` (after `loadStore()`):

```go
// mergeExternalChanges detects if jobs.json was modified externally and merges
// new jobs from disk into the in-memory state. Must be called with cs.mu held.
// Strategy: memory is authoritative for existing jobs, disk adds new ones.
func (cs *CronService) mergeExternalChanges() {
	info, err := os.Stat(cs.storePath)
	if err != nil {
		return // File doesn't exist or can't be read — nothing to merge
	}

	// No external change if mtime matches our last save
	if info.ModTime().Equal(cs.lastSavedMtime) {
		return
	}

	// File was modified externally — read it
	data, err := os.ReadFile(cs.storePath)
	if err != nil {
		return
	}

	var diskStore CronStore
	if err := json.Unmarshal(data, &diskStore); err != nil {
		log.Printf("[cron] external jobs.json has invalid JSON, skipping merge: %v", err)
		return
	}

	// Build a set of in-memory job IDs for fast lookup
	memoryIDs := make(map[string]bool, len(cs.store.Jobs))
	for _, job := range cs.store.Jobs {
		memoryIDs[job.ID] = true
	}

	// Add any disk jobs that don't exist in memory (new external additions)
	for _, diskJob := range diskStore.Jobs {
		if !memoryIDs[diskJob.ID] {
			cs.store.Jobs = append(cs.store.Jobs, diskJob)
		}
		// If same ID exists in both, memory wins (skip disk version)
	}
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/cron/ -run TestMergeExternal -v`
Expected: All 3 tests PASS

**Step 5: Run all cron tests**

Run: `go test ./pkg/cron/ -v`
Expected: All tests PASS

**Step 6: Commit**

```bash
git add pkg/cron/service.go pkg/cron/service_test.go
git commit -m "feat(cron): add union merge for external file changes"
```

---

### Task 3: Integrate mtime check into saveStoreUnsafe

**Files:**
- Modify: `pkg/cron/service.go:379-396` (`saveStoreUnsafe`)
- Test: `pkg/cron/service_test.go`

**Step 1: Write the failing integration test**

Add to `pkg/cron/service_test.go`:

```go
func TestSavePreservesExternallyAddedJobs(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "cron", "store.json")
	cs := NewCronService(storePath, nil)

	// Add a job via API
	_, err := cs.AddJob(1, "api-job", everySchedule(60000), "api msg", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Simulate external edit: add a new job directly to the file
	data, _ := os.ReadFile(storePath)
	var store CronStore
	json.Unmarshal(data, &store)
	store.Jobs = append(store.Jobs, CronJob{
		ID: "ext-added", UserID: 1, Name: "externally-added", Enabled: true,
		Schedule: everySchedule(30000),
		Payload:  CronPayload{Kind: "agent_turn", Message: "ext"},
	})
	newData, _ := json.MarshalIndent(store, "", "  ")
	os.WriteFile(storePath, newData, 0644)

	// Now trigger a save (which happens every second via checkJobs)
	cs.mu.Lock()
	err = cs.saveStoreUnsafe()
	cs.mu.Unlock()
	if err != nil {
		t.Fatalf("saveStoreUnsafe: %v", err)
	}

	// Read the file back — both jobs should be there
	savedData, _ := os.ReadFile(storePath)
	var savedStore CronStore
	json.Unmarshal(savedData, &savedStore)

	if len(savedStore.Jobs) != 2 {
		t.Fatalf("expected 2 jobs in saved file, got %d", len(savedStore.Jobs))
	}

	names := map[string]bool{}
	for _, j := range savedStore.Jobs {
		names[j.Name] = true
	}
	if !names["api-job"] {
		t.Error("api-job should be preserved")
	}
	if !names["externally-added"] {
		t.Error("externally-added job should be preserved after save")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/cron/ -run TestSavePreservesExternally -v`
Expected: FAIL — externally-added job is lost

**Step 3: Modify `saveStoreUnsafe()` to call merge + update mtime**

Replace `saveStoreUnsafe()` in `pkg/cron/service.go`:

```go
func (cs *CronService) saveStoreUnsafe() error {
	dir := filepath.Dir(cs.storePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Check for external changes before overwriting
	cs.mergeExternalChanges()

	data, err := json.MarshalIndent(cs.store, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: write to temp file then rename to avoid corruption on crash.
	tmpPath := cs.storePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, cs.storePath); err != nil {
		return err
	}

	// Record the mtime of the file we just wrote
	if info, err := os.Stat(cs.storePath); err == nil {
		cs.lastSavedMtime = info.ModTime()
	}

	return nil
}
```

**Step 4: Run the integration test**

Run: `go test ./pkg/cron/ -run TestSavePreservesExternally -v`
Expected: PASS

**Step 5: Run all cron tests**

Run: `go test ./pkg/cron/ -v`
Expected: All tests PASS

**Step 6: Commit**

```bash
git add pkg/cron/service.go pkg/cron/service_test.go
git commit -m "feat(cron): integrate mtime check into saveStoreUnsafe"
```

---

### Task 4: Wire CronTool into web chat agent loop

**Files:**
- Modify: `pkg/web/server.go:1196` (in `handleChatWS()`)

**Step 1: Add CronTool registration after agent manager setup**

In `pkg/web/server.go`, in the `handleChatWS()` function, after line 1196 (`activeAgentLoop := agentMgr.GetActiveAgent()`), add:

```go
	activeAgentLoop := agentMgr.GetActiveAgent()

	// Register CronTool so the agent can manage cron jobs via API
	// instead of falling back to write_file on jobs.json
	if s.multiUserChannelManager != nil {
		if cronSvc, ok := s.multiUserChannelManager.GetCronServiceForUser(userUUID); ok {
			cronTool := tools.NewCronTool(cronSvc, activeAgentLoop, s.msgBus)
			activeAgentLoop.RegisterTool(cronTool)
		}
	}
```

Note: `tools` package is already imported in server.go. Verify with: `grep "pkg/tools" pkg/web/server.go`

**Step 2: Verify the import exists**

Run: `grep -n "sipeed/makoclaw/pkg/tools" pkg/web/server.go`

If the import is missing, add it to the import block. If it exists, proceed.

**Step 3: Verify compilation**

Run: `go build ./cmd/makoclaw/`
Expected: No errors

**Step 4: Commit**

```bash
git add pkg/web/server.go
git commit -m "feat(web): wire CronTool into web chat agent loop"
```

---

### Task 5: Run full test suite and verify

**Files:** None (verification only)

**Step 1: Run cron tests**

Run: `go test ./pkg/cron/ -v`
Expected: All tests PASS

**Step 2: Run web tests**

Run: `go test ./pkg/web/ -v`
Expected: All tests PASS

**Step 3: Run all tests**

Run: `go test ./...`
Expected: All tests PASS (or same failures as before — no new regressions)

**Step 4: Run go vet**

Run: `go vet ./pkg/cron/ ./pkg/web/ ./cmd/makoclaw/`
Expected: No new warnings related to our changes

---

### Task 6: Final commit for InitializeAllUsers fix (already applied)

**Files:**
- Verify: `cmd/makoclaw/main.go:1075-1082`

**Step 1: Verify the InitializeAllUsers change is still in place**

Run: `grep -A5 "InitializeAllUsers" cmd/makoclaw/main.go | head -20`
Expected: Two call sites — one in `gatewayCmd()` and one in `webCmd()`

**Step 2: Commit if not already committed**

```bash
git add cmd/makoclaw/main.go
git commit -m "fix(web): initialize per-user cron services in web mode"
```
