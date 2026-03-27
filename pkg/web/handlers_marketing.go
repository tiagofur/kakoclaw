package web

import (
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
)

// CampaignInfo holds summary metadata for a campaign directory.
type CampaignInfo struct {
	Account      string    `json:"account"`
	Campaign     string    `json:"campaign"`
	BasePath     string    `json:"base_path"`
	CreatedAt    time.Time `json:"created_at"`
	HasStrategy  bool      `json:"has_strategy"`
	HasCopy      bool      `json:"has_copy"`
	HasAssets    bool      `json:"has_assets"`
	HasSchedule  bool      `json:"has_schedule"`
	HasAnalytics bool      `json:"has_analytics"`
}

// CampaignDetail extends CampaignInfo with file content previews.
type CampaignDetail struct {
	CampaignInfo
	Brief    string                `json:"brief,omitempty"`
	Strategy string                `json:"strategy,omitempty"`
	Files    map[string][]FileEntry `json:"files"`
}

// FileEntry describes a single file within a campaign folder.
type FileEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsImage bool   `json:"is_image"`
}

var imageExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".svg": true, ".bmp": true, ".tiff": true,
}

func isImageFile(name string) bool {
	return imageExtensions[strings.ToLower(filepath.Ext(name))]
}

// handleListMarketingCampaigns handles GET /api/v1/marketing/campaigns
func (s *Server) handleListMarketingCampaigns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		logger.ErrorCF("web", "failed to get workspace for marketing list", map[string]interface{}{"error": err.Error()})
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	marketingDir := filepath.Join(userWorkspace, "marketing")

	// If marketing dir doesn't exist return empty list — not an error
	if _, err := os.Stat(marketingDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]interface{}{"campaigns": []CampaignInfo{}})
		return
	}

	var campaigns []CampaignInfo

	// Walk account dirs
	accountEntries, err := os.ReadDir(marketingDir)
	if err != nil {
		logger.ErrorCF("web", "failed to read marketing dir", map[string]interface{}{"error": err.Error()})
		http.Error(w, "failed to read marketing directory", http.StatusInternalServerError)
		return
	}

	for _, accountEntry := range accountEntries {
		if !accountEntry.IsDir() {
			continue
		}
		accountName := accountEntry.Name()
		accountPath := filepath.Join(marketingDir, accountName)

		campaignEntries, err := os.ReadDir(accountPath)
		if err != nil {
			continue
		}

		for _, campaignEntry := range campaignEntries {
			if !campaignEntry.IsDir() {
				continue
			}
			campaignName := campaignEntry.Name()
			campaignPath := filepath.Join(accountPath, campaignName)

			info, err := campaignEntry.Info()
			if err != nil {
				continue
			}

			ci := CampaignInfo{
				Account:   accountName,
				Campaign:  campaignName,
				BasePath:  campaignPath,
				CreatedAt: info.ModTime(),
			}

			// Check which sub-sections exist
			ci.HasStrategy = mktFileExists(filepath.Join(campaignPath, "strategy.md"))
			ci.HasCopy = mktDirExists(filepath.Join(campaignPath, "copy"))
			ci.HasAssets = mktDirExists(filepath.Join(campaignPath, "assets", "images")) ||
				mktDirExists(filepath.Join(campaignPath, "assets"))
			ci.HasSchedule = mktDirExists(filepath.Join(campaignPath, "schedules")) ||
				mktFileExists(filepath.Join(campaignPath, "schedule.json"))
			ci.HasAnalytics = mktDirExists(filepath.Join(campaignPath, "analytics"))

			campaigns = append(campaigns, ci)
		}
	}

	if campaigns == nil {
		campaigns = []CampaignInfo{}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"campaigns": campaigns})
}

// handleMarketingCampaignsRouter dispatches /api/v1/marketing/campaigns/{account}/{campaign}
// and /api/v1/marketing/campaigns/{account}/{campaign}/files/{path...}
func (s *Server) handleMarketingCampaignsRouter(w http.ResponseWriter, r *http.Request) {
	suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/marketing/campaigns/")
	// Count path segments after campaigns/
	parts := strings.SplitN(suffix, "/", 4)
	// parts[0] = account, parts[1] = campaign, parts[2] = "files" (optional), parts[3] = file path
	if len(parts) >= 3 && parts[2] == "files" {
		s.handleGetMarketingFile(w, r)
		return
	}
	s.handleGetMarketingCampaign(w, r)
}

// handleGetMarketingCampaign handles GET /api/v1/marketing/campaigns/{account}/{campaign}
func (s *Server) handleGetMarketingCampaign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	// Parse account and campaign from URL: /api/v1/marketing/campaigns/{account}/{campaign}
	suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/marketing/campaigns/")
	parts := strings.SplitN(suffix, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.Error(w, "invalid campaign path", http.StatusBadRequest)
		return
	}
	accountName := parts[0]
	campaignName := parts[1]

	// Security: neither part should traverse directories
	if strings.Contains(accountName, "..") || strings.Contains(campaignName, "..") ||
		strings.ContainsAny(accountName, "/\\") || strings.ContainsAny(campaignName, "/\\") {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	campaignPath := filepath.Join(userWorkspace, "marketing", accountName, campaignName)
	absWorkspace, _ := filepath.Abs(filepath.Join(userWorkspace, "marketing"))
	absCampaign, _ := filepath.Abs(campaignPath)
	if !strings.HasPrefix(absCampaign, absWorkspace) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	if _, err := os.Stat(campaignPath); os.IsNotExist(err) {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	}

	info, err := os.Stat(campaignPath)
	if err != nil {
		http.Error(w, "failed to stat campaign", http.StatusInternalServerError)
		return
	}

	detail := CampaignDetail{
		CampaignInfo: CampaignInfo{
			Account:   accountName,
			Campaign:  campaignName,
			BasePath:  campaignPath,
			CreatedAt: info.ModTime(),
		},
		Files: make(map[string][]FileEntry),
	}

	// Check presence flags
	detail.HasStrategy = mktFileExists(filepath.Join(campaignPath, "strategy.md"))
	detail.HasCopy = mktDirExists(filepath.Join(campaignPath, "copy"))
	detail.HasAssets = mktDirExists(filepath.Join(campaignPath, "assets", "images")) ||
		mktDirExists(filepath.Join(campaignPath, "assets"))
	detail.HasSchedule = mktDirExists(filepath.Join(campaignPath, "schedules")) ||
		mktFileExists(filepath.Join(campaignPath, "schedule.json"))
	detail.HasAnalytics = mktDirExists(filepath.Join(campaignPath, "analytics"))

	// Read brief.md
	if brief, err := os.ReadFile(filepath.Join(campaignPath, "brief.md")); err == nil {
		detail.Brief = string(brief)
	}

	// Read strategy.md
	if strategy, err := os.ReadFile(filepath.Join(campaignPath, "strategy.md")); err == nil {
		detail.Strategy = string(strategy)
	}

	// List copy/ files
	detail.Files["copy"] = listFiles(filepath.Join(campaignPath, "copy"), "copy")

	// List assets/images/ (or assets/)
	assetsImagesDir := filepath.Join(campaignPath, "assets", "images")
	if mktDirExists(assetsImagesDir) {
		detail.Files["assets"] = listFiles(assetsImagesDir, "assets/images")
	} else {
		detail.Files["assets"] = listFiles(filepath.Join(campaignPath, "assets"), "assets")
	}

	// List schedules/
	detail.Files["schedules"] = listFiles(filepath.Join(campaignPath, "schedules"), "schedules")

	// List analytics/
	detail.Files["analytics"] = listFiles(filepath.Join(campaignPath, "analytics"), "analytics")

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, detail)
}

// handleGetMarketingFile handles GET /api/v1/marketing/campaigns/{account}/{campaign}/files/{path...}
func (s *Server) handleGetMarketingFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	// URL pattern: /api/v1/marketing/campaigns/{account}/{campaign}/files/{path...}
	suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/marketing/campaigns/")
	// suffix = "{account}/{campaign}/files/{path...}"
	slashIdx := strings.Index(suffix, "/")
	if slashIdx < 0 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	accountName := suffix[:slashIdx]
	rest := suffix[slashIdx+1:]

	slashIdx2 := strings.Index(rest, "/")
	if slashIdx2 < 0 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	campaignName := rest[:slashIdx2]
	afterCampaign := rest[slashIdx2+1:]

	// afterCampaign should start with "files/"
	filePath := strings.TrimPrefix(afterCampaign, "files/")
	if filePath == afterCampaign {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	// Security: verify resolved path is within workspace/marketing/
	marketingBase := filepath.Join(userWorkspace, "marketing")
	targetPath := filepath.Clean(filepath.Join(marketingBase, accountName, campaignName, filePath))
	absBase, _ := filepath.Abs(marketingBase)
	absTarget, _ := filepath.Abs(targetPath)
	if !strings.HasPrefix(absTarget, absBase) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	finfo, err := os.Stat(targetPath)
	if err != nil || finfo.IsDir() {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	ext := strings.ToLower(filepath.Ext(targetPath))
	if imageExtensions[ext] {
		// Serve image with correct content type — no download header
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", mimeType)
		http.ServeFile(w, r, targetPath)
		return
	}

	// Text / JSON files
	data, err := os.ReadFile(targetPath)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	if ext == ".json" {
		// Validate and re-serve as JSON
		var raw json.RawMessage
		if json.Unmarshal(data, &raw) == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(data)
}

// ---- helpers ----

func mktFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func mktDirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// listFiles returns a slice of FileEntry for all non-directory files in dir.
// pathPrefix is prepended to the Path field (relative to campaign root).
func listFiles(dir string, pathPrefix string) []FileEntry {
	var entries []FileEntry
	if !mktDirExists(dir) {
		return entries
	}
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return entries
	}
	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		relPath := pathPrefix + "/" + de.Name()
		entries = append(entries, FileEntry{
			Name:    de.Name(),
			Path:    relPath,
			Size:    info.Size(),
			IsImage: isImageFile(de.Name()),
		})
	}
	return entries
}
