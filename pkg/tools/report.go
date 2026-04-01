package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// SaveReportTool saves a research report or analysis as a Markdown file in the workspace reports/ folder.
type SaveReportTool struct {
	mu        sync.RWMutex
	workspace string
}

func NewSaveReportTool(workspace string) *SaveReportTool {
	return &SaveReportTool{workspace: workspace}
}

func (t *SaveReportTool) SetWorkspace(workspace string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.workspace = workspace
}

func (t *SaveReportTool) Name() string { return "save_report" }

func (t *SaveReportTool) Description() string {
	return `Saves a substantial research report or analysis as a Markdown file in the workspace reports/ folder.

Use this tool when:
- Multiple specialist agents contributed to a research task
- The result is substantial (research, multi-specialist analysis)
- The user explicitly requested a report, analysis, or documentation
- The content would be valuable to reference later

Do NOT use for simple answers, quick lookups, or conversational responses.
When unsure whether to save, ask the user first.`
}

func (t *SaveReportTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Report title — used in filename and as H1 header",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Full Markdown content of the report",
			},
			"filename": map[string]interface{}{
				"type":        "string",
				"description": "Custom filename. If empty, auto-generated as YYYY-MM-DD-{slug}.md where slug is title lowercased with spaces replaced by hyphens",
			},
		},
		"required": []string{"title", "content"},
	}
}

var nonAlphanumHyphen = regexp.MustCompile(`[^a-z0-9-]+`)

func titleToSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = nonAlphanumHyphen.ReplaceAllString(slug, "")
	slug = strings.Trim(slug, "-")
	return slug
}

func (t *SaveReportTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return "", fmt.Errorf("title is required")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content is required")
	}

	filename, _ := args["filename"].(string)
	if filename == "" {
		slug := titleToSlug(title)
		if slug == "" {
			slug = "report"
		}
		filename = fmt.Sprintf("%s-%s.md", time.Now().Format("2006-01-02"), slug)
	}

	t.mu.RLock()
	workspace := t.workspace
	t.mu.RUnlock()

	reportsDir := filepath.Join(workspace, "reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create reports directory: %w", err)
	}

	if !strings.HasPrefix(strings.TrimSpace(content), "#") {
		content = fmt.Sprintf("# %s\n\n%s", title, content)
	}

	targetPath := filepath.Join(reportsDir, filename)
	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write report: %w", err)
	}

	if cb := ReportSavedCallbackFromCtx(ctx); cb != nil {
		_ = cb(ReportSavedEvent{
			Title:     title,
			FilePath:  targetPath,
			Timestamp: time.Now(),
		})
	}

	return fmt.Sprintf("Report saved to reports/%s", filename), nil
}
