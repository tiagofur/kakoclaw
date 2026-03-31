package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
)

func (s *Server) handleMarketingProfiles(w http.ResponseWriter, r *http.Request) {
	// Route: /api/v1/marketing/profiles  and  /api/v1/marketing/profiles/{account}[/research]
	const prefix = "/api/v1/marketing/profiles"
	suffix := strings.TrimPrefix(r.URL.Path, prefix)
	suffix = strings.TrimPrefix(suffix, "/")

	if suffix == "" {
		// Collection endpoint: GET list
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleListCompanyProfiles(w, r)
		return
	}

	parts := strings.SplitN(suffix, "/", 2)
	account := parts[0]
	if account == "" {
		http.Error(w, "account name required", http.StatusBadRequest)
		return
	}

	if len(parts) == 2 && parts[1] == "research" {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleResearchCompany(w, r, account)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetCompanyProfile(w, r, account)
	case http.MethodPut:
		s.handleUpsertCompanyProfile(w, r, account)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleListCompanyProfiles(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	profiles, err := store.ListCompanyProfiles()
	if err != nil {
		http.Error(w, "failed to list profiles", http.StatusInternalServerError)
		return
	}
	if profiles == nil {
		profiles = []storage.CompanyProfile{}
	}
	writeJSONResponse(w, map[string]interface{}{"profiles": profiles})
}

func (s *Server) handleGetCompanyProfile(w http.ResponseWriter, r *http.Request, account string) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	profile, err := store.GetCompanyProfile(account)
	if err != nil {
		http.Error(w, "failed to get profile", http.StatusInternalServerError)
		return
	}
	if profile == nil {
		// Return an empty profile shell so the frontend can render the form
		writeJSONResponse(w, &storage.CompanyProfile{Account: account, SocialLinks: "{}"})
		return
	}
	writeJSONResponse(w, profile)
}

func (s *Server) handleUpsertCompanyProfile(w http.ResponseWriter, r *http.Request, account string) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var p storage.CompanyProfile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	p.Account = account
	saved, err := store.UpsertCompanyProfile(&p)
	if err != nil {
		http.Error(w, "failed to save profile", http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, saved)
}

func (s *Server) handleResearchCompany(w http.ResponseWriter, r *http.Request, account string) {
	store, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Read optional hint: { "website": "..." }
	var hint struct {
		Website string `json:"website"`
	}
	_ = json.NewDecoder(r.Body).Decode(&hint)

	userAgentLoop, err := s.newAgentLoopForUser(userUUID, store)
	if err != nil {
		http.Error(w, "failed to create agent: "+err.Error(), http.StatusInternalServerError)
		return
	}

	websiteHint := ""
	if hint.Website != "" {
		websiteHint = fmt.Sprintf(" Their website is %s.", hint.Website)
	}

	prompt := fmt.Sprintf(`Research the company or brand "%s".%s

Use web_search and web_fetch to find:
1. What they do (industry, product/service description)
2. Their target audience
3. Their main website URL
4. Social media presence (Instagram, Twitter/X, LinkedIn, TikTok, YouTube)
5. Any recent notable news or campaigns

Respond ONLY with a JSON object — no markdown, no extra text:
{
  "website": "...",
  "industry": "...",
  "description": "...",
  "target_audience": "...",
  "social_links": {"instagram": "...", "twitter": "...", "linkedin": "..."},
  "research_notes": "2-3 sentence summary of key findings"
}`, account, websiteHint)

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	sessionKey := "marketing_research_" + account
	result, err := userAgentLoop.ProcessDirect(ctx, prompt, sessionKey)
	if err != nil {
		logger.ErrorCF("web", "company research failed", map[string]interface{}{
			"account": account, "error": err.Error(),
		})
		http.Error(w, "research failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Parse LLM JSON output
	var parsed struct {
		Website        string            `json:"website"`
		Industry       string            `json:"industry"`
		Description    string            `json:"description"`
		TargetAudience string            `json:"target_audience"`
		SocialLinks    map[string]string `json:"social_links"`
		ResearchNotes  string            `json:"research_notes"`
	}

	// Strip any markdown code fences the LLM might have added
	clean := strings.TrimSpace(result)
	if idx := strings.Index(clean, "{"); idx > 0 {
		clean = clean[idx:]
	}
	if idx := strings.LastIndex(clean, "}"); idx >= 0 && idx < len(clean)-1 {
		clean = clean[:idx+1]
	}

	socialLinksJSON := "{}"
	if err := json.Unmarshal([]byte(clean), &parsed); err == nil {
		if parsed.SocialLinks != nil {
			if b, err := json.Marshal(parsed.SocialLinks); err == nil {
				socialLinksJSON = string(b)
			}
		}
	}

	now := time.Now()
	profile := &storage.CompanyProfile{
		Account:        account,
		Website:        parsed.Website,
		Industry:       parsed.Industry,
		Description:    parsed.Description,
		TargetAudience: parsed.TargetAudience,
		SocialLinks:    socialLinksJSON,
		ResearchNotes:  parsed.ResearchNotes,
		ResearchedAt:   &now,
	}
	// Fall back to raw result if parsing failed
	if profile.ResearchNotes == "" {
		profile.ResearchNotes = result
	}

	saved, err := store.UpsertCompanyProfile(profile)
	if err != nil {
		http.Error(w, "failed to save profile", http.StatusInternalServerError)
		return
	}
	writeJSONResponse(w, saved)
}
