package web

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
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

func newPNGFixture(t *testing.T) []byte {
	t.Helper()
	pngData, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAusB9p5nYe0AAAAASUVORK5CYII=")
	if err != nil {
		t.Fatalf("failed to decode png fixture: %v", err)
	}
	return pngData
}

func newMultipartFileRequest(t *testing.T, method, target, fieldName, filename string, data []byte, formValues map[string]string) *http.Request {
	t.Helper()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}

	for key, value := range formValues {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("failed to write form field %s: %v", key, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
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

	t.Run("skips media directories", func(t *testing.T) {
		s, _, workspace := setupMarketingServer(t)
		mediaDir := filepath.Join(workspace, "marketing", "acme", "_MeDiA-library")
		if err := os.MkdirAll(mediaDir, 0755); err != nil {
			t.Fatalf("failed to create media dir: %v", err)
		}

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
			t.Fatalf("expected media directories to be skipped, got %+v", out.Campaigns)
		}
	})
}

func TestHandleMarketingTemplatesCRUD(t *testing.T) {
	s, _, workspace := setupMarketingServer(t)
	if err := os.MkdirAll(filepath.Join(workspace, "marketing", "acme"), 0755); err != nil {
		t.Fatalf("failed to create account dir: %v", err)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/templates/acme", bytes.NewBufferString(`{"slug":"launch-template","name":"Launch","body":"Hola {{first_name}} from {{ company }}","platform":"twitter"}`))
	createReq = withTestUserContext(t, s, createReq)
	createRR := httptest.NewRecorder()

	s.handleMarketingTemplatesRouter(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", createRR.Code, createRR.Body.String())
	}

	var created TemplateInfo
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("invalid create response JSON: %v", err)
	}
	if created.Slug != "launch-template" || created.Name != "Launch" || created.Platform != "twitter" {
		t.Fatalf("unexpected template payload: %+v", created)
	}
	slices.Sort(created.Variables)
	if !slices.Equal(created.Variables, []string{"company", "first_name"}) {
		t.Fatalf("unexpected variables: %+v", created.Variables)
	}

	duplicateReq := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/templates/acme", bytes.NewBufferString(`{"slug":"launch-template","name":"Launch","body":"Hello"}`))
	duplicateReq = withTestUserContext(t, s, duplicateReq)
	duplicateRR := httptest.NewRecorder()

	s.handleMarketingTemplatesRouter(duplicateRR, duplicateReq)

	if duplicateRR.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate template, got %d: %s", duplicateRR.Code, duplicateRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/templates/acme", nil)
	listReq = withTestUserContext(t, s, listReq)
	listRR := httptest.NewRecorder()

	s.handleMarketingTemplatesRouter(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", listRR.Code, listRR.Body.String())
	}

	var listed struct {
		Templates []TemplateInfo `json:"templates"`
	}
	if err := json.Unmarshal(listRR.Body.Bytes(), &listed); err != nil {
		t.Fatalf("invalid list response JSON: %v", err)
	}
	if len(listed.Templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(listed.Templates))
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/v1/marketing/templates/acme/launch-template", bytes.NewBufferString(`{"name":"Launch v2","body":"Hi {{first_name}} - {{promo_code}}","platform":"linkedin"}`))
	updateReq = withTestUserContext(t, s, updateReq)
	updateRR := httptest.NewRecorder()

	s.handleMarketingTemplatesRouter(updateRR, updateReq)

	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", updateRR.Code, updateRR.Body.String())
	}

	var updated TemplateInfo
	if err := json.Unmarshal(updateRR.Body.Bytes(), &updated); err != nil {
		t.Fatalf("invalid update response JSON: %v", err)
	}
	if updated.Name != "Launch v2" || updated.Platform != "linkedin" {
		t.Fatalf("unexpected updated template: %+v", updated)
	}
	slices.Sort(updated.Variables)
	if !slices.Equal(updated.Variables, []string{"first_name", "promo_code"}) {
		t.Fatalf("unexpected updated variables: %+v", updated.Variables)
	}
	if !updated.UpdatedAt.After(updated.CreatedAt) && !updated.UpdatedAt.Equal(updated.CreatedAt) {
		t.Fatalf("expected updated_at to be set, got %+v", updated)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/marketing/templates/acme/launch-template", nil)
	deleteReq = withTestUserContext(t, s, deleteReq)
	deleteRR := httptest.NewRecorder()

	s.handleMarketingTemplatesRouter(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", deleteRR.Code, deleteRR.Body.String())
	}

	listAfterDeleteReq := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/templates/acme", nil)
	listAfterDeleteReq = withTestUserContext(t, s, listAfterDeleteReq)
	listAfterDeleteRR := httptest.NewRecorder()

	s.handleMarketingTemplatesRouter(listAfterDeleteRR, listAfterDeleteReq)

	var afterDelete struct {
		Templates []TemplateInfo `json:"templates"`
	}
	if err := json.Unmarshal(listAfterDeleteRR.Body.Bytes(), &afterDelete); err != nil {
		t.Fatalf("invalid list-after-delete response JSON: %v", err)
	}
	if len(afterDelete.Templates) != 0 {
		t.Fatalf("expected no templates after delete, got %+v", afterDelete.Templates)
	}
}

func TestHandleMarketingMediaCRUDAndCopy(t *testing.T) {
	s, _, workspace := setupMarketingServer(t)
	seedMarketingCampaign(t, workspace, "acme", "launch")
	pngData := newPNGFixture(t)

	uploadReq := newMultipartFileRequest(t, http.MethodPost, "/api/v1/marketing/media/acme", "file", "banner.png", pngData, map[string]string{
		"description": "Hero banner",
		"tags":        "launch,hero",
	})
	uploadReq = withTestUserContext(t, s, uploadReq)
	uploadRR := httptest.NewRecorder()

	s.handleMarketingMediaRouter(uploadRR, uploadReq)

	if uploadRR.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", uploadRR.Code, uploadRR.Body.String())
	}

	var uploaded MediaInfo
	if err := json.Unmarshal(uploadRR.Body.Bytes(), &uploaded); err != nil {
		t.Fatalf("invalid upload response JSON: %v", err)
	}
	if uploaded.Filename != "banner.png" || uploaded.OriginalName != "banner.png" {
		t.Fatalf("unexpected uploaded media: %+v", uploaded)
	}
	if uploaded.Width != 1 || uploaded.Height != 1 {
		t.Fatalf("expected image dimensions 1x1, got %+v", uploaded)
	}
	if !slices.Equal(uploaded.Tags, []string{"launch", "hero"}) {
		t.Fatalf("unexpected tags after upload: %+v", uploaded.Tags)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/media/acme", nil)
	listReq = withTestUserContext(t, s, listReq)
	listRR := httptest.NewRecorder()

	s.handleMarketingMediaRouter(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", listRR.Code, listRR.Body.String())
	}

	var listed struct {
		Media []MediaInfo `json:"media"`
	}
	if err := json.Unmarshal(listRR.Body.Bytes(), &listed); err != nil {
		t.Fatalf("invalid list response JSON: %v", err)
	}
	if len(listed.Media) != 1 {
		t.Fatalf("expected 1 media item, got %d", len(listed.Media))
	}

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/marketing/media/acme/banner.png", bytes.NewBufferString(`{"tags":["updated"],"description":"Updated banner"}`))
	patchReq = withTestUserContext(t, s, patchReq)
	patchRR := httptest.NewRecorder()

	s.handleMarketingMediaRouter(patchRR, patchReq)

	if patchRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", patchRR.Code, patchRR.Body.String())
	}

	copyReq := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/media/acme/copy", bytes.NewBufferString(`{"file":"banner.png","campaign":"launch"}`))
	copyReq = withTestUserContext(t, s, copyReq)
	copyRR := httptest.NewRecorder()

	s.handleMarketingMediaRouter(copyRR, copyReq)

	if copyRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", copyRR.Code, copyRR.Body.String())
	}
	if _, err := os.Stat(filepath.Join(workspace, "marketing", "acme", "launch", "assets", "images", "banner.png")); err != nil {
		t.Fatalf("expected copied asset in campaign assets: %v", err)
	}

	listAfterCopyReq := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/media/acme", nil)
	listAfterCopyReq = withTestUserContext(t, s, listAfterCopyReq)
	listAfterCopyRR := httptest.NewRecorder()

	s.handleMarketingMediaRouter(listAfterCopyRR, listAfterCopyReq)

	var afterCopy struct {
		Media []MediaInfo `json:"media"`
	}
	if err := json.Unmarshal(listAfterCopyRR.Body.Bytes(), &afterCopy); err != nil {
		t.Fatalf("invalid list-after-copy response JSON: %v", err)
	}
	if len(afterCopy.Media) != 1 || !slices.Equal(afterCopy.Media[0].CampaignsUsed, []string{"launch"}) {
		t.Fatalf("expected campaigns_used to include launch, got %+v", afterCopy.Media)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/marketing/media/acme/banner.png", nil)
	deleteReq = withTestUserContext(t, s, deleteReq)
	deleteRR := httptest.NewRecorder()

	s.handleMarketingMediaRouter(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", deleteRR.Code, deleteRR.Body.String())
	}
	if _, err := os.Stat(filepath.Join(workspace, "marketing", "acme", "_media", "banner.png")); !os.IsNotExist(err) {
		t.Fatalf("expected uploaded media file to be removed, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(workspace, "marketing", "acme", "_media", "banner.png.meta.json")); !os.IsNotExist(err) {
		t.Fatalf("expected uploaded media sidecar to be removed, err=%v", err)
	}
}

func TestHandleMarketingAnalyticsSummary(t *testing.T) {
	s, _, workspace := setupMarketingServer(t)
	basePath := seedMarketingCampaign(t, workspace, "acme", "launch")
	analyticsDir := filepath.Join(basePath, "analytics")

	files := map[string]string{
		filepath.Join(analyticsDir, "twitter_post-1_2026-04-01.json"):  `{"impressions":100,"engagements":25}`,
		filepath.Join(analyticsDir, "twitter_post-2_2026-04-01.json"):  `{"impressions":50,"likes":10,"comments":5}`,
		filepath.Join(analyticsDir, "linkedin_post-3_2026-04-02.json"): `{"impressions":200,"engagements":20}`,
		filepath.Join(analyticsDir, "legacy-report.txt"):               "legacy analytics report",
		filepath.Join(analyticsDir, "bad-name.json"):                   `{"impressions":999,"engagements":999}`,
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to seed analytics file %s: %v", path, err)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/campaigns/acme/launch/analytics/summary", nil)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()

	s.handleMarketingCampaignsRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var out AnalyticsSummary
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if out.TotalPosts != 3 || out.TotalImpressions != 350 || out.TotalEngagements != 60 {
		t.Fatalf("unexpected summary totals: %+v", out)
	}
	if math.Abs(out.EngagementRate-(60.0/350.0)) > 0.000001 {
		t.Fatalf("unexpected engagement rate: %+v", out)
	}
	if len(out.ByPlatform) != 2 {
		t.Fatalf("expected 2 platforms, got %+v", out.ByPlatform)
	}
	if got := out.ByPlatform["twitter"]; got.Posts != 2 || got.Impressions != 150 || got.Engagements != 40 {
		t.Fatalf("unexpected twitter stats: %+v", got)
	}
	if got := out.ByPlatform["linkedin"]; got.Posts != 1 || got.Impressions != 200 || got.Engagements != 20 {
		t.Fatalf("unexpected linkedin stats: %+v", got)
	}
	if !slices.Equal(out.TrendDates, []string{"2026-04-01", "2026-04-02"}) {
		t.Fatalf("unexpected trend dates: %+v", out.TrendDates)
	}
	if !slices.Equal(out.TrendImpressions, []int64{150, 200}) {
		t.Fatalf("unexpected trend impressions: %+v", out.TrendImpressions)
	}
	if !out.HasLegacyReports {
		t.Fatalf("expected legacy reports flag to be true")
	}
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
