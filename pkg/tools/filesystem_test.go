package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileTool_Execute(t *testing.T) {
	t.Run("read existing file", func(t *testing.T) {
		workspace := t.TempDir()
		filePath := filepath.Join(workspace, "hello.txt")
		if err := os.WriteFile(filePath, []byte("hello world"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		tool := NewReadFileTool(workspace, true)
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": "hello.txt",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "hello world" {
			t.Errorf("expected %q, got %q", "hello world", result)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewReadFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": "nonexistent.txt",
		})
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read file") {
			t.Errorf("expected 'failed to read file' in error, got: %v", err)
		}
	})

	t.Run("path traversal blocked", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewReadFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": "../../../etc/passwd",
		})
		if err == nil {
			t.Fatal("expected error for path traversal, got nil")
		}
		if !strings.Contains(err.Error(), "access denied") {
			t.Errorf("expected 'access denied' in error, got: %v", err)
		}
	})

	t.Run("unrestricted mode", func(t *testing.T) {
		workspace := t.TempDir()
		siblingDir := t.TempDir()

		outsideFile := filepath.Join(siblingDir, "outside.txt")
		if err := os.WriteFile(outsideFile, []byte("outside content"), 0644); err != nil {
			t.Fatalf("failed to write outside file: %v", err)
		}

		tool := NewReadFileTool(workspace, false)
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": outsideFile,
		})
		if err != nil {
			t.Fatalf("unexpected error in unrestricted mode: %v", err)
		}
		if result != "outside content" {
			t.Errorf("expected %q, got %q", "outside content", result)
		}
	})

	t.Run("missing path parameter", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewReadFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{})
		if err == nil {
			t.Fatal("expected error for missing path, got nil")
		}
		if !strings.Contains(err.Error(), "path is required") {
			t.Errorf("expected 'path is required' in error, got: %v", err)
		}
	})

	t.Run("name and description", func(t *testing.T) {
		tool := NewReadFileTool("", false)
		if tool.Name() != "read_file" {
			t.Errorf("expected name 'read_file', got %q", tool.Name())
		}
		if tool.Description() == "" {
			t.Error("expected non-empty description")
		}
	})
}

func TestWriteFileTool_Execute(t *testing.T) {
	t.Run("write new file", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewWriteFileTool(workspace, true)
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":    "newfile.txt",
			"content": "new content",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "File written successfully") {
			t.Errorf("expected success message, got: %s", result)
		}

		data, err := os.ReadFile(filepath.Join(workspace, "newfile.txt"))
		if err != nil {
			t.Fatalf("failed to read written file: %v", err)
		}
		if string(data) != "new content" {
			t.Errorf("expected %q, got %q", "new content", string(data))
		}
	})

	t.Run("overwrite existing", func(t *testing.T) {
		workspace := t.TempDir()
		filePath := filepath.Join(workspace, "existing.txt")
		if err := os.WriteFile(filePath, []byte("original"), 0644); err != nil {
			t.Fatalf("failed to write initial file: %v", err)
		}

		tool := NewWriteFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":    "existing.txt",
			"content": "overwritten",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(data) != "overwritten" {
			t.Errorf("expected %q, got %q", "overwritten", string(data))
		}
	})

	t.Run("creates subdirectories", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewWriteFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":    "subdir/nested/file.txt",
			"content": "deep content",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(workspace, "subdir", "nested", "file.txt"))
		if err != nil {
			t.Fatalf("failed to read nested file: %v", err)
		}
		if string(data) != "deep content" {
			t.Errorf("expected %q, got %q", "deep content", string(data))
		}
	})

	t.Run("path traversal blocked", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewWriteFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":    "../outside.txt",
			"content": "should not be written",
		})
		if err == nil {
			t.Fatal("expected error for path traversal, got nil")
		}
		if !strings.Contains(err.Error(), "access denied") {
			t.Errorf("expected 'access denied' in error, got: %v", err)
		}
	})

	t.Run("missing path parameter", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewWriteFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"content": "some content",
		})
		if err == nil {
			t.Fatal("expected error for missing path, got nil")
		}
		if !strings.Contains(err.Error(), "path is required") {
			t.Errorf("expected 'path is required' in error, got: %v", err)
		}
	})

	t.Run("missing content parameter", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewWriteFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": "file.txt",
		})
		if err == nil {
			t.Fatal("expected error for missing content, got nil")
		}
		if !strings.Contains(err.Error(), "content is required") {
			t.Errorf("expected 'content is required' in error, got: %v", err)
		}
	})

	t.Run("name and description", func(t *testing.T) {
		tool := NewWriteFileTool("", false)
		if tool.Name() != "write_file" {
			t.Errorf("expected name 'write_file', got %q", tool.Name())
		}
		if tool.Description() == "" {
			t.Error("expected non-empty description")
		}
	})
}

func TestListDirTool_Execute(t *testing.T) {
	t.Run("list files", func(t *testing.T) {
		workspace := t.TempDir()
		// Create files and a subdirectory
		for _, name := range []string{"alpha.txt", "beta.txt", "gamma.txt"} {
			if err := os.WriteFile(filepath.Join(workspace, name), []byte("data"), 0644); err != nil {
				t.Fatalf("failed to create file %s: %v", name, err)
			}
		}
		if err := os.Mkdir(filepath.Join(workspace, "subdir"), 0755); err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}

		tool := NewListDirTool(workspace, true)
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": ".",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, name := range []string{"alpha.txt", "beta.txt", "gamma.txt"} {
			if !strings.Contains(result, name) {
				t.Errorf("expected %q in result, got: %s", name, result)
			}
		}
		if !strings.Contains(result, "FILE:") {
			t.Errorf("expected 'FILE:' prefix in result, got: %s", result)
		}
		if !strings.Contains(result, "DIR:") {
			t.Errorf("expected 'DIR:' prefix for subdir in result, got: %s", result)
		}
		if !strings.Contains(result, "subdir") {
			t.Errorf("expected 'subdir' in result, got: %s", result)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewListDirTool(workspace, true)
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": ".",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "" {
			t.Errorf("expected empty result for empty directory, got: %q", result)
		}
	})

	t.Run("nested directory", func(t *testing.T) {
		workspace := t.TempDir()
		nestedDir := filepath.Join(workspace, "level1", "level2")
		if err := os.MkdirAll(nestedDir, 0755); err != nil {
			t.Fatalf("failed to create nested dirs: %v", err)
		}
		if err := os.WriteFile(filepath.Join(nestedDir, "deep.txt"), []byte("deep"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}

		tool := NewListDirTool(workspace, true)

		// List the nested directory
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": "level1/level2",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "deep.txt") {
			t.Errorf("expected 'deep.txt' in result, got: %s", result)
		}

		// List the parent to see the nested folder
		result, err = tool.Execute(context.Background(), map[string]interface{}{
			"path": "level1",
		})
		if err != nil {
			t.Fatalf("unexpected error listing parent: %v", err)
		}
		if !strings.Contains(result, "level2") {
			t.Errorf("expected 'level2' in result, got: %s", result)
		}
		if !strings.Contains(result, "DIR:") {
			t.Errorf("expected 'DIR:' prefix for nested dir, got: %s", result)
		}
	})

	t.Run("path traversal blocked", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewListDirTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": "../..",
		})
		if err == nil {
			t.Fatal("expected error for path traversal, got nil")
		}
		if !strings.Contains(err.Error(), "access denied") {
			t.Errorf("expected 'access denied' in error, got: %v", err)
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewListDirTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": "does_not_exist",
		})
		if err == nil {
			t.Fatal("expected error for nonexistent directory, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read directory") {
			t.Errorf("expected 'failed to read directory' in error, got: %v", err)
		}
	})

	t.Run("name and description", func(t *testing.T) {
		tool := NewListDirTool("", false)
		if tool.Name() != "list_dir" {
			t.Errorf("expected name 'list_dir', got %q", tool.Name())
		}
		if tool.Description() == "" {
			t.Error("expected non-empty description")
		}
	})
}

func TestValidatePath(t *testing.T) {
	t.Run("empty workspace allows any path", func(t *testing.T) {
		result, err := validatePath("/some/path", "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "/some/path" {
			t.Errorf("expected path to pass through, got: %s", result)
		}
	})

	t.Run("relative path resolved within workspace", func(t *testing.T) {
		workspace := t.TempDir()
		result, err := validatePath("file.txt", workspace, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(workspace, "file.txt")
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("traversal blocked with restrict", func(t *testing.T) {
		workspace := t.TempDir()
		_, err := validatePath("../../etc/passwd", workspace, true)
		if err == nil {
			t.Fatal("expected error for path traversal, got nil")
		}
		if !strings.Contains(err.Error(), "access denied") {
			t.Errorf("expected 'access denied' in error, got: %v", err)
		}
	})

	t.Run("traversal allowed without restrict", func(t *testing.T) {
		workspace := t.TempDir()
		_, err := validatePath("../../etc/passwd", workspace, false)
		if err != nil {
			t.Fatalf("unexpected error without restrict: %v", err)
		}
	})
}

func TestSetWorkspace(t *testing.T) {
	t.Run("ReadFileTool SetWorkspace", func(t *testing.T) {
		workspace1 := t.TempDir()
		workspace2 := t.TempDir()

		// Write a file in workspace2
		if err := os.WriteFile(filepath.Join(workspace2, "test.txt"), []byte("ws2 content"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		tool := NewReadFileTool(workspace1, true)
		tool.SetWorkspace(workspace2)

		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": "test.txt",
		})
		if err != nil {
			t.Fatalf("unexpected error after SetWorkspace: %v", err)
		}
		if result != "ws2 content" {
			t.Errorf("expected %q, got %q", "ws2 content", result)
		}
	})

	t.Run("WriteFileTool SetWorkspace", func(t *testing.T) {
		workspace1 := t.TempDir()
		workspace2 := t.TempDir()

		tool := NewWriteFileTool(workspace1, true)
		tool.SetWorkspace(workspace2)

		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":    "test.txt",
			"content": "written to ws2",
		})
		if err != nil {
			t.Fatalf("unexpected error after SetWorkspace: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(workspace2, "test.txt"))
		if err != nil {
			t.Fatalf("file not found in workspace2: %v", err)
		}
		if string(data) != "written to ws2" {
			t.Errorf("expected %q, got %q", "written to ws2", string(data))
		}
	})

	t.Run("ListDirTool SetWorkspace", func(t *testing.T) {
		workspace1 := t.TempDir()
		workspace2 := t.TempDir()

		if err := os.WriteFile(filepath.Join(workspace2, "inws2.txt"), []byte("data"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		tool := NewListDirTool(workspace1, true)
		tool.SetWorkspace(workspace2)

		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"path": ".",
		})
		if err != nil {
			t.Fatalf("unexpected error after SetWorkspace: %v", err)
		}
		if !strings.Contains(result, "inws2.txt") {
			t.Errorf("expected 'inws2.txt' in listing after SetWorkspace, got: %s", result)
		}
	})
}
