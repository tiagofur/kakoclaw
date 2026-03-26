package web

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/cron"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/mcp"
	"github.com/sipeed/makoclaw/pkg/skills"
	"github.com/sipeed/makoclaw/pkg/storage"
)

// ==================== SKILLS ====================

// newUserSkillsLoader creates a SkillsLoader correctly configured for a user,
// including user-specific, global, and builtin skills paths.
func (s *Server) newUserSkillsLoader(userUUID, userWorkspace string) *skills.SkillsLoader {
	home, _ := os.UserHomeDir()
	globalSkillsDir := filepath.Join(home, ".MakoClaw", "skills")
	loader := skills.NewSkillsLoader(userWorkspace, globalSkillsDir, s.builtinSkillsDir)
	if userUUID != "" {
		userSkillsPath := filepath.Join(home, ".MakoClaw", "users", userUUID, "skills")
		loader.SetUserSkillsPath(userSkillsPath)
	}
	return loader
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		q := r.URL.Query().Get("type")
		w.Header().Set("Content-Type", "application/json")

		if q == "available" {
			// Marketplace: list available skills from remote registry
			if s.skillInstaller == nil {
				writeJSONResponse(w, map[string]interface{}{"skills": []interface{}{}})
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			available, err := s.skillInstaller.ListAvailableSkills(ctx)
			if err != nil {
				writeJSONResponse(w, map[string]interface{}{
					"skills":  []interface{}{},
					"warning": "marketplace unavailable",
				})
				return
			}
			writeJSONResponse(w, map[string]interface{}{"skills": available})
			return
		}

		// Default: list installed skills for this user
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

		userLoader := s.newUserSkillsLoader(userUUID, userWorkspace)
		installed := userLoader.ListSkills()
		writeJSONResponse(w, map[string]interface{}{"skills": installed})
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

		// Multi-user mode: create per-user agent loop
		if s.userMgr == nil || s.fullConfig == nil || s.msgBus == nil || s.centralStore == nil {
			writeJSONError(w, "AI not available - server not fully configured", http.StatusServiceUnavailable)
			return
		}
		userID, uidOK := s.getUserIDFromClaims(r)
		if !uidOK {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userStore, userUUID, userOK := s.getUserStorage(r)
		if !userOK || userUUID == "" || userStore == nil {
			writeJSONError(w, "user storage unavailable", http.StatusServiceUnavailable)
			return
		}
		userAgentLoop, loopErr := s.newAgentLoopForUser(userUUID, userStore)
		if loopErr != nil {
			logger.ErrorCF("web", "Failed to create agent loop for skill generation", map[string]interface{}{"user_uuid": userUUID, "error": loopErr.Error()})
			writeJSONError(w, "AI not available - "+loopErr.Error(), http.StatusServiceUnavailable)
			return
		}
		defer userAgentLoop.Stop()

		prompt := buildSkillGenerationPrompt(skillName, body.Goal, body.Capabilities, body.Constraints, body.Tools, body.Examples)
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		rawDraft, err := userAgentLoop.ProcessDirectWithUserAndModel(ctx, userID, prompt, "web:skills:create:"+skillName, "", "*")
		if err != nil {
			http.Error(w, "failed to generate skill draft", http.StatusInternalServerError)
			return
		}

		draft := normalizeSkillDraft(skillName, rawDraft)
		if err := validateSkillContent(draft); err != nil {
			http.Error(w, "invalid generated skill draft: "+err.Error(), http.StatusBadRequest)
			return
		}
		writeJSONResponse(w, map[string]interface{}{
			"name":  skillName,
			"draft": draft,
		})

	case r.Method == http.MethodPost && strings.HasPrefix(path, "generate-config"):
		var body struct {
			Description string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(body.Description) == "" {
			http.Error(w, "prompt is required", http.StatusBadRequest)
			return
		}

		// Multi-user mode: create per-user agent loop
		if s.userMgr == nil || s.fullConfig == nil || s.msgBus == nil || s.centralStore == nil {
			writeJSONError(w, "AI not available - server not fully configured", http.StatusServiceUnavailable)
			return
		}
		userID, uidOK := s.getUserIDFromClaims(r)
		if !uidOK {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userStore, userUUID, userOK := s.getUserStorage(r)
		if !userOK || userUUID == "" || userStore == nil {
			writeJSONError(w, "user storage unavailable", http.StatusServiceUnavailable)
			return
		}
		userAgentLoop, loopErr := s.newAgentLoopForUser(userUUID, userStore)
		if loopErr != nil {
			logger.ErrorCF("web", "Failed to create agent loop for skill config generation", map[string]interface{}{"user_uuid": userUUID, "error": loopErr.Error()})
			writeJSONError(w, "AI not available - "+loopErr.Error(), http.StatusServiceUnavailable)
			return
		}
		defer userAgentLoop.Stop()

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		prompt := fmt.Sprintf(`You are an expert AI assistant that helps users create configurations for new MakoClaw skills.
Based on the user's request, generate a logical skill name, a short description of what it does, and any additional context or instructions the skill might need.

USER REQUEST: %s

Respond ONLY with a valid JSON object matching this exact structure, with no markdown formatting or extra text:
{
  "name": "a-lowercase-name-with-hyphens",
  "description": "A short, one-sentence summary of the skill's purpose",
  "prompt": "Any additional context, detailed instructions, or tool requirements mentioned by the user."
}`, body.Description)

		rawJSON, genErr := userAgentLoop.ProcessDirectWithUserAndModel(ctx, userID, prompt, "web:skills:generate-config", "", "*")
		if genErr != nil {
			http.Error(w, "failed to generate config", http.StatusInternalServerError)
			return
		}

		// Clean up potential markdown formatting from the response
		rawJSON = strings.TrimSpace(rawJSON)
		rawJSON = strings.TrimPrefix(rawJSON, "```json")
		rawJSON = strings.TrimPrefix(rawJSON, "```")
		rawJSON = strings.TrimSuffix(rawJSON, "```")
		rawJSON = strings.TrimSpace(rawJSON)

		var result map[string]string
		if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
			logger.ErrorCF("web", "Failed to parse generated skill config", map[string]interface{}{"error": err.Error(), "raw": rawJSON})
			http.Error(w, "failed to parse generated config", http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, result)

	case r.Method == http.MethodPost && strings.HasPrefix(path, "refine"):
		// POST /api/v1/skills/refine  body: {"draft": "...", "feedback": "..."}
		var body struct {
			Draft    string `json:"draft"`
			Feedback string `json:"feedback"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(body.Draft) == "" {
			http.Error(w, "draft is required", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(body.Feedback) == "" {
			http.Error(w, "feedback is required", http.StatusBadRequest)
			return
		}

		// Multi-user mode: create per-user agent loop
		if s.userMgr == nil || s.fullConfig == nil || s.msgBus == nil || s.centralStore == nil {
			writeJSONError(w, "AI not available - server not fully configured", http.StatusServiceUnavailable)
			return
		}
		userID, uidOK := s.getUserIDFromClaims(r)
		if !uidOK {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userStore, userUUID, userOK := s.getUserStorage(r)
		if !userOK || userUUID == "" || userStore == nil {
			writeJSONError(w, "user storage unavailable", http.StatusServiceUnavailable)
			return
		}
		userAgentLoop, loopErr := s.newAgentLoopForUser(userUUID, userStore)
		if loopErr != nil {
			logger.ErrorCF("web", "Failed to create agent loop for skill refinement", map[string]interface{}{"user_uuid": userUUID, "error": loopErr.Error()})
			writeJSONError(w, "AI not available - "+loopErr.Error(), http.StatusServiceUnavailable)
			return
		}
		defer userAgentLoop.Stop()

		prompt := buildSkillRefinementPrompt(body.Draft, body.Feedback)
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		rawDraft, err := userAgentLoop.ProcessDirectWithUserAndModel(ctx, userID, prompt, "web:skills:refine", "", "*")
		if err != nil {
			http.Error(w, "failed to refine skill draft", http.StatusInternalServerError)
			return
		}

		// Extract skill name from the original draft frontmatter
		skillName := "refined-skill"
		if idx := strings.Index(body.Draft, "name:"); idx != -1 {
			endIdx := strings.Index(body.Draft[idx:], "\n")
			if endIdx != -1 {
				nameLine := body.Draft[idx : idx+endIdx]
				nameLine = strings.TrimPrefix(nameLine, "name:")
				skillName = strings.TrimSpace(nameLine)
			}
		}

		draft := normalizeSkillDraft(skillName, rawDraft)
		if err := validateSkillContent(draft); err != nil {
			http.Error(w, "invalid refined skill draft: "+err.Error(), http.StatusBadRequest)
			return
		}
		writeJSONResponse(w, map[string]interface{}{
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
			logger.ErrorCF("web", "create skill invalid body", map[string]interface{}{"error": err.Error()})
			writeJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		skillName, ok := sanitizeSkillName(body.Name)
		if !ok {
			logger.ErrorCF("web", "create skill invalid name", map[string]interface{}{"name": body.Name})
			writeJSONError(w, "invalid skill name: '"+body.Name+"'", http.StatusBadRequest)
			return
		}
		content := normalizeSkillDraft(skillName, body.Content)
		if err := validateSkillContent(content); err != nil {
			previewLen := min(len(content), 200)
			logger.ErrorCF("web", "create skill invalid content", map[string]interface{}{"name": skillName, "error": err.Error(), "content_len": len(content), "preview": content[:previewLen]})
			writeJSONError(w, "invalid skill content: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get user-specific workspace
		_, userUUID, ok := s.getUserStorage(r)
		if !ok {
			writeJSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userWorkspace, err := config.EnsureUserWorkspace(userUUID)
		if err != nil {
			writeJSONError(w, "failed to access workspace", http.StatusInternalServerError)
			return
		}

		skillDir := filepath.Join(userWorkspace, "skills", skillName)
		skillPath := filepath.Join(skillDir, "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil && !body.Overwrite {
			writeJSONError(w, "skill already exists", http.StatusConflict)
			return
		}
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			writeJSONError(w, "failed to create skill directory", http.StatusInternalServerError)
			return
		}
		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			writeJSONError(w, "failed to save skill", http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, map[string]interface{}{
			"status": "created",
			"skill": map[string]string{
				"name": skillName,
				"path": filepath.ToSlash(skillPath),
			},
		})

	case r.Method == http.MethodPost && strings.HasPrefix(path, "install"):
		// POST /api/v1/skills/install  body: {"repository": "owner/repo"}
		var body struct {
			Repository string `json:"repository"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Repository == "" {
			http.Error(w, "repository field required", http.StatusBadRequest)
			return
		}

		// Get user-specific workspace
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

		// Create a temporary installer for this user's workspace
		userInstaller := skills.NewSkillInstaller(userWorkspace)

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := userInstaller.InstallFromGitHub(ctx, body.Repository); err != nil {
			http.Error(w, "install failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, map[string]string{"status": "installed", "repository": body.Repository})

	case r.Method == http.MethodPost && strings.HasPrefix(path, "scan"):
		// POST /api/v1/skills/scan  body: {"content": "..."}
		var body struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
			http.Error(w, "content field required", http.StatusBadRequest)
			return
		}

		// Create security scanner with AI reviewer using per-user agent loop
		var aiReviewer skills.AISecurityReviewer
		if s.userMgr != nil && s.fullConfig != nil && s.msgBus != nil && s.centralStore != nil {
			userStore, userUUID, userOK := s.getUserStorage(r)
			if userOK && userUUID != "" && userStore != nil {
				if userAgentLoop, loopErr := s.newAgentLoopForUser(userUUID, userStore); loopErr == nil {
					defer userAgentLoop.Stop()
					aiReviewer = skills.NewLLMSecurityReviewer(userAgentLoop)
				}
			}
		} else if s.agentLoop != nil {
			aiReviewer = skills.NewLLMSecurityReviewer(s.agentLoop)
		}
		scanner := skills.NewSkillSecurityScanner(aiReviewer)

		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()

		result, err := scanner.ScanSkill(ctx, body.Content)
		if err != nil {
			http.Error(w, "scan failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, result)

	case r.Method == http.MethodDelete:
		// DELETE /api/v1/skills/{name}
		skillName := strings.TrimSpace(path)
		if skillName == "" || skillName == "install" {
			http.Error(w, "skill name required", http.StatusBadRequest)
			return
		}

		// Get user-specific workspace
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

		// Create a temporary installer for this user's workspace
		userInstaller := skills.NewSkillInstaller(userWorkspace)

		if err := userInstaller.Uninstall(skillName); err != nil {
			http.Error(w, "uninstall failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, map[string]string{"status": "uninstalled", "name": skillName})

	case r.Method == http.MethodGet:
		// GET /api/v1/skills/{name} — view skill content
		skillName := strings.TrimSpace(path)
		if skillName == "" {
			http.Error(w, "skill name required", http.StatusBadRequest)
			return
		}

		// Get user-specific workspace
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

		userLoader := s.newUserSkillsLoader(userUUID, userWorkspace)
		content, ok := userLoader.LoadSkill(skillName)
		if !ok {
			http.Error(w, "skill not found", http.StatusNotFound)
			return
		}
		writeJSONResponse(w, map[string]string{"name": skillName, "content": content})

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
	return strings.TrimSpace(fmt.Sprintf(`You are an expert skill designer for MakoClaw AI agents. Create a high-quality SKILL.md file.

## SKILL DESIGN PRINCIPLES
1. **Concise is Key**: The context window is shared. Only include what the agent doesn't already know.
2. **Progressive Disclosure**: Keep SKILL.md under 300 lines. Use references/ folder for detailed docs.
3. **Appropriate Freedom**: Match specificity to task fragility (high freedom for flexible tasks, low for fragile ones).
4. **Actionable Instructions**: Every section should provide clear, executable guidance.

## REQUIRED OUTPUT FORMAT
Return ONLY the SKILL.md content (no code fences, no commentary).

Structure:
1. YAML frontmatter with:
   name: %s
   description: one clear sentence describing WHEN to use this skill (e.g., "Use this skill when the user wants to...")
2. H1 title matching the skill name
3. Required sections (in this order):
   - ## When to Use - Brief trigger conditions
   - ## Quick Start - 2-3 practical examples with expected input/output
   - ## Workflow - Step-by-step instructions for the agent
   - ## Safety Constraints - What NOT to do, required confirmations
4. Optional sections (only if needed):
   - ## Prerequisites - Required tools/binaries
   - ## Configuration - Environment variables or settings

## SKILL SPECIFICATION
- Name: %s
- Goal: %s
- Capabilities: %s
- Constraints: %s
- Available Tools: %s
- Example Interactions: %s

## SAFETY REQUIREMENTS (CRITICAL)
- NEVER include destructive commands (rm -rf, format, etc.) without explicit warnings
- NEVER include commands that could expose credentials or secrets
- ALWAYS require user confirmation for file deletions, system changes, or external communications
- NEVER instruct the agent to hide actions from the user
- Include a "Safety Constraints" section documenting:
  * What the agent should NOT do
  * When to ask for user confirmation
  * Potential risks and how to mitigate them

## QUALITY CHECKLIST
Before returning, verify:
[ ] Frontmatter has name and comprehensive description with trigger phrase
[ ] Description starts with "Use this skill when..."
[ ] Quick start section has 2-3 working examples
[ ] Workflow section has numbered steps
[ ] Safety constraints section documents risks
[ ] No placeholder text like "TODO" or "[fill in]"
[ ] Content is under 300 lines
`, name, name, goal, capabilities, constraints, tools, examples))
}

func buildSkillRefinementPrompt(draft, feedback string) string {
	return strings.TrimSpace(fmt.Sprintf(`You are refining a MakoClaw skill based on user feedback.

## CURRENT DRAFT
%s

## USER FEEDBACK
%s

## INSTRUCTIONS
1. Apply the user's feedback to improve the skill
2. Maintain the same structure (frontmatter, sections)
3. Keep the skill concise and actionable
4. Return ONLY the improved SKILL.md content (no code fences, no commentary)
5. If the feedback is unclear, make reasonable improvements based on best practices

Return the complete, refined SKILL.md content.
`, draft, feedback))
}

func normalizeSkillDraft(name, content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```markdown")
	content = strings.TrimPrefix(content, "```md")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		content = fmt.Sprintf("---\nname: %s\ndescription: AI-generated skill\n---\n\n# %s\n\n%s\n", name, toTitleCase(strings.ReplaceAll(name, "-", " ")), content)
	} else {
		// Content already has frontmatter — ensure required fields are present.
		// Find the closing "---" of the frontmatter (look for "\n---" after the opening "---").
		closeOff := strings.Index(content[3:], "\n---")
		if closeOff != -1 {
			closeAbs := 3 + closeOff // absolute index of the '\n' before closing '---'
			endFM := closeAbs + 4    // absolute index right after '\n---'
			fm := content[:endFM]    // frontmatter block including closing "---"
			body := content[endFM:]  // everything after the closing "---"

			// Inject description field if missing.
			if !strings.Contains(fm, "\ndescription:") {
				fm = content[:closeAbs] + "\ndescription: Custom skill" + content[closeAbs:endFM]
			}

			// Inject H1 heading if missing from the full content.
			if !strings.Contains(fm+body, "\n# ") {
				title := toTitleCase(strings.ReplaceAll(name, "-", " "))
				body = "\n\n# " + title + "\n" + strings.TrimLeft(body, "\n")
			}

			content = fm + body
		}
	}
	return strings.TrimSpace(content) + "\n"
}

// toTitleCase capitalises the first letter of each word (ASCII-only, suitable for skill names).
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func validateSkillContent(content string) error {
	content = strings.ReplaceAll(content, "\r\n", "\n")
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

	cronService, userID, err := s.getCronServiceForRequest(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		userUUID := ""
		if claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims); ok && claims != nil {
			userUUID = claims.UUID
		}
		if errors.Is(err, cron.ErrCronNotInitialized) {
			logger.ErrorCF("web", "Per-user cron service not initialized", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		logger.ErrorCF("web", "Failed to resolve cron service for request", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if cronService == nil {
		if r.Method == http.MethodGet {
			writeJSONResponse(w, map[string]interface{}{
				"jobs": []interface{}{},
				"status": map[string]interface{}{
					"enabled": false,
					"reason":  "not_initialized",
				},
			})
			return
		}
		http.Error(w, "cron service not available", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		includeDisabled := r.URL.Query().Get("include_disabled") == "true"
		jobs := cronService.ListJobsForUser(userID, includeDisabled)
		status := cronService.Status()
		writeJSONResponse(w, map[string]interface{}{"jobs": jobs, "status": status})

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
			JobType string `json:"job_type"` // New field: "task" or "reminder"
			Agent   string `json:"agent,omitempty"`
			Deliver bool   `json:"deliver,omitempty"` // Deprecated, for backward compatibility
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

		// Handle job_type with backward compatibility
		jobType := body.JobType
		if jobType == "" {
			// Backward compatibility: convert from deliver field
			if body.Deliver {
				jobType = "reminder"
			} else {
				jobType = "task"
			}
		}
		// Validate job_type
		if jobType != "task" && jobType != "reminder" {
			http.Error(w, "job_type must be 'task' or 'reminder'", http.StatusBadRequest)
			return
		}

		// Import cron schedule type
		schedule := cronScheduleFromBody(body.Schedule.Kind, body.Schedule.AtMS, body.Schedule.EveryMS, body.Schedule.Expr, body.Schedule.TZ)

		job, err := cronService.AddJob(userID, body.Name, schedule, body.Message, jobType, body.Channel, body.To, body.Agent)
		if err != nil {
			// Validation errors from AddJob return 400
			http.Error(w, "failed to create job: "+err.Error(), http.StatusBadRequest)
			return
		}
		writeJSONResponse(w, map[string]interface{}{"job": job})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCronAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/cron/")
	w.Header().Set("Content-Type", "application/json")

	cronService, userID, err := s.getCronServiceForRequest(r)
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		userUUID := ""
		if claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims); ok && claims != nil {
			userUUID = claims.UUID
		}
		if errors.Is(err, cron.ErrCronNotInitialized) {
			logger.ErrorCF("web", "Per-user cron service not initialized", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		logger.ErrorCF("web", "Failed to resolve cron service for request", map[string]interface{}{"user_uuid": userUUID, "error": err.Error()})
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if cronService == nil {
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
		if !cronService.RemoveJobForUser(userID, jobID) {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		writeJSONResponse(w, map[string]string{"status": "removed"})

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
		job := cronService.EnableJobForUser(userID, jobID, *body.Enabled)
		if job == nil {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		writeJSONResponse(w, map[string]interface{}{"job": job})

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
			JobType string `json:"job_type"` // New field: "task" or "reminder"
			Agent   string `json:"agent,omitempty"`
			Deliver bool   `json:"deliver,omitempty"` // Deprecated, for backward compatibility
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

		// Handle job_type with backward compatibility
		jobType := body.JobType
		if jobType == "" {
			// Backward compatibility: convert from deliver field
			if body.Deliver {
				jobType = "reminder"
			} else {
				jobType = "task"
			}
		}
		// Validate job_type
		if jobType != "task" && jobType != "reminder" {
			http.Error(w, "job_type must be 'task' or 'reminder'", http.StatusBadRequest)
			return
		}

		schedule := cronScheduleFromBody(body.Schedule.Kind, body.Schedule.AtMS, body.Schedule.EveryMS, body.Schedule.Expr, body.Schedule.TZ)

		job, err := cronService.UpdateJobForUser(userID, jobID, body.Name, schedule, body.Message, jobType, body.Channel, body.To, body.Agent)
		if err != nil {
			http.Error(w, "invalid job: "+err.Error(), http.StatusBadRequest)
			return
		}
		if job == nil {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		writeJSONResponse(w, map[string]interface{}{"job": job})

	case r.Method == http.MethodPost && strings.HasSuffix(path, "/run"):
		jobID := strings.TrimSpace(strings.TrimSuffix(path, "/run"))
		if jobID == "" {
			http.Error(w, "job id required", http.StatusBadRequest)
			return
		}
		// Execute asynchronously to avoid 504 Gateway Timeout from nginx.
		// Agent processing (LLM calls, tool execution) can take minutes.
		go func() {
			if err := cronService.RunJobForUser(userID, jobID); err != nil {
				logger.ErrorCF("web", "Async cron job execution failed", map[string]interface{}{
					"job_id": jobID,
					"error":  err.Error(),
				})
			}
		}()
		writeJSONResponse(w, map[string]string{"status": "triggered"})

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

	// Try per-user channel manager first (multi-user mode)
	if s.multiUserChannelManager != nil {
		_, userUUID, ok := s.getUserStorage(r)
		if ok && userUUID != "" {
			if mgr, found := s.multiUserChannelManager.GetManagerForUser(userUUID); found {
				status := mgr.GetStatus()
				enabled := mgr.GetEnabledChannels()
				writeJSONResponse(w, map[string]interface{}{
					"channels": status,
					"enabled":  enabled,
				})
				return
			}
		}
	}

	// Fallback to global channel manager
	if s.channelManager == nil {
		writeJSONResponse(w, map[string]interface{}{
			"channels": map[string]interface{}{},
			"enabled":  []string{},
		})
		return
	}

	status := s.channelManager.GetStatus()
	enabled := s.channelManager.GetEnabledChannels()
	writeJSONResponse(w, map[string]interface{}{
		"channels": status,
		"enabled":  enabled,
	})
}

// ==================== CONFIG ====================

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodPost {
		claims, ok := r.Context().Value(userClaimsKey).(*jwtClaims)
		if !ok || claims == nil || claims.Role != "admin" {
			http.Error(w, "forbidden: admin role required", http.StatusForbidden)
			return
		}

		var newConfig map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Validate: user-specific settings must be updated via /api/v1/me/config/update
		userSpecificKeys := []string{"tools", "agents"}
		for _, key := range userSpecificKeys {
			if _, ok := newConfig[key]; ok {
				logger.WarnCF("web", "Attempted to update user-specific setting via global config endpoint", map[string]interface{}{
					"key":              key,
					"endpoint":         "/api/v1/config",
					"correct_endpoint": "/api/v1/me/config/update",
					"user":             claims.Sub,
				})
				http.Error(w, fmt.Sprintf("%s settings must be updated via /api/v1/me/config/update", key), http.StatusBadRequest)
				return
			}
		}

		// Log global config update
		logger.InfoCF("web", "Updating global config", map[string]interface{}{
			"user":             claims.Sub,
			"updated_sections": getMapKeys(newConfig),
		})

		// Update components if present
		updatedAny := false

		if providers, ok := newConfig["providers"].(map[string]interface{}); ok {
			updateProvidersConfig(s.fullConfig, providers)
			updatedAny = true
		}

		if channels, ok := newConfig["channels"].(map[string]interface{}); ok {
			updateChannelConfig(s.fullConfig, channels)
			updatedAny = true
		}

		if updatedAny {
			// Persist config
			configPath := filepath.Join(os.Getenv("HOME"), ".MakoClaw", "config.json")
			if path := os.Getenv("makoclaw_CONFIG_PATH"); path != "" {
				configPath = path
			}
			// Special handling for home dir
			if home, err := os.UserHomeDir(); err == nil && !strings.Contains(configPath, home) && strings.HasPrefix(configPath, "~") {
				configPath = strings.Replace(configPath, "~", home, 1)
			}

			// Re-save using config package
			if err := config.SaveConfig(configPath, s.fullConfig); err != nil {
				logger.ErrorCF("web", "Failed to save config", map[string]interface{}{"error": err.Error()})
				http.Error(w, "failed to save config: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Restart affected channels if channels were updated
			if _, ok := newConfig["channels"].(map[string]interface{}); ok && s.channelManager != nil {
				ctx := context.Background()
				for name := range newConfig["channels"].(map[string]interface{}) {
					if err := s.channelManager.RestartChannel(ctx, name); err != nil {
						logger.ErrorCF("web", "Failed to restart channel", map[string]interface{}{"channel": name, "error": err.Error()})
					}
				}
			}
		}

		// Return updated config (redacted)
	}

	if s.fullConfig == nil {
		writeJSONResponse(w, map[string]interface{}{"config": map[string]interface{}{"error": "config not available"}})
		return
	}

	// Use per-user merged config when possible, fall back to global
	effectiveCfg := s.fullConfig
	_, userUUID, ok := s.getUserStorage(r)
	if ok && userUUID != "" {
		if userCfg, err := config.LoadConfigForUser(userUUID); err == nil {
			effectiveCfg = config.MergeConfigs(s.fullConfig, userCfg)
		}
	}

	// Build a redacted view of the config
	redacted := map[string]interface{}{
		"agents": map[string]interface{}{
			"defaults": map[string]interface{}{
				"workspace":             effectiveCfg.Agents.Defaults.Workspace,
				"restrict_to_workspace": effectiveCfg.Agents.Defaults.RestrictToWorkspace,
				"provider":              effectiveCfg.Agents.Defaults.Provider,
				"model":                 effectiveCfg.Agents.Defaults.Model,
				"max_tokens":            effectiveCfg.Agents.Defaults.MaxTokens,
				"temperature":           effectiveCfg.Agents.Defaults.Temperature,
				"max_tool_iterations":   effectiveCfg.Agents.Defaults.MaxToolIterations,
			},
		},
		"web": map[string]interface{}{
			"enabled": effectiveCfg.Web.Enabled,
			"host":    effectiveCfg.Web.Host,
			"port":    effectiveCfg.Web.Port,
		},
		"gateway": map[string]interface{}{
			"host": effectiveCfg.Gateway.Host,
			"port": effectiveCfg.Gateway.Port,
		},
		"storage": map[string]interface{}{
			"path": effectiveCfg.Storage.Path,
		},
		"providers": redactProviders(effectiveCfg),
		"channels":  redactChannels(effectiveCfg),
		"tools": map[string]interface{}{
			"web": map[string]interface{}{
				"search": map[string]interface{}{
					"api_key":     redactKey(effectiveCfg.Tools.Web.Search.APIKey),
					"max_results": effectiveCfg.Tools.Web.Search.MaxResults,
				},
			},
			"email": map[string]interface{}{
				"enabled": effectiveCfg.Tools.Email.Enabled,
				"host":    effectiveCfg.Tools.Email.Host,
				"port":    effectiveCfg.Tools.Email.Port,
				"to":      effectiveCfg.Tools.Email.To,
				"from":    effectiveCfg.Tools.Email.From,
			},
		},
	}

	writeJSONResponse(w, map[string]interface{}{"config": redacted})
}

// ==================== FILE BROWSER ====================

type fileEntry struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	IsDir      bool      `json:"is_dir"`
	Size       int64     `json:"size"`
	ModTime    time.Time `json:"mod_time"`
	IsSystem   bool      `json:"is_system"`
	IsReadOnly bool      `json:"is_read_only"`
}

// isSystemFile checks if a file path corresponds to a system-managed file
// that should be protected from direct editing
func isSystemFile(relPath string) bool {
	systemPaths := []string{
		"cron/jobs.json",
		"database.db",
		"sessions/",
		"memory/",
		"knowledge/",
		".context/",
	}

	// Bootstrap docs (AGENTS.md, SOUL.md, USER.md, IDENTITY.md) are NOT protected
	// They are user-editable configuration files
	relPath = filepath.ToSlash(relPath)

	for _, sys := range systemPaths {
		if strings.HasPrefix(relPath, sys) || relPath == strings.TrimSuffix(sys, "/") {
			return true
		}
	}

	return false
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Supported methods: GET (list/read/search), POST (upload), DELETE (remove), PUT (create dir/update file), PATCH (rename)
	supportedMethods := map[string]bool{
		http.MethodGet:    true,
		http.MethodPost:   true,
		http.MethodDelete: true,
		http.MethodPut:    true,
		http.MethodPatch:  true,
	}
	if !supportedMethods[r.Method] {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from claims
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user-specific workspace
	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	// Get relative path from URL
	relPath := strings.TrimPrefix(r.URL.Path, "/api/v1/files")
	relPath = strings.TrimPrefix(relPath, "/")

	// Security: resolve and verify path is within user workspace
	fullPath := filepath.Clean(filepath.Join(userWorkspace, relPath))
	absWorkspace, _ := filepath.Abs(userWorkspace)
	absPath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absPath, absWorkspace) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	// Handle Upload (POST)
	if r.Method == http.MethodPost {
		// Limit upload size (e.g., 20MB)
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file field", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// If fullPath is a directory, append filename. Otherwise fullPath is the target file.
		targetPath := fullPath
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			targetPath = filepath.Join(fullPath, header.Filename)
		}

		// Security: verify targetPath is still in workspace (after joining filename)
		absTarget, _ := filepath.Abs(targetPath)
		if !strings.HasPrefix(absTarget, absWorkspace) {
			http.Error(w, "invalid filename", http.StatusForbidden)
			return
		}

		out, err := os.Create(targetPath)
		if err != nil {
			http.Error(w, "failed to create file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, "failed to save file", http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, map[string]string{
			"status": "uploaded",
			"name":   header.Filename,
			"path":   filepath.ToSlash(strings.TrimPrefix(targetPath, userWorkspace)),
		})
		return
	}

	// Handle Delete (DELETE)
	if r.Method == http.MethodDelete {
		// Security: prevent deletion of system files
		if isSystemFile(relPath) {
			http.Error(w, "cannot delete system files", http.StatusForbidden)
			return
		}

		info, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "path not found", http.StatusNotFound)
			} else {
				http.Error(w, "failed to access path", http.StatusInternalServerError)
			}
			return
		}

		// Security: check for symlink attacks - don't follow symlinks outside workspace
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := filepath.EvalSymlinks(fullPath)
			if err != nil {
				http.Error(w, "failed to resolve symlink", http.StatusInternalServerError)
				return
			}
			absLinkTarget, _ := filepath.Abs(linkTarget)
			if !strings.HasPrefix(absLinkTarget, absWorkspace) {
				http.Error(w, "cannot delete symlinks pointing outside workspace", http.StatusForbidden)
				return
			}
		}

		if info.IsDir() {
			// Recursive delete for directories
			if err := os.RemoveAll(fullPath); err != nil {
				http.Error(w, "failed to delete directory: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// Single file delete
			if err := os.Remove(fullPath); err != nil {
				http.Error(w, "failed to delete file: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		writeJSONResponse(w, map[string]interface{}{
			"success": true,
			"message": "Deleted: " + filepath.ToSlash(relPath),
		})
		return
	}

	// Handle Create Folder or Update File (PUT)
	if r.Method == http.MethodPut {
		// Check if this is a mkdir operation
		if r.URL.Query().Get("mkdir") == "true" {
			// Create directory
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				http.Error(w, "failed to create directory: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, map[string]interface{}{
				"success": true,
				"message": "Created directory: " + filepath.ToSlash(relPath),
				"path":    filepath.ToSlash(relPath),
			})
			return
		}

		// Security: prevent modification of system files
		if isSystemFile(relPath) {
			http.Error(w, "cannot modify system files", http.StatusForbidden)
			return
		}

		// Update file content
		// Limit request body to 5MB for text files
		r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

		content, err := io.ReadAll(r.Body)
		if err != nil {
			if strings.Contains(err.Error(), "http: request body too large") {
				http.Error(w, "file content too large (max 5MB)", http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
			}
			return
		}

		// Ensure parent directory exists
		parentDir := filepath.Dir(fullPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			http.Error(w, "failed to create parent directory", http.StatusInternalServerError)
			return
		}

		// Write the file
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			http.Error(w, "failed to write file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, map[string]interface{}{
			"success": true,
			"message": "Updated file: " + filepath.ToSlash(relPath),
			"path":    filepath.ToSlash(relPath),
			"size":    len(content),
		})
		return
	}

	// Handle Rename (PATCH)
	if r.Method == http.MethodPatch {
		var body struct {
			NewName string `json:"new_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(body.NewName) == "" {
			http.Error(w, "new_name is required", http.StatusBadRequest)
			return
		}

		// Validate new_name doesn't contain path separators or dangerous characters
		if strings.ContainsAny(body.NewName, "/\\") {
			http.Error(w, "new_name cannot contain path separators", http.StatusBadRequest)
			return
		}

		// Check source exists
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "path not found", http.StatusNotFound)
			} else {
				http.Error(w, "failed to access path", http.StatusInternalServerError)
			}
			return
		}

		// Build new path (same directory, new name)
		newPath := filepath.Join(filepath.Dir(fullPath), body.NewName)

		// Security: verify new path is still in workspace
		absNewPath, _ := filepath.Abs(newPath)
		if !strings.HasPrefix(absNewPath, absWorkspace) {
			http.Error(w, "invalid new name: would escape workspace", http.StatusForbidden)
			return
		}

		// Check if target already exists
		if _, err := os.Stat(newPath); err == nil {
			http.Error(w, "a file or folder with that name already exists", http.StatusConflict)
			return
		}

		// Perform rename
		if err := os.Rename(fullPath, newPath); err != nil {
			http.Error(w, "failed to rename: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Calculate new relative path
		newRelPath, _ := filepath.Rel(userWorkspace, newPath)
		writeJSONResponse(w, map[string]interface{}{
			"success":  true,
			"message":  "Renamed to: " + body.NewName,
			"old_path": filepath.ToSlash(relPath),
			"new_path": filepath.ToSlash(newRelPath),
		})
		return
	}

	// =====================
	// GET method handling
	// =====================
	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "path not found", http.StatusNotFound)
		return
	}

	isDownload := r.URL.Query().Get("download") == "true"

	if info.IsDir() {
		if isDownload {
			zipName := info.Name()
			if zipName == "." || zipName == "" || zipName == "/" {
				zipName = "workspace"
			}
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", sanitizeFilename(zipName)))

			zw := zip.NewWriter(w)
			defer zw.Close()

			err := filepath.Walk(fullPath, func(path string, walkInfo os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if walkInfo.IsDir() {
					return nil
				}

				relP, err := filepath.Rel(fullPath, path)
				if err != nil {
					return err
				}

				f, err := zw.Create(filepath.ToSlash(relP))
				if err != nil {
					return err
				}

				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()

				_, err = io.Copy(f, file)
				return err
			})
			if err != nil {
				logger.ErrorCF("web", "Failed to create zip", map[string]interface{}{"path": fullPath, "error": err.Error()})
			}
			return
		}

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
			isSystemPath := isSystemFile(entryRelPath)
			files = append(files, fileEntry{
				Name:       e.Name(),
				Path:       filepath.ToSlash(entryRelPath),
				IsDir:      e.IsDir(),
				Size:       eInfo.Size(),
				ModTime:    eInfo.ModTime(),
				IsSystem:   isSystemPath,
				IsReadOnly: isSystemPath,
			})
		}

		// Sort: directories first, then by name
		sort.Slice(files, func(i, j int) bool {
			if files[i].IsDir != files[j].IsDir {
				return files[i].IsDir
			}
			return files[i].Name < files[j].Name
		})

		// Handle search filter if provided
		searchQuery := strings.TrimSpace(r.URL.Query().Get("search"))
		if searchQuery != "" {
			searchLower := strings.ToLower(searchQuery)
			var filtered []fileEntry
			for _, f := range files {
				if strings.Contains(strings.ToLower(f.Name), searchLower) {
					filtered = append(filtered, f)
				}
			}
			files = filtered
		}

		response := map[string]interface{}{
			"path":    filepath.ToSlash(relPath),
			"entries": files,
		}
		if searchQuery != "" {
			response["search"] = searchQuery
		}
		writeJSONResponse(w, response)
		return
	}

	if isDownload {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", sanitizeFilename(info.Name())))
		w.Header().Set("Content-Type", "application/octet-stream")
		// serve file content
		file, err := os.Open(fullPath)
		if err != nil {
			http.Error(w, "failed to read file", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		io.Copy(w, file)
		return
	}

	// It's a file — return content if small enough (<1MB)
	if info.Size() > 1024*1024 {
		writeJSONResponse(w, map[string]interface{}{
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

	writeJSONResponse(w, map[string]interface{}{
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

	store, userID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or storage unavailable", http.StatusUnauthorized)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	tasks, err := store.ListTasksForUser(userID, true) // Include archived
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
		writeJSONResponse(w, map[string]interface{}{
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

	store, userID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or storage unavailable", http.StatusUnauthorized)
		return
	}

	sessionID := r.URL.Query().Get("session_id")

	if sessionID != "" {
		// Export a single session
		messages, err := store.GetMessages(sessionID)
		if err != nil {
			http.Error(w, "failed to get session messages", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=chat_export.json")
		writeJSONResponse(w, map[string]interface{}{
			"exported_at": time.Now().Format(time.RFC3339),
			"session_id":  sessionID,
			"count":       len(messages),
			"messages":    messages,
		})
		return
	}

	// Export all sessions summary
	sessions, err := store.ListSessionsForUser(userID, nil, 0, 0)
	if err != nil {
		http.Error(w, "failed to list sessions", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=sessions_export.json")
	writeJSONResponse(w, map[string]interface{}{
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
			"models":     p.pc.Models,
			"configured": p.pc.APIKey != "" || p.pc.APIBase != "",
		}
	}
	return providers
}

func redactChannels(cfg *config.Config) map[string]interface{} {
	return map[string]interface{}{
		"telegram": map[string]interface{}{
			"enabled":    cfg.Channels.Telegram.Enabled,
			"configured": cfg.Channels.Telegram.Token != "",
			"allow_from": strings.Join(cfg.Channels.Telegram.AllowFrom, ","),
		},
		"discord": map[string]interface{}{
			"enabled":    cfg.Channels.Discord.Enabled,
			"configured": cfg.Channels.Discord.Token != "",
			"allow_from": strings.Join(cfg.Channels.Discord.AllowFrom, ","),
		},
		"slack": map[string]interface{}{
			"enabled":    cfg.Channels.Slack.Enabled,
			"configured": cfg.Channels.Slack.BotToken != "",
		},
		"whatsapp": map[string]interface{}{
			"enabled":    cfg.Channels.WhatsApp.Enabled,
			"configured": cfg.Channels.WhatsApp.BridgeURL != "",
			"bridge_url": cfg.Channels.WhatsApp.BridgeURL,
		},
		"feishu": map[string]interface{}{
			"enabled":    cfg.Channels.Feishu.Enabled,
			"configured": cfg.Channels.Feishu.AppID != "",
			"app_id":     cfg.Channels.Feishu.AppID,
		},
		"dingtalk": map[string]interface{}{
			"enabled":    cfg.Channels.DingTalk.Enabled,
			"configured": cfg.Channels.DingTalk.ClientID != "",
			"client_id":  cfg.Channels.DingTalk.ClientID,
		},
		"qq": map[string]interface{}{
			"enabled":    cfg.Channels.QQ.Enabled,
			"configured": cfg.Channels.QQ.AppID != "",
			"app_id":     cfg.Channels.QQ.AppID,
		},
		"maixcam": map[string]interface{}{
			"enabled": cfg.Channels.MaixCam.Enabled,
			"host":    cfg.Channels.MaixCam.Host,
			"port":    cfg.Channels.MaixCam.Port,
		},
		"signal": map[string]interface{}{
			"enabled":      cfg.Channels.Signal.Enabled,
			"configured":   cfg.Channels.Signal.PhoneNumber != "",
			"phone_number": cfg.Channels.Signal.PhoneNumber,
		},
		"email": map[string]interface{}{
			"enabled":                cfg.Channels.Email.Enabled,
			"configured":            cfg.Channels.Email.IMAPHost != "",
			"imap_host":             cfg.Channels.Email.IMAPHost,
			"imap_port":             cfg.Channels.Email.IMAPPort,
			"smtp_host":             cfg.Channels.Email.SMTPHost,
			"smtp_port":             cfg.Channels.Email.SMTPPort,
			"username":              cfg.Channels.Email.Username,
			"password":              map[string]interface{}{"has_password": cfg.Channels.Email.Password != ""},
			"from":                  cfg.Channels.Email.From,
			"allow_from":            strings.Join(cfg.Channels.Email.AllowFrom, ","),
			"mailbox":               cfg.Channels.Email.Mailbox,
			"poll_interval_seconds": cfg.Channels.Email.PollIntervalSecs,
			"mark_as_read":          cfg.Channels.Email.MarkAsRead,
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

	switch r.Method {
	case http.MethodGet:
		s.handleKnowledgeList(w, r)
	case http.MethodPost:
		s.handleKnowledgeUpload(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSONResponse(w, map[string]string{"error": "method not allowed"})
	}
}

// handleKnowledgeAction handles DELETE /api/v1/knowledge/{id}
func (s *Server) handleKnowledgeAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store, _, ok := s.getUserStorage(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONResponse(w, map[string]string{"error": "unauthorized"})
		return
	}

	// Extract ID from path: /api/v1/knowledge/{id}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{"error": "document id required"})
		return
	}

	var docID int64
	if _, err := fmt.Sscanf(pathParts[0], "%d", &docID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{"error": "invalid document id"})
		return
	}

	if r.Method == http.MethodDelete {
		if err := store.DeleteKnowledgeDocument(0, docID); err != nil {
			w.WriteHeader(http.StatusNotFound)
			writeJSONResponse(w, map[string]string{"error": "document not found"})
			return
		}
		writeJSONResponse(w, map[string]string{"status": "deleted"})
		return
	}

	if len(pathParts) > 1 && pathParts[1] == "chunks" && r.Method == http.MethodGet {
		chunks, err := store.GetKnowledgeDocumentChunks(0, docID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSONResponse(w, map[string]string{"error": "failed to get chunks: " + err.Error()})
			return
		}
		writeJSONResponse(w, map[string]any{"chunks": chunks})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	writeJSONResponse(w, map[string]string{"error": "method not allowed"})
}

// handleKnowledgeChunkAction handles PUT /api/v1/knowledge/chunks/{id}
func (s *Server) handleKnowledgeChunkAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store, _, ok := s.getUserStorage(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONResponse(w, map[string]string{"error": "unauthorized"})
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge/chunks/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{"error": "chunk id required"})
		return
	}

	var chunkID int64
	if _, err := fmt.Sscanf(pathParts[0], "%d", &chunkID); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{"error": "invalid chunk id"})
		return
	}

	if r.Method == http.MethodPut {
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Content) == "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSONResponse(w, map[string]string{"error": "valid chunk content required"})
			return
		}

		if err := store.UpdateKnowledgeChunk(0, chunkID, req.Content); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSONResponse(w, map[string]string{"error": "failed to update chunk: " + err.Error()})
			return
		}
		writeJSONResponse(w, map[string]string{"status": "updated"})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	writeJSONResponse(w, map[string]string{"error": "method not allowed"})
}

// handleKnowledgeSearch handles GET /api/v1/knowledge/search?q=...
func (s *Server) handleKnowledgeSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store, _, ok := s.getUserStorage(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONResponse(w, map[string]string{"error": "unauthorized"})
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{"error": "query parameter 'q' is required"})
		return
	}

	results, err := store.SearchKnowledge(0, query, 10)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]string{"error": err.Error()})
		return
	}

	writeJSONResponse(w, map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	})
}

func (s *Server) handleKnowledgeList(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONResponse(w, map[string]string{"error": "unauthorized"})
		return
	}

	docs, err := store.ListKnowledgeDocuments(0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]string{"error": err.Error()})
		return
	}
	if docs == nil {
		docs = []storage.KnowledgeDocument{}
	}
	writeJSONResponse(w, map[string]interface{}{"documents": docs})
}

func (s *Server) handleKnowledgeUpload(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONResponse(w, map[string]string{"error": "unauthorized"})
		return
	}

	// Limit upload to 25MB
	r.Body = http.MaxBytesReader(w, r.Body, 25<<20)
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{"error": "file too large (max 25MB)"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{"error": "file field 'file' is required"})
		return
	}
	defer file.Close()

	// Read the file content
	content := make([]byte, header.Size)
	n, err := file.Read(content)
	if err != nil && n == 0 {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]string{"error": "failed to read file"})
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
		writeJSONResponse(w, map[string]string{"error": "file contains no extractable text"})
		return
	}

	doc, err := store.SaveKnowledgeDocument(0, header.Filename, mimeType, header.Size, chunks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]string{"error": "failed to save document: " + err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, doc)
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
	switch r.Method {
	case http.MethodGet:
		// List all MCP servers
		w.Header().Set("Content-Type", "application/json")

		if s.mcpManager == nil {
			writeJSONResponse(w, map[string]interface{}{
				"servers": []interface{}{},
				"message": "MCP not configured",
			})
			return
		}

		servers := s.mcpManager.ServerStatus()
		writeJSONResponse(w, map[string]interface{}{
			"servers": servers,
		})

	case http.MethodPost:
		// Create new MCP server
		s.handleMCPCreate(w, r)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleMCPServerAction(w http.ResponseWriter, r *http.Request) {
	// Extract server name from path: /api/v1/mcp/{name} or /api/v1/mcp/{name}/action
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/mcp/")
	parts := strings.SplitN(path, "/", 2)
	serverName := parts[0]

	if serverName == "" {
		http.Error(w, "Server name required", http.StatusBadRequest)
		return
	}

	// Check for action suffix
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	// Route based on action and method
	switch {
	case action == "reconnect" && r.Method == http.MethodPost:
		s.handleMCPReconnect(w, r, serverName)

	case action == "test" && r.Method == http.MethodPost:
		s.handleMCPTest(w, r, serverName)

	case action == "" && r.Method == http.MethodGet:
		s.handleMCPServerInfo(w, r, serverName)

	case action == "" && r.Method == http.MethodPut:
		s.handleMCPUpdate(w, r, serverName)

	case action == "" && r.Method == http.MethodDelete:
		s.handleMCPDelete(w, r, serverName)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleMCPServerInfo(w http.ResponseWriter, r *http.Request, name string) {
	w.Header().Set("Content-Type", "application/json")

	if s.mcpManager == nil {
		http.Error(w, "MCP not configured", http.StatusNotFound)
		return
	}

	for _, srv := range s.mcpManager.ServerStatus() {
		if srv.Name == name {
			writeJSONResponse(w, srv)
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
		writeJSONResponse(w, map[string]interface{}{
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

	writeJSONResponse(w, map[string]interface{}{
		"ok":      true,
		"message": fmt.Sprintf("Reconnected to MCP server %q", name),
	})
}

// handleMCPReconnectAll handles POST /api/v1/mcp/reconnect-all - reconnects all MCP servers
func (s *Server) handleMCPReconnectAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.mcpManager == nil {
		writeJSONResponse(w, map[string]interface{}{
			"success":     false,
			"reconnected": []string{},
			"failed":      []string{},
			"error":       "MCP not configured",
		})
		return
	}

	servers := s.mcpManager.ServerStatus()
	if len(servers) == 0 {
		writeJSONResponse(w, map[string]interface{}{
			"success":     true,
			"reconnected": []string{},
			"failed":      []string{},
			"message":     "No MCP servers configured",
		})
		return
	}

	var reconnected []string
	var failed []string

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	for _, srv := range servers {
		if err := s.mcpManager.Reconnect(ctx, srv.Name); err != nil {
			failed = append(failed, srv.Name)
			logger.WarnC("mcp", "Failed to reconnect MCP server "+srv.Name+": "+err.Error())
		} else {
			reconnected = append(reconnected, srv.Name)
		}
	}

	// Re-register MCP tools on the agent loop
	if s.agentLoop != nil {
		for _, tool := range s.mcpManager.GetTools() {
			s.agentLoop.RegisterTool(tool)
		}
	}

	writeJSONResponse(w, map[string]interface{}{
		"success":     len(failed) == 0,
		"reconnected": reconnected,
		"failed":      failed,
	})
}

// handleMCPCreate handles POST /api/v1/mcp - adds a new MCP server
func (s *Server) handleMCPCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get user UUID for multi-user config
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Log the userUUID for debugging
	logger.InfoCF("mcp", "handleMCPCreate called", map[string]interface{}{
		"user_uuid":   userUUID,
		"legacy_mode": userUUID == "",
	})

	// Parse request body
	var req struct {
		Name    string            `json:"name"`
		Enabled bool              `json:"enabled"`
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
		SaveTo  string            `json:"save_to"` // "config" (default), "user_folder", "global_folder"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	// Validate name
	if !isValidMCPServerName(req.Name) {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "invalid server name: must be alphanumeric with hyphens/underscores only, 1-64 characters",
		})
		return
	}

	// Validate command
	if strings.TrimSpace(req.Command) == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "command is required",
		})
		return
	}

	// Create server config
	serverCfg := config.MCPServerConfig{
		Enabled: req.Enabled,
		Command: req.Command,
		Args:    req.Args,
		Env:     req.Env,
	}

	// Determine save location
	saveTo := req.SaveTo
	if saveTo == "" {
		saveTo = "config" // Default to config.json
	}

	var source mcp.MCPServerSource
	var savePath string
	var err error

	dataDir := config.GetDataDir()

	switch saveTo {
	case "user_folder":
		if userUUID == "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSONResponse(w, map[string]interface{}{
				"ok":    false,
				"error": "user_folder requires multi-user mode",
			})
			return
		}
		loader := mcp.NewLoader(dataDir)
		userMCPPath := filepath.Join(dataDir, "users", userUUID, "mcp")
		loader.SetUserMCPPath(userMCPPath)
		err = loader.SaveToFolder(userMCPPath, req.Name, serverCfg)
		source = mcp.SourceUserFolder
		savePath = filepath.Join(userMCPPath, req.Name, "mcp.json")

	case "global_folder":
		loader := mcp.NewLoader(dataDir)
		globalMCPPath := loader.GetGlobalMCPPath()
		err = loader.SaveToFolder(globalMCPPath, req.Name, serverCfg)
		source = mcp.SourceGlobalFolder
		savePath = filepath.Join(globalMCPPath, req.Name, "mcp.json")

	default: // "config"
		// Load config - use user-specific or global depending on mode
		var cfg *config.Config
		if userUUID != "" {
			cfg, err = config.LoadConfigForUser(userUUID)
			source = mcp.SourceUserConfig
			savePath = config.GetUserConfigPath(userUUID)
		} else {
			cfg, err = config.LoadConfig(config.GetGlobalConfigPath())
			source = mcp.SourceGlobalConfig
			savePath = config.GetGlobalConfigPath()
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSONResponse(w, map[string]interface{}{
				"ok":    false,
				"error": "failed to load config: " + err.Error(),
			})
			return
		}

		// Initialize MCP config if needed
		if cfg.Tools.MCP.Servers == nil {
			cfg.Tools.MCP.Servers = make(map[string]config.MCPServerConfig)
		}

		// Check if server already exists
		if _, exists := cfg.Tools.MCP.Servers[req.Name]; exists {
			w.WriteHeader(http.StatusConflict)
			writeJSONResponse(w, map[string]interface{}{
				"ok":    false,
				"error": "MCP server with this name already exists",
			})
			return
		}

		// Add to config
		cfg.Tools.MCP.Servers[req.Name] = serverCfg

		// Save config - use user-specific or global depending on mode
		if userUUID != "" {
			err = config.SaveConfigForUser(userUUID, cfg)
		} else {
			err = config.SaveConfig(config.GetGlobalConfigPath(), cfg)
		}
	}
	if err != nil {
		logger.ErrorCF("mcp", "Failed to save MCP config", map[string]interface{}{
			"user_uuid":   userUUID,
			"server_name": req.Name,
			"error":       err.Error(),
		})
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "failed to save config: " + err.Error(),
		})
		return
	}

	logger.InfoCF("mcp", "MCP server config saved successfully", map[string]interface{}{
		"user_uuid":   userUUID,
		"server_name": req.Name,
		"save_to":     saveTo,
		"source":      source,
		"path":        savePath,
	})

	// Add to running MCP manager if available
	if s.mcpManager != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := s.mcpManager.AddServerWithSource(ctx, req.Name, serverCfg, true, source, savePath); err != nil {
			// Log but don't fail - config is saved
			logger.WarnCF("mcp", "Failed to add server to running manager", map[string]interface{}{
				"name":  req.Name,
				"error": err.Error(),
			})
		}
	}

	writeJSONResponse(w, map[string]interface{}{
		"ok":     true,
		"source": source,
		"path":   savePath,
		"server": map[string]interface{}{
			"name":    req.Name,
			"enabled": serverCfg.Enabled,
			"command": serverCfg.Command,
			"args":    serverCfg.Args,
		},
	})
}

// handleMCPUpdate handles PUT /api/v1/mcp/{name} - updates an existing MCP server
func (s *Server) handleMCPUpdate(w http.ResponseWriter, r *http.Request, name string) {
	w.Header().Set("Content-Type", "application/json")

	// Get user UUID for multi-user config
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Enabled *bool             `json:"enabled"`
		Command string            `json:"command"`
		Args    []string          `json:"args"`
		Env     map[string]string `json:"env"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	// Load config - use user-specific or global depending on mode
	var cfg *config.Config
	var err error
	if userUUID != "" {
		cfg, err = config.LoadConfigForUser(userUUID)
	} else {
		cfg, err = config.LoadConfig(config.GetGlobalConfigPath())
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "failed to load config: " + err.Error(),
		})
		return
	}

	// Check if server exists
	existingCfg, exists := cfg.Tools.MCP.Servers[name]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "MCP server not found",
		})
		return
	}

	// Update fields
	if req.Enabled != nil {
		existingCfg.Enabled = *req.Enabled
	}
	if req.Command != "" {
		existingCfg.Command = req.Command
	}
	if req.Args != nil {
		existingCfg.Args = req.Args
	}
	if req.Env != nil {
		existingCfg.Env = req.Env
	}

	// Save updated config - use user-specific or global depending on mode
	cfg.Tools.MCP.Servers[name] = existingCfg
	if userUUID != "" {
		err = config.SaveConfigForUser(userUUID, cfg)
	} else {
		err = config.SaveConfig(config.GetGlobalConfigPath(), cfg)
	}
	if err != nil {
		logger.ErrorCF("mcp", "Failed to save updated MCP config", map[string]interface{}{
			"user_uuid":   userUUID,
			"server_name": name,
			"error":       err.Error(),
		})
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "failed to save config: " + err.Error(),
		})
		return
	}

	// Update running MCP manager if available
	if s.mcpManager != nil && s.mcpManager.HasServer(name) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		if err := s.mcpManager.UpdateServer(ctx, name, existingCfg); err != nil {
			logger.WarnCF("mcp", "Failed to update server in running manager", map[string]interface{}{
				"name":  name,
				"error": err.Error(),
			})
		}
	}

	writeJSONResponse(w, map[string]interface{}{
		"ok": true,
		"server": map[string]interface{}{
			"name":    name,
			"enabled": existingCfg.Enabled,
			"command": existingCfg.Command,
			"args":    existingCfg.Args,
		},
	})
}

// handleMCPDelete handles DELETE /api/v1/mcp/{name} - removes an MCP server
func (s *Server) handleMCPDelete(w http.ResponseWriter, r *http.Request, name string) {
	w.Header().Set("Content-Type", "application/json")

	// Get user UUID for multi-user config
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Load config - use user-specific or global depending on mode
	var cfg *config.Config
	var err error
	if userUUID != "" {
		cfg, err = config.LoadConfigForUser(userUUID)
	} else {
		cfg, err = config.LoadConfig(config.GetGlobalConfigPath())
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "failed to load config: " + err.Error(),
		})
		return
	}

	// Check if server exists
	if cfg.Tools.MCP.Servers == nil {
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "MCP server not found",
		})
		return
	}
	if _, exists := cfg.Tools.MCP.Servers[name]; !exists {
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "MCP server not found",
		})
		return
	}

	// Remove from config
	delete(cfg.Tools.MCP.Servers, name)

	// Save config - use user-specific or global depending on mode
	if userUUID != "" {
		err = config.SaveConfigForUser(userUUID, cfg)
	} else {
		err = config.SaveConfig(config.GetGlobalConfigPath(), cfg)
	}
	if err != nil {
		logger.ErrorCF("mcp", "Failed to save config after MCP delete", map[string]interface{}{
			"user_uuid":   userUUID,
			"server_name": name,
			"error":       err.Error(),
		})
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]interface{}{
			"ok":    false,
			"error": "failed to save config: " + err.Error(),
		})
		return
	}

	// Remove from running MCP manager if available
	if s.mcpManager != nil && s.mcpManager.HasServer(name) {
		if err := s.mcpManager.RemoveServer(name); err != nil {
			logger.WarnCF("mcp", "Failed to remove server from running manager", map[string]interface{}{
				"name":  name,
				"error": err.Error(),
			})
		}
	}

	writeJSONResponse(w, map[string]interface{}{
		"ok":      true,
		"message": fmt.Sprintf("MCP server %q removed", name),
	})
}

// handleMCPTest handles POST /api/v1/mcp/{name}/test - tests connection to an MCP server
func (s *Server) handleMCPTest(w http.ResponseWriter, r *http.Request, name string) {
	w.Header().Set("Content-Type", "application/json")

	// Get user UUID for multi-user config
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Load config - use user-specific or global depending on mode
	var cfg *config.Config
	var err error
	if userUUID != "" {
		cfg, err = config.LoadConfigForUser(userUUID)
	} else {
		cfg, err = config.LoadConfig(config.GetGlobalConfigPath())
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]interface{}{
			"success":     false,
			"tools_count": 0,
			"error":       "failed to load config: " + err.Error(),
		})
		return
	}

	// Get server config
	serverCfg, exists := cfg.Tools.MCP.Servers[name]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		writeJSONResponse(w, map[string]interface{}{
			"success":     false,
			"tools_count": 0,
			"error":       "MCP server not found",
		})
		return
	}

	// Create a temporary client for testing
	env := make([]string, 0, len(serverCfg.Env))
	for k, v := range serverCfg.Env {
		env = append(env, k+"="+v)
	}

	testClient := mcp.NewClient(name, serverCfg.Command, serverCfg.Args, env)
	defer testClient.Close()

	// Try to connect with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	if err := testClient.Connect(ctx); err != nil {
		writeJSONResponse(w, map[string]interface{}{
			"success":     false,
			"tools_count": 0,
			"error":       err.Error(),
		})
		return
	}

	// Success - return tool count
	toolCount := len(testClient.Tools())
	writeJSONResponse(w, map[string]interface{}{
		"success":     true,
		"tools_count": toolCount,
		"error":       "",
	})
}

// isValidMCPServerName validates MCP server names (alphanumeric + hyphens + underscores, 1-64 chars)
func isValidMCPServerName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// ==================== IMPORT CONVERSATIONS ====================

func (s *Server) handleImportChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or storage unavailable", http.StatusUnauthorized)
		return
	}

	// Limit body to 50MB
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)

	var payload struct {
		Format string          `json:"format"` // "makoclaw", "chatgpt", "claude", "auto"
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
		count, err := store.ImportMessages(sess.SessionID, sess.Messages)
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
	writeJSONResponse(w, map[string]interface{}{
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
	case "makoclaw":
		return parsemakoclawExport(data)
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
		// makoclaw: has "messages" and "session_id"
		if _, ok := obj["messages"]; ok {
			return "makoclaw"
		}
		// makoclaw all-sessions export: has "sessions"
		if _, ok := obj["sessions"]; ok {
			return "makoclaw"
		}
	}

	return "makoclaw" // default fallback
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

// Helper to update channels config from generic map
func updateChannelConfig(cfg *config.Config, updates map[string]interface{}) {
	for name, val := range updates {
		data, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		// Only update fields if they are provided
		if enabled, ok := data["enabled"].(bool); ok {
			switch name {
			case "telegram":
				cfg.Channels.Telegram.Enabled = enabled
			case "discord":
				cfg.Channels.Discord.Enabled = enabled
			case "whatsapp":
				cfg.Channels.WhatsApp.Enabled = enabled
			case "slack":
				cfg.Channels.Slack.Enabled = enabled
			case "feishu":
				cfg.Channels.Feishu.Enabled = enabled
			case "dingtalk":
				cfg.Channels.DingTalk.Enabled = enabled
			case "qq":
				cfg.Channels.QQ.Enabled = enabled
			case "signal":
				cfg.Channels.Signal.Enabled = enabled
			case "maixcam":
				cfg.Channels.MaixCam.Enabled = enabled
			}
		}

		switch name {
		case "telegram":
			if token, ok := data["token"].(string); ok && token != "" {
				cfg.Channels.Telegram.Token = token
			}
			if allow, ok := data["allow_from"].(string); ok {
				if allow == "" {
					cfg.Channels.Telegram.AllowFrom = []string{}
				} else {
					cfg.Channels.Telegram.AllowFrom = strings.Split(allow, ",")
				}
			}
		case "discord":
			if token, ok := data["token"].(string); ok && token != "" {
				cfg.Channels.Discord.Token = token
			}
			if allow, ok := data["allow_from"].(string); ok {
				if allow == "" {
					cfg.Channels.Discord.AllowFrom = []string{}
				} else {
					cfg.Channels.Discord.AllowFrom = strings.Split(allow, ",")
				}
			}
		case "whatsapp":
			if url, ok := data["bridge_url"].(string); ok && url != "" {
				cfg.Channels.WhatsApp.BridgeURL = url
			}
		case "slack":
			if botToken, ok := data["bot_token"].(string); ok && botToken != "" {
				cfg.Channels.Slack.BotToken = botToken
			}
		case "feishu":
			if appId, ok := data["app_id"].(string); ok && appId != "" {
				cfg.Channels.Feishu.AppID = appId
			}
			if appSecret, ok := data["app_secret"].(string); ok && appSecret != "" {
				cfg.Channels.Feishu.AppSecret = appSecret
			}
		case "dingtalk":
			if clientId, ok := data["client_id"].(string); ok && clientId != "" {
				cfg.Channels.DingTalk.ClientID = clientId
			}
			if clientSecret, ok := data["client_secret"].(string); ok && clientSecret != "" {
				cfg.Channels.DingTalk.ClientSecret = clientSecret
			}
		case "qq":
			if appId, ok := data["app_id"].(string); ok && appId != "" {
				cfg.Channels.QQ.AppID = appId
			}
			if appSecret, ok := data["app_secret"].(string); ok && appSecret != "" {
				cfg.Channels.QQ.AppSecret = appSecret
			}
		case "signal":
			if phone, ok := data["phone_number"].(string); ok && phone != "" {
				cfg.Channels.Signal.PhoneNumber = phone
			}
		case "maixcam":
			if host, ok := data["host"].(string); ok && host != "" {
				cfg.Channels.MaixCam.Host = host
			}
			if port, ok := data["port"].(float64); ok {
				cfg.Channels.MaixCam.Port = int(port)
			}
		}
	}
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

// parsemakoclawExport parses makoclaw's own export format.
// Single session: { session_id, messages: [...] }
// All sessions: { sessions: [ { session_id, ... } ] } — re-imports summary only (not useful), so we accept the single format.
func parsemakoclawExport(data json.RawMessage) ([]importSession, error) {
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
			sessionID = fmt.Sprintf("import:makoclaw:%d", time.Now().UnixMilli())
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
		sessionID := fmt.Sprintf("import:makoclaw:%d", time.Now().UnixMilli())
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

	return nil, fmt.Errorf("could not parse makoclaw format")
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
		writeJSONResponse(w, map[string]interface{}{"tools": []string{}})
		return
	}
	registry := s.agentLoop.ToolRegistry()
	if registry == nil {
		writeJSONResponse(w, map[string]interface{}{"tools": []string{}})
		return
	}
	names := registry.List()
	if names == nil {
		names = []string{}
	}
	writeJSONResponse(w, map[string]interface{}{"tools": names})
}

// ==================== WORKFLOWS ====================

// writeJSONError writes a consistent JSON error response with the given status code.
func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	writeJSONResponse(w, map[string]string{"error": msg})
}

// handleWorkflows handles GET (list) and POST (create) for workflows.
func (s *Server) handleWorkflows(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store, _, ok := s.getUserStorage(r)
	if !ok {
		writeJSONError(w, "unauthorized or storage unavailable", http.StatusUnauthorized)
		return
	}
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		workflows, err := store.ListWorkflows(userID)
		if err != nil {
			writeJSONError(w, "failed to list workflows: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if workflows == nil {
			workflows = []storage.Workflow{}
		}
		writeJSONResponse(w, map[string]interface{}{"workflows": workflows})

	case http.MethodPost:
		var body struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			Steps       json.RawMessage `json:"steps"`
			Parameters  json.RawMessage `json:"parameters"`
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
		id, err := store.CreateWorkflow(userID, body.Name, body.Description, body.Steps, body.Parameters, body.Schedule)
		if err != nil {
			writeJSONError(w, "failed to create workflow: "+err.Error(), http.StatusInternalServerError)
			return
		}
		wf, _ := store.GetWorkflow(id, userID)
		w.WriteHeader(http.StatusCreated)
		writeJSONResponse(w, wf)

	default:
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleWorkflowAction handles GET/PUT/DELETE /api/v1/workflows/{id} and POST /api/v1/workflows/{id}/run
func (s *Server) handleWorkflowAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store, _, ok := s.getUserStorage(r)
	if !ok {
		writeJSONError(w, "unauthorized or storage unavailable", http.StatusUnauthorized)
		return
	}
	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
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
		s.handleWorkflowRun(w, r, workflowID, userID, store)

	case action == "runs" && r.Method == http.MethodGet:
		runs, err := store.ListWorkflowRuns(workflowID, 20)
		if err != nil {
			writeJSONError(w, "failed to list runs: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if runs == nil {
			runs = []storage.WorkflowRun{}
		}
		writeJSONResponse(w, map[string]interface{}{"runs": runs})

	case action == "" && r.Method == http.MethodGet:
		wf, err := store.GetWorkflow(workflowID, userID)
		if err != nil {
			writeJSONError(w, "workflow not found", http.StatusNotFound)
			return
		}
		writeJSONResponse(w, wf)

	case action == "" && r.Method == http.MethodPut:
		var body struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			Enabled     *bool           `json:"enabled"`
			Steps       json.RawMessage `json:"steps"`
			Parameters  json.RawMessage `json:"parameters"`
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
			existing, err := store.GetWorkflow(workflowID, userID)
			if err == nil {
				enabled = existing.Enabled
			}
		}
		wf, err := store.UpdateWorkflow(workflowID, userID, body.Name, body.Description, enabled, body.Steps, body.Parameters, body.Schedule)
		if err != nil {
			writeJSONError(w, "failed to update workflow: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, wf)

	case action == "" && r.Method == http.MethodDelete:
		if err := store.DeleteWorkflow(workflowID, userID); err != nil {
			writeJSONError(w, "failed to delete: "+err.Error(), http.StatusNotFound)
			return
		}
		writeJSONResponse(w, map[string]string{"status": "deleted"})

	default:
		writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleWorkflowRun executes a workflow and returns the results.
// Note: The engine's Run() method internally creates and updates the workflow run
// record in the database, so no additional persistence is needed here.
func (s *Server) handleWorkflowRun(w http.ResponseWriter, r *http.Request, workflowID int64, userID int64, store *storage.Storage) {
	if s.workflowEngine == nil {
		writeJSONError(w, "workflow engine not available", http.StatusServiceUnavailable)
		return
	}

	// Parse runtime parameters from request body
	var reqBody struct {
		Parameters map[string]string `json:"parameters"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
	}

	wf, err := store.GetWorkflow(workflowID, userID)
	if err != nil {
		writeJSONError(w, "workflow not found", http.StatusNotFound)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	results, err := s.workflowEngine.RunWithParams(ctx, wf, reqBody.Parameters)
	if err != nil {
		writeJSONError(w, "execution failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]interface{}{
		"ok":      true,
		"results": results,
	})
}

// handleWorkflowApprovals lists pending workflow approvals.
func (s *Server) handleWorkflowApprovals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	store, _, ok := s.getUserStorage(r)
	if !ok {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	approvals, err := store.GetPendingApprovals()
	if err != nil {
		writeJSONError(w, "failed to get approvals: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, map[string]interface{}{"approvals": approvals})
}

// handleWorkflowApprovalAction resolves a workflow approval (approve/reject).
func (s *Server) handleWorkflowApprovalAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	store, _, ok := s.getUserStorage(r)
	if !ok {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract approval ID from path: /api/v1/workflows/approvals/{id}/resolve
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/approvals/")
	path = strings.TrimSuffix(path, "/resolve")
	approvalID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		writeJSONError(w, "invalid approval ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status     string `json:"status"` // "approved" or "rejected"
		ApprovedBy string `json:"approved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Status != "approved" && req.Status != "rejected" {
		writeJSONError(w, "status must be 'approved' or 'rejected'", http.StatusBadRequest)
		return
	}

	if err := store.ResolveWorkflowApproval(approvalID, req.Status, req.ApprovedBy); err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSONResponse(w, map[string]interface{}{"ok": true, "status": req.Status})
}

func updateProvidersConfig(cfg *config.Config, updates map[string]interface{}) {
	for name, val := range updates {
		data, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		var p *config.ProviderConfig
		switch name {
		case "anthropic":
			p = &cfg.Providers.Anthropic
		case "openai":
			p = &cfg.Providers.OpenAI
		case "openrouter":
			p = &cfg.Providers.OpenRouter
		case "groq":
			p = &cfg.Providers.Groq
		case "zhipu":
			p = &cfg.Providers.Zhipu
		case "vllm":
			p = &cfg.Providers.VLLM
		case "gemini":
			p = &cfg.Providers.Gemini
		case "nvidia":
			p = &cfg.Providers.Nvidia
		case "moonshot":
			p = &cfg.Providers.Moonshot
		case "ollama":
			p = &cfg.Providers.Ollama
		}

		if p != nil {
			if apiKey, ok := data["api_key"].(string); ok && apiKey != "" && !strings.Contains(apiKey, "****") {
				p.APIKey = apiKey
			}
			if apiBase, ok := data["api_base"].(string); ok {
				p.APIBase = apiBase
			}
			if models, ok := data["models"].([]interface{}); ok {
				p.Models = make([]string, 0, len(models))
				for _, m := range models {
					if ms, ok := m.(string); ok && ms != "" {
						p.Models = append(p.Models, ms)
					}
				}
			}
		}
	}
}

func updateAgentsConfig(cfg *config.Config, updates map[string]interface{}) {
	if defaults, ok := updates["defaults"].(map[string]interface{}); ok {
		if provider, ok := defaults["provider"].(string); ok {
			cfg.Agents.Defaults.Provider = provider
		}
		if model, ok := defaults["model"].(string); ok {
			cfg.Agents.Defaults.Model = model
		}
		if temp, ok := defaults["temperature"].(float64); ok {
			cfg.Agents.Defaults.Temperature = temp
		}
		if tokens, ok := defaults["max_tokens"].(float64); ok {
			cfg.Agents.Defaults.MaxTokens = int(tokens)
		}
		if iterations, ok := defaults["max_tool_iterations"].(float64); ok {
			cfg.Agents.Defaults.MaxToolIterations = int(iterations)
		}
	}

	if orchestrator, ok := updates["orchestrator"].(map[string]interface{}); ok {
		if enabled, ok := orchestrator["enabled"].(bool); ok {
			cfg.Agents.Orchestrator.Enabled = enabled
		}
		if provider, ok := orchestrator["provider"].(string); ok {
			cfg.Agents.Orchestrator.Provider = provider
		}
		if model, ok := orchestrator["model"].(string); ok {
			cfg.Agents.Orchestrator.Model = model
		}
		if temp, ok := orchestrator["temperature"].(float64); ok {
			cfg.Agents.Orchestrator.Temperature = temp
		}
		if tokens, ok := orchestrator["max_tokens"].(float64); ok {
			cfg.Agents.Orchestrator.MaxTokens = int(tokens)
		}
		if retries, ok := orchestrator["max_delegation_retries"].(float64); ok {
			cfg.Agents.Orchestrator.MaxDelegationRetries = int(retries)
		}
		if fallback, ok := orchestrator["fallback_to_default"].(bool); ok {
			cfg.Agents.Orchestrator.FallbackToDefault = fallback
		}
		if desc, ok := orchestrator["description"].(string); ok {
			cfg.Agents.Orchestrator.Description = desc
		}
	}
}

func updateToolsConfig(cfg *config.Config, updates map[string]interface{}) {
	if web, ok := updates["web"].(map[string]interface{}); ok {
		if search, ok := web["search"].(map[string]interface{}); ok {
			if apiKey, ok := search["api_key"].(string); ok && apiKey != "" && !strings.Contains(apiKey, "****") {
				cfg.Tools.Web.Search.APIKey = apiKey
			}
			if max, ok := search["max_results"].(float64); ok {
				cfg.Tools.Web.Search.MaxResults = int(max)
			}
		}
	}

	if email, ok := updates["email"].(map[string]interface{}); ok {
		if enabled, ok := email["enabled"].(bool); ok {
			cfg.Tools.Email.Enabled = enabled
		}
		if host, ok := email["host"].(string); ok {
			cfg.Tools.Email.Host = host
		}
		if port, ok := email["port"].(float64); ok {
			cfg.Tools.Email.Port = int(port)
		}
		if user, ok := email["username"].(string); ok {
			cfg.Tools.Email.Username = user
		}
		if pass, ok := email["password"].(string); ok && pass != "" && !strings.Contains(pass, "****") {
			cfg.Tools.Email.Password = pass
		}
		if from, ok := email["from"].(string); ok {
			cfg.Tools.Email.From = from
		}
		if to, ok := email["to"].(string); ok {
			cfg.Tools.Email.To = to
		}
	}
}

// getMapKeys returns the keys of a map as a string slice
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ==================== REPORTS EMAIL ====================

// handleSendReportEmail handles POST /api/v1/reports/email - sends an email report directly
func (s *Server) handleSendReportEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Get user UUID for multi-user config
	_, userUUID, ok := s.getUserStorage(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	// Parse request body
	var req struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"error":   "invalid request body: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Subject) == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"error":   "subject is required",
		})
		return
	}
	if strings.TrimSpace(req.Body) == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"error":   "body is required",
		})
		return
	}

	// Load user config (merged with global)
	userCfg, _ := config.LoadConfigForUser(userUUID)
	var emailCfg config.EmailToolsConfig
	if userCfg != nil && s.fullConfig != nil {
		merged := config.MergeConfigs(s.fullConfig, userCfg)
		emailCfg = merged.Tools.Email
	} else if s.fullConfig != nil {
		s.fullConfig.RLock()
		emailCfg = s.fullConfig.Tools.Email
		s.fullConfig.RUnlock()
	} else if userCfg != nil {
		emailCfg = userCfg.Tools.Email
	}

	// Check if email is enabled
	if !emailCfg.Enabled {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"error":   "email is not configured - please enable email in your settings",
		})
		return
	}

	// Check required SMTP settings
	if emailCfg.Host == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"error":   "SMTP host is not configured",
		})
		return
	}

	// Determine recipient
	to := strings.TrimSpace(req.To)
	if to == "" {
		to = emailCfg.To
	}
	if to == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"error":   "no recipient specified and no default recipient configured",
		})
		return
	}

	// Sanitize subject and recipient to prevent email header injection
	subject := strings.ReplaceAll(req.Subject, "\r", "")
	subject = strings.ReplaceAll(subject, "\n", "")
	to = strings.ReplaceAll(to, "\r", "")
	to = strings.ReplaceAll(to, "\n", "")

	// Validate recipient email format
	if _, err := mail.ParseAddress(to); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"error":   "invalid recipient email address: " + err.Error(),
		})
		return
	}

	// Determine from address
	from := emailCfg.From
	if from == "" {
		from = emailCfg.Username
	}
	envelopeFrom, fromHeader, err := parseReportFromAddress(from, emailCfg.Username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"error":   "invalid from address: " + err.Error(),
		})
		return
	}

	// Normalize password (Gmail app passwords may have spaces)
	password := emailCfg.Password
	if strings.EqualFold(strings.TrimSpace(emailCfg.Host), "smtp.gmail.com") {
		password = strings.ReplaceAll(password, " ", "")
	}

	// Construct email message
	msg := []byte("From: " + fromHeader + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		req.Body + "\r\n")

	auth := smtp.PlainAuth("", emailCfg.Username, password, emailCfg.Host)
	addr := fmt.Sprintf("%s:%d", emailCfg.Host, emailCfg.Port)

	// Send email
	if err := smtp.SendMail(addr, auth, envelopeFrom, []string{to}, msg); err != nil {
		logger.ErrorCF("web", "Failed to send report email", map[string]interface{}{
			"error": err.Error(),
			"host":  emailCfg.Host,
			"port":  emailCfg.Port,
			"to":    to,
		})
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONResponse(w, map[string]interface{}{
			"success": false,
			"error":   "failed to send email: " + err.Error(),
		})
		return
	}

	logger.InfoCF("web", "Report email sent successfully", map[string]interface{}{
		"to":      to,
		"subject": subject,
	})

	writeJSONResponse(w, map[string]interface{}{
		"success": true,
		"message": "Email sent successfully to " + to,
	})
}

// parseReportFromAddress parses and validates the from address
func parseReportFromAddress(from, fallback string) (envelope, header string, err error) {
	trimmedFrom := strings.TrimSpace(from)
	if trimmedFrom == "" {
		trimmedFrom = strings.TrimSpace(fallback)
	}
	if trimmedFrom == "" {
		return "", "", fmt.Errorf("from address is required")
	}

	parsed, parseErr := mail.ParseAddress(trimmedFrom)
	if parseErr == nil && parsed != nil {
		return parsed.Address, parsed.String(), nil
	}
	if strings.Contains(trimmedFrom, "<") || strings.Contains(trimmedFrom, ">") {
		return "", "", fmt.Errorf("invalid from address %q", from)
	}
	return trimmedFrom, trimmedFrom, nil
}

// handleAIFixJson uses AI to fix and validate JSON content
func (s *Server) handleAIFixJson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Content string `json:"content"`
		Type    string `json:"type"` // "cron_job", "config", "generic"
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(body.Content) == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}

	// Multi-user mode: create per-user agent loop
	if s.userMgr == nil || s.fullConfig == nil || s.msgBus == nil || s.centralStore == nil {
		http.Error(w, "AI not available - server not fully configured", http.StatusServiceUnavailable)
		return
	}

	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userStore, userUUID, userOK := s.getUserStorage(r)
	if !userOK || userUUID == "" || userStore == nil {
		http.Error(w, "user storage unavailable", http.StatusServiceUnavailable)
		return
	}

	userAgentLoop, err := s.newAgentLoopForUser(userUUID, userStore)
	if err != nil {
		logger.ErrorCF("web", "Failed to create agent loop for AI fix JSON", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		http.Error(w, "AI not available - "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer userAgentLoop.Stop()

	prompt := fmt.Sprintf(`You are a JSON validator and fixer.

The user has provided JSON for: %s

Your task:
1. Parse and validate the JSON
2. Fix any syntax errors (missing commas, quotes, brackets)
3. Fix any structural issues
4. Preserve all data and intent
5. Return ONLY the fixed JSON, no explanations

If the JSON is already valid, return it unchanged.

JSON to fix:
%s`, body.Type, body.Content)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	response, aiErr := userAgentLoop.ProcessDirectWithUserAndModel(
		ctx, userID, prompt,
		fmt.Sprintf("web:ai:fixjson:%s", body.Type), "", "*",
	)
	if aiErr != nil {
		http.Error(w, "AI processing failed: "+aiErr.Error(), http.StatusInternalServerError)
		return
	}

	// Extract JSON from response (may be in code blocks)
	fixedJson := extractJsonFromResponse(response)

	// Validate
	var validation interface{}
	if err := json.Unmarshal([]byte(fixedJson), &validation); err != nil {
		http.Error(w, "AI produced invalid JSON", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, map[string]interface{}{
		"fixed_json": fixedJson,
		"changes":    detectJsonChanges(body.Content, fixedJson),
		"valid":      true,
	})
}

// extractJsonFromResponse extracts JSON content from AI response, handling code blocks
func extractJsonFromResponse(response string) string {
	response = strings.TrimSpace(response)
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json")
		end := strings.LastIndex(response, "```")
		if start != -1 && end != -1 && end > start {
			response = response[start+7 : end]
		}
	} else if strings.Contains(response, "```") {
		start := strings.Index(response, "```")
		end := strings.LastIndex(response, "```")
		if start != -1 && end != -1 && end > start {
			response = response[start+3 : end]
		}
	}
	return strings.TrimSpace(response)
}

// detectJsonChanges provides a simple description of changes made
func detectJsonChanges(original, fixed string) string {
	if strings.TrimSpace(original) == strings.TrimSpace(fixed) {
		return ""
	}
	var origMap, fixedMap interface{}
	origErr := json.Unmarshal([]byte(original), &origMap)
	fixedErr := json.Unmarshal([]byte(fixed), &fixedMap)
	if origErr != nil && fixedErr == nil {
		return "Fixed syntax errors"
	}
	return "Applied fixes"
}

// handleAICreateCron generates cron jobs from natural language descriptions
func (s *Server) handleAICreateCron(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var body struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Multi-user mode: create per-user agent loop
	if s.userMgr == nil || s.fullConfig == nil || s.msgBus == nil || s.centralStore == nil {
		http.Error(w, "AI not available - server not fully configured", http.StatusServiceUnavailable)
		return
	}

	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userStore, userUUID, userOK := s.getUserStorage(r)
	if !userOK || userUUID == "" || userStore == nil {
		http.Error(w, "user storage unavailable", http.StatusServiceUnavailable)
		return
	}

	userAgentLoop, err := s.newAgentLoopForUser(userUUID, userStore)
	if err != nil {
		logger.ErrorCF("web", "Failed to create agent loop for AI cron create", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		http.Error(w, "AI not available - "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer userAgentLoop.Stop()

	aiPrompt := fmt.Sprintf(`You are a cron job creation assistant for MakoClaw.

The user wants to create a cron job with this description:
"%s"

Your task: Generate a complete cron job configuration.

Return format:
EXPLANATION: [Brief 1-sentence explanation]

JSON:
{
  "name": "descriptive name",
  "message": "what the agent should do",
  "job_type": "task" or "reminder",
  "channel": "telegram" (optional),
  "to": "chat_id" (optional),
  "schedule": {
    "kind": "cron" | "every" | "at",
    "expr": "0 9 * * *" (if cron),
    "everyMs": milliseconds (if every),
    "atMs": timestamp (if at),
    "tz": "America/New_York" (optional)
  }
}

Guidelines:
- Use "task" by default, "reminder" only if user wants direct message
- Daily: cron "0 9 * * *"
- Weekly: cron "0 9 * * 1" (Monday)
- Interval: every with everyMs
- One-time: at with atMs

Return only explanation and JSON.`, body.Prompt)

	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	response, aiErr := userAgentLoop.ProcessDirectWithUserAndModel(
		ctx, userID, aiPrompt, "web:ai:create-cron", "", "*",
	)
	if aiErr != nil {
		http.Error(w, "AI processing failed: "+aiErr.Error(), http.StatusInternalServerError)
		return
	}

	jsonStr := extractJsonFromResponse(response)
	var jobDraft map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(jsonStr), &jobDraft); jsonErr != nil {
		http.Error(w, "AI produced invalid cron job format", http.StatusInternalServerError)
		return
	}

	if jobDraft["name"] == nil || jobDraft["schedule"] == nil {
		http.Error(w, "AI generated incomplete cron job", http.StatusBadRequest)
		return
	}

	// Ensure message field exists (from payload or top-level)
	if jobDraft["message"] == nil {
		if payload, ok := jobDraft["payload"].(map[string]interface{}); ok {
			if payload["message"] == nil {
				http.Error(w, "AI generated cron job missing message", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "AI generated cron job missing message", http.StatusBadRequest)
			return
		}
	}

	writeJSONResponse(w, map[string]interface{}{
		"job":         jobDraft,
		"explanation": extractExplanation(response, jsonStr),
	})
}

// extractExplanation pulls the explanation text from the AI response
func extractExplanation(fullResponse, jsonPart string) string {
	before := strings.Split(fullResponse, jsonPart)[0]
	if strings.Contains(before, "EXPLANATION:") {
		parts := strings.Split(before, "EXPLANATION:")
		if len(parts) > 1 {
			explanation := strings.TrimSpace(parts[1])
			lines := strings.Split(explanation, "\n")
			if len(lines) > 0 {
				return strings.TrimSpace(lines[0])
			}
		}
	}
	return "AI-generated cron job"
}

func (s *Server) handleSkillGenerateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		writeJSONError(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	// Multi-user mode: create per-user agent loop
	if s.userMgr == nil || s.fullConfig == nil || s.msgBus == nil || s.centralStore == nil {
		writeJSONError(w, "AI not available - server not fully configured", http.StatusServiceUnavailable)
		return
	}

	userID, ok := s.getUserIDFromClaims(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userStore, userUUID, userOK := s.getUserStorage(r)
	if !userOK || userUUID == "" || userStore == nil {
		writeJSONError(w, "user storage unavailable", http.StatusServiceUnavailable)
		return
	}

	userAgentLoop, err := s.newAgentLoopForUser(userUUID, userStore)
	if err != nil {
		logger.ErrorCF("web", "Failed to create agent loop for skill config generation", map[string]interface{}{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		writeJSONError(w, "AI not available - "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer userAgentLoop.Stop()

	aiPrompt := fmt.Sprintf(`You are a skill configuration generator for MakoClaw.
Your task is to propose a skill name, description, and additional context based on the user's request.
A skill is a set of instructions (SKILL.md) that teaches the agent how to perform specific tasks.

Rules:
1. Name: Use a short, descriptive name with lowercase letters and hyphens (e.g., "github-helper", "system-monitor").
2. Description: A concise summary of what the skill does.
3. Additional Context: Detailed instructions or requirements for the skill generator.

YOU MUST RESPOND ONLY WITH A JSON OBJECT:
{
  "name": "proposed-skill-name",
  "description": "Short description",
  "prompt": "Proactive details for the SKILL.md template generation"
}

Generate a skill configuration for: %s`, req.Prompt)

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Use ProcessDirectWithUserAndModel without tools — we only need a JSON text response
	response, aiErr := userAgentLoop.ProcessDirectWithUserAndModel(
		ctx, userID, aiPrompt, "web:skills:generate-config", "", "*",
	)
	if aiErr != nil {
		writeJSONError(w, "AI generation failed: "+aiErr.Error(), http.StatusInternalServerError)
		return
	}

	// Try to extract JSON from the response
	content := extractJsonFromResponse(response)
	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(content), &result); jsonErr != nil {
		writeJSONError(w, "Failed to parse AI response as JSON", http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, result)
}

// handleSkillAnalytics returns per-user skill usage statistics for the past 30 days.
func (s *Server) handleSkillAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userStorage, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	stats, err := userStorage.GetSkillUsageStats(30)
	if err != nil {
		http.Error(w, "failed to fetch analytics", http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, map[string]interface{}{
		"stats": stats,
		"days":  30,
	})
}
