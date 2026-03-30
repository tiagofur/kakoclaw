package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var imageHTTPClient = &http.Client{
	Timeout: 60 * time.Second,
}

var reNonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

type ImageGenerateTool struct {
	provider  ImageProvider
	workspace string
	restrict  bool
}

func NewImageGenerateTool(provider ImageProvider, workspace string, restrict bool) *ImageGenerateTool {
	return &ImageGenerateTool{
		provider:  provider,
		workspace: workspace,
		restrict:  restrict,
	}
}

func (t *ImageGenerateTool) Name() string {
	return "image_generate"
}

func (t *ImageGenerateTool) Description() string {
	return "Generate images using AI. The generated image will be automatically displayed as an inline visual preview in the chat — the user will see it and can download it directly. Provide a detailed text description of the image to create."
}

func (t *ImageGenerateTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"prompt": map[string]any{
				"type":        "string",
				"description": "Detailed description of the image to generate",
			},
			"size": map[string]any{
				"type":        "string",
				"description": "Image dimensions",
				"enum":        []string{"1024x1024", "1792x1024", "1024x1792"},
			},
			"style": map[string]any{
				"type":        "string",
				"description": "Image style - vivid for hyper-real, natural for more natural",
				"enum":        []string{"vivid", "natural"},
			},
			"quality": map[string]any{
				"type":        "string",
				"description": "Image quality",
				"enum":        []string{"standard", "hd"},
			},
			"filename": map[string]any{
				"type":        "string",
				"description": "Custom filename for the saved image",
			},
			"output_dir": map[string]any{
				"type":        "string",
				"description": "Relative path within workspace to save the image. Defaults to 'generated-images'.",
			},
		},
		"required": []string{"prompt"},
	}
}

func (t *ImageGenerateTool) SetWorkspace(workspace string) {
	t.workspace = workspace
}

func (t *ImageGenerateTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	if t.provider == nil {
		return "Error: No image provider configured. Set MAKOCLAW_TOOLS_IMAGE_API_KEY to enable image generation.", nil
	}

	prompt, ok := args["prompt"].(string)
	if !ok || strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("prompt is required")
	}

	size := "1024x1024"
	if s, ok := args["size"].(string); ok && s != "" {
		size = s
	}

	style := "vivid"
	if s, ok := args["style"].(string); ok && s != "" {
		style = s
	}

	quality := "standard"
	if q, ok := args["quality"].(string); ok && q != "" {
		quality = q
	}

	result, err := t.provider.Generate(ctx, prompt, ImageOptions{
		Size:    size,
		Style:   style,
		Quality: quality,
	})
	if err != nil {
		return "", fmt.Errorf("image generation failed: %w", err)
	}

	// Create output directory
	outputSubDir := "generated-images"
	if v, ok := args["output_dir"].(string); ok && v != "" {
		outputSubDir = v
	}
	saveDir := filepath.Join(t.workspace, outputSubDir)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine filename
	var filename string
	if customName, ok := args["filename"].(string); ok && customName != "" {
		filename = customName
		if !strings.HasSuffix(strings.ToLower(filename), ".png") {
			filename += ".png"
		}
	} else {
		filename = fmt.Sprintf("%d_%s.png", time.Now().Unix(), sanitizeFilename(prompt))
	}

	if result.LocalPath != "" && result.URL == "" {
		// Google provider returns a temp LocalPath instead of a URL — move it into workspace
		if err := googleImageTempToWorkspace(result, saveDir, filename); err != nil {
			return "", fmt.Errorf("failed to move google image to workspace: %w", err)
		}
	} else {
		// Download the image from URL
		imgReq, err := http.NewRequestWithContext(ctx, "GET", result.URL, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create download request: %w", err)
		}

		imgResp, err := imageHTTPClient.Do(imgReq)
		if err != nil {
			return "", fmt.Errorf("failed to download image: %w", err)
		}
		defer imgResp.Body.Close()

		const maxImageSize = 20 * 1024 * 1024 // 20 MB cap
		imgData, err := io.ReadAll(io.LimitReader(imgResp.Body, maxImageSize))
		if err != nil {
			return "", fmt.Errorf("failed to read image data: %w", err)
		}

		outputPath := filepath.Join(saveDir, filename)
		if err := os.WriteFile(outputPath, imgData, 0644); err != nil {
			return "", fmt.Errorf("failed to save image: %w", err)
		}
		result.LocalPath = outputPath
	}

	// Return JSON result
	output := map[string]any{
		"url":             result.URL,
		"local_path":      result.LocalPath,
		"revised_prompt":  result.Revised,
	}

	resultJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(resultJSON), nil
}

// ShowImageTool displays an existing workspace image as an inline preview in chat.
type ShowImageTool struct {
	workspace string
}

func NewShowImageTool(workspace string) *ShowImageTool {
	return &ShowImageTool{workspace: workspace}
}

func (t *ShowImageTool) Name() string { return "show_image" }

func (t *ShowImageTool) Description() string {
	return "Display an existing image file from the workspace as an inline visual preview in the chat. The user will see the image and can download it. Provide the relative path to the image file (e.g. 'generated-images/photo.png')."
}

func (t *ShowImageTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Relative path to the image file within the workspace",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ShowImageTool) SetWorkspace(workspace string) { t.workspace = workspace }

func (t *ShowImageTool) Execute(_ context.Context, args map[string]any) (string, error) {
	relPath, ok := args["path"].(string)
	if !ok || strings.TrimSpace(relPath) == "" {
		return "", fmt.Errorf("path is required")
	}

	absPath := filepath.Join(t.workspace, relPath)
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Sprintf("Error: file not found at path %q", relPath), nil
	}

	ext := strings.ToLower(filepath.Ext(absPath))
	imageExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true}
	if !imageExts[ext] {
		return fmt.Sprintf("Error: %q is not a supported image file (png, jpg, jpeg, gif, webp, svg)", relPath), nil
	}

	output := map[string]any{
		"url":            "",
		"local_path":     absPath,
		"revised_prompt": "",
	}
	resultJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(resultJSON), nil
}

// sanitizeFilename creates a safe filename from the prompt text.
func sanitizeFilename(prompt string) string {
	name := strings.ToLower(prompt)
	name = reNonAlphanumeric.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if len(name) > 50 {
		name = name[:50]
	}
	return name
}
