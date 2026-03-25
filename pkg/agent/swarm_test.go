package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
)

// ---------------------------------------------------------------------------
// SwarmRunner creation
// ---------------------------------------------------------------------------

func TestNewSwarmRunner(t *testing.T) {
	reg := NewSpecialistRegistry()
	ct := NewAgentCostTracker()
	sr := NewSwarmRunner(reg, ct)
	if sr == nil {
		t.Fatal("expected non-nil SwarmRunner")
	}
	if sr.registry != reg {
		t.Error("registry not set")
	}
	if sr.costTracker != ct {
		t.Error("costTracker not set")
	}
}

// ---------------------------------------------------------------------------
// RunSwarm validation
// ---------------------------------------------------------------------------

func TestRunSwarm_EmptyMembers(t *testing.T) {
	reg := NewSpecialistRegistry()
	sr := NewSwarmRunner(reg, nil)

	cfg := config.SwarmConfig{
		Name:    "empty",
		Members: []string{},
		Mode:    "sequential",
	}

	_, err := sr.RunSwarm(context.Background(), cfg, "task")
	if err == nil {
		t.Fatal("expected error for empty members")
	}
}

func TestRunSwarm_UnknownMember(t *testing.T) {
	reg := NewSpecialistRegistry()
	sr := NewSwarmRunner(reg, nil)

	cfg := config.SwarmConfig{
		Name:    "test",
		Members: []string{"nonexistent"},
		Mode:    "sequential",
	}

	_, err := sr.RunSwarm(context.Background(), cfg, "task")
	if err == nil {
		t.Fatal("expected error for unknown member")
	}
}

// ---------------------------------------------------------------------------
// Budget enforcement
// ---------------------------------------------------------------------------

func TestCheckBudget_Unlimited(t *testing.T) {
	sr := NewSwarmRunner(NewSpecialistRegistry(), nil)
	exec := &SwarmExecution{
		Config:  config.SwarmConfig{MaxBudget: 0},
		Results: map[string]*SwarmMemberResult{},
	}
	if err := sr.checkBudget(exec); err != nil {
		t.Fatalf("unexpected error for unlimited budget: %v", err)
	}
}

func TestCheckBudget_UnderLimit(t *testing.T) {
	sr := NewSwarmRunner(NewSpecialistRegistry(), nil)
	exec := &SwarmExecution{
		Config: config.SwarmConfig{MaxBudget: 1.0},
		Results: map[string]*SwarmMemberResult{
			"dev": {CostUSD: 0.5},
		},
	}
	if err := sr.checkBudget(exec); err != nil {
		t.Fatalf("unexpected error when under budget: %v", err)
	}
}

func TestCheckBudget_Exceeded(t *testing.T) {
	sr := NewSwarmRunner(NewSpecialistRegistry(), nil)
	exec := &SwarmExecution{
		SwarmName: "test",
		Config:    config.SwarmConfig{MaxBudget: 1.0},
		Results: map[string]*SwarmMemberResult{
			"dev":      {CostUSD: 0.6},
			"reviewer": {CostUSD: 0.5},
		},
	}
	if err := sr.checkBudget(exec); err == nil {
		t.Fatal("expected budget exceeded error")
	}
}

// ---------------------------------------------------------------------------
// Timeout enforcement
// ---------------------------------------------------------------------------

func TestRunSwarm_Timeout(t *testing.T) {
	reg := NewSpecialistRegistry()
	// Register a mock specialist — it won't actually run since timeout is near-instant
	sa := &SpecialistAgent{
		name:        "sleeper",
		description: "Slow agent",
	}
	_ = reg.RegisterSpecialist(sa)

	sr := NewSwarmRunner(reg, nil)

	cfg := config.SwarmConfig{
		Name:    "timeout-test",
		Members: []string{"sleeper"},
		Mode:    "sequential",
		Timeout: 1, // 1 second timeout
	}

	// Use an already-cancelled context to test immediate timeout behavior
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, _ := sr.RunSwarm(ctx, cfg, "do something slow")
	// The result may fail or succeed depending on timing, but should not hang
	if result != nil && result.Duration > 5*time.Second {
		t.Error("swarm took too long despite timeout")
	}
}

// ---------------------------------------------------------------------------
// Aggregate results
// ---------------------------------------------------------------------------

func TestAggregateResults_Sequential(t *testing.T) {
	sr := NewSwarmRunner(NewSpecialistRegistry(), nil)
	exec := &SwarmExecution{
		SwarmName: "test",
		Config: config.SwarmConfig{
			Name:    "test",
			Mode:    "sequential",
			Members: []string{"dev", "reviewer"},
		},
		Results: map[string]*SwarmMemberResult{
			"dev":      {Member: "dev", Status: "completed", Result: "Code written"},
			"reviewer": {Member: "reviewer", Status: "completed", Result: "Code reviewed"},
		},
		StartTime:   time.Now().Add(-time.Second),
		EndTime:     time.Now(),
		TotalCost:   0.05,
		TeamContext: &TeamContext{},
	}

	text := sr.aggregateResults(exec)
	if text == "" {
		t.Fatal("expected non-empty aggregated text")
	}
	if !strings.Contains(text, "test") {
		t.Error("aggregated text should contain swarm name")
	}
}

func TestAggregateResults_WithFailedMember(t *testing.T) {
	sr := NewSwarmRunner(NewSpecialistRegistry(), nil)
	exec := &SwarmExecution{
		SwarmName: "test",
		Config: config.SwarmConfig{
			Name:    "test",
			Mode:    "parallel",
			Members: []string{"dev", "broken"},
		},
		Results: map[string]*SwarmMemberResult{
			"dev":    {Member: "dev", Status: "completed", Result: "Done"},
			"broken": {Member: "broken", Status: "failed", Error: "provider error"},
		},
		StartTime:   time.Now().Add(-time.Second),
		EndTime:     time.Now(),
		TeamContext: &TeamContext{},
	}

	text := sr.aggregateResults(exec)
	if !strings.Contains(text, "provider error") {
		t.Error("aggregated text should contain error from failed member")
	}
}

// ---------------------------------------------------------------------------
// buildTaskWithSharedNotes
// ---------------------------------------------------------------------------

func TestBuildTaskWithSharedNotes_NoNotes(t *testing.T) {
	sr := NewSwarmRunner(NewSpecialistRegistry(), nil)
	exec := &SwarmExecution{
		Task: "original task",
		TeamContext: &TeamContext{
			SharedNotes: []TeamNote{},
		},
	}

	result := sr.buildTaskWithSharedNotes(exec, "dev")
	if result != "original task" {
		t.Errorf("expected original task, got %q", result)
	}
}

func TestBuildTaskWithSharedNotes_WithNotes(t *testing.T) {
	sr := NewSwarmRunner(NewSpecialistRegistry(), nil)
	exec := &SwarmExecution{
		Task: "review the code",
		TeamContext: &TeamContext{
			SharedNotes: []TeamNote{
				{Author: "dev", Content: "I wrote a REST API"},
			},
		},
	}

	result := sr.buildTaskWithSharedNotes(exec, "reviewer")
	if !strings.Contains(result, "dev") || !strings.Contains(result, "REST API") {
		t.Error("shared notes should be included in task")
	}
	if !strings.Contains(result, "review the code") {
		t.Error("original task should be included")
	}
}

func TestBuildTaskWithSharedNotes_SkipsSelf(t *testing.T) {
	sr := NewSwarmRunner(NewSpecialistRegistry(), nil)
	exec := &SwarmExecution{
		Task: "task",
		TeamContext: &TeamContext{
			SharedNotes: []TeamNote{
				{Author: "dev", Content: "my own note"},
			},
		},
	}

	result := sr.buildTaskWithSharedNotes(exec, "dev")
	// When only the current member's notes exist and are skipped, fall through
	if strings.Contains(result, "my own note") {
		t.Error("should not include own notes in context")
	}
}

// ---------------------------------------------------------------------------
// BuiltInSwarmTemplates
// ---------------------------------------------------------------------------

func TestBuiltInSwarmTemplates(t *testing.T) {
	if len(BuiltInSwarmTemplates) == 0 {
		t.Fatal("expected at least one built-in template")
	}

	for key, tmpl := range BuiltInSwarmTemplates {
		if tmpl.Name == "" {
			t.Errorf("template %q has empty name", key)
		}
		if len(tmpl.Members) == 0 {
			t.Errorf("template %q has no members", key)
		}
		if tmpl.Mode == "" {
			t.Errorf("template %q has empty mode", key)
		}
	}
}

// ---------------------------------------------------------------------------
// SwarmConfig defaults
// ---------------------------------------------------------------------------

func TestSwarmConfig_DefaultTimeout(t *testing.T) {
	cfg := config.SwarmConfig{
		Name:    "test",
		Members: []string{"dev"},
	}
	if cfg.Timeout != 0 {
		t.Error("timeout should default to 0 (handled at runtime as 300s)")
	}
}

func TestSwarmConfig_DefaultMode(t *testing.T) {
	cfg := config.SwarmConfig{
		Name:    "test",
		Members: []string{"dev"},
	}
	if cfg.Mode != "" {
		t.Error("mode should default to empty (handled at runtime as sequential)")
	}
}

// ---------------------------------------------------------------------------
// AgentManager swarm config management
// ---------------------------------------------------------------------------

func TestAgentManager_SwarmConfigCRUD(t *testing.T) {
	am := NewAgentManager(nil)

	// Initially empty
	list := am.GetSwarmsList()
	if len(list) != 0 {
		t.Errorf("expected empty swarms list, got %d", len(list))
	}

	// Set config
	cfg := config.SwarmConfig{
		Name:    "Test Swarm",
		Members: []string{"dev", "tester"},
		Mode:    "sequential",
	}
	am.SetSwarmConfig("test", cfg)

	// Get config
	got, ok := am.GetSwarmConfig("test")
	if !ok {
		t.Fatal("expected to find swarm config")
	}
	if got.Name != "Test Swarm" {
		t.Errorf("expected 'Test Swarm', got %q", got.Name)
	}

	// List
	list = am.GetSwarmsList()
	if len(list) != 1 {
		t.Errorf("expected 1 swarm, got %d", len(list))
	}

	// Delete
	deleted := am.DeleteSwarmConfig("test")
	if !deleted {
		t.Error("expected successful deletion")
	}

	// Delete nonexistent
	deleted = am.DeleteSwarmConfig("nonexistent")
	if deleted {
		t.Error("expected false for nonexistent deletion")
	}

	// Verify empty
	list = am.GetSwarmsList()
	if len(list) != 0 {
		t.Errorf("expected empty after delete, got %d", len(list))
	}
}

func TestAgentManager_GetSwarmConfig_NotFound(t *testing.T) {
	am := NewAgentManager(nil)
	_, ok := am.GetSwarmConfig("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestSwarmRunnerPropagatesUserContextToChildSpecialists(t *testing.T) {
	dataDir := t.TempDir()
	previousDataDir := config.GetDataDir()
	config.InitDataDir(dataDir)
	t.Cleanup(func() { config.InitDataDir(previousDataDir) })

	registry := NewSpecialistRegistry()
	provider := &staticProvider{response: "swarm complete"}
	cfg := newAgentTestConfig(filepath.Join(dataDir, "global-workspace"))
	specialist, err := NewSpecialistAgent("worker", &config.SpecialistConfig{
		Name:        "worker",
		Description: "test worker",
	}, cfg, nil, provider, nil)
	if err != nil {
		t.Fatalf("NewSpecialistAgent failed: %v", err)
	}
	if err := registry.RegisterSpecialist(specialist); err != nil {
		t.Fatalf("RegisterSpecialist failed: %v", err)
	}

	runner := NewSwarmRunner(registry, nil)
	ctx := ContextWithTeamContext(context.Background(), &TeamContext{
		UserUUID: "aaa-111",
		UserID:   7,
	})

	result, err := runner.RunSwarm(ctx, config.SwarmConfig{
		Name:         "isolation",
		Members:      []string{"worker"},
		Mode:         "sequential",
		SharedMemory: true,
	}, "write into the user workspace")
	if err != nil {
		t.Fatalf("RunSwarm failed: %v", err)
	}

	if specialist.userUUID != "aaa-111" {
		t.Fatalf("expected propagated user UUID aaa-111, got %q", specialist.userUUID)
	}
	if specialist.userID != 7 {
		t.Fatalf("expected propagated user ID 7, got %d", specialist.userID)
	}
	if result.TeamContext == nil || result.TeamContext.UserUUID != "aaa-111" || result.TeamContext.UserID != 7 {
		t.Fatalf("expected swarm result team context to keep propagated user identity, got %+v", result.TeamContext)
	}

	userSession := filepath.Join(dataDir, "users", "aaa-111", "workspace", "sessions", "specialist_worker.json")
	if _, err := os.Stat(userSession); err != nil {
		t.Fatalf("expected swarm specialist session in user workspace: %v", err)
	}
	otherUserSession := filepath.Join(dataDir, "users", "bbb-222", "workspace", "sessions", "specialist_worker.json")
	if _, err := os.Stat(otherUserSession); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no swarm session in another user's workspace, got err=%v", err)
	}
}
