package agent

import "testing"

func TestGetCostTracker(t *testing.T) {
	// Nil manager
	var nilMgr *AgentManager
	if nilMgr.GetCostTracker() != nil {
		t.Error("expected nil cost tracker from nil manager")
	}

	// Manager with nil default agent
	mgr := NewAgentManager(nil)
	if mgr.GetCostTracker() != nil {
		t.Error("expected nil cost tracker from manager with nil default agent")
	}
}

func TestCostSummary(t *testing.T) {
	tracker := NewAgentCostTracker()

	// Empty tracker
	summary := tracker.GetCostSummary()
	if summary.TotalCost != 0 {
		t.Errorf("expected 0 total cost, got %f", summary.TotalCost)
	}

	// Record some calls
	tracker.RecordAPICall("agent-a", 1000, 500, "gpt-4")
	tracker.RecordAPICall("agent-b", 2000, 1000, "claude-haiku")
	tracker.RecordAPICall("agent-a", 500, 200, "gpt-4")

	summary = tracker.GetCostSummary()
	if summary.TotalCost == 0 {
		t.Error("expected non-zero total cost after recording calls")
	}
	if len(summary.AgentCosts) != 2 {
		t.Errorf("expected 2 agents, got %d", len(summary.AgentCosts))
	}
	if summary.MostExpensive != "agent-a" {
		t.Errorf("expected most expensive to be agent-a, got %s", summary.MostExpensive)
	}

	// Test GetAllMetrics
	all := tracker.GetAllMetrics()
	if len(all) != 2 {
		t.Errorf("expected 2 agents in all metrics, got %d", len(all))
	}
	if all["agent-a"].APICallsCount != 2 {
		t.Errorf("expected 2 API calls for agent-a, got %d", all["agent-a"].APICallsCount)
	}
	if all["agent-b"].APICallsCount != 1 {
		t.Errorf("expected 1 API call for agent-b, got %d", all["agent-b"].APICallsCount)
	}
}

func TestCostReset(t *testing.T) {
	tracker := NewAgentCostTracker()

	tracker.RecordAPICall("agent-a", 1000, 500, "gpt-4")
	tracker.RecordAPICall("agent-b", 2000, 1000, "claude-haiku")

	if tracker.GetTotalCost() == 0 {
		t.Error("expected non-zero cost before reset")
	}

	tracker.Reset()

	if tracker.GetTotalCost() != 0 {
		t.Error("expected zero cost after reset")
	}
	all := tracker.GetAllMetrics()
	if len(all) != 0 {
		t.Errorf("expected 0 agents after reset, got %d", len(all))
	}
}
