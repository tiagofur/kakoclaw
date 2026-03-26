package tools

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/cron"
)

func newTestCronTool(t *testing.T) *CronTool {
	t.Helper()
	service := cron.NewCronService(filepath.Join(t.TempDir(), "cron", "jobs.json"), nil)
	tool := NewCronTool(service, nil, nil)
	tool.SetUserID(7)
	tool.SetContext("web", "chat")
	return tool
}

func TestCronToolUpdateJob(t *testing.T) {
	tool := newTestCronTool(t)
	ctx := context.Background()

	created, err := tool.Execute(ctx, map[string]interface{}{
		"action":        "add",
		"message":       "original message",
		"every_seconds": float64(60),
		"job_type":      "task",
	})
	if err != nil {
		t.Fatalf("add returned error: %v", err)
	}
	if !strings.Contains(created, "Created job") {
		t.Fatalf("unexpected create result: %s", created)
	}

	jobs := tool.cronService.ListJobsForUser(7, true)
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}

	result, err := tool.Execute(ctx, map[string]interface{}{
		"action":        "update",
		"job_id":        jobs[0].ID,
		"message":       "updated message",
		"every_seconds": float64(120),
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}
	if !strings.Contains(result, "Updated job") {
		t.Fatalf("unexpected update result: %s", result)
	}

	updatedJobs := tool.cronService.ListJobsForUser(7, true)
	if got := updatedJobs[0].Payload.Message; got != "updated message" {
		t.Fatalf("expected updated message, got %q", got)
	}
	if updatedJobs[0].Schedule.EveryMS == nil || *updatedJobs[0].Schedule.EveryMS != 120000 {
		t.Fatalf("expected every_ms to be 120000, got %#v", updatedJobs[0].Schedule.EveryMS)
	}
	if got := updatedJobs[0].Name; got != "updated message" {
		t.Fatalf("expected derived name to follow updated message, got %q", got)
	}
}

func TestCronToolEditPreservesExistingFields(t *testing.T) {
	tool := newTestCronTool(t)
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"action":    "add",
		"message":   "keep schedule",
		"cron_expr": "0 9 * * *",
		"job_type":  "reminder",
	})
	if err != nil {
		t.Fatalf("add returned error: %v", err)
	}

	jobs := tool.cronService.ListJobsForUser(7, true)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"action": "edit",
		"job_id": jobs[0].ID,
		"name":   "renamed job",
	})
	if err != nil {
		t.Fatalf("edit returned error: %v", err)
	}
	if !strings.Contains(result, "Updated job") {
		t.Fatalf("unexpected edit result: %s", result)
	}

	updated := tool.cronService.ListJobsForUser(7, true)[0]
	if updated.Name != "renamed job" {
		t.Fatalf("expected custom name to persist, got %q", updated.Name)
	}
	if updated.Payload.Message != "keep schedule" {
		t.Fatalf("expected existing message to be preserved, got %q", updated.Payload.Message)
	}
	if updated.Schedule.Expr != "0 9 * * *" {
		t.Fatalf("expected existing cron expr to be preserved, got %q", updated.Schedule.Expr)
	}
	if updated.Payload.JobType != "reminder" {
		t.Fatalf("expected existing job type to be preserved, got %q", updated.Payload.JobType)
	}
}

func TestCronToolConsolidateRemovesOriginals(t *testing.T) {
	tool := newTestCronTool(t)
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"action":        "add",
		"message":       "job one",
		"every_seconds": float64(60),
	})
	if err != nil {
		t.Fatalf("first add returned error: %v", err)
	}
	_, err = tool.Execute(ctx, map[string]interface{}{
		"action":        "add",
		"message":       "job two",
		"every_seconds": float64(120),
	})
	if err != nil {
		t.Fatalf("second add returned error: %v", err)
	}

	sourceJobs := tool.cronService.ListJobsForUser(7, true)
	if len(sourceJobs) != 2 {
		t.Fatalf("expected 2 source jobs, got %d", len(sourceJobs))
	}

	result, err := tool.Execute(ctx, map[string]interface{}{
		"action":        "consolidate",
		"job_ids":       []interface{}{sourceJobs[0].ID, sourceJobs[1].ID},
		"message":       "combined job",
		"every_seconds": float64(300),
		"remove_originals": true,
		"confirmed":     true,
	})
	if err != nil {
		t.Fatalf("consolidate returned error: %v", err)
	}
	if !strings.Contains(result, "Consolidated") {
		t.Fatalf("unexpected consolidate result: %s", result)
	}

	finalJobs := tool.cronService.ListJobsForUser(7, true)
	if len(finalJobs) != 1 {
		t.Fatalf("expected 1 final job after consolidation, got %d", len(finalJobs))
	}
	if finalJobs[0].Payload.Message != "combined job" {
		t.Fatalf("expected consolidated message, got %q", finalJobs[0].Payload.Message)
	}
	if finalJobs[0].Schedule.EveryMS == nil || *finalJobs[0].Schedule.EveryMS != 300000 {
		t.Fatalf("expected consolidated schedule every 300000ms, got %#v", finalJobs[0].Schedule.EveryMS)
	}
}

func TestCronToolConsolidateKeepsOriginalsWhenRequested(t *testing.T) {
	tool := newTestCronTool(t)
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"action":    "add",
		"message":   "job alpha",
		"cron_expr": "0 8 * * *",
	})
	if err != nil {
		t.Fatalf("first add returned error: %v", err)
	}
	_, err = tool.Execute(ctx, map[string]interface{}{
		"action":    "add",
		"message":   "job beta",
		"cron_expr": "0 9 * * *",
	})
	if err != nil {
		t.Fatalf("second add returned error: %v", err)
	}

	sourceJobs := tool.cronService.ListJobsForUser(7, true)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"action":           "consolidate",
		"job_ids":          []interface{}{sourceJobs[0].ID, sourceJobs[1].ID},
		"name":             "morning consolidated",
		"cron_expr":        "0 10 * * *",
		"remove_originals": false,
	})
	if err != nil {
		t.Fatalf("consolidate returned error: %v", err)
	}
	if !strings.Contains(result, "Original jobs kept") {
		t.Fatalf("unexpected consolidate keep result: %s", result)
	}

	finalJobs := tool.cronService.ListJobsForUser(7, true)
	if len(finalJobs) != 3 {
		t.Fatalf("expected 3 jobs when keeping originals, got %d", len(finalJobs))
	}
}

func TestCronToolRemoveRequiresConfirmation(t *testing.T) {
	tool := newTestCronTool(t)
	ctx := WithSensitiveConfirmationRequired(context.Background(), true)

	_, err := tool.Execute(ctx, map[string]interface{}{
		"action":        "add",
		"message":       "job to remove",
		"every_seconds": float64(60),
	})
	if err != nil {
		t.Fatalf("add returned error: %v", err)
	}

	jobs := tool.cronService.ListJobsForUser(7, true)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"action": "remove",
		"job_id": jobs[0].ID,
	})
	if err != nil {
		t.Fatalf("remove returned error: %v", err)
	}
	if !strings.Contains(result, "Confirmation required") {
		t.Fatalf("expected confirmation requirement, got: %s", result)
	}

	confirmedResult, err := tool.Execute(ctx, map[string]interface{}{
		"action":    "remove",
		"job_id":    jobs[0].ID,
		"confirmed": true,
	})
	if err != nil {
		t.Fatalf("confirmed remove returned error: %v", err)
	}
	if !strings.Contains(confirmedResult, "Removed job") {
		t.Fatalf("expected successful confirmed remove, got: %s", confirmedResult)
	}
}

func TestCronToolConsolidateRequiresConfirmationWhenRemovingOriginals(t *testing.T) {
	tool := newTestCronTool(t)
	ctx := WithSensitiveConfirmationRequired(context.Background(), true)

	_, err := tool.Execute(ctx, map[string]interface{}{
		"action":        "add",
		"message":       "job first",
		"every_seconds": float64(60),
	})
	if err != nil {
		t.Fatalf("first add returned error: %v", err)
	}
	_, err = tool.Execute(ctx, map[string]interface{}{
		"action":        "add",
		"message":       "job second",
		"every_seconds": float64(90),
	})
	if err != nil {
		t.Fatalf("second add returned error: %v", err)
	}

	sourceJobs := tool.cronService.ListJobsForUser(7, true)
	result, err := tool.Execute(ctx, map[string]interface{}{
		"action":           "consolidate",
		"job_ids":          []interface{}{sourceJobs[0].ID, sourceJobs[1].ID},
		"message":          "merged",
		"every_seconds":    float64(300),
		"remove_originals": true,
	})
	if err != nil {
		t.Fatalf("consolidate returned error: %v", err)
	}
	if !strings.Contains(result, "Confirmation required") {
		t.Fatalf("expected confirmation requirement, got: %s", result)
	}
}