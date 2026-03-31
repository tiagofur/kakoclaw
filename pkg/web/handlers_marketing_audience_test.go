package web

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sipeed/makoclaw/pkg/storage"
)

// --- Contacts CRUD ---

func TestHandleAudienceContacts_CRUD(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	// Create contact
	body := `{"email":"alice@example.com","first_name":"Alice","last_name":"Smith","status":"active","tags":"vip,early-adopter"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/contacts", strings.NewReader(body))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var created storage.Contact
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("invalid create response: %v", err)
	}
	if created.Email != "alice@example.com" {
		t.Fatalf("email = %q, want alice@example.com", created.Email)
	}
	if created.FirstName != "Alice" {
		t.Fatalf("first_name = %q, want Alice", created.FirstName)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	// List contacts
	req = httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/contacts", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var listResp struct {
		Contacts []storage.Contact `json:"contacts"`
		Total    int               `json:"total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("invalid list response: %v", err)
	}
	if listResp.Total != 1 {
		t.Fatalf("total = %d, want 1", listResp.Total)
	}

	// Get contact by ID
	req = httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/contacts/1", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Update contact
	updateBody := `{"first_name":"Alicia","last_name":"Johnson","status":"inactive"}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/marketing/audience/contacts/1", strings.NewReader(updateBody))
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var updated storage.Contact
	if err := json.Unmarshal(rr.Body.Bytes(), &updated); err != nil {
		t.Fatalf("invalid update response: %v", err)
	}
	if updated.FirstName != "Alicia" {
		t.Fatalf("first_name after update = %q, want Alicia", updated.FirstName)
	}

	// Delete contact
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/marketing/audience/contacts/1", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify deleted
	req = httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/contacts", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	var afterDelete struct {
		Total int `json:"total"`
	}
	json.Unmarshal(rr.Body.Bytes(), &afterDelete)
	if afterDelete.Total != 0 {
		t.Fatalf("total after delete = %d, want 0", afterDelete.Total)
	}
}

func TestHandleAudienceContacts_CreateRequiresEmail(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	body := `{"first_name":"NoEmail"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/contacts", strings.NewReader(body))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing email, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAudienceContacts_AuthRequired(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/contacts", nil)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestHandleAudienceContacts_GetNotFound(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/contacts/999", nil)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAudienceContacts_InvalidID(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/contacts/abc", nil)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// --- CSV Import/Export ---

func TestHandleAudienceContacts_ImportExport(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	// Import CSV
	csvData := "email,first_name,last_name,status\nalice@test.com,Alice,A,active\nbob@test.com,Bob,B,active\n"

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "contacts.csv")
	io.Copy(part, strings.NewReader(csvData))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/contacts/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("import: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var importResp struct {
		Inserted int `json:"inserted"`
		Skipped  int `json:"skipped"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &importResp); err != nil {
		t.Fatalf("invalid import response: %v", err)
	}
	if importResp.Inserted != 2 {
		t.Fatalf("inserted = %d, want 2", importResp.Inserted)
	}

	// Export CSV
	req = httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/contacts/export", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("export: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	if ct := rr.Header().Get("Content-Type"); ct != "text/csv" {
		t.Fatalf("Content-Type = %q, want text/csv", ct)
	}

	reader := csv.NewReader(strings.NewReader(rr.Body.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse exported CSV: %v", err)
	}
	// Header + 2 data rows
	if len(records) < 3 {
		t.Fatalf("expected at least 3 rows (header + 2 contacts), got %d", len(records))
	}
}

// --- Contact Lists CRUD ---

func TestHandleAudienceLists_CRUD(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	// Create list
	body := `{"name":"VIP Customers","description":"Top customers","account":"acme"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/lists", strings.NewReader(body))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("create list: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var createdList storage.ContactList
	if err := json.Unmarshal(rr.Body.Bytes(), &createdList); err != nil {
		t.Fatalf("invalid create response: %v", err)
	}
	if createdList.Name != "VIP Customers" {
		t.Fatalf("name = %q, want 'VIP Customers'", createdList.Name)
	}

	// List all lists
	req = httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/lists?account=acme", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", rr.Code)
	}

	var listsResp struct {
		Lists []storage.ContactList `json:"lists"`
	}
	json.Unmarshal(rr.Body.Bytes(), &listsResp)
	if len(listsResp.Lists) != 1 {
		t.Fatalf("lists count = %d, want 1", len(listsResp.Lists))
	}

	// Get list by ID
	req = httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/lists/1", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("get list: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Update list
	updateBody := `{"name":"Premium Customers","description":"Updated"}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/marketing/audience/lists/1", strings.NewReader(updateBody))
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("update list: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Delete list
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/marketing/audience/lists/1", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete list: expected 204, got %d", rr.Code)
	}
}

func TestHandleAudienceLists_CreateRequiresName(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	body := `{"description":"no name"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/lists", strings.NewReader(body))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// --- List Members ---

func TestHandleAudienceListMembers(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	// Create contact
	contactBody := `{"email":"member@test.com","first_name":"Member"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/contacts", strings.NewReader(contactBody))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create contact: expected 201, got %d", rr.Code)
	}

	// Create list
	listBody := `{"name":"Test List","account":"acme"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/lists", strings.NewReader(listBody))
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create list: expected 201, got %d", rr.Code)
	}

	// Add member to list
	addBody := `{"contact_id":1}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/lists/1/members", strings.NewReader(addBody))
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("add member: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Remove member from list
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/marketing/audience/lists/1/members/1", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("remove member: expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

// --- Segments CRUD ---

func TestHandleAudienceSegments_CRUD(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	// Create segment
	rules := `[{"field":"status","operator":"equals","value":"active"}]`
	segBody, _ := json.Marshal(map[string]interface{}{
		"name":        "Active Users",
		"description": "All active contacts",
		"account":     "acme",
		"rules":       rules,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/segments", bytes.NewReader(segBody))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("create segment: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var createdSeg storage.Segment
	if err := json.Unmarshal(rr.Body.Bytes(), &createdSeg); err != nil {
		t.Fatalf("invalid create response: %v", err)
	}
	if createdSeg.Name != "Active Users" {
		t.Fatalf("name = %q, want 'Active Users'", createdSeg.Name)
	}

	// List all segments
	req = httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/segments?account=acme", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("list segments: expected 200, got %d", rr.Code)
	}

	var segResp struct {
		Segments []storage.Segment `json:"segments"`
	}
	json.Unmarshal(rr.Body.Bytes(), &segResp)
	if len(segResp.Segments) != 1 {
		t.Fatalf("segments count = %d, want 1", len(segResp.Segments))
	}

	// Get segment
	req = httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/segments/1", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("get segment: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Update segment
	updateSeg, _ := json.Marshal(map[string]interface{}{
		"name":        "Active VIPs",
		"description": "Updated",
	})
	req = httptest.NewRequest(http.MethodPut, "/api/v1/marketing/audience/segments/1", bytes.NewReader(updateSeg))
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("update segment: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Delete segment
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/marketing/audience/segments/1", nil)
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete segment: expected 204, got %d", rr.Code)
	}
}

func TestHandleAudienceSegments_CreateRequiresName(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	body := `{"description":"no name"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/segments", strings.NewReader(body))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// --- Deliveries ---

func TestHandleAudienceDeliveries_EmptyList(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/deliveries", nil)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Deliveries []storage.EmailDelivery `json:"deliveries"`
		Total      int                     `json:"total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if len(resp.Deliveries) != 0 {
		t.Fatalf("expected empty deliveries, got %d", len(resp.Deliveries))
	}
}

// --- Send to List ---

func TestHandleAudienceSendToList(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	// Create contact
	contactBody := `{"email":"send-test@example.com","first_name":"Test","status":"active"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/contacts", strings.NewReader(contactBody))
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	// Create list
	listBody := `{"name":"Send Test List","account":"acme"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/lists", strings.NewReader(listBody))
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	// Add contact to list
	addBody := `{"contact_id":1}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/lists/1/members", strings.NewReader(addBody))
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	// Send to list
	sendBody := `{"list_id":1,"campaign_account":"acme","campaign_name":"launch","subject":"Hello!","body":"Welcome"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/send", strings.NewReader(sendBody))
	req = withTestUserContext(t, s, req)
	rr = httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("send: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var sendResp struct {
		Sent    int `json:"sent"`
		Skipped int `json:"skipped"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &sendResp); err != nil {
		t.Fatalf("invalid send response: %v", err)
	}
	if sendResp.Sent != 1 {
		t.Fatalf("sent = %d, want 1", sendResp.Sent)
	}
}

func TestHandleAudienceSendToList_RequiresFields(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	tests := []struct {
		name string
		body string
	}{
		{"missing list_id", `{"subject":"Hi","body":"Hello"}`},
		{"missing subject", `{"list_id":1,"body":"Hello"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/marketing/audience/send", strings.NewReader(tt.body))
			req = withTestUserContext(t, s, req)
			rr := httptest.NewRecorder()
			s.handleMarketingAudienceRouter(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rr.Code, rr.Body.String())
			}
		})
	}
}

// --- Router edge cases ---

func TestHandleAudienceRouter_InvalidPath(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/unknown", nil)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandleAudienceRouter_EmptyPath(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/audience/", nil)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleAudienceContacts_MethodNotAllowed(t *testing.T) {
	s, _, _ := setupMarketingServer(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/marketing/audience/contacts", nil)
	req = withTestUserContext(t, s, req)
	rr := httptest.NewRecorder()
	s.handleMarketingAudienceRouter(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}
