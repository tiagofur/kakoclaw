package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMarketingCampaignTool_ExecuteCreatesWorkspace(t *testing.T) {
	workspace := t.TempDir()
	tool := NewMarketingCampaignTool(workspace)

	result, err := tool.Execute(context.Background(), map[string]any{
		"account":     "Acme Corp",
		"campaign":    "Spring Launch 2026",
		"description": "Launch our new developer workflow to early adopters.",
		"platforms":   "twitter,bluesky,linkedin,facebook",
		"objective":   "Product launch",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var out map[string]string
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("result should be valid JSON: %v", err)
	}

	basePath := filepath.Join(workspace, "marketing", "acme-corp", "spring-launch-2026")
	for _, dir := range []string{
		basePath,
		filepath.Join(basePath, "copy"),
		filepath.Join(basePath, "assets", "images"),
		filepath.Join(basePath, "assets", "videos"),
		filepath.Join(basePath, "schedules"),
		filepath.Join(basePath, "analytics"),
	} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected %s to exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}

	if out["status"] != "initialized" {
		t.Fatalf("expected initialized status, got %q", out["status"])
	}
	if out["base_path"] != "marketing/acme-corp/spring-launch-2026" {
		t.Fatalf("unexpected base_path: %q", out["base_path"])
	}
	if out["brief_path"] != "marketing/acme-corp/spring-launch-2026/brief.md" {
		t.Fatalf("unexpected brief_path: %q", out["brief_path"])
	}
	if out["status_path"] != "marketing/acme-corp/spring-launch-2026/status.json" {
		t.Fatalf("unexpected status_path: %q", out["status_path"])
	}
	if out["campaign_status"] != "draft" {
		t.Fatalf("expected draft campaign_status, got %q", out["campaign_status"])
	}

	briefPath := filepath.Join(basePath, "brief.md")
	brief, err := os.ReadFile(briefPath)
	if err != nil {
		t.Fatalf("expected brief.md to exist: %v", err)
	}
	briefText := string(brief)
	for _, snippet := range []string{
		"# Campaign Brief: acme-corp / spring-launch-2026",
		"**Objective:** Product launch",
		"**Platforms:** twitter,bluesky,linkedin,facebook",
		"Launch our new developer workflow to early adopters.",
		"marketing/acme-corp/spring-launch-2026/",
	} {
		if !strings.Contains(briefText, snippet) {
			t.Fatalf("expected brief to contain %q, got:\n%s", snippet, briefText)
		}
	}

	statusPath := filepath.Join(basePath, "status.json")
	statusData, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("expected status.json to exist: %v", err)
	}
	var status CampaignStatus
	if err := json.Unmarshal(statusData, &status); err != nil {
		t.Fatalf("status.json should be valid JSON: %v", err)
	}
	if status.Status != "draft" {
		t.Fatalf("expected draft status, got %q", status.Status)
	}
	if status.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at to be set")
	}
}

func TestMarketingCampaignTool_ExecuteSanitizesSlugs(t *testing.T) {
	workspace := t.TempDir()
	tool := NewMarketingCampaignTool(workspace)

	result, err := tool.Execute(context.Background(), map[string]any{
		"account":  " ACME & Co. ",
		"campaign": "Launch!!! 2026 / Beta",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	var out map[string]string
	if err := json.Unmarshal([]byte(result), &out); err != nil {
		t.Fatalf("result should be valid JSON: %v", err)
	}

	if out["account"] != "acme-co" {
		t.Fatalf("expected sanitized account, got %q", out["account"])
	}
	if out["campaign"] != "launch-2026-beta" {
		t.Fatalf("expected sanitized campaign, got %q", out["campaign"])
	}
	if _, err := os.Stat(filepath.Join(workspace, out["base_path"])); err != nil {
		t.Fatalf("expected sanitized base path to exist: %v", err)
	}
}

func TestCreateCampaignDir(t *testing.T) {
	workspace := t.TempDir()

	result, err := CreateCampaignDir(workspace, " ACME & Co. ", "Launch!!! 2026 / Beta", "desc", "objective", "twitter,linkedin")
	if err != nil {
		t.Fatalf("CreateCampaignDir failed: %v", err)
	}

	if result["account"] != "acme-co" {
		t.Fatalf("expected sanitized account, got %q", result["account"])
	}
	if result["campaign"] != "launch-2026-beta" {
		t.Fatalf("expected sanitized campaign, got %q", result["campaign"])
	}
	if result["campaign_status"] != "draft" {
		t.Fatalf("expected draft campaign status, got %q", result["campaign_status"])
	}

	statusPath := filepath.Join(workspace, result["status_path"])
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("expected status.json to exist: %v", err)
	}

	var status CampaignStatus
	if err := json.Unmarshal(data, &status); err != nil {
		t.Fatalf("invalid status.json: %v", err)
	}
	if status.Status != "draft" {
		t.Fatalf("expected draft status, got %q", status.Status)
	}
	if time.Since(status.UpdatedAt) > time.Minute {
		t.Fatalf("expected recent updated_at, got %v", status.UpdatedAt)
	}
}

func TestMarketingCampaignTool_ExecuteRejectsEmptyAccountOrCampaign(t *testing.T) {
	tool := NewMarketingCampaignTool(t.TempDir())

	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "empty account",
			args: map[string]any{"account": "", "campaign": "launch"},
		},
		{
			name: "empty campaign",
			args: map[string]any{"account": "acme", "campaign": "   "},
		},
		{
			name: "sanitized account becomes empty",
			args: map[string]any{"account": "!!!", "campaign": "launch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tool.Execute(context.Background(), tt.args)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), "account and campaign are required") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMarketingSanitizeSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "Spring Launch", want: "spring-launch"},
		{input: "Acme!!! #1", want: "acme-1"},
		{input: "mañana launch", want: "ma-ana-launch"},
		{input: "  ---Already-Slugged---  ", want: "already-slugged"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := sanitizeSlug(tt.input); got != tt.want {
				t.Fatalf("sanitizeSlug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
