package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/tools"
)

// CampaignInfo holds summary metadata for a campaign directory.
type CampaignInfo struct {
	Account      string    `json:"account"`
	Campaign     string    `json:"campaign"`
	BasePath     string    `json:"base_path"`
	Status       string    `json:"status"`
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
	Brief    string                 `json:"brief,omitempty"`
	Strategy string                 `json:"strategy,omitempty"`
	Files    map[string][]FileEntry `json:"files"`
}

// FileEntry describes a single file within a campaign folder.
type FileEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsImage bool   `json:"is_image"`
}

type CreateCampaignRequest struct {
	Account     string   `json:"account"`
	Campaign    string   `json:"campaign"`
	Objective   string   `json:"objective"`
	Platforms   []string `json:"platforms"`
	Description string   `json:"description"`
}

type renameCampaignRequest struct {
	NewName string `json:"new_name"`
}

type updateCampaignStatusRequest struct {
	Status string `json:"status"`
}

type MediaMeta struct {
	OriginalName  string    `json:"original_name"`
	Tags          []string  `json:"tags"`
	Description   string    `json:"description"`
	UploadedAt    time.Time `json:"uploaded_at"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	CampaignsUsed []string  `json:"campaigns_used"`
}

type MediaInfo struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	MediaMeta
}

type updateMediaRequest struct {
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

type copyMediaRequest struct {
	File     string `json:"file"`
	Campaign string `json:"campaign"`
}

var allowedCampaignStatuses = map[string]struct{}{
	"draft":     {},
	"active":    {},
	"paused":    {},
	"completed": {},
	"archived":  {},
}

var imageExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".svg": true, ".bmp": true, ".tiff": true,
}

func isImageFile(name string) bool {
	return imageExtensions[strings.ToLower(filepath.Ext(name))]
}

// handleListMarketingCampaigns handles GET/POST /api/v1/marketing/campaigns
func (s *Server) handleListMarketingCampaigns(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// continue below
	case http.MethodPost:
		s.handleCreateMarketingCampaign(w, r)
		return
	default:
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
			if strings.HasPrefix(strings.ToLower(campaignName), "_media") {
				continue
			}
			campaignPath := filepath.Join(accountPath, campaignName)

			info, err := campaignEntry.Info()
			if err != nil {
				continue
			}

			ci := CampaignInfo{
				Account:   accountName,
				Campaign:  campaignName,
				BasePath:  mktToRelative(userWorkspace, campaignPath),
				Status:    readCampaignStatus(campaignPath).Status,
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
// and nested files/status endpoints.
func (s *Server) handleMarketingCampaignsRouter(w http.ResponseWriter, r *http.Request) {
	suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/marketing/campaigns/")
	// Count path segments after campaigns/
	parts := strings.SplitN(suffix, "/", 4)
	// parts[0] = account, parts[1] = campaign, parts[2] = "files" (optional), parts[3] = file path
	if len(parts) >= 3 && parts[2] == "files" {
		s.handleGetMarketingFile(w, r)
		return
	}
	if len(parts) >= 3 && parts[2] == "status" {
		s.handleUpdateCampaignStatus(w, r)
		return
	}
	if len(parts) >= 4 && parts[2] == "analytics" && parts[3] == "summary" {
		s.handleMarketingAnalyticsSummary(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetMarketingCampaign(w, r)
	case http.MethodPatch:
		s.handleRenameMarketingCampaign(w, r)
	case http.MethodDelete:
		s.handleDeleteMarketingCampaign(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleMarketingMediaRouter dispatches /api/v1/marketing/media/{account}[/{filename}|/copy]
func (s *Server) handleMarketingMediaRouter(w http.ResponseWriter, r *http.Request) {
	suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/marketing/media/")
	parts := strings.SplitN(suffix, "/", 3)
	if len(parts) < 1 || strings.TrimSpace(parts[0]) == "" {
		http.Error(w, "invalid media path", http.StatusBadRequest)
		return
	}
	if len(parts) >= 2 && parts[1] == "copy" {
		s.handleCopyMarketingMedia(w, r)
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.handleListMarketingMedia(w, r)
		case http.MethodPost:
			s.handleUploadMarketingMedia(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	switch r.Method {
	case http.MethodPatch:
		s.handleUpdateMarketingMedia(w, r)
	case http.MethodDelete:
		s.handleDeleteMarketingMedia(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleListMarketingMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userWorkspace, accountName, ok := s.marketingAccountWorkspace(w, r, "/api/v1/marketing/media/")
	if !ok {
		return
	}

	mediaDir := filepath.Join(userWorkspace, "marketing", accountName, "_media")
	if _, err := os.Stat(mediaDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]interface{}{"media": []MediaInfo{}})
		return
	}

	entries, err := os.ReadDir(mediaDir)
	if err != nil {
		http.Error(w, "failed to read media directory", http.StatusInternalServerError)
		return
	}

	media := make([]MediaInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(strings.ToLower(entry.Name()), ".meta.json") {
			continue
		}
		info, err := buildMediaInfo(userWorkspace, accountName, mediaDir, entry.Name())
		if err != nil {
			continue
		}
		media = append(media, info)
	}
	sort.Slice(media, func(i, j int) bool { return media[i].Filename < media[j].Filename })

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"media": media})
}

func (s *Server) handleUploadMarketingMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userWorkspace, accountName, ok := s.marketingAccountWorkspace(w, r, "/api/v1/marketing/media/")
	if !ok {
		return
	}

	const maxUploadSize = 25 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large or invalid form (max 25MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing 'file' field in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Base(strings.TrimSpace(header.Filename))
	if filename == "." || filename == "" || isUnsafeMarketingFilename(filename) {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
	if err != nil {
		http.Error(w, "failed to read uploaded file", http.StatusInternalServerError)
		return
	}

	mediaDir := filepath.Join(userWorkspace, "marketing", accountName, "_media")
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		http.Error(w, "failed to create media directory", http.StatusInternalServerError)
		return
	}

	targetPath := filepath.Join(mediaDir, filename)
	if _, err := os.Stat(targetPath); err == nil {
		http.Error(w, "media file already exists", http.StatusConflict)
		return
	} else if err != nil && !os.IsNotExist(err) {
		http.Error(w, "failed to access media file", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		http.Error(w, "failed to store media file", http.StatusInternalServerError)
		return
	}

	width, height := imageDimensions(data)
	meta := MediaMeta{
		OriginalName:  filename,
		Tags:          splitAndCleanCSV(r.FormValue("tags")),
		Description:   strings.TrimSpace(r.FormValue("description")),
		UploadedAt:    time.Now().UTC(),
		Width:         width,
		Height:        height,
		CampaignsUsed: []string{},
	}
	if err := writeMarketingJSON(mediaMetaPath(mediaDir, filename), meta); err != nil {
		_ = os.Remove(targetPath)
		http.Error(w, "failed to store media metadata", http.StatusInternalServerError)
		return
	}

	stored, err := buildMediaInfo(userWorkspace, accountName, mediaDir, filename)
	if err != nil {
		http.Error(w, "failed to load uploaded media", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, stored)
}

func (s *Server) handleUpdateMarketingMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userWorkspace, accountName, filename, ok := s.marketingMediaItemWorkspace(w, r)
	if !ok {
		return
	}

	mediaDir := filepath.Join(userWorkspace, "marketing", accountName, "_media")
	mediaPath := filepath.Join(mediaDir, filename)
	if _, err := os.Stat(mediaPath); os.IsNotExist(err) {
		http.Error(w, "media file not found", http.StatusNotFound)
		return
	}

	var req updateMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	meta, err := readMediaMeta(mediaDir, filename)
	if err != nil {
		http.Error(w, "failed to load media metadata", http.StatusInternalServerError)
		return
	}
	meta.Tags = sanitizeStringList(req.Tags)
	meta.Description = strings.TrimSpace(req.Description)
	if err := writeMarketingJSON(mediaMetaPath(mediaDir, filename), meta); err != nil {
		http.Error(w, "failed to update media metadata", http.StatusInternalServerError)
		return
	}

	updated, err := buildMediaInfo(userWorkspace, accountName, mediaDir, filename)
	if err != nil {
		http.Error(w, "failed to load updated media", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updated)
}

func (s *Server) handleDeleteMarketingMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userWorkspace, accountName, filename, ok := s.marketingMediaItemWorkspace(w, r)
	if !ok {
		return
	}

	mediaDir := filepath.Join(userWorkspace, "marketing", accountName, "_media")
	mediaPath := filepath.Join(mediaDir, filename)
	if _, err := os.Stat(mediaPath); os.IsNotExist(err) {
		http.Error(w, "media file not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(mediaPath); err != nil {
		http.Error(w, "failed to delete media file", http.StatusInternalServerError)
		return
	}
	if err := os.Remove(mediaMetaPath(mediaDir, filename)); err != nil && !os.IsNotExist(err) {
		http.Error(w, "failed to delete media metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCopyMarketingMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userWorkspace, accountName, ok := s.marketingAccountWorkspace(w, r, "/api/v1/marketing/media/")
	if !ok {
		return
	}

	var req copyMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	filename := strings.TrimSpace(req.File)
	campaignName := strings.TrimSpace(req.Campaign)
	if filename == "" || campaignName == "" {
		http.Error(w, "file and campaign are required", http.StatusBadRequest)
		return
	}
	if isUnsafeMarketingFilename(filename) || isUnsafeMarketingSegment(campaignName) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	mediaDir := filepath.Join(userWorkspace, "marketing", accountName, "_media")
	mediaPath := filepath.Join(mediaDir, filename)
	if _, err := os.Stat(mediaPath); os.IsNotExist(err) {
		http.Error(w, "media file not found", http.StatusNotFound)
		return
	}

	campaignPath := filepath.Join(userWorkspace, "marketing", accountName, campaignName)
	if _, err := os.Stat(campaignPath); os.IsNotExist(err) {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	}

	assetDir := filepath.Join(campaignPath, "assets")
	if isImageFile(filename) {
		assetDir = filepath.Join(assetDir, "images")
	}
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		http.Error(w, "failed to prepare asset directory", http.StatusInternalServerError)
		return
	}

	data, err := os.ReadFile(mediaPath)
	if err != nil {
		http.Error(w, "failed to read media file", http.StatusInternalServerError)
		return
	}
	if err := os.WriteFile(filepath.Join(assetDir, filename), data, 0644); err != nil {
		http.Error(w, "failed to copy media file", http.StatusInternalServerError)
		return
	}

	meta, err := readMediaMeta(mediaDir, filename)
	if err != nil {
		http.Error(w, "failed to load media metadata", http.StatusInternalServerError)
		return
	}
	meta.CampaignsUsed = appendUniqueString(meta.CampaignsUsed, campaignName)
	if err := writeMarketingJSON(mediaMetaPath(mediaDir, filename), meta); err != nil {
		http.Error(w, "failed to update media metadata", http.StatusInternalServerError)
		return
	}

	updated, err := buildMediaInfo(userWorkspace, accountName, mediaDir, filename)
	if err != nil {
		http.Error(w, "failed to load copied media", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updated)
}

// handleCreateMarketingCampaign handles POST /api/v1/marketing/campaigns
func (s *Server) handleCreateMarketingCampaign(w http.ResponseWriter, r *http.Request) {
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

	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Account) == "" || strings.TrimSpace(req.Campaign) == "" {
		http.Error(w, "account and campaign are required", http.StatusBadRequest)
		return
	}
	if isUnsafeMarketingSegment(req.Account) || isUnsafeMarketingSegment(req.Campaign) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	accountSlug := tools.SanitizeMarketingSlug(req.Account)
	campaignSlug := tools.SanitizeMarketingSlug(req.Campaign)
	if accountSlug == "" || campaignSlug == "" {
		http.Error(w, "account and campaign are required", http.StatusBadRequest)
		return
	}

	campaignPath := filepath.Join(userWorkspace, "marketing", accountSlug, campaignSlug)
	if _, err := os.Stat(campaignPath); err == nil {
		http.Error(w, "campaign already exists", http.StatusConflict)
		return
	} else if err != nil && !os.IsNotExist(err) {
		http.Error(w, "failed to access campaign", http.StatusInternalServerError)
		return
	}

	if _, err := tools.CreateCampaignDir(
		userWorkspace,
		req.Account,
		req.Campaign,
		req.Description,
		req.Objective,
		strings.Join(req.Platforms, ","),
	); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdInfo, err := buildCampaignInfo(userWorkspace, accountSlug, campaignSlug, campaignPath)
	if err != nil {
		http.Error(w, "failed to load created campaign", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, createdInfo)
}

// handleGetMarketingCampaign handles GET /api/v1/marketing/campaigns/{account}/{campaign}
func (s *Server) handleGetMarketingCampaign(w http.ResponseWriter, r *http.Request) {
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
	if isUnsafeMarketingSegment(accountName) || isUnsafeMarketingSegment(campaignName) {
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

	detail, err := buildCampaignDetail(userWorkspace, accountName, campaignName, campaignPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, detail)
}

// handleRenameMarketingCampaign handles PATCH /api/v1/marketing/campaigns/{account}/{campaign}
func (s *Server) handleRenameMarketingCampaign(w http.ResponseWriter, r *http.Request) {
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

	accountName, campaignName, err := parseMarketingCampaignPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if isUnsafeMarketingSegment(accountName) || isUnsafeMarketingSegment(campaignName) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	var req renameCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.NewName) == "" {
		http.Error(w, "new_name is required", http.StatusBadRequest)
		return
	}
	if isUnsafeMarketingSegment(req.NewName) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	newCampaignName := tools.SanitizeMarketingSlug(req.NewName)
	if newCampaignName == "" {
		http.Error(w, "new_name is required", http.StatusBadRequest)
		return
	}

	oldPath := filepath.Join(userWorkspace, "marketing", accountName, campaignName)
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "failed to access campaign", http.StatusInternalServerError)
		return
	}

	if newCampaignName != campaignName {
		newPath := filepath.Join(userWorkspace, "marketing", accountName, newCampaignName)
		if _, err := os.Stat(newPath); err == nil {
			http.Error(w, "campaign already exists", http.StatusConflict)
			return
		} else if err != nil && !os.IsNotExist(err) {
			http.Error(w, "failed to access campaign", http.StatusInternalServerError)
			return
		}

		if err := os.Rename(oldPath, newPath); err != nil {
			http.Error(w, "failed to rename campaign", http.StatusInternalServerError)
			return
		}
		oldPath = newPath
	}

	if _, err := ensureCampaignStatus(oldPath); err != nil {
		http.Error(w, "failed to update campaign status", http.StatusInternalServerError)
		return
	}

	updatedInfo, err := buildCampaignInfo(userWorkspace, accountName, newCampaignName, oldPath)
	if err != nil {
		http.Error(w, "failed to load renamed campaign", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updatedInfo)
}

// handleDeleteMarketingCampaign handles DELETE /api/v1/marketing/campaigns/{account}/{campaign}
func (s *Server) handleDeleteMarketingCampaign(w http.ResponseWriter, r *http.Request) {
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

	accountName, campaignName, err := parseMarketingCampaignPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if isUnsafeMarketingSegment(accountName) || isUnsafeMarketingSegment(campaignName) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	campaignPath := filepath.Join(userWorkspace, "marketing", accountName, campaignName)
	if _, err := os.Stat(campaignPath); os.IsNotExist(err) {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "failed to access campaign", http.StatusInternalServerError)
		return
	}

	if err := os.RemoveAll(campaignPath); err != nil {
		http.Error(w, "failed to delete campaign", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleUpdateCampaignStatus handles PATCH /api/v1/marketing/campaigns/{account}/{campaign}/status
func (s *Server) handleUpdateCampaignStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
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

	accountName, campaignName, err := parseMarketingCampaignStatusPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if isUnsafeMarketingSegment(accountName) || isUnsafeMarketingSegment(campaignName) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	var req updateCampaignStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	status := strings.ToLower(strings.TrimSpace(req.Status))
	if _, ok := allowedCampaignStatuses[status]; !ok {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	campaignPath := filepath.Join(userWorkspace, "marketing", accountName, campaignName)
	if _, err := os.Stat(campaignPath); os.IsNotExist(err) {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "failed to access campaign", http.StatusInternalServerError)
		return
	}

	updatedStatus, err := writeCampaignStatus(campaignPath, status)
	if err != nil {
		http.Error(w, "failed to update campaign status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updatedStatus)
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

	if isUnsafeMarketingSegment(accountName) || isUnsafeMarketingSegment(campaignName) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	// afterCampaign should start with "files/"
	filePath := strings.TrimPrefix(afterCampaign, "files/")
	if filePath == afterCampaign {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	if unescaped, err := url.PathUnescape(filePath); err == nil {
		filePath = unescaped
	}
	if isUnsafeMarketingFilePath(filePath) {
		http.Error(w, "access denied", http.StatusForbidden)
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

func isUnsafeMarketingSegment(value string) bool {
	if unescaped, err := url.PathUnescape(value); err == nil {
		value = unescaped
	}
	return strings.Contains(value, "..") || strings.ContainsAny(value, "/\\")
}

func isUnsafeMarketingFilePath(value string) bool {
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "\\") {
		return true
	}
	value = strings.ReplaceAll(value, "\\", "/")
	for _, part := range strings.Split(value, "/") {
		if part == ".." {
			return true
		}
	}
	return false
}

func isUnsafeMarketingFilename(value string) bool {
	if unescaped, err := url.PathUnescape(value); err == nil {
		value = unescaped
	}
	value = strings.TrimSpace(value)
	return value == "" || strings.Contains(value, "..") || strings.ContainsAny(value, "/\\")
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

func parseMarketingCampaignPath(path string) (string, string, error) {
	suffix := strings.TrimPrefix(path, "/api/v1/marketing/campaigns/")
	parts := strings.SplitN(suffix, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid campaign path")
	}
	return parts[0], parts[1], nil
}

func parseMarketingCampaignStatusPath(path string) (string, string, error) {
	suffix := strings.TrimPrefix(path, "/api/v1/marketing/campaigns/")
	parts := strings.SplitN(suffix, "/", 3)
	if len(parts) != 3 || parts[2] != "status" || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid campaign path")
	}
	return parts[0], parts[1], nil
}

func parseMarketingAccountPath(path, prefix string) (string, error) {
	suffix := strings.TrimPrefix(path, prefix)
	parts := strings.SplitN(suffix, "/", 2)
	if len(parts) < 1 || strings.TrimSpace(parts[0]) == "" {
		return "", fmt.Errorf("invalid account path")
	}
	return parts[0], nil
}

func parseMarketingAccountItemPath(path, prefix string) (string, string, error) {
	suffix := strings.TrimPrefix(path, prefix)
	parts := strings.Split(suffix, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("invalid path")
	}
	return parts[0], parts[1], nil
}

func mktToRelative(workspace, absPath string) string {
	rel, err := filepath.Rel(workspace, absPath)
	if err != nil {
		return absPath
	}
	return filepath.ToSlash(rel)
}

func (s *Server) marketingAccountWorkspace(w http.ResponseWriter, r *http.Request, prefix string) (string, string, bool) {
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", "", false
	}

	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return "", "", false
	}

	accountName, err := parseMarketingAccountPath(r.URL.Path, prefix)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", "", false
	}
	if isUnsafeMarketingSegment(accountName) {
		http.Error(w, "access denied", http.StatusForbidden)
		return "", "", false
	}

	return userWorkspace, accountName, true
}

func (s *Server) marketingMediaItemWorkspace(w http.ResponseWriter, r *http.Request) (string, string, string, bool) {
	userWorkspace, accountName, ok := s.marketingAccountWorkspace(w, r, "/api/v1/marketing/media/")
	if !ok {
		return "", "", "", false
	}
	_, filename, err := parseMarketingAccountItemPath(r.URL.Path, "/api/v1/marketing/media/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", "", "", false
	}
	if isUnsafeMarketingFilename(filename) || filename == "copy" {
		http.Error(w, "access denied", http.StatusForbidden)
		return "", "", "", false
	}
	return userWorkspace, accountName, filename, true
}

func mediaMetaPath(mediaDir, filename string) string {
	return filepath.Join(mediaDir, filename+".meta.json")
}

func readMediaMeta(mediaDir, filename string) (MediaMeta, error) {
	data, err := os.ReadFile(mediaMetaPath(mediaDir, filename))
	if err != nil {
		if os.IsNotExist(err) {
			return MediaMeta{OriginalName: filename, Tags: []string{}, CampaignsUsed: []string{}}, nil
		}
		return MediaMeta{}, err
	}
	var meta MediaMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return MediaMeta{}, err
	}
	meta.Tags = sanitizeStringList(meta.Tags)
	meta.CampaignsUsed = sanitizeStringList(meta.CampaignsUsed)
	if meta.OriginalName == "" {
		meta.OriginalName = filename
	}
	return meta, nil
}

func buildMediaInfo(userWorkspace, accountName, mediaDir, filename string) (MediaInfo, error) {
	info, err := os.Stat(filepath.Join(mediaDir, filename))
	if err != nil {
		return MediaInfo{}, err
	}
	meta, err := readMediaMeta(mediaDir, filename)
	if err != nil {
		return MediaInfo{}, err
	}
	return MediaInfo{
		Filename:  filename,
		Path:      mktToRelative(userWorkspace, filepath.Join(userWorkspace, "marketing", accountName, "_media", filename)),
		Size:      info.Size(),
		MediaMeta: meta,
	}, nil
}

func sanitizeStringList(values []string) []string {
	if values == nil {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	clean := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		clean = append(clean, value)
	}
	return clean
}

func splitAndCleanCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	return sanitizeStringList(strings.Split(value, ","))
}

func appendUniqueString(values []string, value string) []string {
	values = append(values, value)
	return sanitizeStringList(values)
}

func imageDimensions(data []byte) (int, int) {
	reader := bytes.NewReader(data)
	cfg, _, err := image.DecodeConfig(reader)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

func writeMarketingJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func readCampaignStatus(campaignPath string) tools.CampaignStatus {
	status, err := ensureCampaignStatus(campaignPath)
	if err != nil {
		return tools.CampaignStatus{Status: "draft"}
	}
	return status
}

func ensureCampaignStatus(campaignPath string) (tools.CampaignStatus, error) {
	statusPath := filepath.Join(campaignPath, "status.json")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		if os.IsNotExist(err) {
			status := tools.CampaignStatus{Status: "draft"}
			return writeCampaignStatus(campaignPath, status.Status)
		}
		return tools.CampaignStatus{}, err
	}

	var status tools.CampaignStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return writeCampaignStatus(campaignPath, "draft")
	}
	status.Status = strings.ToLower(strings.TrimSpace(status.Status))
	if status.Status == "" {
		return writeCampaignStatus(campaignPath, "draft")
	}
	return status, nil
}

func writeCampaignStatus(campaignPath, status string) (tools.CampaignStatus, error) {
	payload := tools.CampaignStatus{
		Status:    status,
		UpdatedAt: time.Now().UTC(),
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return tools.CampaignStatus{}, err
	}
	if err := os.WriteFile(filepath.Join(campaignPath, "status.json"), data, 0644); err != nil {
		return tools.CampaignStatus{}, err
	}
	return payload, nil
}

func buildCampaignInfo(userWorkspace, accountName, campaignName, campaignPath string) (CampaignInfo, error) {
	info, err := os.Stat(campaignPath)
	if err != nil {
		return CampaignInfo{}, err
	}

	ci := CampaignInfo{
		Account:   accountName,
		Campaign:  campaignName,
		BasePath:  mktToRelative(userWorkspace, campaignPath),
		Status:    readCampaignStatus(campaignPath).Status,
		CreatedAt: info.ModTime(),
	}
	ci.HasStrategy = mktFileExists(filepath.Join(campaignPath, "strategy.md"))
	ci.HasCopy = mktDirExists(filepath.Join(campaignPath, "copy"))
	ci.HasAssets = mktDirExists(filepath.Join(campaignPath, "assets", "images")) ||
		mktDirExists(filepath.Join(campaignPath, "assets"))
	ci.HasSchedule = mktDirExists(filepath.Join(campaignPath, "schedules")) ||
		mktFileExists(filepath.Join(campaignPath, "schedule.json"))
	ci.HasAnalytics = mktDirExists(filepath.Join(campaignPath, "analytics"))
	return ci, nil
}

func buildCampaignDetail(userWorkspace, accountName, campaignName, campaignPath string) (CampaignDetail, error) {
	info, err := buildCampaignInfo(userWorkspace, accountName, campaignName, campaignPath)
	if err != nil {
		return CampaignDetail{}, err
	}

	detail := CampaignDetail{
		CampaignInfo: info,
		Files:        make(map[string][]FileEntry),
	}

	if brief, err := os.ReadFile(filepath.Join(campaignPath, "brief.md")); err == nil {
		detail.Brief = string(brief)
	}
	if strategy, err := os.ReadFile(filepath.Join(campaignPath, "strategy.md")); err == nil {
		detail.Strategy = string(strategy)
	}

	detail.Files["copy"] = listFiles(filepath.Join(campaignPath, "copy"), "copy")
	assetsImagesDir := filepath.Join(campaignPath, "assets", "images")
	if mktDirExists(assetsImagesDir) {
		detail.Files["assets"] = listFiles(assetsImagesDir, "assets/images")
	} else {
		detail.Files["assets"] = listFiles(filepath.Join(campaignPath, "assets"), "assets")
	}
	detail.Files["schedules"] = listFiles(filepath.Join(campaignPath, "schedules"), "schedules")
	detail.Files["analytics"] = listFiles(filepath.Join(campaignPath, "analytics"), "analytics")

	return detail, nil
}
