package agent

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

func TestInitializeOrchestratorSetsUserContextForAllSpecialists(t *testing.T) {
	dataDir := t.TempDir()
	previousDataDir := config.GetDataDir()
	config.InitDataDir(dataDir)
	t.Cleanup(func() { config.InitDataDir(previousDataDir) })

	cfg := newAgentTestConfig(filepath.Join(dataDir, "global-workspace"))
	cfg.Agents.Specialists = map[string]config.SpecialistConfig{
		"developer": {
			Name:        "developer",
			Description: "writes code",
		},
	}

	defaultAgent := NewAgentLoop(cfg, nil, &staticProvider{response: "ok"})
	defaultAgent.SetUserForAgent("aaa-111", 21)

	am := NewAgentManager(defaultAgent)
	if err := am.InitializeOrchestrator(cfg, nil, nil); err != nil {
		t.Fatalf("InitializeOrchestrator failed: %v", err)
	}

	if am.orchestrator == nil {
		t.Fatal("expected orchestrator to be initialized")
	}
	if am.orchestrator.AgentLoop.userUUID != "aaa-111" || am.orchestrator.AgentLoop.userID != 21 {
		t.Fatalf("expected orchestrator user context aaa-111/21, got %q/%d", am.orchestrator.AgentLoop.userUUID, am.orchestrator.AgentLoop.userID)
	}

	for name, specialist := range am.specialistReg.ListSpecialists() {
		if specialist.userUUID != "aaa-111" || specialist.userID != 21 {
			t.Fatalf("expected specialist %s user context aaa-111/21, got %q/%d", name, specialist.userUUID, specialist.userID)
		}
	}
}

func TestInitializeOrchestratorSpecialistWritesToUserWorkspace(t *testing.T) {
	dataDir := t.TempDir()
	previousDataDir := config.GetDataDir()
	config.InitDataDir(dataDir)
	t.Cleanup(func() { config.InitDataDir(previousDataDir) })

	cfg := newAgentTestConfig(filepath.Join(dataDir, "global-workspace"))
	cfg.Agents.Specialists = map[string]config.SpecialistConfig{
		"developer": {
			Name:        "developer",
			Description: "writes code",
		},
	}

	defaultAgent := NewAgentLoop(cfg, nil, &staticProvider{response: "ok"})
	defaultAgent.SetUserForAgent("aaa-111", 21)

	am := NewAgentManager(defaultAgent)
	if err := am.InitializeOrchestrator(cfg, nil, nil); err != nil {
		t.Fatalf("InitializeOrchestrator failed: %v", err)
	}

	specialist, err := am.specialistReg.GetSpecialist("developer")
	if err != nil {
		t.Fatalf("GetSpecialist failed: %v", err)
	}

	specialist.sessions.AddMessage("specialist-session", "user", "hello")
	session := specialist.sessions.GetOrCreate("specialist-session")
	if err := specialist.sessions.Save(session); err != nil {
		t.Fatalf("failed to save specialist session: %v", err)
	}

	userSessionPath := filepath.Join(dataDir, "users", "aaa-111", "workspace", "sessions", "specialist-session.json")
	if _, err := os.Stat(userSessionPath); err != nil {
		t.Fatalf("expected session file in user workspace: %v", err)
	}

	globalSessionPath := filepath.Join(cfg.WorkspacePath(), "sessions", "specialist-session.json")
	if _, err := os.Stat(globalSessionPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no specialist session in global workspace, got err=%v", err)
	}
}
