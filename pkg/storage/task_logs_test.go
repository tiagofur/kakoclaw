package storage

import "testing"

func TestGetTaskLogsLegacyReturnsEmptyForCrossUser(t *testing.T) {
	s := newTestStorage(t)

	taskID, err := s.CreateTaskForUser(int64(1), "owned task", "", "todo", "")
	if err != nil {
		t.Fatalf("CreateTaskForUser: %v", err)
	}
	if err := s.AddTaskLog(taskID, "created", "task was created"); err != nil {
		t.Fatalf("AddTaskLog: %v", err)
	}

	logs, err := s.GetTaskLogs(taskID, 2)
	if err != nil {
		t.Fatalf("GetTaskLogs cross-user: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("expected no logs for cross-user access, got %d", len(logs))
	}

	ownerLogs, err := s.GetTaskLogs(taskID, 1)
	if err != nil {
		t.Fatalf("GetTaskLogs owner: %v", err)
	}
	if len(ownerLogs) != 1 {
		t.Fatalf("expected owner to receive 1 log, got %d", len(ownerLogs))
	}
}
