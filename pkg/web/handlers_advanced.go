package web

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/cron"
	"github.com/sipeed/picoclaw/pkg/storage"
)

// ==================== SKILLS ====================

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		q := r.URL.Query().Get("type")
		w.Header().Set("Content-Type", "application/json")

		if q == "available" {
			// Marketplace: list available skills from remote registry
			if s.skillInstaller == nil {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"skills": []interface{}{}})
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			available, err := s.skillInstaller.ListAvailableSkills(ctx)
			if err != nil {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"skills":  []interface{}{},
					"warning": "marketplace unavailable",
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"skills": available})
			return
		}

		// Default: list installed skills
		if s.skillsLoader == nil {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"skills": []interface{}{}})
			return
		}
		installed := s.skillsLoader.ListSkills()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"skills": installed})
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleSkillAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/skills/")
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodPost && strings.HasPrefix(path, "generate"):
		var body struct {
			Name         string `json:"name"`
			Goal         string `json:"goal"`
			Capabilities string `json:"capabilities"`
			Constraints  string `json:"constraints"`
			Tools        string `json:"tools"`
			Examples     string `json:"examples"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		skillName, ok := sanitizeSkillName(body.Name)
		if !ok {
			http.Error(w, "invalid skill name", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(body.Goal) == "" {
			http.Error(w, "goal is required", http.StatusBadRequest)
			return
		}
		if s.agentLoop == nil {
			http.Error(w, "agent loop unavailable", http.StatusServiceUnavailable)
			return
		}

		prompt := buildSkillGenerationPrompt(skillName, body.Goal, body.Capabilities, body.Constraints, body.Tools, body.Examples)
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		rawDraft, err := s.agentLoop.ProcessDirect(ctx, prompt, "web:skills:create:"+skillName)
		if err != nil {
			http.Error(w, "failed to generate skill draft", http.StatusInternalServerError)
			return
		}

		draft := normalizeSkillDraft(skillName, rawDraft)
		if err := validateSkillContent(draft); err != nil {
			http.Error(w, "invalid generated skill draft: "+err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"name":  skillName,
			"draft": draft,
		})

	case r.Method == http.MethodPost && strings.HasPrefix(path, "create"):
		var body struct {
			Name      string `json:"name"`
			Content   string `json:"content"`
			Overwrite bool   `json:"overwrite"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		skillName, ok := sanitizeSkillName(body.Name)
		if !ok {
			http.Error(w, "invalid skill name", http.StatusBadRequest)
			return
		}
		content := normalizeSkillDraft(skillName, body.Content)
		if err := validateSkillContent(content); err != nil {
			http.Error(w, "invalid skill content: "+err.Error(), http.StatusBadRequest)
			return
		}

		skillDir := filepath.Join(s.workspace, "skills", skillName)
		skillPath := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil && !body.Overwrite {
			http.Error(w, "skill already exists", http.StatusConflict)
			return
		}
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			http.Error(w, "failed to create skill directory", http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			http.Error(w, "failed to save skill", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "created",
			"skill": map[string]string{
				"name": skillName,
				"path": filepath.ToSlash(skillPath),
			},
		})

	case r.Method == http.MethodPost && strings.HasPrefix(path, "install"):
		// POST /api/v1/skills/install  body: {"repository": "owner/repo"}
		if s.skillInstaller == nil {
			http.Error(w, "skills installer not available", http.StatusServiceUnavailable)
			return
		}
		var body struct {
			Repository string `json:"repository"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Repository == "" {
			http.Error(w, "repository field required", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := s.skillInstaller.InstallFromGitHub(ctx, body.Repository); err != nil {
			http.Error(w, "install failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "installed", "repository": body.Repository})

	case r.Method == http.MethodDelete:
		// DELETE /api/v1/skills/{name}
		skillName := strings.TrimSpace(path)
		if skillName == "" || skillName == "install" {
			http.Error(w, "skill name required", http.StatusBadRequest)
			return
		}
		if s.skillInstaller == nil {
			http.Error(w, "skills installer not available", http.StatusServiceUnavailable)
			return
		}
		if err := s.skillInstaller.Uninstall(skillName); err != nil {
			http.Error(w, "uninstall failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "uninstalled", "name": skillName})

	case r.Method == http.MethodGet:
		// GET /api/v1/skills/{name} — view skill content
		skillName := strings.TrimSpace(path)
		if skillName == "" {
			http.Error(w, "skill name required", http.StatusBadRequest)
			return
		}
		if s.skillsLoader == nil {
			http.Error(w, "skills loader not available", http.StatusServiceUnavailable)
			return
		}
		content, ok := s.skillsLoader.LoadSkill(skillName)
		if !ok {
			http.Error(w, "skill not found", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"name": skillName, "content": content})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func sanitizeSkillName(name string) (string, bool) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return "", false
	}
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if r == '-' || r == '_' || unicode.IsSpace(r) {
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
			continue
		}
		return "", false
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" || len(slug) > 64 {
		return "", false
	}
	return slug, true
}

func buildSkillGenerationPrompt(name, goal, capabilities, constraints, tools, examples string) string {
	return strings.TrimSpace(fmt.Sprintf(`Create a PicoClaw skill markdown file.
Return only the content for SKILL.md (no code fences, no extra commentary).

Requirements:
- Include YAML frontmatter with at least:
  name: %s
  description: one concise sentence
- Include an H1 title.
- Include sections:
  - When to use
  - Quick start
  - Safety constraints
- Keep instructions practical and concise.
- Do not include destructive or unsafe commands by default.

User input:
- Goal: %s
- Capabilities: %s
- Constraints: %s
- Tools available: %s
- Example interactions: %s
`, name, goal, capabilities, constraints, tools, examples))
}

func normalizeSkillDraft(name, content string) string {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```markdown")
	content = strings.TrimPrefix(content, "```md")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		content = fmt.Sprintf("---\nname: %s\ndescription: AI-generated skill\n---\n\n# %s\n\n%s\n", name, strings.Title(strings.ReplaceAll(name, "-", " ")), content)
	}
	return strings.TrimSpace(content) + "\n"
}

func validateSkillContent(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return fmt.Errorf("content is empty")
	}
	if len(trimmed) > 100_000 {
		return fmt.Errorf("content too large")
	}
	if !strings.HasPrefix(trimmed, "---") {
		return fmt.Errorf("frontmatter is required")
	}
	if !strings.Contains(trimmed, "\ndescription:") {
		return fmt.Errorf("frontmatter description is required")
	}
	if !strings.Contains(trimmed, "\n# ") {
		return fmt.Errorf("title heading is required")
	}
	if strings.Contains(strings.ToLower(trimmed), "rm -rf /") {
		return fmt.Errorf("unsafe command pattern detected")
	}
	return nil
}

// ==================== CRON ====================

func (s *Server) handleCron(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.cronService == nil {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"jobs": []interface{}{}, "status": map[string]interface{}{"enabled": false}})
			return
		}
		http.Error(w, "cron service not available", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		includeDisabled := r.URL.Query().Get("include_disabled") == "true"
		jobs := s.cronService.ListJobs(includeDisabled)
		status := s.cronService.Status()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"jobs": jobs, "status": status})

	case http.MethodPost:
		var body struct {
			Name     string `json:"name"`
			Schedule struct {
				Kind    string `json:"kind"`
				AtMS    *int64 `json:"atMs,omitempty"`
				EveryMS *int64 `json:"everyMs,omitempty"`
				Expr    string `json:"expr,omitempty"`
				TZ      string `json:"tz,omitempty"`
			} `json:"schedule"`
			Message string `json:"message"`
			Deliver bool   `json:"deliver"`
			Channel string `json:"channel,omitempty"`
			To      string `json:"to,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if body.Name == "" || body.Message == "" {
			http.Error(w, "name and message are required", http.StatusBadRequest)
			return
		}

		// Import cron schedule type
		schedule := cronScheduleFromBody(body.Schedule.Kind, body.Schedule.AtMS, body.Schedule.EveryMS, body.Schedule.Expr, body.Schedule.TZ)

		job, err := s.cronService.AddJob(body.Name, schedule, body.Message, body.Deliver, body.Channel, body.To)
		if err != nil {
			// Validation errors from AddJob return 400
			http.Error(w, "failed to create job: "+err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"job": job})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCronAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/cron/")
	w.Header().Set("Content-Type", "application/json")

	if s.cronService == nil {
		http.Error(w, "cron service not available", http.StatusServiceUnavailable)
		return
	}

	switch {
	case r.Method == http.MethodDelete:
		// DELETE /api/v1/cron/{id}
		jobID := strings.TrimSpace(path)
		if jobID == "" {
			http.Error(w, "job id required", http.StatusBadRequest)
			return
		}
		if !s.cronService.RemoveJob(jobID) {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "removed"})

	case r.Method == http.MethodPatch:
		// PATCH /api/v1/cron/{id}  body: {"enabled": true/false}
		jobID := strings.TrimSpace(path)
		if jobID == "" {
			http.Error(w, "job id required", http.StatusBadRequest)
			return
		}
		var body struct {
			Enabled *bool `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Enabled == nil {
			http.Error(w, "enabled field required", http.StatusBadRequest)
			return
		}
		job := s.cronService.EnableJob(jobID, *body.Enabled)
		if job == nil {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"job": job})

	case r.Method == http.MethodPut:
		// PUT /api/v1/cron/{id}  — full update of a job
		jobID := strings.TrimSpace(path)
		if jobID == "" {
			http.Error(w, "job id required", http.StatusBadRequest)
			return
		}
		var body struct {
			Name     string `json:"name"`
			Schedule struct {
				Kind    string `json:"kind"`
				AtMS    *int64 `json:"atMs,omitempty"`
				EveryMS *int64 `json:"everyMs,omitempty"`
				Expr    string `json:"expr,omitempty"`
				TZ      string `json:"tz,omitempty"`
			} `json:"schedule"`
			Message string `json:"message"`
			Deliver bool   `json:"deliver"`
			Channel string `json:"channel,omitempty"`
			To      string `json:"to,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if body.Name == "" || body.Message == "" {
			http.Error(w, "name and message are required", http.StatusBadRequest)
			return
		}

		schedule := cronScheduleFromBody(body.Schedule.Kind, body.Schedule.AtMS, body.Schedule.EveryMS, body.Schedule.Expr, body.Schedule.TZ)

		job, err := s.cronService.UpdateJob(jobID, body.Name, schedule, body.Message, body.Deliver, body.Channel, body.To)
		if err != nil {
			http.Error(w, "invalid job: "+err.Error(), http.StatusBadRequest)
			return
		}
		if job == nil {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"job": job})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ==================== CHANNELS ====================

func (s *Server) handleChannels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.channelManager == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"channels": map[string]interface{}{},
			"enabled":  []string{},
		})
		return
	}

	status := s.channelManager.GetStatus()
	enabled := s.channelManager.GetEnabledChannels()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"channels": status,
		"enabled":  enabled,
	})
}

// ==================== CONFIG ====================

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.fullConfig == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"config": map[string]interface{}{"error": "config not available"}})
		return
	}

	// Build a redacted view of the config
	redacted := map[string]interface{}{
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"workspace":             s.fullConfig.Agents.Defaults.Workspace,
				"restrict_to_workspace": s.fullConfig.Agents.Defaults.RestrictToWorkspace,
				"provider":              s.fullConfig.Agents.Defaults.Provider,
				"model":                 s.fullConfig.Agents.Defaults.Model,
				"max_tokens":            s.fullConfig.Agents.Defaults.MaxTokens,
				"temperature":           s.fullConfig.Agents.Defaults.Temperature,
				"max_tool_iterations":   s.fullConfig.Agents.Defaults.MaxToolIterations,
			},
		},
		"web": map[string]interface{}{
			"enabled": s.fullConfig.Web.Enabled,
			"host":    s.fullConfig.Web.Host,
			"port":    s.fullConfig.Web.Port,
		},
		"gateway": map[string]interface{}{
			"host": s.fullConfig.Gateway.Host,
			"port": s.fullConfig.Gateway.Port,
		},
		"storage": map[string]interface{}{
			"path": s.fullConfig.Storage.Path,
		},
		"providers": redactProviders(s.fullConfig),
		"channels":  redactChannels(s.fullConfig),
		"tools": map[string]interface{}{
			"web": map[string]interface{}{
				"search": map[string]interface{}{
					"api_key":     redactKey(s.fullConfig.Tools.Web.Search.APIKey),
					"max_results": s.fullConfig.Tools.Web.Search.MaxResults,
				},
			},
			"email": map[string]interface{}{
				"enabled": s.fullConfig.Tools.Email.Enabled,
				"host":    s.fullConfig.Tools.Email.Host,
				"port":    s.fullConfig.Tools.Email.Port,
			},
		},
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{"config": redacted})
}

// ==================== FILE BROWSER ====================

type fileEntry struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	IsDir   bool      `json:"is_dir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get relative path from URL
	relPath := strings.TrimPrefix(r.URL.Path, "/api/v1/files")
	relPath = strings.TrimPrefix(relPath, "/")

	// Security: resolve and verify path is within workspace
	fullPath := filepath.Clean(filepath.Join(s.workspace, relPath))
	absWorkspace, _ := filepath.Abs(s.workspace)
	absPath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absPath, absWorkspace) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "path not found", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			http.Error(w, "failed to read directory", http.StatusInternalServerError)
			return
		}

		var files []fileEntry
		for _, e := range entries {
			eInfo, _ := e.Info()
			if eInfo == nil {
				continue
			}
			entryRelPath := filepath.Join(relPath, e.Name())
			files = append(files, fileEntry{
				Name:    e.Name(),
				Path:    filepath.ToSlash(entryRelPath),
				IsDir:   e.IsDir(),
				Size:    eInfo.Size(),
				ModTime: eInfo.ModTime(),
			})
		}

		// Sort: directories first, then by name
		sort.Slice(files, func(i, j int) bool {
			if files[i].IsDir != files[j].IsDir {
				return files[i].IsDir
			}
			return files[i].Name < files[j].Name
		})

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"path":    filepath.ToSlash(relPath),
			"entries": files,
		})
		return
	}

	// It's a file — return content if small enough (<1MB)
	if info.Size() > 1024*1024 {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"path":  filepath.ToSlash(relPath),
			"name":  info.Name(),
			"size":  info.Size(),
			"error": "file too large to display (>1MB)",
		})
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"path":    filepath.ToSlash(relPath),
		"name":    info.Name(),
		"size":    info.Size(),
		"content": string(content),
	})
}

// ==================== EXPORT ====================

func (s *Server) handleExportTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.store == nil {
		http.Error(w, "storage not available", http.StatusServiceUnavailable)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	tasks, err := s.store.ListTasks(true) // Include archived
	if err != nil {
		http.Error(w, "failed to list tasks", http.StatusInternalServerError)
		return
	}

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=tasks_export.csv")
		writer := csv.NewWriter(w)
		_ = writer.Write([]string{"ID", "Title", "Description", "Status", "Result", "Archived", "Created At", "Updated At"})
		for _, t := range tasks {
			archived := "false"
			if t.Archived {
				archived = "true"
			}
			_ = writer.Write([]string{
				fmt.Sprintf("%d", t.ID),
				t.Title,
				t.Description,
				t.Status,
				t.Result,
				archived,
				t.CreatedAt.Format(time.RFC3339),
				t.UpdatedAt.Format(time.RFC3339),
			})
		}
		writer.Flush()

	default: // json
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=tasks_export.json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"exported_at": time.Now().Format(time.RFC3339),
			"count":       len(tasks),
			"tasks":       tasks,
		})
	}
}

func (s *Server) handleExportChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.store == nil {
		http.Error(w, "storage not available", http.StatusServiceUnavailable)
		return
	}

	sessionID := r.URL.Query().Get("session_id")

	if sessionID != "" {
		// Export a single session
		messages, err := s.store.GetMessages(sessionID)
		if err != nil {
			http.Error(w, "failed to get session messages", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=chat_export.json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"exported_at": time.Now().Format(time.RFC3339),
			"session_id":  sessionID,
			"count":       len(messages),
			"messages":    messages,
		})
		return
	}

	// Export all sessions summary
	sessions, err := s.store.ListSessions(nil, 0, 0)
	if err != nil {
		http.Error(w, "failed to list sessions", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=sessions_export.json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"count":       len(sessions),
		"sessions":    sessions,
	})
}

// ==================== HELPERS ====================

func redactKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func redactProviders(cfg *config.Config) map[string]interface{} {
	providers := make(map[string]interface{})
	providerList := []struct {
		name string
		pc   config.ProviderConfig
	}{
		{"anthropic", cfg.Providers.Anthropic},
		{"openai", cfg.Providers.OpenAI},
		{"openrouter", cfg.Providers.OpenRouter},
		{"groq", cfg.Providers.Groq},
		{"zhipu", cfg.Providers.Zhipu},
		{"vllm", cfg.Providers.VLLM},
		{"gemini", cfg.Providers.Gemini},
		{"nvidia", cfg.Providers.Nvidia},
		{"moonshot", cfg.Providers.Moonshot},
		{"ollama", cfg.Providers.Ollama},
	}
	for _, p := range providerList {
		providers[p.name] = map[string]interface{}{
			"api_key":    redactKey(p.pc.APIKey),
			"api_base":   p.pc.APIBase,
			"configured": p.pc.APIKey != "",
		}
	}
	return providers
}

func redactChannels(cfg *config.Config) map[string]interface{} {
	return map[string]interface{}{
		"telegram": map[string]interface{}{
			"enabled":    cfg.Channels.Telegram.Enabled,
			"configured": cfg.Channels.Telegram.Token != "",
		},
		"discord": map[string]interface{}{
			"enabled":    cfg.Channels.Discord.Enabled,
			"configured": cfg.Channels.Discord.Token != "",
		},
		"slack": map[string]interface{}{
			"enabled":    cfg.Channels.Slack.Enabled,
			"configured": cfg.Channels.Slack.BotToken != "",
		},
		"whatsapp": map[string]interface{}{
			"enabled":    cfg.Channels.WhatsApp.Enabled,
			"configured": cfg.Channels.WhatsApp.BridgeURL != "",
		},
		"feishu": map[string]interface{}{
			"enabled":    cfg.Channels.Feishu.Enabled,
			"configured": cfg.Channels.Feishu.AppID != "",
		},
		"dingtalk": map[string]interface{}{
			"enabled":    cfg.Channels.DingTalk.Enabled,
			"configured": cfg.Channels.DingTalk.ClientID != "",
		},
		"qq": map[string]interface{}{
			"enabled":    cfg.Channels.QQ.Enabled,
			"configured": cfg.Channels.QQ.AppID != "",
		},
		"maixcam": map[string]interface{}{
			"enabled": cfg.Channels.MaixCam.Enabled,
		},
		"signal": map[string]interface{}{
			"enabled":    cfg.Channels.Signal.Enabled,
			"configured": cfg.Channels.Signal.PhoneNumber != "",
		},
	}
}

// cronScheduleFromBody constructs a cron.CronSchedule from request body fields
func cronScheduleFromBody(kind string, atMS, everyMS *int64, expr, tz string) cron.CronSchedule {
	return cron.CronSchedule{
		Kind:    kind,
		AtMS:    atMS,
		EveryMS: everyMS,
		Expr:    expr,
		TZ:      tz,
	}
}

// ==================== KNOWLEDGE BASE (RAG) ====================

// handleKnowledge handles GET (list) and POST (upload) for knowledge documents.
func (s *Server) handleKnowledge(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.store == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "storage not configured"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleKnowledgeList(w, r)
	case http.MethodPost:
		s.handleKnowledgeUpload(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}

// handleKnowledgeAction handles DELETE /api/v1/knowledge/{id}
func (s *Server) handleKnowledgeAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.store == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "storage not configured"})
		return
	}

	// Extract ID from path: /api/v1/knowledge/{id}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "document id required"})
		return
	}

	var docID int64
	if _, err := fmt.Sscanf(pathParts[0], "%d", &docID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid document id"})
		return
	}

	if r.Method == http.MethodDelete {
		if err := s.store.DeleteKnowledgeDocument(docID); err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "document not found"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
}

// handleKnowledgeSearch handles GET /api/v1/knowledge/search?q=...
func (s *Server) handleKnowledgeSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.store == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "storage not configured"})
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "query parameter 'q' is required"})
		return
	}

	results, err := s.store.SearchKnowledge(query, 10)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	})
}

func (s *Server) handleKnowledgeList(w http.ResponseWriter, r *http.Request) {
	docs, err := s.store.ListKnowledgeDocuments()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if docs == nil {
		docs = []storage.KnowledgeDocument{}
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"documents": docs})
}

func (s *Server) handleKnowledgeUpload(w http.ResponseWriter, r *http.Request) {
	// Limit upload to 25MB
	r.Body = http.MaxBytesReader(w, r.Body, 25<<20)
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "file too large (max 25MB)"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "file field 'file' is required"})
		return
	}
	defer file.Close()

	// Read the file content
	content := make([]byte, header.Size)
	n, err := file.Read(content)
	if err != nil && n == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to read file"})
		return
	}
	content = content[:n]

	// Determine mime type from extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	mimeType := "text/plain"
	switch ext {
	case ".md", ".markdown":
		mimeType = "text/markdown"
	case ".pdf":
		// We only extract text from PDFs if they are text-based
		// For now, treat as plain text (user should upload extracted text)
		mimeType = "application/pdf"
	case ".txt":
		mimeType = "text/plain"
	case ".html", ".htm":
		mimeType = "text/html"
	case ".json":
		mimeType = "application/json"
	case ".csv":
		mimeType = "text/csv"
	}

	// Chunk the content (split by paragraphs/double-newlines, with a max chunk size)
	text := string(content)
	chunks := chunkText(text, 1000, 200)

	if len(chunks) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "file contains no extractable text"})
		return
	}

	doc, err := s.store.SaveKnowledgeDocument(header.Filename, mimeType, header.Size, chunks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to save document: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(doc)
}

// chunkText splits text into chunks of approximately maxChunkSize characters,
// preferring to break at paragraph boundaries (double newlines). If a paragraph
// is too large, it is split at sentence boundaries or hard-wrapped.
// overlap specifies how many characters from the end of one chunk to prepend to the next.
func chunkText(text string, maxChunkSize, overlap int) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")

	// Split into paragraphs (double newline)
	paragraphs := strings.Split(text, "\n\n")

	var chunks []string
	var current strings.Builder

	flush := func() {
		s := strings.TrimSpace(current.String())
		if s != "" {
			chunks = append(chunks, s)
		}
		// Apply overlap: keep the last `overlap` characters for the next chunk
		if overlap > 0 && len(s) > overlap {
			current.Reset()
			current.WriteString(s[len(s)-overlap:])
		} else {
			current.Reset()
		}
	}

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// If adding this paragraph would exceed the limit, flush first
		if current.Len() > 0 && current.Len()+len(para)+2 > maxChunkSize {
			flush()
		}

		// If the paragraph itself is too big, split it further
		if len(para) > maxChunkSize {
			// Split by sentences (period + space or newline)
			sentences := splitSentences(para)
			for _, sent := range sentences {
				if current.Len() > 0 && current.Len()+len(sent)+1 > maxChunkSize {
					flush()
				}
				if current.Len() > 0 {
					current.WriteString(" ")
				}
				current.WriteString(sent)
			}
		} else {
			if current.Len() > 0 {
				current.WriteString("\n\n")
			}
			current.WriteString(para)
		}
	}

	flush()
	return chunks
}

// splitSentences does a simple sentence split on ". ", "! ", "? ", or newlines.
func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		current.WriteRune(runes[i])
		// Check for sentence-ending punctuation followed by space or end
		if (runes[i] == '.' || runes[i] == '!' || runes[i] == '?') &&
			(i+1 >= len(runes) || runes[i+1] == ' ' || runes[i+1] == '\n') {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		} else if runes[i] == '\n' {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}
	// Remaining text
	s := strings.TrimSpace(current.String())
	if s != "" {
		sentences = append(sentences, s)
	}
	return sentences
}

// ==================== MCP SERVERS ====================

func (s *Server) handleMCPServers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if s.mcpManager == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"servers": []interface{}{},
			"message": "MCP not configured",
		})
		return
	}

	servers := s.mcpManager.ServerStatus()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"servers": servers,
	})
}

func (s *Server) handleMCPServerAction(w http.ResponseWriter, r *http.Request) {
	// Extract server name from path: /api/v1/mcp/{name}/reconnect
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp/")
	parts := strings.SplitN(path, "/", 2)
	serverName := parts[0]

	if serverName == "" {
		http.Error(w, "Server name required", http.StatusBadRequest)
		return
	}

	// Check for reconnect action
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	if action == "reconnect" && r.Method == http.MethodPost {
		s.handleMCPReconnect(w, r, serverName)
		return
	}

	// Default: return info for specific server
	if r.Method == http.MethodGet {
		s.handleMCPServerInfo(w, r, serverName)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleMCPServerInfo(w http.ResponseWriter, r *http.Request, name string) {
	w.Header().Set("Content-Type", "application/json")

	if s.mcpManager == nil {
		http.Error(w, "MCP not configured", http.StatusNotFound)
		return
	}

	for _, srv := range s.mcpManager.ServerStatus() {
		if srv.Name == name {
			_ = json.NewEncoder(w).Encode(srv)
			return
		}
	}

	http.Error(w, "MCP server not found", http.StatusNotFound)
}

func (s *Server) handleMCPReconnect(w http.ResponseWriter, r *http.Request, name string) {
	w.Header().Set("Content-Type", "application/json")

	if s.mcpManager == nil {
		http.Error(w, "MCP not configured", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if err := s.mcpManager.Reconnect(ctx, name); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	// Re-register MCP tools on the agent loop
	if s.agentLoop != nil {
		for _, tool := range s.mcpManager.GetTools() {
			s.agentLoop.RegisterTool(tool)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"message": fmt.Sprintf("Reconnected to MCP server %q", name),
	})
}

// ==================== IMPORT CONVERSATIONS ====================

func (s *Server) handleImportChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.store == nil {
		http.Error(w, "storage not available", http.StatusServiceUnavailable)
		return
	}

	// Limit body to 50MB
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	var payload struct {
		Format string          `json:"format"` // "picoclaw", "chatgpt", "claude", "auto"
		Data   json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if payload.Format == "" {
		payload.Format = "auto"
	}

	sessions, err := parseImportData(payload.Format, payload.Data)
	if err != nil {
		http.Error(w, "failed to parse import data: "+err.Error(), http.StatusBadRequest)
		return
	}

	totalImported := 0
	sessionsImported := 0
	for _, sess := range sessions {
		count, err := s.store.ImportMessages(sess.SessionID, sess.Messages)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to import session %s: %v", sess.SessionID, err), http.StatusInternalServerError)
			return
		}
		totalImported += count
		if count > 0 {
			sessionsImported++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":                true,
		"sessions_imported": sessionsImported,
		"messages_imported": totalImported,
	})
}

type importSession struct {
	SessionID string
	Messages  []storage.ImportMessage
}

// parseImportData detects and parses import data from various formats.
func parseImportData(format string, data json.RawMessage) ([]importSession, error) {
	if format == "auto" {
		format = detectImportFormat(data)
	}

	switch format {
	case "chatgpt":
		return parseChatGPTExport(data)
	case "claude":
		return parseClaudeExport(data)
	case "picoclaw":
		return parsePicoClawExport(data)
	default:
		return nil, fmt.Errorf("unsupported import format: %s", format)
	}
}

// detectImportFormat guesses the format from the JSON structure.
func detectImportFormat(data json.RawMessage) string {
	// Try array first (ChatGPT exports are arrays of conversations)
	var arr []json.RawMessage
	if json.Unmarshal(data, &arr) == nil && len(arr) > 0 {
		// Check if first element has "mapping" field (ChatGPT)
		var first map[string]json.RawMessage
		if json.Unmarshal(arr[0], &first) == nil {
			if _, ok := first["mapping"]; ok {
				return "chatgpt"
			}
		}
	}

	// Try object
	var obj map[string]json.RawMessage
	if json.Unmarshal(data, &obj) == nil {
		// Claude: has "chat_messages" array
		if _, ok := obj["chat_messages"]; ok {
			return "claude"
		}
		// PicoClaw: has "messages" and "session_id"
		if _, ok := obj["messages"]; ok {
			return "picoclaw"
		}
		// PicoClaw all-sessions export: has "sessions"
		if _, ok := obj["sessions"]; ok {
			return "picoclaw"
		}
	}

	return "picoclaw" // default fallback
}

// parseChatGPTExport parses the ChatGPT conversations.json export format.
// Structure: array of { title, mapping: { id: { message: { author: { role }, content: { parts: [] }, create_time } } } }
func parseChatGPTExport(data json.RawMessage) ([]importSession, error) {
	var conversations []struct {
		Title   string `json:"title"`
		Mapping map[string]struct {
			Message *struct {
				Author struct {
					Role string `json:"role"`
				} `json:"author"`
				Content struct {
					Parts []interface{} `json:"parts"`
				} `json:"content"`
				CreateTime float64 `json:"create_time"`
			} `json:"message"`
			Parent   string   `json:"parent"`
			Children []string `json:"children"`
		} `json:"mapping"`
	}

	if err := json.Unmarshal(data, &conversations); err != nil {
		return nil, fmt.Errorf("ChatGPT format parse error: %w", err)
	}

	var sessions []importSession
	for i, conv := range conversations {
		sessionID := fmt.Sprintf("import:chatgpt:%d:%s", i, sanitizeSessionID(conv.Title))
		var msgs []storage.ImportMessage

		// Build ordered list by following parent->children chain
		type nodeInfo struct {
			role       string
			content    string
			createTime float64
		}
		var ordered []nodeInfo

		// Find the root node (no parent or parent not in mapping)
		rootID := ""
		for id, node := range conv.Mapping {
			if node.Parent == "" {
				rootID = id
				break
			}
			if _, exists := conv.Mapping[node.Parent]; !exists {
				rootID = id
				break
			}
		}

		// Walk the tree following first child
		visited := make(map[string]bool)
		current := rootID
		for current != "" && !visited[current] {
			visited[current] = true
			node, ok := conv.Mapping[current]
			if !ok {
				break
			}
			if node.Message != nil && node.Message.Author.Role != "" && node.Message.Author.Role != "system" {
				content := ""
				for _, part := range node.Message.Content.Parts {
					if s, ok := part.(string); ok {
						content += s
					}
				}
				if content != "" {
					ordered = append(ordered, nodeInfo{
						role:       mapChatGPTRole(node.Message.Author.Role),
						content:    content,
						createTime: node.Message.CreateTime,
					})
				}
			}
			current = ""
			if len(node.Children) > 0 {
				current = node.Children[0]
			}
		}

		for _, n := range ordered {
			msg := storage.ImportMessage{
				Role:    n.role,
				Content: n.content,
			}
			if n.createTime > 0 {
				msg.CreatedAt = time.Unix(int64(n.createTime), 0)
			}
			msgs = append(msgs, msg)
		}

		if len(msgs) > 0 {
			sessions = append(sessions, importSession{SessionID: sessionID, Messages: msgs})
		}
	}

	return sessions, nil
}

func mapChatGPTRole(role string) string {
	switch role {
	case "assistant":
		return "assistant"
	case "user":
		return "user"
	case "tool":
		return "assistant"
	default:
		return role
	}
}

// parseClaudeExport parses Claude's export format.
// Structure: { chat_messages: [ { sender, text, created_at } ] } or array of conversations
func parseClaudeExport(data json.RawMessage) ([]importSession, error) {
	// Single conversation format: { uuid, name, chat_messages: [...] }
	var single struct {
		UUID     string `json:"uuid"`
		Name     string `json:"name"`
		Messages []struct {
			Sender    string `json:"sender"`
			Text      string `json:"text"`
			CreatedAt string `json:"created_at"`
		} `json:"chat_messages"`
	}

	if err := json.Unmarshal(data, &single); err == nil && len(single.Messages) > 0 {
		sessionID := fmt.Sprintf("import:claude:%s", sanitizeSessionID(single.Name))
		if single.UUID != "" {
			sessionID = fmt.Sprintf("import:claude:%s", single.UUID)
		}
		var msgs []storage.ImportMessage
		for _, m := range single.Messages {
			role := "user"
			if m.Sender == "assistant" || m.Sender == "human" {
				if m.Sender == "human" {
					role = "user"
				} else {
					role = "assistant"
				}
			}
			msg := storage.ImportMessage{Role: role, Content: m.Text}
			if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
				msg.CreatedAt = t
			}
			msgs = append(msgs, msg)
		}
		return []importSession{{SessionID: sessionID, Messages: msgs}}, nil
	}

	// Array of conversations
	var convs []struct {
		UUID     string `json:"uuid"`
		Name     string `json:"name"`
		Messages []struct {
			Sender    string `json:"sender"`
			Text      string `json:"text"`
			CreatedAt string `json:"created_at"`
		} `json:"chat_messages"`
	}

	if err := json.Unmarshal(data, &convs); err != nil {
		return nil, fmt.Errorf("Claude format parse error: %w", err)
	}

	var sessions []importSession
	for _, conv := range convs {
		sessionID := fmt.Sprintf("import:claude:%s", sanitizeSessionID(conv.Name))
		if conv.UUID != "" {
			sessionID = fmt.Sprintf("import:claude:%s", conv.UUID)
		}
		var msgs []storage.ImportMessage
		for _, m := range conv.Messages {
			role := "user"
			if m.Sender == "assistant" {
				role = "assistant"
			}
			msg := storage.ImportMessage{Role: role, Content: m.Text}
			if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
				msg.CreatedAt = t
			}
			msgs = append(msgs, msg)
		}
		if len(msgs) > 0 {
			sessions = append(sessions, importSession{SessionID: sessionID, Messages: msgs})
		}
	}

	return sessions, nil
}

// parsePicoClawExport parses PicoClaw's own export format.
// Single session: { session_id, messages: [...] }
// All sessions: { sessions: [ { session_id, ... } ] } — re-imports summary only (not useful), so we accept the single format.
func parsePicoClawExport(data json.RawMessage) ([]importSession, error) {
	// Single session
	var single struct {
		SessionID string `json:"session_id"`
		Messages  []struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			CreatedAt string `json:"created_at"`
		} `json:"messages"`
	}

	if err := json.Unmarshal(data, &single); err == nil && len(single.Messages) > 0 {
		sessionID := single.SessionID
		if sessionID == "" {
			sessionID = fmt.Sprintf("import:picoclaw:%d", time.Now().UnixMilli())
		}
		var msgs []storage.ImportMessage
		for _, m := range single.Messages {
			msg := storage.ImportMessage{Role: m.Role, Content: m.Content}
			if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
				msg.CreatedAt = t
			}
			msgs = append(msgs, msg)
		}
		return []importSession{{SessionID: sessionID, Messages: msgs}}, nil
	}

	// Generic array of messages (simple format)
	var msgs []struct {
		Role      string `json:"role"`
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(data, &msgs); err == nil && len(msgs) > 0 {
		sessionID := fmt.Sprintf("import:picoclaw:%d", time.Now().UnixMilli())
		var importMsgs []storage.ImportMessage
		for _, m := range msgs {
			msg := storage.ImportMessage{Role: m.Role, Content: m.Content}
			if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
				msg.CreatedAt = t
			}
			importMsgs = append(importMsgs, msg)
		}
		return []importSession{{SessionID: sessionID, Messages: importMsgs}}, nil
	}

	return nil, fmt.Errorf("could not parse PicoClaw format")
}

func sanitizeSessionID(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	// Keep only alphanumeric, dash, underscore
	var result strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result.WriteRune(c)
		}
	}
	out := result.String()
	if len(out) > 60 {
		out = out[:60]
	}
	if out == "" {
		out = fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	return out
}

// ==================== TOOLS LIST ====================

// handleToolsList returns the list of available tool names for the workflow builder.
func (s *Server) handleToolsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.agentLoop == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"tools": []string{}})
		return
	}
	registry := s.agentLoop.ToolRegistry()
	if registry == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"tools": []string{}})
		return
	}
	names := registry.List()
	if names == nil {
		names = []string{}
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"tools": names})
}

// ==================== WORKFLOWS ====================

// writeJSONError writes a consistent JSON error response with the given status code.
func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// handleWorkflows handles GET (list) and POST (create) for workflows.
func (s *Server) handleWorkflows(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.store == nil {
		writeJSONError(w, "storage not available", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		workflows, err := s.store.ListWorkflows()
		if err != nil {
			writeJSONError(w, "failed to list workflows: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if workflows == nil {
			workflows = []storage.Workflow{}
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"workflows": workflows})

	case http.MethodPost:
		var body struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			Steps       json.RawMessage `json:"steps"`
			Schedule    json.RawMessage `json:"schedule,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if body.Name == "" {
			writeJSONError(w, "name is required", http.StatusBadRequest)
			return
		}
		id, err := s.store.CreateWorkflow(body.Name, body.Description, body.Steps, body.Schedule)
		if err != nil {
			writeJSONError(w, "failed to create workflow: "+err.Error(), http.StatusInternalServerError)
			return
		}
		wf, _ := s.store.GetWorkflow(id)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(wf)

	default:
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleWorkflowAction handles GET/PUT/DELETE /api/v1/workflows/{id} and POST /api/v1/workflows/{id}/run
func (s *Server) handleWorkflowAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.store == nil {
		writeJSONError(w, "storage not available", http.StatusServiceUnavailable)
		return
	}

	// Parse path: /api/v1/workflows/{id} or /api/v1/workflows/{id}/run or /api/v1/workflows/{id}/runs
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		writeJSONError(w, "workflow id required", http.StatusBadRequest)
		return
	}

	workflowID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeJSONError(w, "invalid workflow id", http.StatusBadRequest)
		return
	}

	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch {
	case action == "run" && r.Method == http.MethodPost:
		s.handleWorkflowRun(w, r, workflowID)

	case action == "runs" && r.Method == http.MethodGet:
		runs, err := s.store.ListWorkflowRuns(workflowID, 20)
		if err != nil {
			writeJSONError(w, "failed to list runs: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if runs == nil {
			runs = []storage.WorkflowRun{}
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"runs": runs})

	case action == "" && r.Method == http.MethodGet:
		wf, err := s.store.GetWorkflow(workflowID)
		if err != nil {
			writeJSONError(w, "workflow not found", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(wf)

	case action == "" && r.Method == http.MethodPut:
		var body struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			Enabled     *bool           `json:"enabled"`
			Steps       json.RawMessage `json:"steps"`
			Schedule    json.RawMessage `json:"schedule,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if body.Name == "" {
			writeJSONError(w, "name is required", http.StatusBadRequest)
			return
		}
		// Resolve enabled: use provided value, or fetch existing workflow's value as default
		enabled := true
		if body.Enabled != nil {
			enabled = *body.Enabled
		} else {
			existing, err := s.store.GetWorkflow(workflowID)
			if err == nil {
				enabled = existing.Enabled
			}
		}
		wf, err := s.store.UpdateWorkflow(workflowID, body.Name, body.Description, enabled, body.Steps, body.Schedule)
		if err != nil {
			writeJSONError(w, "failed to update workflow: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(wf)

	case action == "" && r.Method == http.MethodDelete:
		if err := s.store.DeleteWorkflow(workflowID); err != nil {
			writeJSONError(w, "failed to delete: "+err.Error(), http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})

	default:
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleWorkflowRun executes a workflow and returns the results.
// Note: The engine's Run() method internally creates and updates the workflow run
// record in the database, so no additional persistence is needed here.
func (s *Server) handleWorkflowRun(w http.ResponseWriter, r *http.Request, workflowID int64) {
	if s.workflowEngine == nil {
		writeJSONError(w, "workflow engine not available", http.StatusServiceUnavailable)
		return
	}

	wf, err := s.store.GetWorkflow(workflowID)
	if err != nil {
		writeJSONError(w, "workflow not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	results, err := s.workflowEngine.Run(ctx, wf)
	if err != nil {
		writeJSONError(w, "execution failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"results": results,
	})
}
