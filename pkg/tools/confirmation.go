package tools

import (
	"context"
	"path/filepath"
	"strings"
)

type confirmationContextKey struct{}

// WithSensitiveConfirmationRequired marks a context to enforce explicit confirmations
// for sensitive or destructive tool actions.
func WithSensitiveConfirmationRequired(ctx context.Context, required bool) context.Context {
	return context.WithValue(ctx, confirmationContextKey{}, required)
}

// IsSensitiveConfirmationRequired reports whether sensitive confirmation checks are enabled.
func IsSensitiveConfirmationRequired(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	required, ok := ctx.Value(confirmationContextKey{}).(bool)
	return ok && required
}

// IsSensitiveWorkspacePath reports whether the path targets bootstrap identity
// files or memory files in the workspace.
func IsSensitiveWorkspacePath(path, workspace string) bool {
	if path == "" || workspace == "" {
		return false
	}

	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(absWorkspace, absPath)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return false
	}

	rel = strings.ReplaceAll(filepath.Clean(rel), "\\", "/")

	for _, name := range []string{"AGENTS.md", "SOUL.md", "USER.md", "IDENTITY.md"} {
		if strings.EqualFold(rel, name) {
			return true
		}
	}

	return strings.HasPrefix(strings.ToLower(rel), "memory/")
}
