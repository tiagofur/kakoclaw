package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// validatePath ensures the given path is within the workspace if restrict is true.
func validatePath(path, workspace string, restrict bool) (string, error) {
	if workspace == "" {
		return path, nil
	}

	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace path: %w", err)
	}

	var absPath string
	if filepath.IsAbs(path) {
		absPath = filepath.Clean(path)
	} else {
		absPath, err = filepath.Abs(filepath.Join(absWorkspace, path))
		if err != nil {
			return "", fmt.Errorf("failed to resolve file path: %w", err)
		}
	}

	if restrict {
		// Resolve symlinks to prevent symlink-based path traversal
		realWorkspace, err := filepath.EvalSymlinks(absWorkspace)
		if err != nil {
			// If workspace doesn't exist yet, fall back to Abs-based check
			realWorkspace = absWorkspace
		}

		// Try to resolve symlinks on the target path; if the file doesn't
		// exist yet (e.g. write_file creating a new file), resolve the parent.
		realPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			parentDir := filepath.Dir(absPath)
			realParent, err2 := filepath.EvalSymlinks(parentDir)
			if err2 != nil {
				realPath = absPath
			} else {
				realPath = filepath.Join(realParent, filepath.Base(absPath))
			}
		}

		// Use filepath.Rel to avoid prefix-based bypasses like /workspace-hack
		rel, err := filepath.Rel(realWorkspace, realPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", fmt.Errorf("access denied: path is outside the workspace")
		}
	}

	return absPath, nil
}

type ReadFileTool struct {
	mu        sync.RWMutex
	workspace string
	restrict  bool
}

func NewReadFileTool(workspace string, restrict bool) *ReadFileTool {
	return &ReadFileTool{workspace: workspace, restrict: restrict}
}

func (t *ReadFileTool) SetWorkspace(workspace string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.workspace = workspace
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file"
}

func (t *ReadFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path is required")
	}

	t.mu.RLock()
	workspace := t.workspace
	restrict := t.restrict
	t.mu.RUnlock()

	resolvedPath, err := validatePath(path, workspace, restrict)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

type WriteFileTool struct {
	mu        sync.RWMutex
	workspace string
	restrict  bool
}

func NewWriteFileTool(workspace string, restrict bool) *WriteFileTool {
	return &WriteFileTool{workspace: workspace, restrict: restrict}
}

func (t *WriteFileTool) SetWorkspace(workspace string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.workspace = workspace
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Write content to a file"
}

func (t *WriteFileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
			"confirmed": map[string]interface{}{
				"type":        "boolean",
				"description": "Set to true only after explicit user confirmation for sensitive targets.",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content is required")
	}

	t.mu.RLock()
	workspace := t.workspace
	restrict := t.restrict
	t.mu.RUnlock()

	resolvedPath, err := validatePath(path, workspace, restrict)
	if err != nil {
		return "", err
	}

	if IsSensitiveConfirmationRequired(ctx) && IsSensitiveWorkspacePath(resolvedPath, workspace) && !isConfirmed(args) {
		return "Confirmation required: write_file targets a sensitive file. Re-run with confirmed=true after user approval.", nil
	}

	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(resolvedPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("File written successfully to %s", resolvedPath), nil
}

type ListDirTool struct {
	mu        sync.RWMutex
	workspace string
	restrict  bool
}

func NewListDirTool(workspace string, restrict bool) *ListDirTool {
	return &ListDirTool{workspace: workspace, restrict: restrict}
}

func (t *ListDirTool) SetWorkspace(workspace string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.workspace = workspace
}

func (t *ListDirTool) Name() string {
	return "list_dir"
}

func (t *ListDirTool) Description() string {
	return "List files and directories in a path"
}

func (t *ListDirTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to list",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ListDirTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		path = "."
	}

	t.mu.RLock()
	workspace := t.workspace
	restrict := t.restrict
	t.mu.RUnlock()

	resolvedPath, err := validatePath(path, workspace, restrict)
	if err != nil {
		return "", err
	}

	entries, err := os.ReadDir(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	result := ""
	for _, entry := range entries {
		if entry.IsDir() {
			result += "DIR:  " + entry.Name() + "\n"
		} else {
			result += "FILE: " + entry.Name() + "\n"
		}
	}

	return result, nil
}
