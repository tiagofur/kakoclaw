package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImageGenerateTool_Name(t *testing.T) {
	tool := NewImageGenerateTool(nil, "", false)
	if tool.Name() != "image_generate" {
		t.Fatalf("expected name 'image_generate', got %q", tool.Name())
	}
}

func TestImageGenerateTool_NilProvider(t *testing.T) {
	tool := NewImageGenerateTool(nil, "", false)
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]any{
		"prompt": "a cat sitting on a cloud",
	})
	if err != nil {
		t.Fatalf("expected no error with nil provider, got: %v", err)
	}
	if !strings.Contains(result, "No image provider configured") {
		t.Fatalf("expected helpful error message, got: %s", result)
	}
}

func TestImageGenerateTool_Generate(t *testing.T) {
	// Mock image server that returns a small PNG
	imageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Minimal valid PNG (1x1 pixel)
		png := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
			0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
			0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
			0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC,
			0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
			0x44, 0xAE, 0x42, 0x60, 0x82,
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(png)
	}))
	defer imageServer.Close()

	// Mock OpenAI API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/images/generations" {
			t.Errorf("expected /images/generations, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		var reqBody map[string]any
		json.NewDecoder(r.Body).Decode(&reqBody)

		if reqBody["prompt"] != "a cat sitting on a cloud" {
			t.Errorf("expected prompt 'a cat sitting on a cloud', got %v", reqBody["prompt"])
		}

		resp := map[string]any{
			"data": []map[string]any{
				{
					"url":            imageServer.URL + "/image.png",
					"revised_prompt": "A fluffy cat sitting on a white cloud in a blue sky",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer apiServer.Close()

	// Create temp workspace
	tmpDir, err := os.MkdirTemp("", "image-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	provider := NewOpenAIImageProvider("test-key", apiServer.URL, "dall-e-3")
	tool := NewImageGenerateTool(provider, tmpDir, false)

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]any{
		"prompt": "a cat sitting on a cloud",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var output map[string]any
	if err := json.Unmarshal([]byte(result), &output); err != nil {
		t.Fatalf("failed to parse result JSON: %v", err)
	}

	if output["url"] == nil || output["url"] == "" {
		t.Error("expected url in result")
	}
	if output["local_path"] == nil || output["local_path"] == "" {
		t.Error("expected local_path in result")
	}
	if output["revised_prompt"] != "A fluffy cat sitting on a white cloud in a blue sky" {
		t.Errorf("unexpected revised_prompt: %v", output["revised_prompt"])
	}

	// Verify file was actually saved
	localPath, ok := output["local_path"].(string)
	if !ok {
		t.Fatal("local_path is not a string")
	}
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		t.Errorf("image file was not saved at %s", localPath)
	}

	// Verify it's in the generated-images directory
	if !strings.Contains(localPath, "generated-images") {
		t.Errorf("expected file in generated-images directory, got %s", localPath)
	}
}

func TestImageGenerateTool_SetWorkspace(t *testing.T) {
	tool := NewImageGenerateTool(nil, "", false)
	tool.SetWorkspace("/tmp/test-workspace")

	if tool.workspace != "/tmp/test-workspace" {
		t.Fatalf("expected workspace '/tmp/test-workspace', got %q", tool.workspace)
	}

	// Verify it implements WorkspaceTool
	var _ WorkspaceTool = tool
}

func TestImageGenerateTool_CustomFilename(t *testing.T) {
	// Mock image server
	imageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte{0x89, 0x50, 0x4E, 0x47}) // minimal bytes
	}))
	defer imageServer.Close()

	// Mock API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"data": []map[string]any{
				{
					"url":            imageServer.URL + "/image.png",
					"revised_prompt": "revised",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer apiServer.Close()

	tmpDir, err := os.MkdirTemp("", "image-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	provider := NewOpenAIImageProvider("test-key", apiServer.URL, "dall-e-3")
	tool := NewImageGenerateTool(provider, tmpDir, false)

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]any{
		"prompt":   "test image",
		"filename": "my-custom-image",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var output map[string]any
	json.Unmarshal([]byte(result), &output)

	localPath := output["local_path"].(string)
	expectedFilename := "my-custom-image.png"
	if filepath.Base(localPath) != expectedFilename {
		t.Errorf("expected filename %q, got %q", expectedFilename, filepath.Base(localPath))
	}
}

func TestOpenAIImageProvider_Generate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/images/generations") {
			t.Errorf("expected path ending with /images/generations, got %s", r.URL.Path)
		}

		var reqBody map[string]any
		json.NewDecoder(r.Body).Decode(&reqBody)

		if reqBody["model"] != "dall-e-3" {
			t.Errorf("expected model dall-e-3, got %v", reqBody["model"])
		}
		if reqBody["size"] != "1792x1024" {
			t.Errorf("expected size 1792x1024, got %v", reqBody["size"])
		}
		if reqBody["style"] != "natural" {
			t.Errorf("expected style natural, got %v", reqBody["style"])
		}
		if reqBody["quality"] != "hd" {
			t.Errorf("expected quality hd, got %v", reqBody["quality"])
		}

		resp := map[string]any{
			"data": []map[string]any{
				{
					"url":            "https://example.com/generated.png",
					"revised_prompt": "A revised prompt here",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIImageProvider("test-api-key", server.URL, "dall-e-3")
	ctx := context.Background()

	result, err := provider.Generate(ctx, "a landscape painting", ImageOptions{
		Size:    "1792x1024",
		Style:   "natural",
		Quality: "hd",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result.URL != "https://example.com/generated.png" {
		t.Errorf("expected URL 'https://example.com/generated.png', got %q", result.URL)
	}
	if result.Revised != "A revised prompt here" {
		t.Errorf("expected revised prompt, got %q", result.Revised)
	}
}

func TestOpenAIImageProvider_Defaults(t *testing.T) {
	provider := NewOpenAIImageProvider("key", "", "")

	if provider.apiBase != "https://api.openai.com/v1" {
		t.Errorf("expected default apiBase 'https://api.openai.com/v1', got %q", provider.apiBase)
	}
	if provider.model != "dall-e-3" {
		t.Errorf("expected default model 'dall-e-3', got %q", provider.model)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"A Cat On A Cloud", "a-cat-on-a-cloud"},
		{"hello world!", "hello-world"},
		{"test@#$%image", "test-image"},
		{"", ""},
		{
			"this is a very long prompt that should be truncated because it exceeds the fifty character limit for filenames",
			"this-is-a-very-long-prompt-that-should-be-truncate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
