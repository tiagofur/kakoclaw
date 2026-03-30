package agent

import (
	"testing"
)

func TestLLMTaskRegistered(t *testing.T) {
	cfg := newAgentTestConfig(t.TempDir())
	loop := NewAgentLoop(cfg, nil, nil)
	tool, ok := loop.tools.Get("llm_task")
	if !ok || tool == nil {
		t.Fatal("llm_task tool not registered in AgentLoop")
	}
}
