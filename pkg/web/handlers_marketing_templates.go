package web

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
)

type TemplateInfo struct {
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	Variables []string  `json:"variables"`
	Platform  string    `json:"platform,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type createTemplateRequest struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Body     string `json:"body"`
	Platform string `json:"platform"`
}

type updateTemplateRequest struct {
	Name     string `json:"name"`
	Body     string `json:"body"`
	Platform string `json:"platform"`
}

var marketingTemplateSlugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)
var marketingTemplateVariablePattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_\.:-]+)\s*\}\}`)

func (s *Server) handleMarketingTemplatesRouter(w http.ResponseWriter, r *http.Request) {
	suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/marketing/templates/")
	parts := strings.Split(suffix, "/")
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.handleListMarketingTemplates(w, r)
		case http.MethodPost:
			s.handleCreateMarketingTemplate(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	if len(parts) == 2 {
		switch r.Method {
		case http.MethodPut:
			s.handleUpdateMarketingTemplate(w, r)
		case http.MethodDelete:
			s.handleDeleteMarketingTemplate(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	http.Error(w, "invalid template path", http.StatusBadRequest)
}

func (s *Server) handleListMarketingTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userWorkspace, accountName, ok := s.marketingAccountWorkspace(w, r, "/api/v1/marketing/templates/")
	if !ok {
		return
	}

	templatesDir := filepath.Join(userWorkspace, "marketing", accountName, "templates")
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		writeJSONResponse(w, map[string]interface{}{"templates": []TemplateInfo{}})
		return
	}

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		http.Error(w, "failed to read templates directory", http.StatusInternalServerError)
		return
	}

	templates := make([]TemplateInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".json" {
			continue
		}
		tpl, err := readTemplateInfo(filepath.Join(templatesDir, entry.Name()))
		if err != nil {
			continue
		}
		templates = append(templates, tpl)
	}
	sort.Slice(templates, func(i, j int) bool { return templates[i].Slug < templates[j].Slug })

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"templates": templates})
}

func (s *Server) handleCreateMarketingTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userWorkspace, accountName, ok := s.marketingAccountWorkspace(w, r, "/api/v1/marketing/templates/")
	if !ok {
		return
	}

	var req createTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	tpl, err := newTemplateInfo(req.Slug, req.Name, req.Body, req.Platform, time.Now().UTC(), time.Time{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	templatesDir := filepath.Join(userWorkspace, "marketing", accountName, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		http.Error(w, "failed to create templates directory", http.StatusInternalServerError)
		return
	}

	templatePath := filepath.Join(templatesDir, tpl.Slug+".json")
	if _, err := os.Stat(templatePath); err == nil {
		http.Error(w, "template already exists", http.StatusConflict)
		return
	} else if err != nil && !os.IsNotExist(err) {
		http.Error(w, "failed to access template", http.StatusInternalServerError)
		return
	}

	if err := writeMarketingJSON(templatePath, tpl); err != nil {
		http.Error(w, "failed to save template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, tpl)
}

func (s *Server) handleUpdateMarketingTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userWorkspace, accountName, slug, ok := s.marketingTemplateItemWorkspace(w, r)
	if !ok {
		return
	}

	var req updateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	templatePath := filepath.Join(userWorkspace, "marketing", accountName, "templates", slug+".json")
	existing, err := readTemplateInfo(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "template not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to read template", http.StatusInternalServerError)
		return
	}

	updated, err := newTemplateInfo(slug, req.Name, req.Body, req.Platform, existing.CreatedAt, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := writeMarketingJSON(templatePath, updated); err != nil {
		http.Error(w, "failed to update template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updated)
}

func (s *Server) handleDeleteMarketingTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userWorkspace, accountName, slug, ok := s.marketingTemplateItemWorkspace(w, r)
	if !ok {
		return
	}

	templatePath := filepath.Join(userWorkspace, "marketing", accountName, "templates", slug+".json")
	if err := os.Remove(templatePath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "template not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete template", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) marketingTemplateItemWorkspace(w http.ResponseWriter, r *http.Request) (string, string, string, bool) {
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", "", "", false
	}

	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return "", "", "", false
	}

	accountName, slug, err := parseMarketingAccountItemPath(r.URL.Path, "/api/v1/marketing/templates/")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return "", "", "", false
	}
	if isUnsafeMarketingSegment(accountName) || !isValidTemplateSlug(slug) {
		http.Error(w, "access denied", http.StatusForbidden)
		return "", "", "", false
	}

	return userWorkspace, accountName, slug, true
}

func newTemplateInfo(slug, name, body, platform string, createdAt, updatedAt time.Time) (TemplateInfo, error) {
	slug = strings.TrimSpace(slug)
	if !isValidTemplateSlug(slug) {
		return TemplateInfo{}, httpErrorf("invalid slug")
	}
	if strings.TrimSpace(name) == "" {
		return TemplateInfo{}, httpErrorf("name is required")
	}
	if strings.TrimSpace(body) == "" {
		return TemplateInfo{}, httpErrorf("body is required")
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	return TemplateInfo{
		Slug:      slug,
		Name:      strings.TrimSpace(name),
		Body:      body,
		Variables: extractTemplateVariables(body),
		Platform:  strings.TrimSpace(platform),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func readTemplateInfo(path string) (TemplateInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TemplateInfo{}, err
	}
	var tpl TemplateInfo
	if err := json.Unmarshal(data, &tpl); err != nil {
		return TemplateInfo{}, err
	}
	tpl.Variables = extractTemplateVariables(tpl.Body)
	if tpl.UpdatedAt.IsZero() {
		tpl.UpdatedAt = tpl.CreatedAt
	}
	return tpl, nil
}

func extractTemplateVariables(body string) []string {
	matches := marketingTemplateVariablePattern.FindAllStringSubmatch(body, -1)
	seen := make(map[string]struct{}, len(matches))
	variables := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		name := strings.TrimSpace(match[1])
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		variables = append(variables, name)
	}
	sort.Strings(variables)
	return variables
}

func isValidTemplateSlug(slug string) bool {
	return !isUnsafeMarketingSegment(slug) && marketingTemplateSlugPattern.MatchString(strings.TrimSpace(slug))
}

type httpError string

func (e httpError) Error() string { return string(e) }

func httpErrorf(message string) error { return httpError(message) }
