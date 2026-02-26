package utils

import "context"

type agentNameKey struct{}

// WithAgentName injects the name of the active agent/specialist into the context.
func WithAgentName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, agentNameKey{}, name)
}

// AgentNameFrom retrieves the active agent/specialist name from the context.
// Returns an empty string if not set.
func AgentNameFrom(ctx context.Context) string {
	if val, ok := ctx.Value(agentNameKey{}).(string); ok {
		return val
	}
	return ""
}
