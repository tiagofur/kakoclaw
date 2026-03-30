package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// MarketingCampaignTool initializes a marketing campaign workspace with organized folder structure.
type MarketingCampaignTool struct {
	workspace string
}

// CampaignStatus tracks the current lifecycle state of a marketing campaign.
type CampaignStatus struct {
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewMarketingCampaignTool creates a MarketingCampaignTool for the given workspace.
func NewMarketingCampaignTool(workspace string) *MarketingCampaignTool {
	return &MarketingCampaignTool{workspace: workspace}
}

func (t *MarketingCampaignTool) SetWorkspace(workspace string) { t.workspace = workspace }

func (t *MarketingCampaignTool) Name() string { return "marketing_init_campaign" }

func (t *MarketingCampaignTool) Description() string {
	return "Initializes a marketing campaign workspace with organized folder structure. Creates directories for copy, visual assets, schedules, and analytics. Returns the paths for each folder so agents know exactly where to save their outputs."
}

func (t *MarketingCampaignTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"account": map[string]any{
				"type":        "string",
				"description": "Company or brand identifier slug (e.g. 'acme-corp', 'nike'). Used as top-level folder name.",
			},
			"campaign": map[string]any{
				"type":        "string",
				"description": "Campaign identifier slug (e.g. 'summer-2026', 'product-launch'). Used as sub-folder name.",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Optional campaign description saved to brief.md",
			},
			"platforms": map[string]any{
				"type":        "string",
				"description": "Comma-separated list of target platforms (e.g. 'twitter,linkedin,instagram')",
			},
			"objective": map[string]any{
				"type":        "string",
				"description": "Optional campaign objective (e.g. 'brand awareness', 'product launch', 'lead generation')",
			},
		},
		"required": []string{"account", "campaign"},
	}
}

func (t *MarketingCampaignTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	account, _ := args["account"].(string)
	campaign, _ := args["campaign"].(string)
	description, _ := args["description"].(string)
	platforms, _ := args["platforms"].(string)
	objective, _ := args["objective"].(string)

	result, err := CreateCampaignDir(t.workspace, account, campaign, description, objective, platforms)
	if err != nil {
		return "", err
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

// CreateCampaignDir creates the marketing campaign folder structure and seed files.
func CreateCampaignDir(workspace, account, campaign, description, objective, platforms string) (map[string]string, error) {
	account = sanitizeSlug(account)
	campaign = sanitizeSlug(campaign)
	if account == "" || campaign == "" {
		return nil, fmt.Errorf("account and campaign are required and must be non-empty")
	}

	basePath := filepath.Join(workspace, "marketing", account, campaign)
	dirs := []string{
		basePath,
		filepath.Join(basePath, "copy"),
		filepath.Join(basePath, "assets", "images"),
		filepath.Join(basePath, "assets", "videos"),
		filepath.Join(basePath, "schedules"),
		filepath.Join(basePath, "analytics"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write brief.md
	brief := buildBriefContent(account, campaign, description, objective, platforms)
	briefPath := filepath.Join(basePath, "brief.md")
	if err := os.WriteFile(briefPath, []byte(brief), 0644); err != nil {
		return nil, fmt.Errorf("failed to write brief.md: %w", err)
	}

	statusPath := filepath.Join(basePath, "status.json")
	status := CampaignStatus{
		Status:    "draft",
		UpdatedAt: time.Now().UTC(),
	}
	statusData, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal status.json: %w", err)
	}
	if err := os.WriteFile(statusPath, statusData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write status.json: %w", err)
	}

	result := map[string]string{
		"base_path":       toRelative(workspace, basePath),
		"brief_path":      toRelative(workspace, briefPath),
		"status_path":     toRelative(workspace, statusPath),
		"copy_path":       toRelative(workspace, filepath.Join(basePath, "copy")),
		"images_path":     toRelative(workspace, filepath.Join(basePath, "assets", "images")),
		"videos_path":     toRelative(workspace, filepath.Join(basePath, "assets", "videos")),
		"schedules_path":  toRelative(workspace, filepath.Join(basePath, "schedules")),
		"analytics_path":  toRelative(workspace, filepath.Join(basePath, "analytics")),
		"account":         account,
		"campaign":        campaign,
		"status":          "initialized",
		"campaign_status": status.Status,
	}

	return result, nil
}

var reSlugInvalid = regexp.MustCompile(`[^a-z0-9\-_]`)
var reSlugMultiDash = regexp.MustCompile(`-{2,}`)

func sanitizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = reSlugInvalid.ReplaceAllString(s, "-")
	s = reSlugMultiDash.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// SanitizeMarketingSlug normalizes account and campaign names into safe slugs.
func SanitizeMarketingSlug(s string) string {
	return sanitizeSlug(s)
}

func toRelative(workspace, absPath string) string {
	rel, err := filepath.Rel(workspace, absPath)
	if err != nil {
		return absPath
	}
	return filepath.ToSlash(rel)
}

func buildBriefContent(account, campaign, description, objective, platforms string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Campaign Brief: %s / %s\n\n", account, campaign)
	fmt.Fprintf(&sb, "**Created:** %s\n\n", time.Now().Format("2006-01-02"))
	if objective != "" {
		fmt.Fprintf(&sb, "**Objective:** %s\n\n", objective)
	}
	if platforms != "" {
		fmt.Fprintf(&sb, "**Platforms:** %s\n\n", platforms)
	}
	if description != "" {
		sb.WriteString("## Description\n\n")
		sb.WriteString(description + "\n\n")
	}
	sb.WriteString("## Folder Structure\n\n")
	sb.WriteString("```\n")
	fmt.Fprintf(&sb, "marketing/%s/%s/\n", account, campaign)
	sb.WriteString("  brief.md          — this file\n")
	sb.WriteString("  strategy.md       — content strategy (by content_strategist)\n")
	sb.WriteString("  copy/             — marketing copy per platform\n")
	sb.WriteString("  assets/images/    — generated visual assets\n")
	sb.WriteString("  assets/videos/    — video assets\n")
	sb.WriteString("  schedules/        — posting schedules\n")
	sb.WriteString("  analytics/        — performance reports\n")
	sb.WriteString("```\n")
	return sb.String()
}
