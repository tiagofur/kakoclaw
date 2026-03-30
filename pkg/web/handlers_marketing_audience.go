package web

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/sipeed/makoclaw/pkg/storage"
)

const audiencePrefix = "/api/v1/marketing/audience/"

func (s *Server) handleMarketingAudienceRouter(w http.ResponseWriter, r *http.Request) {
	suffix := strings.TrimPrefix(r.URL.Path, audiencePrefix)
	if suffix == r.URL.Path {
		http.Error(w, "invalid audience path", http.StatusBadRequest)
		return
	}
	parts := strings.SplitN(suffix, "/", 4)
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "invalid audience path", http.StatusBadRequest)
		return
	}

	resource := parts[0]
	switch resource {
	case "contacts":
		s.audienceContactsRouter(w, r, parts)
	case "lists":
		s.audienceListsRouter(w, r, parts)
	case "segments":
		s.audienceSegmentsRouter(w, r, parts)
	case "deliveries":
		if len(parts) == 1 {
			s.handleListDeliveries(w, r)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	case "send":
		if r.Method == http.MethodPost {
			s.handleSendToList(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

func (s *Server) audienceContactsRouter(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.handleListContacts(w, r)
		case http.MethodPost:
			s.handleCreateContact(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	sub := parts[1]
	if sub == "export" {
		s.handleExportContacts(w, r)
		return
	}
	if sub == "import" {
		s.handleImportContacts(w, r)
		return
	}

	idStr := sub
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetContact(w, r)
	case http.MethodPut:
		s.handleUpdateContact(w, r)
	case http.MethodDelete:
		s.handleDeleteContact(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) audienceListsRouter(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.handleListLists(w, r)
		case http.MethodPost:
			s.handleCreateList(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	idStr := parts[1]
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid list id", http.StatusBadRequest)
		return
	}

	if len(parts) >= 3 && parts[2] == "members" {
		if len(parts) == 4 && parts[3] != "" {
			switch r.Method {
			case http.MethodDelete:
				s.handleRemoveListMember(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		switch r.Method {
		case http.MethodPost:
			s.handleAddListMember(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetList(w, r)
	case http.MethodPut:
		s.handleUpdateList(w, r)
	case http.MethodDelete:
		s.handleDeleteList(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) audienceSegmentsRouter(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.handleListSegments(w, r)
		case http.MethodPost:
			s.handleCreateSegment(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	idStr := parts[1]
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid segment id", http.StatusBadRequest)
		return
	}

	if len(parts) >= 3 && parts[2] == "preview" {
		s.handlePreviewSegment(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetSegment(w, r)
	case http.MethodPut:
		s.handleUpdateSegment(w, r)
	case http.MethodDelete:
		s.handleDeleteSegment(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleListContacts(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := storage.ContactFilter{
		Q:      q.Get("q"),
		Status: q.Get("status"),
		Tags:   q.Get("tags"),
		Page:   page,
		Limit:  limit,
	}

	contacts, total, err := store.ListContacts(filter)
	if err != nil {
		http.Error(w, "failed to list contacts", http.StatusInternalServerError)
		return
	}
	if contacts == nil {
		contacts = []storage.Contact{}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"contacts": contacts,
		"total":    total,
		"page":     filter.Page,
		"limit":    filter.Limit,
	})
}

func (s *Server) handleGetContact(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}

	contact, err := store.GetContactByID(id)
	if err != nil {
		http.Error(w, "failed to get contact", http.StatusInternalServerError)
		return
	}
	if contact == nil {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, contact)
}

func (s *Server) handleCreateContact(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var c storage.Contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(c.Email) == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	created, err := store.CreateContact(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, created)
}

func (s *Server) handleUpdateContact(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}

	var c storage.Contact
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	c.ID = id

	if err := store.UpdateContact(&c); err != nil {
		http.Error(w, "failed to update contact", http.StatusInternalServerError)
		return
	}

	updated, err := store.GetContactByID(id)
	if err != nil || updated == nil {
		http.Error(w, "failed to get updated contact", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updated)
}

func (s *Server) handleDeleteContact(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}

	if err := store.DeleteContact(id); err != nil {
		http.Error(w, "failed to delete contact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleImportContacts(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	const maxUploadSize = 25 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large or invalid form (max 25MB)", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing 'file' field in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	if err != nil {
		http.Error(w, "failed to read CSV", http.StatusBadRequest)
		return
	}
	if len(records) == 0 {
		http.Error(w, "CSV file is empty", http.StatusBadRequest)
		return
	}

	headers := records[0]
	dataRows := records[1:]

	inserted, skipped, err := store.ImportContactsFromCSV(dataRows, headers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"inserted": inserted,
		"skipped":  skipped,
	})
}

func (s *Server) handleExportContacts(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	filter := storage.ContactFilter{
		Q:      q.Get("q"),
		Status: q.Get("status"),
		Tags:   q.Get("tags"),
		Page:   page,
		Limit:  limit,
	}

	data, err := store.ExportContactsCSV(filter)
	if err != nil {
		http.Error(w, "failed to export contacts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=contacts.csv")
	w.Write(data)
}

func (s *Server) handleListLists(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	account := r.URL.Query().Get("account")
	lists, err := store.ListContactLists(account)
	if err != nil {
		http.Error(w, "failed to list contact lists", http.StatusInternalServerError)
		return
	}
	if lists == nil {
		lists = []storage.ContactList{}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"lists": lists})
}

func (s *Server) handleGetList(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid list id", http.StatusBadRequest)
		return
	}

	list, err := store.GetContactListByID(id)
	if err != nil {
		http.Error(w, "failed to get contact list", http.StatusInternalServerError)
		return
	}
	if list == nil {
		http.Error(w, "contact list not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, list)
}

func (s *Server) handleCreateList(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var l storage.ContactList
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(l.Name) == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	created, err := store.CreateContactList(&l)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, created)
}

func (s *Server) handleUpdateList(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid list id", http.StatusBadRequest)
		return
	}

	var l storage.ContactList
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	l.ID = id

	if err := store.UpdateContactList(&l); err != nil {
		http.Error(w, "failed to update contact list", http.StatusInternalServerError)
		return
	}

	updated, err := store.GetContactListByID(id)
	if err != nil || updated == nil {
		http.Error(w, "failed to get updated contact list", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updated)
}

func (s *Server) handleDeleteList(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid list id", http.StatusBadRequest)
		return
	}

	if err := store.DeleteContactList(id); err != nil {
		http.Error(w, "failed to delete contact list", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAddListMember(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	listID, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid list id", http.StatusBadRequest)
		return
	}

	var req struct {
		ContactID int64 `json:"contact_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.ContactID == 0 {
		http.Error(w, "contact_id is required", http.StatusBadRequest)
		return
	}

	if err := store.AddListMember(req.ContactID, listID); err != nil {
		http.Error(w, "failed to add list member", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]string{"status": "ok"})
}

func (s *Server) handleRemoveListMember(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	suffix := strings.TrimPrefix(r.URL.Path, audiencePrefix+"lists/")
	pathParts := strings.SplitN(suffix, "/", 4)
	if len(pathParts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	listID, err := strconv.ParseInt(pathParts[0], 10, 64)
	if err != nil {
		http.Error(w, "invalid list id", http.StatusBadRequest)
		return
	}
	contactID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		http.Error(w, "invalid contact id", http.StatusBadRequest)
		return
	}

	if err := store.RemoveListMember(contactID, listID); err != nil {
		http.Error(w, "failed to remove list member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListSegments(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	account := r.URL.Query().Get("account")
	segments, err := store.ListSegments(account)
	if err != nil {
		http.Error(w, "failed to list segments", http.StatusInternalServerError)
		return
	}
	if segments == nil {
		segments = []storage.Segment{}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{"segments": segments})
}

func (s *Server) handleGetSegment(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid segment id", http.StatusBadRequest)
		return
	}

	segment, err := store.GetSegmentByID(id)
	if err != nil {
		http.Error(w, "failed to get segment", http.StatusInternalServerError)
		return
	}
	if segment == nil {
		http.Error(w, "segment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, segment)
}

func (s *Server) handleCreateSegment(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var seg storage.Segment
	if err := json.NewDecoder(r.Body).Decode(&seg); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(seg.Name) == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	created, err := store.CreateSegment(&seg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeJSONResponse(w, created)
}

func (s *Server) handleUpdateSegment(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid segment id", http.StatusBadRequest)
		return
	}

	var seg storage.Segment
	if err := json.NewDecoder(r.Body).Decode(&seg); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	seg.ID = id

	if err := store.UpdateSegment(&seg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := store.GetSegmentByID(id)
	if err != nil || updated == nil {
		http.Error(w, "failed to get updated segment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, updated)
}

func (s *Server) handleDeleteSegment(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, ok := audienceIDFromPath(r)
	if !ok {
		http.Error(w, "invalid segment id", http.StatusBadRequest)
		return
	}

	if err := store.DeleteSegment(id); err != nil {
		http.Error(w, "failed to delete segment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handlePreviewSegment(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Rules string `json:"rules"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Rules == "" {
		req.Rules = "[]"
	}

	var rules []storage.SegmentRule
	if err := json.Unmarshal([]byte(req.Rules), &rules); err != nil {
		http.Error(w, "invalid rules format", http.StatusBadRequest)
		return
	}

	contactIDs, err := store.EvaluateSegmentRules(rules)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"contact_count": len(contactIDs),
	})
}

func (s *Server) handleListDeliveries(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	deliveries, total, err := store.ListDeliveries(q.Get("account"), q.Get("campaign"), page, limit)
	if err != nil {
		http.Error(w, "failed to list deliveries", http.StatusInternalServerError)
		return
	}
	if deliveries == nil {
		deliveries = []storage.EmailDelivery{}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"deliveries": deliveries,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

func (s *Server) handleSendToList(w http.ResponseWriter, r *http.Request) {
	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ListID          int64  `json:"list_id"`
		CampaignAccount string `json:"campaign_account"`
		CampaignName    string `json:"campaign_name"`
		Subject         string `json:"subject"`
		Body            string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.ListID == 0 {
		http.Error(w, "list_id is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Subject) == "" {
		http.Error(w, "subject is required", http.StatusBadRequest)
		return
	}

	sendFunc := func(to, subject, body string) error {
		// TODO: integrate with per-user email provider config
		return nil
	}

	sent, skipped, err := store.SendToList(req.ListID, req.CampaignAccount, req.CampaignName, req.Subject, req.Body, sendFunc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, map[string]interface{}{
		"sent":    sent,
		"skipped": skipped,
	})
}

func audienceIDFromPath(r *http.Request) (int64, bool) {
	suffix := strings.TrimPrefix(r.URL.Path, audiencePrefix)
	parts := strings.SplitN(suffix, "/", 4)
	if len(parts) < 2 {
		return 0, false
	}
	idStr := parts[1]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
