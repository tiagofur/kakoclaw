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
	return "Generate images using AI. Provide a detailed text description and get back an image saved to your workspace."
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

	// Download the image
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

	// Create output directory
	outputDir := filepath.Join(t.workspace, "generated-images")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
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

	outputPath := filepath.Join(outputDir, filename)

	// Save image to file
	if err := os.WriteFile(outputPath, imgData, 0644); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	result.LocalPath = outputPath

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
