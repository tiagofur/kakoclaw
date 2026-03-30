package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
)

func setupMarketingServer(t *testing.T) (*Server, *storage.User, string) {
	t.Helper()

	s := newTestServer(t)
	config.InitDataDir(s.workspace)
	t.Cleanup(func() { config.InitDataDir("") })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_ = withTestUserContext(t, s, req)

	user, err := s.centralStore.GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("failed to get admin user: %v", err)
	}
	workspace, err := config.EnsureUserWorkspace(user.UUID)
	if err != nil {
		t.Fatalf("EnsureUserWorkspace failed: %v", err)
	}

	return s, user, workspace
}

func seedMarketingCampaign(t *testing.T, workspace, account, campaign string) string {
	t.Helper()

	basePath := filepath.Join(workspace, "marketing", account, campaign)
	for _, dir := range []string{
		filepath.Join(basePath, "copy"),
		filepath.Join(basePath, "assets", "images"),
		filepath.Join(basePath, "schedules"),
		filepath.Join(basePath, "analytics"),
	} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create %s: %v", dir, err)
		}
	}

	files := map[string]string{
		filepath.Join(basePath, "brief.md"):                "# Brief\n\nLaunch week plan",
		filepath.Join(basePath, "strategy.md"):             "# Strategy\n\nFocus on developers",
		filepath.Join(basePath, "copy", "post.txt"):        "Ship faster with MakoClaw",
		filepath.Join(basePath, "schedules", "plan.json"):  `{"time":"2026-04-01T09:00:00Z"}`,
		filepath.Join(basePath, "analytics", "report.txt"): "CTR up 14%",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	statusData, err := json.Marshal(tools.CampaignStatus{Status: "active"})
	if err != nil {
		t.Fatalf("failed to marshal status: %v", err)
	}
	if err := os.WriteFile(filepath.Join(basePath, "status.json"), statusData, 0644); err != nil {
		t.Fatalf("failed to write status.json: %v", err)
	}

	pngData, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9p5nYe0AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatalf("failed to decode png fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(basePath, "assets", "images", "banner.png"), pngData, 0644); err != nil {
		t.Fatalf("failed to write image fixture: %v", err)
	}

	return basePath
}

func TestHandleListMarketingCampaigns(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		s, _, _ := setupMarketingServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns", nil)
		req = withTestUserContext(t, s, req)
		rr := httptest.NewRecorder()

		s.handleListMarketingCampaigns(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var out struct {
			Campaigns []CampaignInfo `json:"campaigns"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
			t.Fatalf("invalid response JSON: %v", err)
		}
		if len(out.Campaigns) != 0 {
			t.Fatalf("expected empty campaign list, got %d", len(out.Campaigns))
		}
	})

	t.Run("campaigns found", func(t *testing.T) {
		s, _, workspace := setupMarketingServer(t)
		seedMarketingCampaign(t, workspace, "acme", "launch")

		req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns", nil)
		req = withTestUserContext(t, s, req)
		rr := httptest.NewRecorder()

		s.handleListMarketingCampaigns(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var out struct {
			Campaigns []CampaignInfo `json:"campaigns"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
			t.Fatalf("invalid response JSON: %v", err)
		}
		if len(out.Campaigns) != 1 {
			t.Fatalf("expected 1 campaign, got %d", len(out.Campaigns))
		}

		campaign := out.Campaigns[0]
		if campaign.Account != "acme" || campaign.Campaign != "launch" {
			t.Fatalf("unexpected campaign payload: %+v", campaign)
		}
		if campaign.Status != "active" {
			t.Fatalf("expected active status, got %+v", campaign)
		}
		if !campaign.HasStrategy || !campaign.HasCopy || !campaign.HasAssets || !campaign.HasSchedule || !campaign.HasAnalytics {
			t.Fatalf("expected all presence flags to be true, got %+v", campaign)
		}
	})

	t.Run("auth required", func(t *testing.T) {
		s, _, _ := setupMarketingServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns", nil)
		rr := httptest.NewRecorder()

		s.handleListMarketingCampaigns(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})
}

func TestHandleGetMarketingCampaign(t *testing.T) {
	t.Run("valid campaign", func(t *testing.T) {
		s, _, workspace := setupMarketingServer(t)
		seedMarketingCampaign(t, workspace, "acme", "launch")

		req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns/acme/launch", nil)
		req = withTestUserContext(t, s, req)
		rr := httptest.NewRecorder()

		s.handleGetMarketingCampaign(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var detail CampaignDetail
		if err := json.Unmarshal(rr.Body.Bytes(), &detail); err != nil {
			t.Fatalf("invalid response JSON: %v", err)
		}
		if detail.Account != "acme" || detail.Campaign != "launch" {
			t.Fatalf("unexpected campaign detail: %+v", detail)
		}
		if detail.Status != "active" {
			t.Fatalf("expected active status, got %+v", detail)
		}
		if !strings.Contains(detail.Brief, "Launch week plan") {
			t.Fatalf("expected brief content, got %q", detail.Brief)
		}
		if !strings.Contains(detail.Strategy, "Focus on developers") {
			t.Fatalf("expected strategy content, got %q", detail.Strategy)
		}
		if len(detail.Files["copy"]) != 1 || len(detail.Files["assets"]) != 1 {
			t.Fatalf("expected copy and asset files, got %+v", detail.Files)
		}
		if !detail.Files["assets"][0].IsImage {
			t.Fatalf("expected image flag on assets entry, got %+v", detail.Files["assets"][0])
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, _, _ := setupMarketingServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns/acme/missing", nil)
		req = withTestUserContext(t, s, req)
		rr := httptest.NewRecorder()

		s.handleGetMarketingCampaign(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})

	t.Run("auth required", func(t *testing.T) {
		s, _, _ := setupMarketingServer(t)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns/acme/launch", nil)
		rr := httptest.NewRecorder()

		s.handleGetMarketingCampaign(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rr.Code)
		}
	})
}

func TestHandleGetMarketingFile(t *testing.T) {
	t.Run("serve text file", func(t *testing.T) {
		s, _, workspace := setupMarketingServer(t)
		seedMarketingCampaign(t, workspace, "acme", "launch")

		req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns/acme/launch/files/copy/post.txt", nil)
		req = withTestUserContext(t, s, req)
		rr := httptest.NewRecorder()

		s.handleGetMarketingFile(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
			t.Fatalf("expected text/plain content type, got %q", ct)
		}
		if body := rr.Body.String(); !strings.Contains(body, "Ship faster with MakoClaw") {
			t.Fatalf("unexpected body: %q", body)
		}
	})

	t.Run("serve image file", func(t *testing.T) {
		s, _, workspace := setupMarketingServer(t)
		basePath := seedMarketingCampaign(t, workspace, "acme", "launch")
		expected, err := os.ReadFile(filepath.Join(basePath, "assets", "images", "banner.png"))
		if err != nil {
			t.Fatalf("failed to read image fixture: %v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns/acme/launch/files/assets/images/banner.png", nil)
		req = withTestUserContext(t, s, req)
		rr := httptest.NewRecorder()

		s.handleGetMarketingFile(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
		if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "image/png") {
			t.Fatalf("expected image/png content type, got %q", ct)
		}
		if got := rr.Body.Bytes(); string(got) != string(expected) {
			t.Fatalf("expected served image bytes to match fixture")
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, _, workspace := setupMarketingServer(t)
		seedMarketingCampaign(t, workspace, "acme", "launch")

		req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns/acme/launch/files/copy/missing.txt", nil)
		req = withTestUserContext(t, s, req)
		rr := httptest.NewRecorder()

		s.handleGetMarketingFile(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rr.Code)
		}
	})
}

func TestHandleGetMarketingFile_PathTraversal(t *testing.T) {
	s, _, workspace := setupMarketingServer(t)
	seedMarketingCampaign(t, workspace, "acme", "launch")

	tests := []struct {
		name   string
		target string
	}{
		{
			name:   "account traversal",
			target: "/api/v1/marketing/campaigns/../../../etc/passwd/launch/files/brief.md",
		},
		{
			name:   "campaign traversal",
			target: "/api/v1/marketing/campaigns/acme/../../files/brief.md",
		},
		{
			name:   "url encoded file traversal",
			target: "/api/v1/marketing/campaigns/acme/launch/files/..%2F..%2Fsecret.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			req = withTestUserContext(t, s, req)
			rr := httptest.NewRecorder()

			s.handleGetMarketingFile(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Fatalf("expected 403 for %s, got %d: %s", tt.name, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestHandleCreateMarketingCampaign(t *testing.T) {
	t.Run("creates campaign", func(t *testing.T) {
		s, _, workspace := setupMarketingServer(t)

		body := `{"account":"Acme Corp","campaign":"Launch 2026","objective":"Ship","platforms":["Twitter","LinkedIn"],"description":"Hello"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/campaigns", bytes.NewBufferString(body))
		req = withTestUserContext(t, s, req)
		rr := httptest.NewRecorder()

		s.handleListMarketingCampaigns(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
		}

		var out CampaignInfo
		if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
			t.Fatalf("invalid response JSON: %v", err)
		}
		if out.Account != "acme-corp" || out.Campaign != "launch-2026" {
			t.Fatalf("unexpected campaign info: %+v", out)
		}
		if out.Status != "draft" {
			t.Fatalf("expected draft status, got %+v", out)
		}
		if _, err := os.Stat(filepath.Join(workspace, "marketing", "acme-corp", "launch-2026", "status.json")); err != nil {
			t.Fatalf("expected status.json to exist: %v", err)
		}
	})

	t.Run("rejects duplicate campaign", func(t *testing.T) {
		s, _, workspace := setupMarketingServer(t)
		seedMarketingCampaign(t, workspace, "acme", "launch")

		body := `{"account":"acme","campaign":"launch"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/campaigns", bytes.NewBufferString(body))
		req = withTestUserContext(t, s, req)
		rr := httptest.NewRecorder()

		s.handleListMarketingCampaigns(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d: %s", rr.Code, rr.Body.String())
		}
	})
}

func TestHandleRenameMarketingCampaign(t *testing.T) {
	s, _, workspace := setupMarketingServer(t)
	seedMarketingCampaign(t, workspace, "acme", "launch")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/marketing/campaigns/acme/launch", bytes.NewBufferString(`{"new_name":"Launch 2027"}`))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleMarketingCampaignsRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var out CampaignInfo
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if out.Campaign != "launch-2027" {
		t.Fatalf("expected renamed campaign, got %+v", out)
	}
	if _, err := os.Stat(filepath.Join(workspace, "marketing", "acme", "launch-2027", "status.json")); err != nil {
		t.Fatalf("expected status.json in renamed path: %v", err)
	}
}

func TestHandleDeleteMarketingCampaign(t *testing.T) {
	s, _, workspace := setupMarketingServer(t)
	seedMarketingCampaign(t, workspace, "acme", "launch")

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/marketing/campaigns/acme/launch", nil)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleMarketingCampaignsRouter(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
	if _, err := os.Stat(filepath.Join(workspace, "marketing", "acme", "launch")); !os.IsNotExist(err) {
		t.Fatalf("expected campaign directory removed, err=%v", err)
	}
}

func TestHandleUpdateCampaignStatus(t *testing.T) {
	s, _, workspace := setupMarketingServer(t)
	basePath := seedMarketingCampaign(t, workspace, "acme", "launch")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/marketing/campaigns/acme/launch/status", bytes.NewBufferString(`{"status":"paused"}`))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleMarketingCampaignsRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var out tools.CampaignStatus
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if out.Status != "paused" {
		t.Fatalf("expected paused status, got %+v", out)
	}

	data, err := os.ReadFile(filepath.Join(basePath, "status.json"))
	if err != nil {
		t.Fatalf("failed to read status.json: %v", err)
	}
	var stored tools.CampaignStatus
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("invalid status.json: %v", err)
	}
	if stored.Status != "paused" {
		t.Fatalf("expected persisted paused status, got %+v", stored)
	}
}

func TestHandleMarketingCampaign_DefaultsMissingStatusToDraft(t *testing.T) {
	s, _, workspace := setupMarketingServer(t)
	basePath := seedMarketingCampaign(t, workspace, "acme", "launch")
	if err := os.Remove(filepath.Join(basePath, "status.json")); err != nil {
		t.Fatalf("failed to remove status.json: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns/acme/launch", nil)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleGetMarketingCampaign(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var detail CampaignDetail
	if err := json.Unmarshal(rr.Body.Bytes(), &detail); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if detail.Status != "draft" {
		t.Fatalf("expected default draft status, got %+v", detail)
	}
}
