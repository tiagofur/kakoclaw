package cron

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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
	_, _ = cs.AddJob("test", everySchedule(60000), "hello", false, "", "")

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
	_, _ = cs.AddJob("test", everySchedule(60000), "msg", false, "", "")

	tmpPath := cs.storePath + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatal("temp file should not exist after successful save")
	}
}

// --- CRUD ---

func TestAddAndListJobs(t *testing.T) {
	cs := newTestCronService(t)
	job, err := cs.AddJob("alarm", everySchedule(30000), "wake up", false, "", "")
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
	job, _ := cs.AddJob("old name", everySchedule(60000), "msg", false, "", "")

	updated, err := cs.UpdateJob(job.ID, "new name", everySchedule(120000), "new msg", true, "ch", "to")
	if err != nil {
		t.Fatalf("UpdateJob: %v", err)
	}
	if updated.Name != "new name" {
		t.Fatalf("expected 'new name', got %q", updated.Name)
	}
	if updated.Payload.Message != "new msg" || !updated.Payload.Deliver {
		t.Fatalf("unexpected payload: %+v", updated.Payload)
	}
}

func TestRemoveJob(t *testing.T) {
	cs := newTestCronService(t)
	job, _ := cs.AddJob("removeme", everySchedule(60000), "msg", false, "", "")

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
	_, _ = cs1.AddJob("persist-me", everySchedule(60000), "msg", false, "", "")

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
	_, err := cs.AddJob("bad", CronSchedule{Kind: "every", EveryMS: &zero}, "msg", false, "", "")
	if err == nil {
		t.Fatal("expected error for everyMs=0")
	}
}

func TestValidateScheduleAtRequiresAtMs(t *testing.T) {
	cs := newTestCronService(t)
	_, err := cs.AddJob("bad", CronSchedule{Kind: "at"}, "msg", false, "", "")
	if err == nil {
		t.Fatal("expected error for missing atMs")
	}
}
