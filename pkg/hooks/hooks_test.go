package hooks_test

import (
	"context"
	"testing"

	"github.com/sipeed/makoclaw/pkg/hooks"
	"github.com/stretchr/testify/require"
)

type mockHandler struct {
	priority int
	action   hooks.HookAction
	reason   string
	called   *bool
}

func (m *mockHandler) Priority() int { return m.priority }
func (m *mockHandler) Handle(_ context.Context, _ hooks.HookContext) hooks.HookResult {
	*m.called = true
	return hooks.HookResult{Action: m.action, Reason: m.reason}
}

func TestHookRegistry_AllowPassthrough(t *testing.T) {
	r := hooks.NewHookRegistry()
	called := false
	r.Register(&mockHandler{priority: 10, action: hooks.HookAllow, called: &called})
	result := r.Run(context.Background(), "before_tool_call", hooks.HookContext{ToolName: "exec"})
	require.Equal(t, hooks.HookAllow, result.Action)
	require.True(t, called)
}

func TestHookRegistry_BlockShortCircuits(t *testing.T) {
	r := hooks.NewHookRegistry()
	secondCalled := false
	r.Register(&mockHandler{priority: 10, action: hooks.HookBlock, reason: "denied", called: new(bool)})
	r.Register(&mockHandler{priority: 20, action: hooks.HookAllow, called: &secondCalled})
	result := r.Run(context.Background(), "before_tool_call", hooks.HookContext{ToolName: "exec"})
	require.Equal(t, hooks.HookBlock, result.Action)
	require.Equal(t, "denied", result.Reason)
	require.False(t, secondCalled, "second handler must not run after block")
}

func TestHookRegistry_PriorityOrder(t *testing.T) {
	r := hooks.NewHookRegistry()
	order := []int{}
	r.Register(&mockOrderHandler{p: 30, order: &order})
	r.Register(&mockOrderHandler{p: 10, order: &order})
	r.Register(&mockOrderHandler{p: 20, order: &order})
	r.Run(context.Background(), "before_tool_call", hooks.HookContext{})
	require.Equal(t, []int{10, 20, 30}, order)
}

type mockOrderHandler struct {
	p     int
	order *[]int
}

func (m *mockOrderHandler) Priority() int { return m.p }
func (m *mockOrderHandler) Handle(_ context.Context, _ hooks.HookContext) hooks.HookResult {
	*m.order = append(*m.order, m.p)
	return hooks.HookResult{Action: hooks.HookAllow}
}

func TestHookRegistry_PanicRecovered(t *testing.T) {
	r := hooks.NewHookRegistry()
	r.Register(&panicHandler{})
	result := r.Run(context.Background(), "before_tool_call", hooks.HookContext{ToolName: "exec"})
	// After panic recovery, default is allow
	require.Equal(t, hooks.HookAllow, result.Action)
}

type panicHandler struct{}

func (p *panicHandler) Priority() int { return 10 }
func (p *panicHandler) Handle(_ context.Context, _ hooks.HookContext) hooks.HookResult {
	panic("simulated hook panic")
}

func TestHookRegistry_EmptyRegistryAllows(t *testing.T) {
	r := hooks.NewHookRegistry()
	result := r.Run(context.Background(), "before_tool_call", hooks.HookContext{})
	require.Equal(t, hooks.HookAllow, result.Action)
}

func TestHookRegistry_PanicAfterBlockPreservesBlock(t *testing.T) {
	r := hooks.NewHookRegistry()
	r.Register(&panicHandler{}) // priority 10, runs first, panics -> result stays HookAllow
	r.Register(&mockHandler{priority: 20, action: hooks.HookBlock, reason: "blocked", called: new(bool)})
	result := r.Run(context.Background(), "before_tool_call", hooks.HookContext{})
	// panic handler runs first (p=10), panics, result stays HookAllow
	// block handler runs second (p=20), sets result to HookBlock, short-circuits
	require.Equal(t, hooks.HookBlock, result.Action)
	require.Equal(t, "blocked", result.Reason)
}
