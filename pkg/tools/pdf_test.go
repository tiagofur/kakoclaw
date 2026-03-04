package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParsePageRange(t *testing.T) {
	tests := []struct {
		input   string
		want    map[int]bool
		wantErr bool
	}{
		{"", nil, false},
		{"3", map[int]bool{3: true}, false},
		{"1-3", map[int]bool{1: true, 2: true, 3: true}, false},
		{"1,3,5", map[int]bool{1: true, 3: true, 5: true}, false},
		{"1-2,5", map[int]bool{1: true, 2: true, 5: true}, false},
		{"abc", nil, true},
		{"0", nil, true},
		{"5-2", nil, true},
	}

	for _, tt := range tests {
		got, err := parsePageRange(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parsePageRange(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if tt.want == nil && got != nil {
			t.Errorf("parsePageRange(%q) = %v, want nil", tt.input, got)
		}
		if tt.want != nil {
			for k := range tt.want {
				if !got[k] {
					t.Errorf("parsePageRange(%q) missing page %d", tt.input, k)
				}
			}
		}
	}
}

func TestPDFToolExecute_NonExistentFile(t *testing.T) {
	tool := NewPDFTool(t.TempDir(), false)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"file_path": "/nonexistent/file.pdf",
	})
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestPDFToolExecute_NonPDFFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pdf")
	os.WriteFile(path, []byte("this is not a pdf"), 0644)

	tool := NewPDFTool(dir, false)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"file_path": path,
	})
	if err == nil {
		t.Error("expected error for non-PDF file")
	}
}

func TestPDFToolExecute_MissingFilePath(t *testing.T) {
	tool := NewPDFTool(t.TempDir(), false)
	_, err := tool.Execute(context.Background(), map[string]interface{}{})
	if err == nil {
		t.Error("expected error for missing file_path")
	}
}

func TestPDFToolExecute_InvalidPageRange(t *testing.T) {
	tool := NewPDFTool(t.TempDir(), false)
	_, err := tool.Execute(context.Background(), map[string]interface{}{
		"file_path": "/some/file.pdf",
		"pages":     "abc",
	})
	if err == nil {
		t.Error("expected error for invalid page range")
	}
}
