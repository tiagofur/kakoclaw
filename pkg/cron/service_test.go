package cron

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func newTestCronService(t *testing.T) *CronService {
	t.Helper()
	storePath := filepath.Join(t.TempDir(), "cron", "store.json")
	cs := NewCronService(storePath, nil)
	return cs
}

func everySchedule(ms int64) CronSchedule {
	return CronSchedule{Kind: "every", EveryMS: &ms}
}

// --- Atomic write ---

func TestSaveStoreCreatesFile(t *testing.T) {
	cs := newTestCronService(t)
	cs.mu.Lock()
	err := cs.saveStoreUnsafe()
	cs.mu.Unlock()
	if err != nil {
		t.Fatalf("saveStoreUnsafe: %v", err)
	}
	if _, err := os.Stat(cs.storePath); os.IsNotExist(err) {
		t.Fatal("store file should exist after save")
	}
}

func TestSaveStoreIsValidJSON(t *testing.T) {
	cs := newTestCronService(t)
	_, _ = cs.AddJob(0, "test", everySchedule(60000), "hello", "", "", "", "")

	data, err := os.ReadFile(cs.storePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var store CronStore
	if err := json.Unmarshal(data, &store); err != nil {
		t.Fatalf("store file is not valid JSON: %v", err)
	}
	if len(store.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(store.Jobs))
	}
}

func TestSaveStoreNoTempFileLeft(t *testing.T) {
	cs := newTestCronService(t)
	_, _ = cs.AddJob(0, "test", everySchedule(60000), "msg", "", "", "", "")

	tmpPath := cs.storePath + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatal("temp file should not exist after successful save")
	}
}

// --- CRUD ---

func TestAddAndListJobs(t *testing.T) {
	cs := newTestCronService(t)
	job, err := cs.AddJob(0, "alarm", everySchedule(30000), "wake up", "", "", "", "")
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}
	if job.Name != "alarm" || !job.Enabled {
		t.Fatalf("unexpected job: %+v", job)
	}

	jobs := cs.ListJobs(true)
	if len(jobs) != 1 || jobs[0].ID != job.ID {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
}

func TestUpdateJob(t *testing.T) {
	cs := newTestCronService(t)
	job, _ := cs.AddJob(0, "old name", everySchedule(60000), "msg", "", "", "", "")

	updated, err := cs.UpdateJob(job.ID, "new name", everySchedule(120000), "new msg", "reminder", "ch", "to", "")
	if err != nil {
		t.Fatalf("UpdateJob: %v", err)
	}
	if updated.Name != "new name" {
		t.Fatalf("expected 'new name', got %q", updated.Name)
	}
	if updated.Payload.Message != "new msg" || updated.Payload.JobType != "reminder" {
		t.Fatalf("unexpected payload: %+v", updated.Payload)
	}
}

func TestRemoveJob(t *testing.T) {
	cs := newTestCronService(t)
	job, _ := cs.AddJob(0, "removeme", everySchedule(60000), "msg", "", "", "", "")

	removed := cs.RemoveJob(job.ID)
	if !removed {
		t.Fatal("expected job to be removed")
	}
	if len(cs.ListJobs(true)) != 0 {
		t.Fatal("expected 0 jobs after remove")
	}
}

func TestRemoveNonexistentJob(t *testing.T) {
	cs := newTestCronService(t)
	if cs.RemoveJob("nonexistent") {
		t.Fatal("should not be able to remove nonexistent job")
	}
}

// --- Persistence across reload ---

func TestPersistenceAcrossReload(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "cron", "store.json")
	cs1 := NewCronService(storePath, nil)
	_, _ = cs1.AddJob(0, "persist-me", everySchedule(60000), "msg", "", "", "", "")

	// Create a new service from the same path
	cs2 := NewCronService(storePath, nil)
	jobs := cs2.ListJobs(true)
	if len(jobs) != 1 || jobs[0].Name != "persist-me" {
		t.Fatalf("expected persisted job, got %v", jobs)
	}
}

// --- Schedule validation ---

func TestValidateScheduleEveryRequiresPositiveMs(t *testing.T) {
	cs := newTestCronService(t)
	zero := int64(0)
	_, err := cs.AddJob(0, "bad", CronSchedule{Kind: "every", EveryMS: &zero}, "msg", "", "", "", "")
	if err == nil {
		t.Fatal("expected error for everyMs=0")
	}
}

func TestValidateScheduleAtRequiresAtMs(t *testing.T) {
	cs := newTestCronService(t)
	_, err := cs.AddJob(0, "bad", CronSchedule{Kind: "at"}, "msg", "", "", "", "")
	if err == nil {
		t.Fatal("expected error for missing atMs")
	}
}

// --- computeNextRun ---

func TestComputeNextRun_AtType_Future(t *testing.T) {
	cs := newTestCronService(t)
	now := time.Now().UnixMilli()
	futureMS := now + 60000 // 1 minute in the future

	schedule := &CronSchedule{
		Kind: "at",
		AtMS: &futureMS,
	}

	nextRun := cs.computeNextRun(schedule, now)
	if nextRun == nil {
		t.Fatal("computeNextRun returned nil for future 'at' schedule")
	}
	if *nextRun != futureMS {
		t.Errorf("NextRunAtMS = %d, want %d", *nextRun, futureMS)
	}
}

func TestComputeNextRun_AtType_Past(t *testing.T) {
	cs := newTestCronService(t)
	now := time.Now().UnixMilli()
	pastMS := now - 60000 // 1 minute in the past

	schedule := &CronSchedule{
		Kind: "at",
		AtMS: &pastMS,
	}

	nextRun := cs.computeNextRun(schedule, now)
	if nextRun != nil {
		t.Errorf("expected nil for past 'at' schedule, got %d", *nextRun)
	}
}

func TestComputeNextRun_EveryType(t *testing.T) {
	cs := newTestCronService(t)
	now := time.Now().UnixMilli()
	everyMS := int64(30000) // 30 seconds

	schedule := &CronSchedule{
		Kind:    "every",
		EveryMS: &everyMS,
	}

	nextRun := cs.computeNextRun(schedule, now)
	if nextRun == nil {
		t.Fatal("computeNextRun returned nil for 'every' schedule")
	}

	expected := now + everyMS
	if *nextRun != expected {
		t.Errorf("NextRunAtMS = %d, want %d", *nextRun, expected)
	}
}

func TestComputeNextRun_EveryType_NilEveryMS(t *testing.T) {
	cs := newTestCronService(t)
	now := time.Now().UnixMilli()

	schedule := &CronSchedule{
		Kind:    "every",
		EveryMS: nil,
	}

	nextRun := cs.computeNextRun(schedule, now)
	if nextRun != nil {
		t.Errorf("expected nil for nil EveryMS, got %d", *nextRun)
	}
}

func TestComputeNextRun_EveryType_ZeroEveryMS(t *testing.T) {
	cs := newTestCronService(t)
	now := time.Now().UnixMilli()
	zero := int64(0)

	schedule := &CronSchedule{
		Kind:    "every",
		EveryMS: &zero,
	}

	nextRun := cs.computeNextRun(schedule, now)
	if nextRun != nil {
		t.Errorf("expected nil for zero EveryMS, got %d", *nextRun)
	}
}

func TestComputeNextRun_CronType(t *testing.T) {
	cs := newTestCronService(t)
	now := time.Now().UnixMilli()

	// "every minute" cron expression
	schedule := &CronSchedule{
		Kind: "cron",
		Expr: "* * * * *",
	}

	nextRun := cs.computeNextRun(schedule, now)
	if nextRun == nil {
		t.Fatal("computeNextRun returned nil for valid cron expression")
	}

	// Next run should be in the future
	if *nextRun <= now {
		t.Errorf("next run %d should be after now %d", *nextRun, now)
	}

	// For "every minute", the next run should be within the next ~60 seconds
	maxExpected := now + 61000
	if *nextRun > maxExpected {
		t.Errorf("next run %d is too far in the future (max expected %d)", *nextRun, maxExpected)
	}
}

func TestComputeNextRun_CronType_EmptyExpr(t *testing.T) {
	cs := newTestCronService(t)
	now := time.Now().UnixMilli()

	schedule := &CronSchedule{
		Kind: "cron",
		Expr: "",
	}

	nextRun := cs.computeNextRun(schedule, now)
	if nextRun != nil {
		t.Errorf("expected nil for empty cron expression, got %d", *nextRun)
	}
}

func TestComputeNextRun_UnknownKind(t *testing.T) {
	cs := newTestCronService(t)
	now := time.Now().UnixMilli()

	schedule := &CronSchedule{
		Kind: "unknown",
	}

	nextRun := cs.computeNextRun(schedule, now)
	if nextRun != nil {
		t.Errorf("expected nil for unknown schedule kind, got %d", *nextRun)
	}
}

// --- RunJob ---

func TestRunJob_ExecutesHandler(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "cron", "store.json")

	var mu sync.Mutex
	var calledWith []string

	handler := func(job *CronJob) (string, error) {
		mu.Lock()
		calledWith = append(calledWith, job.ID)
		mu.Unlock()
		return "done", nil
	}

	cs := NewCronService(storePath, handler)
	job, err := cs.AddJob(1, "test-run", everySchedule(60000), "hello", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	if err := cs.RunJob(job.ID); err != nil {
		t.Fatalf("RunJob: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(calledWith) != 1 {
		t.Fatalf("expected handler called once, got %d calls", len(calledWith))
	}
	if calledWith[0] != job.ID {
		t.Errorf("handler called with job ID %q, want %q", calledWith[0], job.ID)
	}
}

func TestRunJob_NotFound(t *testing.T) {
	cs := newTestCronService(t)
	if err := cs.RunJob("nonexistent"); err == nil {
		t.Fatal("expected error for non-existent job ID")
	}
}

func TestRunJob_UpdatesState(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "cron", "store.json")
	handler := func(job *CronJob) (string, error) {
		return "ok", nil
	}
	cs := NewCronService(storePath, handler)

	job, err := cs.AddJob(1, "state-test", everySchedule(60000), "msg", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	before := time.Now().UnixMilli()
	if err := cs.RunJob(job.ID); err != nil {
		t.Fatalf("RunJob: %v", err)
	}

	// Check state was updated
	jobs := cs.ListJobs(true)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	updated := jobs[0]
	if updated.State.LastRunAtMS == nil {
		t.Fatal("LastRunAtMS should be set after RunJob")
	}
	if *updated.State.LastRunAtMS < before {
		t.Errorf("LastRunAtMS %d should be >= %d", *updated.State.LastRunAtMS, before)
	}
	if updated.State.LastStatus != "ok" {
		t.Errorf("LastStatus = %q, want %q", updated.State.LastStatus, "ok")
	}
}

// --- ListJobsForUser ---

func TestListJobsForUser(t *testing.T) {
	cs := newTestCronService(t)

	// Add jobs for different users
	_, err := cs.AddJob(1, "user1-job-a", everySchedule(60000), "msg1", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob user1-a: %v", err)
	}
	_, err = cs.AddJob(1, "user1-job-b", everySchedule(60000), "msg2", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob user1-b: %v", err)
	}
	_, err = cs.AddJob(2, "user2-job", everySchedule(60000), "msg3", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob user2: %v", err)
	}

	// List for user 1
	user1Jobs := cs.ListJobsForUser(1, true)
	if len(user1Jobs) != 2 {
		t.Fatalf("expected 2 jobs for user 1, got %d", len(user1Jobs))
	}
	for _, j := range user1Jobs {
		if j.UserID != 1 {
			t.Errorf("expected UserID 1, got %d for job %q", j.UserID, j.Name)
		}
	}

	// List for user 2
	user2Jobs := cs.ListJobsForUser(2, true)
	if len(user2Jobs) != 1 {
		t.Fatalf("expected 1 job for user 2, got %d", len(user2Jobs))
	}
	if user2Jobs[0].Name != "user2-job" {
		t.Errorf("expected job name %q, got %q", "user2-job", user2Jobs[0].Name)
	}

	// List for user 3 (no jobs)
	user3Jobs := cs.ListJobsForUser(3, true)
	if len(user3Jobs) != 0 {
		t.Fatalf("expected 0 jobs for user 3, got %d", len(user3Jobs))
	}
}

func TestListJobsForUser_ExcludesDisabled(t *testing.T) {
	cs := newTestCronService(t)

	job, err := cs.AddJob(1, "to-disable", everySchedule(60000), "msg", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}
	_, err = cs.AddJob(1, "stays-enabled", everySchedule(60000), "msg", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Disable the first job
	cs.EnableJob(job.ID, false)

	// Without including disabled
	enabledOnly := cs.ListJobsForUser(1, false)
	if len(enabledOnly) != 1 {
		t.Fatalf("expected 1 enabled job for user 1, got %d", len(enabledOnly))
	}
	if enabledOnly[0].Name != "stays-enabled" {
		t.Errorf("expected enabled job name %q, got %q", "stays-enabled", enabledOnly[0].Name)
	}

	// Including disabled
	allJobs := cs.ListJobsForUser(1, true)
	if len(allJobs) != 2 {
		t.Fatalf("expected 2 total jobs for user 1, got %d", len(allJobs))
	}
}

func TestListJobsForUser_ZeroNormalizesToOne(t *testing.T) {
	cs := newTestCronService(t)

	// AddJob with userID=0 normalizes to 1
	_, err := cs.AddJob(0, "zero-user", everySchedule(60000), "msg", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// ListJobsForUser with userID=0 should also normalize to 1
	jobs := cs.ListJobsForUser(0, true)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job for userID 0 (normalized to 1), got %d", len(jobs))
	}
}

// --- EnableJob / DisableJob ---

func TestEnableDisableJob(t *testing.T) {
	cs := newTestCronService(t)

	job, err := cs.AddJob(1, "toggle-me", everySchedule(60000), "msg", "task", "", "", "")
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}

	// Job should be enabled initially
	if !job.Enabled {
		t.Fatal("expected job to be enabled initially")
	}

	// Disable it
	disabled := cs.EnableJob(job.ID, false)
	if disabled == nil {
		t.Fatal("EnableJob returned nil")
	}
	if disabled.Enabled {
		t.Error("expected job to be disabled after EnableJob(false)")
	}
	if disabled.State.NextRunAtMS != nil {
		t.Error("expected NextRunAtMS to be nil when disabled")
	}

	// Verify in list
	jobs := cs.ListJobs(true)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Enabled {
		t.Error("expected job to remain disabled in list")
	}

	// Re-enable
	enabled := cs.EnableJob(job.ID, true)
	if enabled == nil {
		t.Fatal("EnableJob returned nil")
	}
	if !enabled.Enabled {
		t.Error("expected job to be enabled after EnableJob(true)")
	}
	if enabled.State.NextRunAtMS == nil {
		t.Error("expected NextRunAtMS to be set when re-enabled")
	}
}

func TestEnableJob_NonExistent(t *testing.T) {
	cs := newTestCronService(t)
	result := cs.EnableJob("nonexistent", true)
	if result != nil {
		t.Errorf("expected nil for non-existent job, got %+v", result)
	}
}
