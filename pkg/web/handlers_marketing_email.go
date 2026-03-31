package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sipeed/makoclaw/pkg/config"
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

	emailCfg := s.resolveEmailConfig(userUUID)
	if !emailCfg.Enabled || emailCfg.Host == "" {
		http.Error(w, "SMTP not configured", http.StatusBadRequest)
		return
	}

	baseURL := fmt.Sprintf("http://%s", s.cfg.Host+":"+strconv.Itoa(s.cfg.Port))
	sender := marketing.NewSender(emailCfg, baseURL)

	campaign.Status = "sending"
	store.UpdateCampaign(campaign)

	contacts, err := store.GetListMembers(campaign.ListID)
	if err != nil {
		http.Error(w, "failed to load list members", http.StatusInternalServerError)
		return
	}

	sent := 0
	skipped := 0

	for _, contact := range contacts {
		if contact.Status != "active" {
			skipped++
			continue
		}

		contactFields := buildContactFields(&contact)
		html := marketing.Render(campaign.BodyHTML, contactFields)
		text := marketing.Render(campaign.BodyText, contactFields)

		token := sender.GenerateUnsubscribeToken(contact.ID, campaign.ID, unsubscribeSecret)
		unsubURL := baseURL + "/api/v1/marketing/unsubscribe/" + token
		html, text = marketing.AppendUnsubscribeFooter(html, text, unsubURL)

		headers := map[string]string{
			"List-Unsubscribe": "<" + unsubURL + ">",
		}

		sendErr := sender.Send(contact.Email, campaign.Subject, html, text, headers)

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
			Subject:         campaign.Subject,
			Status:          deliveryStatus,
			Error:           deliveryError,
		})
	}

	campaign.SentCount = sent
	campaign.SkippedCount = skipped
	campaign.Status = "sent"
	store.UpdateCampaign(campaign)

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"sent":    sent,
		"skipped": skipped,
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
