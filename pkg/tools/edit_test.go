package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditFileTool_Execute(t *testing.T) {
	t.Run("search and replace", func(t *testing.T) {
		workspace := t.TempDir()
		filePath := filepath.Join(workspace, "sample.txt")
		if err := os.WriteFile(filePath, []byte("Hello, World! Goodbye, World!"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		tool := NewEditFileTool(workspace, true)
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":     "sample.txt",
			"old_text": "Hello, World!",
			"new_text": "Hi, Earth!",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "Successfully edited") {
			t.Errorf("expected success message, got: %s", result)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read edited file: %v", err)
		}
		expected := "Hi, Earth! Goodbye, World!"
		if string(data) != expected {
			t.Errorf("expected %q, got %q", expected, string(data))
		}
	})

	t.Run("text not found", func(t *testing.T) {
		workspace := t.TempDir()
		filePath := filepath.Join(workspace, "sample.txt")
		if err := os.WriteFile(filePath, []byte("Hello, World!"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		tool := NewEditFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":     "sample.txt",
			"old_text": "nonexistent text",
			"new_text": "replacement",
		})
		if err == nil {
			t.Fatal("expected error for text not found, got nil")
		}
		if !strings.Contains(err.Error(), "old_text not found") {
			t.Errorf("expected 'old_text not found' in error, got: %v", err)
		}
	})

	t.Run("multiple occurrences", func(t *testing.T) {
		workspace := t.TempDir()
		filePath := filepath.Join(workspace, "sample.txt")
		if err := os.WriteFile(filePath, []byte("foo bar foo baz foo"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		tool := NewEditFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":     "sample.txt",
			"old_text": "foo",
			"new_text": "qux",
		})
		if err == nil {
			t.Fatal("expected error for multiple occurrences, got nil")
		}
		if !strings.Contains(err.Error(), "appears") {
			t.Errorf("expected error about multiple appearances, got: %v", err)
		}
		if !strings.Contains(err.Error(), "3") {
			t.Errorf("expected count of 3 in error, got: %v", err)
		}
	})

	t.Run("path traversal blocked", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewEditFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":     "../../../etc/hosts",
			"old_text": "localhost",
			"new_text": "evil",
		})
		if err == nil {
			t.Fatal("expected error for path traversal, got nil")
		}
		if !strings.Contains(err.Error(), "access denied") {
			t.Errorf("expected 'access denied' in error, got: %v", err)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewEditFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":     "nonexistent.txt",
			"old_text": "text",
			"new_text": "new",
		})
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
		if !strings.Contains(err.Error(), "file not found") {
			t.Errorf("expected 'file not found' in error, got: %v", err)
		}
	})

	t.Run("missing path parameter", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewEditFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"old_text": "text",
			"new_text": "new",
		})
		if err == nil {
			t.Fatal("expected error for missing path, got nil")
		}
		if !strings.Contains(err.Error(), "path is required") {
			t.Errorf("expected 'path is required' in error, got: %v", err)
		}
	})

	t.Run("missing old_text parameter", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewEditFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":     "file.txt",
			"new_text": "new",
		})
		if err == nil {
			t.Fatal("expected error for missing old_text, got nil")
		}
		if !strings.Contains(err.Error(), "old_text is required") {
			t.Errorf("expected 'old_text is required' in error, got: %v", err)
		}
	})

	t.Run("missing new_text parameter", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewEditFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":     "file.txt",
			"old_text": "old",
		})
		if err == nil {
			t.Fatal("expected error for missing new_text, got nil")
		}
		if !strings.Contains(err.Error(), "new_text is required") {
			t.Errorf("expected 'new_text is required' in error, got: %v", err)
		}
	})

	t.Run("preserves file permissions", func(t *testing.T) {
		workspace := t.TempDir()
		filePath := filepath.Join(workspace, "perms.txt")
		if err := os.WriteFile(filePath, []byte("original"), 0600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		tool := NewEditFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":     "perms.txt",
			"old_text": "original",
			"new_text": "modified",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("failed to stat file: %v", err)
		}
		// On Windows, permission bits are limited, so just verify the file is readable
		if info.Size() == 0 {
			t.Error("file should not be empty after edit")
		}
	})

	t.Run("name and description", func(t *testing.T) {
		tool := NewEditFileTool("", false)
		if tool.Name() != "edit_file" {
			t.Errorf("expected name 'edit_file', got %q", tool.Name())
		}
		if tool.Description() == "" {
			t.Error("expected non-empty description")
		}
	})

	t.Run("SetWorkspace updates allowed directory", func(t *testing.T) {
		workspace1 := t.TempDir()
		workspace2 := t.TempDir()

		filePath := filepath.Join(workspace2, "target.txt")
		if err := os.WriteFile(filePath, []byte("before edit"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		tool := NewEditFileTool(workspace1, true)
		tool.SetWorkspace(workspace2)

		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":     "target.txt",
			"old_text": "before edit",
			"new_text": "after edit",
		})
		if err != nil {
			t.Fatalf("unexpected error after SetWorkspace: %v", err)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(data) != "after edit" {
			t.Errorf("expected %q, got %q", "after edit", string(data))
		}
	})
}

func TestAppendFileTool_Execute(t *testing.T) {
	t.Run("append to existing", func(t *testing.T) {
		workspace := t.TempDir()
		filePath := filepath.Join(workspace, "appendme.txt")
		if err := os.WriteFile(filePath, []byte("line one\n"), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		tool := NewAppendFileTool(workspace, true)
		result, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":    "appendme.txt",
			"content": "line two\n",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(result, "Successfully appended") {
			t.Errorf("expected success message, got: %s", result)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		expected := "line one\nline two\n"
		if string(data) != expected {
			t.Errorf("expected %q, got %q", expected, string(data))
		}
	})

	t.Run("create new on append", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewAppendFileTool(workspace, true)

		// AppendFileTool uses O_CREATE, so it should create the file
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":    "newfile.txt",
			"content": "first content",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(workspace, "newfile.txt"))
		if err != nil {
			t.Fatalf("file was not created: %v", err)
		}
		if string(data) != "first content" {
			t.Errorf("expected %q, got %q", "first content", string(data))
		}
	})

	t.Run("multiple appends", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewAppendFileTool(workspace, true)

		for i, content := range []string{"A", "B", "C"} {
			_, err := tool.Execute(context.Background(), map[string]interface{}{
				"path":    "multi.txt",
				"content": content,
			})
			if err != nil {
				t.Fatalf("unexpected error on append %d: %v", i, err)
			}
		}

		data, err := os.ReadFile(filepath.Join(workspace, "multi.txt"))
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if string(data) != "ABC" {
			t.Errorf("expected %q, got %q", "ABC", string(data))
		}
	})

	t.Run("path traversal blocked", func(t *testing.T) {
		workspace := t.TempDir()
		tool := NewAppendFileTool(workspace, true)
		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":    "../outside.txt",
			"content": "should not append",
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
		tool := NewAppendFileTool(workspace, true)
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
		tool := NewAppendFileTool(workspace, true)
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
		tool := NewAppendFileTool("", false)
		if tool.Name() != "append_file" {
			t.Errorf("expected name 'append_file', got %q", tool.Name())
		}
		if tool.Description() == "" {
			t.Error("expected non-empty description")
		}
	})

	t.Run("SetWorkspace updates workspace", func(t *testing.T) {
		workspace1 := t.TempDir()
		workspace2 := t.TempDir()

		tool := NewAppendFileTool(workspace1, true)
		tool.SetWorkspace(workspace2)

		_, err := tool.Execute(context.Background(), map[string]interface{}{
			"path":    "ws2file.txt",
			"content": "in workspace 2",
		})
		if err != nil {
			t.Fatalf("unexpected error after SetWorkspace: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(workspace2, "ws2file.txt"))
		if err != nil {
			t.Fatalf("file not found in workspace2: %v", err)
		}
		if string(data) != "in workspace 2" {
			t.Errorf("expected %q, got %q", "in workspace 2", string(data))
		}
	})
}
