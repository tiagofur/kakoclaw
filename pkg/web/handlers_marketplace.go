package web

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/skills"
	"github.com/sipeed/makoclaw/pkg/storage"
)

// PublicSkillResponse is the public-facing representation of a marketplace skill
// (without full content). Used by handleMarketplaceSkills and reused by future features.
type PublicSkillResponse struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Slug          string   `json:"slug"`
	Version       string   `json:"version"`
	Description   string   `json:"description"`
	Author        string   `json:"author"`
	Tags          []string `json:"tags"`
	Category      string   `json:"category"`
	SecurityScore int      `json:"security_score"`
	InstallCount  int      `json:"install_count"`
	PublishedAt   *string  `json:"published_at"`
	Visibility    string   `json:"visibility"`
}

// ==================== MARKETPLACE ====================

// handleMarketplaceSkills lists approved skills from the marketplace
func (s *Server) handleMarketplaceSkills(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.centralStore == nil {
		http.Error(w, "marketplace unavailable", http.StatusServiceUnavailable)
		return
	}

	category := r.URL.Query().Get("category")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	// Determine caller's user ID (0 if unauthenticated; private skills will still be filtered to the owner)
	var callerUserID int64
	if _, userUUID, ok := s.getUserStorage(r); ok {
		callerUserID, _ = s.getUserIDFromUUID(userUUID)
	}

	submissions, err := s.centralStore.GetApprovedSkillSubmissions(category, limit, offset, callerUserID)
	if err != nil {
		http.Error(w, "failed to fetch skills", http.StatusInternalServerError)
		return
	}

	// Convert to public format (without full content)
	publicSkills := make([]PublicSkillResponse, 0, len(submissions))
	for _, sub := range submissions {
		publicSkills = append(publicSkills, PublicSkillResponse{
			ID:            sub.ID,
			Name:          sub.SkillName,
			Slug:          sub.SkillSlug,
			Version:       sub.Version,
			Description:   sub.Description,
			Author:        sub.Author,
			Tags:          sub.Tags,
			Category:      sub.Category,
			SecurityScore: sub.SecurityScore,
			InstallCount:  sub.InstallCount,
			PublishedAt:   sub.PublishedAt,
			Visibility:    sub.Visibility,
		})
	}

	writeJSONResponse(w, map[string]interface{}{
		"skills": publicSkills,
		"page":   page,
	})
}

// handleMarketplaceSkillDetail returns full details of a marketplace skill
func (s *Server) handleMarketplaceSkillDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/marketplace/skills/")
	slug = strings.TrimSuffix(slug, "/install")

	if s.centralStore == nil {
		http.Error(w, "marketplace unavailable", http.StatusServiceUnavailable)
		return
	}

	sub, err := s.centralStore.GetSkillSubmissionBySlug(slug)
	if err != nil {
		http.Error(w, "failed to fetch skill", http.StatusInternalServerError)
		return
	}
	if sub == nil || sub.Status != "approved" {
		http.Error(w, "skill not found", http.StatusNotFound)
		return
	}

	// Private skills are only accessible to their owner
	if sub.Visibility == "private" {
		_, userUUID, ok := s.getUserStorage(r)
		if !ok {
			http.Error(w, "skill not found", http.StatusNotFound)
			return
		}
		callerUserID, _ := s.getUserIDFromUUID(userUUID)
		if callerUserID != sub.UserID {
			http.Error(w, "skill not found", http.StatusNotFound)
			return
		}
	}

	writeJSONResponse(w, sub)
}

// handleMarketplaceInstall installs a skill from the marketplace
func (s *Server) handleMarketplaceInstall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/marketplace/skills/")
	slug = strings.TrimSuffix(slug, "/install")

	if s.centralStore == nil {
		http.Error(w, "marketplace unavailable", http.StatusServiceUnavailable)
		return
	}

	sub, err := s.centralStore.GetSkillSubmissionBySlug(slug)
	if err != nil {
		http.Error(w, "failed to fetch skill", http.StatusInternalServerError)
		return
	}
	if sub == nil || sub.Status != "approved" {
		http.Error(w, "skill not found", http.StatusNotFound)
		return
	}

	// Get user workspace and install the skill
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Private skills can only be installed by their owner
	if sub.Visibility == "private" {
		callerUserID, _ := s.getUserIDFromUUID(userUUID)
		if callerUserID != sub.UserID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	// Create skill directory and write SKILL.md
	installer := skills.NewSkillInstaller(userWorkspace)
	if err := installer.InstallFromContent(sub.SkillSlug, sub.Content); err != nil {
		http.Error(w, "failed to install skill: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Increment install count
	_ = s.centralStore.IncrementSkillInstallCount(sub.ID)

	writeJSONResponse(w, map[string]interface{}{
		"status":  "installed",
		"skill":   sub.SkillSlug,
		"version": sub.Version,
	})
}

// handleMarketplaceSubmit handles skill submission to the marketplace
func (s *Server) handleMarketplaceSubmit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		Content       string   `json:"content"`
		Version       string   `json:"version"`
		Category      string   `json:"category"`
		Tags          []string `json:"tags"`
		RepositoryURL string   `json:"repository_url"`
		Visibility    string   `json:"visibility"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Default and validate visibility
	if body.Visibility == "" {
		body.Visibility = "public"
	}
	if body.Visibility != "public" && body.Visibility != "private" {
		http.Error(w, "visibility must be 'public' or 'private'", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if strings.TrimSpace(body.Name) == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Content) == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}

	// Get user info
	store, userUUID, ok := s.getUserStorage(r)
	if !ok || store == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := s.getUserIDFromUUID(userUUID)
	if userID == 0 {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	user, err := s.centralStore.GetUserByID(userID)
	if err != nil || user == nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Validate skill content
	if err := validateSkillContent(body.Content); err != nil {
		http.Error(w, "invalid skill content: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create slug from name
	slug := slugify(body.Name)

	// Check for existing submission with same slug
	existing, _ := s.centralStore.GetSkillSubmissionBySlug(slug)
	if existing != nil {
		http.Error(w, "skill with this name already exists", http.StatusConflict)
		return
	}

	// Set defaults
	version := body.Version
	if version == "" {
		version = "1.0.0"
	}
	category := body.Category
	if category == "" {
		category = "general"
	}
	description := body.Description
	if description == "" {
		// Extract from frontmatter
		if idx := strings.Index(body.Content, "description:"); idx != -1 {
			endIdx := strings.Index(body.Content[idx:], "\n")
			if endIdx != -1 {
				description = strings.TrimSpace(strings.TrimPrefix(body.Content[idx:idx+endIdx], "description:"))
				description = strings.Trim(description, "\"'")
			}
		}
	}

	now := time.Now().Format(time.RFC3339)

	var securityScore int
	var findingsJSON []byte
	var securityScanAt *string
	var status string

	if body.Visibility == "private" {
		// Private skills bypass the review queue and are auto-approved immediately
		securityScore = 100
		findingsJSON = []byte("[]")
		securityScanAt = nil
		status = "approved"
	} else {
		// Run security scan for public submissions
		var aiReviewer skills.AISecurityReviewer
		if s.agentLoop != nil {
			aiReviewer = skills.NewLLMSecurityReviewer(s.agentLoop)
		}
		scanner := skills.NewSkillSecurityScanner(aiReviewer)

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		scanResult, err := scanner.ScanSkill(ctx, body.Content)
		if err != nil {
			http.Error(w, "security scan failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		securityScore = scanResult.Score
		findingsJSON, _ = json.Marshal(scanResult.Findings)
		securityScanAt = &now

		// Determine submission status based on scan
		if !scanResult.Passed {
			status = "rejected"
		} else if scanResult.AutoApprove {
			status = "approved"
		} else {
			status = "needs_review"
		}
	}

	sub := &storage.SkillSubmission{
		UserID:           userID,
		SkillName:        body.Name,
		SkillSlug:        slug,
		Version:          version,
		Description:      description,
		Content:          body.Content,
		Author:           user.Username,
		Tags:             body.Tags,
		Category:         category,
		RepositoryURL:    body.RepositoryURL,
		SecurityScore:    securityScore,
		SecurityFindings: string(findingsJSON),
		SecurityScanAt:   securityScanAt,
		Status:           status,
		Visibility:       body.Visibility,
	}

	// Private auto-approved skills get published_at set immediately
	if body.Visibility == "private" {
		sub.PublishedAt = &now
	}

	id, err := s.centralStore.CreateSkillSubmission(sub)
	if err != nil {
		http.Error(w, "failed to create submission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If public and auto-approved, set published_at via ApproveSkillSubmission
	if body.Visibility == "public" && status == "approved" {
		_ = s.centralStore.ApproveSkillSubmission(id, userID, "Auto-approved: passed security scan")
	}

	writeJSONResponse(w, map[string]interface{}{
		"id":             id,
		"status":         status,
		"security_score": securityScore,
		"findings":       json.RawMessage(findingsJSON),
		"message":        getSubmissionMessage(status, securityScore),
	})
}

// handleMarketplaceMySubmissions returns the user's own submissions
func (s *Server) handleMarketplaceMySubmissions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, _ := s.getUserIDFromUUID(userUUID)
	if userID == 0 {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	submissions, err := s.centralStore.GetUserSkillSubmissions(userID)
	if err != nil {
		http.Error(w, "failed to fetch submissions", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]interface{}{
		"submissions": submissions,
	})
}

// handleMarketplaceCategories returns all skill categories
func (s *Server) handleMarketplaceCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.centralStore == nil {
		http.Error(w, "marketplace unavailable", http.StatusServiceUnavailable)
		return
	}

	categories, err := s.centralStore.GetSkillCategories()
	if err != nil {
		http.Error(w, "failed to fetch categories", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]interface{}{
		"categories": categories,
	})
}

// ==================== ADMIN ENDPOINTS ====================

// handleAdminPendingSubmissions returns submissions needing review
func (s *Server) handleAdminPendingSubmissions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check admin role
	if !s.isAdmin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	submissions, err := s.centralStore.GetPendingSkillSubmissions()
	if err != nil {
		http.Error(w, "failed to fetch submissions", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]interface{}{
		"submissions": submissions,
	})
}

// handleAdminApproveSubmission approves a skill submission
func (s *Server) handleAdminApproveSubmission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.isAdmin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/submissions/")
	idStr = strings.TrimSuffix(idStr, "/approve")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid submission ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	_, userUUID, _ := s.getUserStorage(r)
	reviewerID, _ := s.getUserIDFromUUID(userUUID)

	if err := s.centralStore.ApproveSkillSubmission(id, reviewerID, body.Notes); err != nil {
		http.Error(w, "failed to approve submission", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]string{"status": "approved"})
}

// handleAdminRejectSubmission rejects a skill submission
func (s *Server) handleAdminRejectSubmission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.isAdmin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/submissions/")
	idStr = strings.TrimSuffix(idStr, "/reject")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid submission ID", http.StatusBadRequest)
		return
	}

	var body struct {
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Notes == "" {
		http.Error(w, "rejection notes are required", http.StatusBadRequest)
		return
	}

	_, userUUID, _ := s.getUserStorage(r)
	reviewerID, _ := s.getUserIDFromUUID(userUUID)

	if err := s.centralStore.RejectSkillSubmission(id, reviewerID, body.Notes); err != nil {
		http.Error(w, "failed to reject submission", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]string{"status": "rejected"})
}

// handleMarketplaceSkillAction routes to detail or install based on path
func (s *Server) handleMarketplaceSkillAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/marketplace/skills/")

	if strings.HasSuffix(path, "/install") {
		s.handleMarketplaceInstall(w, r)
		return
	}

	s.handleMarketplaceSkillDetail(w, r)
}

// handleAdminSubmissionAction routes to approve or reject based on path
func (s *Server) handleAdminSubmissionAction(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasSuffix(path, "/approve") {
		s.handleAdminApproveSubmission(w, r)
		return
	}

	if strings.HasSuffix(path, "/reject") {
		s.handleAdminRejectSubmission(w, r)
		return
	}

	http.Error(w, "not found", http.StatusNotFound)
}

// ==================== HELPERS ====================

func slugify(name string) string {
	name = strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug := reg.ReplaceAllString(name, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 64 {
		slug = slug[:64]
	}
	return slug
}

func getSubmissionMessage(status string, score int) string {
	switch status {
	case "approved":
		return "Your skill has been approved and is now available in the marketplace."
	case "rejected":
		return "Your skill could not be approved due to security concerns. Please review the findings and resubmit."
	case "needs_review":
		return "Your skill has been submitted for manual review. An admin will review it shortly."
	default:
		return "Your skill has been submitted and is pending review."
	}
}

func (s *Server) isAdmin(r *http.Request) bool {
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		return false
	}

	userID, _ := s.getUserIDFromUUID(userUUID)
	if userID == 0 {
		return false
	}

	user, err := s.centralStore.GetUserByID(userID)
	if err != nil || user == nil {
		return false
	}

	return user.Role == "admin"
}

func (s *Server) getUserIDFromUUID(uuid string) (int64, error) {
	if s.centralStore == nil {
		return 0, nil
	}
	user, err := s.centralStore.GetUserByUUID(uuid)
	if err != nil || user == nil {
		return 0, err
	}
	return user.ID, nil
}
