package hooks

import (
	"context"
	"sort"

	"github.com/sipeed/makoclaw/pkg/logger"
)

// HookAction is the outcome of a HookHandler invocation.
type HookAction string

const (
	HookAllow           HookAction = "allow"
	HookBlock           HookAction = "block"
	HookRequireApproval HookAction = "require_approval"
)

// HookContext carries event metadata to each handler.
type HookContext struct {
	Event    string                 // "before_tool_call" | "before_install" | "message_sending"
	ToolName string                 // populated for before_tool_call
	Args     map[string]interface{} // populated for before_tool_call
	Message  string                 // populated for message_sending
	UserID   int64
}

// HookResult is the outcome returned by a handler.
type HookResult struct {
	Action HookAction
	Reason string
}

// HookHandler is the interface each security hook must implement.
type HookHandler interface {
	Priority() int
	Handle(ctx context.Context, hc HookContext) HookResult
}

// HookRegistry holds ordered HookHandler entries.
type HookRegistry struct {
	handlers []HookHandler
}

// NewHookRegistry creates an empty registry.
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{}
}

// Register adds a handler and re-sorts by ascending priority.
// Not safe for concurrent use — register all handlers before calling Run.
func (r *HookRegistry) Register(h HookHandler) {
	r.handlers = append(r.handlers, h)
	sort.Slice(r.handlers, func(i, j int) bool {
		return r.handlers[i].Priority() < r.handlers[j].Priority()
	})
}

// Run executes all handlers for the given event in priority order.
// Stops on the first block or require_approval result.
// Panics inside handlers are recovered and logged. A panicking handler does not
// update result, so the previous result is preserved: if no prior handler ran,
// the default HookAllow is kept (fail-open); if a prior handler returned HookBlock,
// that result is preserved (fail-secure). All handlers always see the event field
// set on HookContext.
func (r *HookRegistry) Run(ctx context.Context, event string, hc HookContext) (result HookResult) {
	result = HookResult{Action: HookAllow}
	hc.Event = event

	for _, h := range r.handlers {
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					// result is intentionally not reset: preserves prior handler's decision
					// (fail-open on first handler panic, fail-secure if prior returned block)
					logger.WarnCF("hooks", "handler panicked, defaulting to allow", map[string]interface{}{
						"event":   event,
						"recover": rec,
					})
				}
			}()
			result = h.Handle(ctx, hc)
		}()

		if result.Action != HookAllow {
			return result
		}
	}
	return result
}
