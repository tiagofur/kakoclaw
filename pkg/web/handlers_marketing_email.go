package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/marketing"
	"github.com/sipeed/makoclaw/pkg/storage"
)

const unsubscribeSecret = "makoclaw-marketing-secret-v1"

func (s *Server) audienceEmailCampaignsRouter(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.handleListEmailCampaigns(w, r)
		case http.MethodPost:
			s.handleCreateEmailCampaign(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	idStr := parts[1]
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}

	if len(parts) >= 3 && parts[2] == "send" {
		if r.Method == http.MethodPost {
			s.handleSendEmailCampaign(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(parts) >= 3 && parts[2] == "progress" {
		if r.Method == http.MethodGet {
			s.handleGetCampaignProgress(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(parts) >= 3 && parts[2] == "variants" {
		s.audienceVariantsRouter(w, r, parts)
		return
	}

	if len(parts) >= 3 && parts[2] == "memory" {
		s.audienceCampaignMemoryRouter(w, r, parts)
		return
	}

	if len(parts) >= 3 && parts[2] == "versions" {
		s.audienceCampaignVersionsRouter(w, r, parts)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetEmailCampaign(w, r)
	case http.MethodPut:
		s.handleUpdateEmailCampaign(w, r)
	case http.MethodDelete:
		s.handleDeleteEmailCampaign(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleListEmailCampaigns(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	campaigns, total, err := store.ListCampaigns(q.Get("account"), q.Get("status"), page, limit)
	if err != nil {
		http.Error(w, "failed to list campaigns", http.StatusInternalServerError)
		return
	}
	if campaigns == nil {
		campaigns = []storage.EmailCampaign{}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"campaigns": campaigns,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

func (s *Server) handleCreateEmailCampaign(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var c storage.EmailCampaign
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(c.Name) == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	created, err := store.CreateCampaign(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, created)
}

func (s *Server) handleGetEmailCampaign(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}

	campaign, err := store.GetCampaignByID(id)
	if err != nil {
		http.Error(w, "failed to get campaign", http.StatusInternalServerError)
		return
	}
	if campaign == nil {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, campaign)
}

func (s *Server) handleUpdateEmailCampaign(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}

	var c storage.EmailCampaign
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	c.ID = id

	if err := store.UpdateCampaign(&c); err != nil {
		http.Error(w, "failed to update campaign", http.StatusInternalServerError)
		return
	}

	updated, err := store.GetCampaignByID(id)
	if err != nil || updated == nil {
		http.Error(w, "failed to get updated campaign", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updated)
}

func (s *Server) handleDeleteEmailCampaign(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}

	if err := store.DeleteCampaign(id); err != nil {
		http.Error(w, "failed to delete campaign", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type deliveryProgress struct {
	Total     int    `json:"total"`
	Sent      int    `json:"sent"`
	Skipped   int    `json:"skipped"`
	Pct       int    `json:"pct"`
	StartedAt string `json:"started_at"`
}

func (s *Server) handleSendEmailCampaign(w http.ResponseWriter, r *http.Request) {
	store, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}

	campaign, err := store.GetCampaignByID(id)
	if err != nil {
		http.Error(w, "failed to get campaign", http.StatusInternalServerError)
		return
	}
	if campaign == nil {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	}
	if campaign.Status == "delivering" || campaign.Status == "sending" {
		http.Error(w, "campaign is already being delivered", http.StatusConflict)
		return
	}

	emailCfg := s.resolveEmailConfig(userUUID)
	if !emailCfg.Enabled || emailCfg.Host == "" {
		http.Error(w, "SMTP not configured", http.StatusBadRequest)
		return
	}

	contacts, err := store.GetListMembers(campaign.ListID)
	if err != nil {
		http.Error(w, "failed to load list members", http.StatusInternalServerError)
		return
	}

	// Mark as delivering immediately and return 202
	progress := deliveryProgress{
		Total:     len(contacts),
		StartedAt: time.Now().Format(time.RFC3339),
	}
	progressJSON, _ := json.Marshal(progress)
	campaign.Status = "delivering"
	campaign.DeliveryProgress = string(progressJSON)
	store.UpdateCampaign(campaign)
	_ = store.UpdateCampaignDeliveryProgress(campaign.ID, string(progressJSON))

	w.WriteHeader(http.StatusAccepted)
	writeJSONResponse(w, map[string]interface{}{
		"status": "delivering",
		"total":  len(contacts),
	})

	// Async delivery
	go func() {
		baseURL := fmt.Sprintf("http://%s", s.cfg.Host+":"+strconv.Itoa(s.cfg.Port))
		sender := marketing.NewSender(emailCfg, baseURL)
		variants, _ := store.ListVariants(campaign.ID)

		sent := 0
		skipped := 0
		total := len(contacts)

		for i, contact := range contacts {
			if contact.Status != "active" {
				skipped++
			} else {
				subject := campaign.Subject
				html := campaign.BodyHTML
				text := campaign.BodyText
				var variantID int64

				if len(variants) > 0 {
					bucket := contact.ID % 100
					cumulative := 0
					for _, v := range variants {
						cumulative += v.SplitPercent
						if int(bucket) < cumulative {
							variantID = v.ID
							if v.Subject != "" {
								subject = v.Subject
							}
							if v.BodyHTML != "" {
								html = v.BodyHTML
							}
							if v.BodyText != "" {
								text = v.BodyText
							}
							break
						}
					}
				}

				contactFields := buildContactFields(&contact)
				html = marketing.Render(html, contactFields)
				text = marketing.Render(text, contactFields)

				token := sender.GenerateUnsubscribeToken(contact.ID, campaign.ID, unsubscribeSecret)
				unsubURL := baseURL + "/api/v1/marketing/unsubscribe/" + token
				html, text = marketing.AppendUnsubscribeFooter(html, text, unsubURL)

				headers := map[string]string{"List-Unsubscribe": "<" + unsubURL + ">"}
				sendErr := sender.Send(contact.Email, subject, html, text, headers)

				deliveryStatus := "sent"
				deliveryError := ""
				if sendErr != nil {
					deliveryStatus = "failed"
					deliveryError = sendErr.Error()
					skipped++
				} else {
					sent++
				}

				store.CreateDelivery(&storage.EmailDelivery{
					CampaignAccount: campaign.Account,
					CampaignName:    campaign.Name,
					ContactID:       &contact.ID,
					VariantID:       variantID,
					Subject:         subject,
					Status:          deliveryStatus,
					Error:           deliveryError,
				})
			}

			// Update progress every 10 contacts or on last
			if i%10 == 0 || i == total-1 {
				pct := 0
				if total > 0 {
					pct = ((sent + skipped) * 100) / total
				}
				p := deliveryProgress{Total: total, Sent: sent, Skipped: skipped, Pct: pct, StartedAt: progress.StartedAt}
				if b, err := json.Marshal(p); err == nil {
					_ = store.UpdateCampaignDeliveryProgress(campaign.ID, string(b))
				}
			}
		}

		c, _ := store.GetCampaignByID(campaign.ID)
		if c != nil {
			c.SentCount = sent
			c.SkippedCount = skipped
			c.Status = "sent"
			finalProgress := deliveryProgress{Total: total, Sent: sent, Skipped: skipped, Pct: 100, StartedAt: progress.StartedAt}
			if b, err := json.Marshal(finalProgress); err == nil {
				c.DeliveryProgress = string(b)
			}
			if err := store.UpdateCampaign(c); err != nil {
				logger.ErrorCF("web", "failed to finalize campaign delivery", map[string]interface{}{
					"campaign_id": campaign.ID, "error": err.Error(),
				})
			}
		}
	}()
}

func (s *Server) handleGetCampaignProgress(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	campaign, err := store.GetCampaignByID(id)
	if err != nil || campaign == nil {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	}
	writeJSONResponse(w, map[string]interface{}{
		"status":   campaign.Status,
		"progress": campaign.DeliveryProgress,
	})
}

func (s *Server) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/api/v1/marketing/unsubscribe/")
	if token == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<html><body><h2>Invalid unsubscribe link</h2></body></html>`))
		return
	}

	sender := &marketing.SMTPSender{}
	contactID, _, err := sender.VerifyUnsubscribeToken(token, unsubscribeSecret)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<html><body><h2>Invalid or expired unsubscribe link</h2></body></html>`))
		return
	}

	store, _, ok := s.getUserStorage(r)
	if !ok {
		store = s.store
	}
	if store == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<html><body><h2>Error processing request</h2></body></html>`))
		return
	}

	contact, err := store.GetContactByID(contactID)
	if err != nil || contact == nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`<html><body><h2>Contact not found</h2></body></html>`))
		return
	}

	contact.Status = "unsubscribed"
	if err := store.UpdateContact(contact); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<html><body><h2>Error updating subscription</h2></body></html>`))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<html><body><h2>You've been unsubscribed</h2><p>You will no longer receive marketing emails from us.</p></body></html>`))
}

func (s *Server) resolveEmailConfig(userUUID string) config.EmailToolsConfig {
	if userUUID != "" {
		userCfg, _ := config.LoadConfigForUser(userUUID)
		if userCfg != nil && s.fullConfig != nil {
			merged := config.MergeConfigs(s.fullConfig, userCfg)
			return merged.Tools.Email
		}
		if userCfg != nil && userCfg.Tools.Email.Enabled && userCfg.Tools.Email.Host != "" {
			return userCfg.Tools.Email
		}
	}
	if s.fullConfig != nil {
		s.fullConfig.RLock()
		cfg := s.fullConfig.Tools.Email
		s.fullConfig.RUnlock()
		return cfg
	}
	return config.EmailToolsConfig{}
}

func buildContactFields(c *storage.Contact) map[string]string {
	return map[string]string{
		"email":      c.Email,
		"first_name": c.FirstName,
		"last_name":  c.LastName,
		"phone":      c.Phone,
		"company":    c.Company,
		"title":      c.Title,
		"tags":       c.Tags,
		"source":     c.Source,
		"status":     c.Status,
	}
}

func (s *Server) audienceVariantsRouter(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) < 3 {
		http.Error(w, "invalid variants path", http.StatusBadRequest)
		return
	}

	if len(parts) == 3 {
		switch r.Method {
		case http.MethodGet:
			s.handleListVariants(w, r)
		case http.MethodPost:
			s.handleCreateVariant(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	subParts := strings.SplitN(parts[3], "/", 2)
	_, err := strconv.ParseInt(subParts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid variant id", http.StatusBadRequest)
		return
	}

	if len(subParts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.handleGetVariant(w, r)
		case http.MethodPut:
			s.handleUpdateVariant(w, r)
		case http.MethodDelete:
			s.handleDeleteVariant(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	switch subParts[1] {
	case "metrics":
		s.handleVariantMetrics(w, r)
	case "winner":
		s.handleSetWinner(w, r)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (s *Server) handleListVariants(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}

	variants, err := store.ListVariants(campaignID)
	if err != nil {
		http.Error(w, "failed to list variants", http.StatusInternalServerError)
		return
	}
	if variants == nil {
		variants = []storage.ExperimentVariant{}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"variants": variants})
}

func (s *Server) handleCreateVariant(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}

	var v storage.ExperimentVariant
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	v.CampaignID = campaignID

	created, err := store.CreateVariant(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, created)
}

func (s *Server) handleGetVariant(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	variantID, ok := variantIDFromPath(r)
	if !ok {
		http.Error(w, "invalid variant id", http.StatusBadRequest)
		return
	}

	variant, err := store.GetVariantByID(variantID)
	if err != nil {
		http.Error(w, "failed to get variant", http.StatusInternalServerError)
		return
	}
	if variant == nil {
		http.Error(w, "variant not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, variant)
}

func (s *Server) handleUpdateVariant(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	variantID, ok := variantIDFromPath(r)
	if !ok {
		http.Error(w, "invalid variant id", http.StatusBadRequest)
		return
	}

	var v storage.ExperimentVariant
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	v.ID = variantID

	if err := store.UpdateVariant(&v); err != nil {
		http.Error(w, "failed to update variant", http.StatusInternalServerError)
		return
	}

	updated, err := store.GetVariantByID(variantID)
	if err != nil || updated == nil {
		http.Error(w, "failed to get updated variant", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updated)
}

func (s *Server) handleDeleteVariant(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	variantID, ok := variantIDFromPath(r)
	if !ok {
		http.Error(w, "invalid variant id", http.StatusBadRequest)
		return
	}

	if err := store.DeleteVariant(variantID); err != nil {
		http.Error(w, "failed to delete variant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleVariantMetrics(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	variantID, ok := variantIDFromPath(r)
	if !ok {
		http.Error(w, "invalid variant id", http.StatusBadRequest)
		return
	}

	sent, opened, clicked, err := store.GetVariantMetrics(variantID)
	if err != nil {
		http.Error(w, "failed to get variant metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"sent":    sent,
		"opened":  opened,
		"clicked": clicked,
	})
}

func (s *Server) handleSetWinner(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	variantID, ok := variantIDFromPath(r)
	if !ok {
		http.Error(w, "invalid variant id", http.StatusBadRequest)
		return
	}

	if err := store.SetWinner(variantID); err != nil {
		http.Error(w, "failed to set winner", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]string{"status": "ok"})
}

func variantIDFromPath(r *http.Request) (int64, bool) {
	suffix := strings.TrimPrefix(r.URL.Path, audiencePrefix)
	parts := strings.SplitN(suffix, "/", 4)
	if len(parts) < 4 {
		return 0, false
	}
	if parts[2] != "variants" {
		return 0, false
	}
	subParts := strings.SplitN(parts[3], "/", 2)
	id, err := strconv.ParseInt(subParts[0], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (s *Server) audienceAutomationsRouter(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.handleListAutomations(w, r)
		case http.MethodPost:
			s.handleCreateAutomation(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	idStr := parts[1]
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid automation id", http.StatusBadRequest)
		return
	}

	if len(parts) >= 3 && parts[2] == "runs" {
		if r.Method == http.MethodGet {
			s.handleListAutomationRuns(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetAutomation(w, r)
	case http.MethodPut:
		s.handleUpdateAutomation(w, r)
	case http.MethodDelete:
		s.handleDeleteAutomation(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleListAutomations(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	account := r.URL.Query().Get("account")
	automations, err := store.ListAutomations(account)
	if err != nil {
		http.Error(w, "failed to list automations", http.StatusInternalServerError)
		return
	}
	if automations == nil {
		automations = []storage.Automation{}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"automations": automations})
}

func (s *Server) handleCreateAutomation(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var a storage.Automation
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(a.Name) == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	created, err := store.CreateAutomation(&a)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, created)
}

func (s *Server) handleGetAutomation(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := automationIDFromPath(r)
	if !ok {
		http.Error(w, "invalid automation id", http.StatusBadRequest)
		return
	}

	automation, err := store.GetAutomationByID(id)
	if err != nil {
		http.Error(w, "failed to get automation", http.StatusInternalServerError)
		return
	}
	if automation == nil {
		http.Error(w, "automation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, automation)
}

func (s *Server) handleUpdateAutomation(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := automationIDFromPath(r)
	if !ok {
		http.Error(w, "invalid automation id", http.StatusBadRequest)
		return
	}

	var a storage.Automation
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	a.ID = id

	if err := store.UpdateAutomation(&a); err != nil {
		http.Error(w, "failed to update automation", http.StatusInternalServerError)
		return
	}

	updated, err := store.GetAutomationByID(id)
	if err != nil || updated == nil {
		http.Error(w, "failed to get updated automation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updated)
}

func (s *Server) handleDeleteAutomation(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := automationIDFromPath(r)
	if !ok {
		http.Error(w, "invalid automation id", http.StatusBadRequest)
		return
	}

	if err := store.DeleteAutomation(id); err != nil {
		http.Error(w, "failed to delete automation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListAutomationRuns(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	automationID, ok := automationIDFromPath(r)
	if !ok {
		http.Error(w, "invalid automation id", http.StatusBadRequest)
		return
	}

	runs, err := store.ListRunsByAutomation(automationID)
	if err != nil {
		http.Error(w, "failed to list automation runs", http.StatusInternalServerError)
		return
	}
	if runs == nil {
		runs = []storage.AutomationRun{}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"runs": runs})
}

func automationIDFromPath(r *http.Request) (int64, bool) {
	suffix := strings.TrimPrefix(r.URL.Path, audiencePrefix)
	parts := strings.SplitN(suffix, "/", 4)
	if len(parts) < 2 {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

// ── Campaign Memory ──────────────────────────────────────────────────────────

func (s *Server) audienceCampaignMemoryRouter(w http.ResponseWriter, r *http.Request, parts []string) {
	// parts[0]=campaignID, parts[1]=campaignID value, parts[2]="memory", parts[3]=optional entryID
	if len(parts) == 3 {
		switch r.Method {
		case http.MethodGet:
			s.handleListCampaignMemory(w, r)
		case http.MethodPost:
			s.handleAddCampaignMemoryEntry(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if len(parts) >= 4 && r.Method == http.MethodDelete {
		s.handleDeleteCampaignMemoryEntry(w, r, parts[3])
		return
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) handleListCampaignMemory(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	campaignID, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	entries, err := store.ListCampaignMemory(campaignID)
	if err != nil {
		http.Error(w, "failed to list campaign memory", http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []storage.CampaignMemoryEntry{}
	}
	writeJSONResponse(w, map[string]interface{}{"entries": entries})
}

func (s *Server) handleAddCampaignMemoryEntry(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	campaignID, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	var body struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Content) == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	entry, err := store.AddCampaignMemoryEntry(campaignID, body.Role, body.Content)
	if err != nil {
		http.Error(w, "failed to add memory entry", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, entry)
}

func (s *Server) handleDeleteCampaignMemoryEntry(w http.ResponseWriter, r *http.Request, entryIDStr string) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	entryID, err := strconv.ParseInt(entryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid entry id", http.StatusBadRequest)
		return
	}
	if err := store.DeleteCampaignMemoryEntry(entryID); err != nil {
		http.Error(w, "entry not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Campaign Versions ────────────────────────────────────────────────────────

func (s *Server) audienceCampaignVersionsRouter(w http.ResponseWriter, r *http.Request, parts []string) {
	// SplitN(..., 4) so parts = ["email-campaigns", campaignID, "versions", "rest/of/path"]
	// len(parts)==3 → collection endpoint
	// len(parts)==4 → parts[3] = "{versionID}" or "{versionID}/restore"
	if len(parts) == 3 {
		switch r.Method {
		case http.MethodGet:
			s.handleListCampaignVersions(w, r)
		case http.MethodPost:
			s.handleSaveCampaignVersion(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if len(parts) == 4 {
		subParts := strings.SplitN(parts[3], "/", 2)
		versionIDStr := subParts[0]
		subResource := ""
		if len(subParts) == 2 {
			subResource = subParts[1]
		}
		if subResource == "restore" && r.Method == http.MethodPost {
			s.handleRestoreCampaignVersion(w, r, versionIDStr)
			return
		}
	}

	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) handleListCampaignVersions(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	campaignID, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	versions, err := store.ListCampaignVersions(campaignID)
	if err != nil {
		http.Error(w, "failed to list versions", http.StatusInternalServerError)
		return
	}
	if versions == nil {
		versions = []storage.CampaignVersion{}
	}
	writeJSONResponse(w, map[string]interface{}{"versions": versions})
}

func (s *Server) handleSaveCampaignVersion(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	campaignID, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	campaign, err := store.GetCampaignByID(campaignID)
	if err != nil || campaign == nil {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	}
	note := body.Note
	if note == "" {
		note = "manual snapshot"
	}
	version, err := store.SaveCampaignVersion(campaignID, campaign.Subject, campaign.BodyHTML, campaign.BodyText, note)
	if err != nil {
		http.Error(w, "failed to save version", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, version)
}

func (s *Server) handleRestoreCampaignVersion(w http.ResponseWriter, r *http.Request, versionIDStr string) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	campaignID, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid campaign id", http.StatusBadRequest)
		return
	}
	versionID, err := strconv.ParseInt(versionIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid version id", http.StatusBadRequest)
		return
	}
	version, err := store.GetCampaignVersion(versionID)
	if err != nil || version == nil {
		http.Error(w, "version not found", http.StatusNotFound)
		return
	}
	campaign, err := store.GetCampaignByID(campaignID)
	if err != nil || campaign == nil {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	}
	campaign.Subject = version.Subject
	campaign.BodyHTML = version.BodyHTML
	campaign.BodyText = version.BodyText
	if err := store.UpdateCampaign(campaign); err != nil {
		http.Error(w, "failed to restore version", http.StatusInternalServerError)
		return
	}
	updated, _ := store.GetCampaignByID(campaignID)
	writeJSONResponse(w, updated)
}
