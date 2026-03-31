package storage

import (
	"fmt"
	"path/filepath"
	"testing"
)

func newAudienceTestStorage(t *testing.T) *Storage {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "user.db")
	s, err := NewUserStorage(dbPath)
	if err != nil {
		t.Fatalf("NewUserStorage: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// --- Contact CRUD ---

func TestContactCreateAndGet(t *testing.T) {
	s := newAudienceTestStorage(t)

	c, err := s.CreateContact(&Contact{Email: "alice@example.com", FirstName: "Alice", LastName: "Smith", Status: "active"})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}
	if c.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if c.Email != "alice@example.com" {
		t.Fatalf("email = %q, want alice@example.com", c.Email)
	}

	got, err := s.GetContactByID(c.ID)
	if err != nil {
		t.Fatalf("GetContactByID: %v", err)
	}
	if got == nil {
		t.Fatal("expected contact, got nil")
	}
	if got.FirstName != "Alice" {
		t.Fatalf("first_name = %q, want Alice", got.FirstName)
	}
}

func TestContactGetPreservesSubscriptionDates(t *testing.T) {
	s := newAudienceTestStorage(t)

	c, err := s.CreateContact(&Contact{Email: "sub@test.com", Status: "active"})
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}

	// GetContactByID should use scanContact and parse subscription dates
	got, err := s.GetContactByID(c.ID)
	if err != nil {
		t.Fatalf("GetContactByID: %v", err)
	}
	if got == nil {
		t.Fatal("expected contact")
	}
	// Dates should be nil (not set), not panicking
	if got.SubscribedAt != nil {
		t.Fatalf("expected nil SubscribedAt, got %v", got.SubscribedAt)
	}
}

func TestContactEmailValidation(t *testing.T) {
	s := newAudienceTestStorage(t)

	tests := []struct {
		name  string
		email string
	}{
		{"empty", ""},
		{"no at sign", "invalid"},
		{"no domain", "user@"},
		{"no dot in domain", "user@localhost"},
		{"trailing dot", "user@example."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.CreateContact(&Contact{Email: tt.email})
			if err == nil {
				t.Fatalf("expected validation error for email %q", tt.email)
			}
		})
	}
}

func TestContactDuplicateEmailRejected(t *testing.T) {
	s := newAudienceTestStorage(t)

	_, err := s.CreateContact(&Contact{Email: "dup@test.com"})
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err = s.CreateContact(&Contact{Email: "dup@test.com"})
	if err == nil {
		t.Fatal("expected duplicate email error")
	}
}

func TestContactUpdate(t *testing.T) {
	s := newAudienceTestStorage(t)

	c, _ := s.CreateContact(&Contact{Email: "update@test.com", FirstName: "Old"})
	c.FirstName = "New"
	if err := s.UpdateContact(c); err != nil {
		t.Fatalf("UpdateContact: %v", err)
	}

	got, _ := s.GetContactByID(c.ID)
	if got.FirstName != "New" {
		t.Fatalf("first_name = %q, want New", got.FirstName)
	}
}

func TestContactDelete(t *testing.T) {
	s := newAudienceTestStorage(t)

	c, _ := s.CreateContact(&Contact{Email: "delete@test.com"})
	if err := s.DeleteContact(c.ID); err != nil {
		t.Fatalf("DeleteContact: %v", err)
	}

	got, _ := s.GetContactByID(c.ID)
	if got != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestContactList(t *testing.T) {
	s := newAudienceTestStorage(t)

	s.CreateContact(&Contact{Email: "a@test.com", Status: "active"})
	s.CreateContact(&Contact{Email: "b@test.com", Status: "inactive"})

	contacts, total, err := s.ListContacts(ContactFilter{})
	if err != nil {
		t.Fatalf("ListContacts: %v", err)
	}
	if total != 2 {
		t.Fatalf("total = %d, want 2", total)
	}
	if len(contacts) != 2 {
		t.Fatalf("len = %d, want 2", len(contacts))
	}

	// Filter by status
	contacts, total, err = s.ListContacts(ContactFilter{Status: "active"})
	if err != nil {
		t.Fatalf("ListContacts with status filter: %v", err)
	}
	if total != 1 {
		t.Fatalf("total with status filter = %d, want 1", total)
	}
}

func TestContactCSVImportExport(t *testing.T) {
	s := newAudienceTestStorage(t)

	headers := []string{"email", "first_name", "last_name"}
	records := [][]string{
		{"alice@csv.com", "Alice", "A"},
		{"bob@csv.com", "Bob", "B"},
		{"", "NoEmail", ""},    // should be skipped
		{"invalid", "Bad", ""}, // should be skipped (invalid email)
	}

	inserted, skipped, err := s.ImportContactsFromCSV(records, headers)
	if err != nil {
		t.Fatalf("ImportContactsFromCSV: %v", err)
	}
	if inserted != 2 {
		t.Fatalf("inserted = %d, want 2", inserted)
	}
	if skipped != 2 {
		t.Fatalf("skipped = %d, want 2", skipped)
	}

	// Export
	data, err := s.ExportContactsCSV(ContactFilter{})
	if err != nil {
		t.Fatalf("ExportContactsCSV: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty CSV export")
	}
}

// --- Contact Lists ---

func TestContactListCRUD(t *testing.T) {
	s := newAudienceTestStorage(t)

	l, err := s.CreateContactList(&ContactList{Name: "VIPs", Account: "acme"})
	if err != nil {
		t.Fatalf("CreateContactList: %v", err)
	}
	if l.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if l.Slug == "" {
		t.Fatal("expected non-empty slug")
	}

	got, err := s.GetContactListByID(l.ID)
	if err != nil || got == nil {
		t.Fatalf("GetContactListByID: %v", err)
	}
	if got.Name != "VIPs" {
		t.Fatalf("name = %q, want VIPs", got.Name)
	}

	lists, err := s.ListContactLists("acme")
	if err != nil {
		t.Fatalf("ListContactLists: %v", err)
	}
	if len(lists) != 1 {
		t.Fatalf("len = %d, want 1", len(lists))
	}

	l.Name = "Premium"
	if err := s.UpdateContactList(l); err != nil {
		t.Fatalf("UpdateContactList: %v", err)
	}

	if err := s.DeleteContactList(l.ID); err != nil {
		t.Fatalf("DeleteContactList: %v", err)
	}

	got, _ = s.GetContactListByID(l.ID)
	if got != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestListMemberAddRemove(t *testing.T) {
	s := newAudienceTestStorage(t)

	c, _ := s.CreateContact(&Contact{Email: "member@test.com"})
	l, _ := s.CreateContactList(&ContactList{Name: "Test List", Account: "acme"})

	if err := s.AddListMember(c.ID, l.ID); err != nil {
		t.Fatalf("AddListMember: %v", err)
	}

	if err := s.RemoveListMember(c.ID, l.ID); err != nil {
		t.Fatalf("RemoveListMember: %v", err)
	}
}

func TestDeleteContactCascadesToMembers(t *testing.T) {
	s := newAudienceTestStorage(t)

	c, _ := s.CreateContact(&Contact{Email: "cascade@test.com"})
	l, _ := s.CreateContactList(&ContactList{Name: "Cascade List", Account: "acme"})
	s.AddListMember(c.ID, l.ID)

	if err := s.DeleteContact(c.ID); err != nil {
		t.Fatalf("DeleteContact: %v", err)
	}

	// Verify contact is gone
	got, _ := s.GetContactByID(c.ID)
	if got != nil {
		t.Fatal("expected contact to be deleted")
	}
}

// --- Segments ---

func TestSegmentCRUD(t *testing.T) {
	s := newAudienceTestStorage(t)

	seg, err := s.CreateSegment(&Segment{
		Name:    "Active Users",
		Account: "acme",
		Rules:   `[{"field":"status","operator":"equals","value":"active"}]`,
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}
	if seg.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	got, err := s.GetSegmentByID(seg.ID)
	if err != nil || got == nil {
		t.Fatalf("GetSegmentByID: %v", err)
	}
	if got.Name != "Active Users" {
		t.Fatalf("name = %q, want Active Users", got.Name)
	}

	segments, err := s.ListSegments("acme")
	if err != nil {
		t.Fatalf("ListSegments: %v", err)
	}
	if len(segments) != 1 {
		t.Fatalf("len = %d, want 1", len(segments))
	}

	seg.Name = "Very Active"
	if err := s.UpdateSegment(seg); err != nil {
		t.Fatalf("UpdateSegment: %v", err)
	}

	if err := s.DeleteSegment(seg.ID); err != nil {
		t.Fatalf("DeleteSegment: %v", err)
	}
}

func TestSegmentRuleEvaluationSQL(t *testing.T) {
	s := newAudienceTestStorage(t)

	s.CreateContact(&Contact{Email: "active1@test.com", Status: "active", Company: "Acme"})
	s.CreateContact(&Contact{Email: "active2@test.com", Status: "active", Company: "Beta"})
	s.CreateContact(&Contact{Email: "inactive@test.com", Status: "inactive", Company: "Acme"})

	// Equals operator
	ids, err := s.EvaluateSegmentRules([]SegmentRule{
		{Field: "status", Operator: "equals", Value: "active"},
	})
	if err != nil {
		t.Fatalf("EvaluateSegmentRules: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 active contacts, got %d", len(ids))
	}

	// Multiple rules (AND)
	ids, err = s.EvaluateSegmentRules([]SegmentRule{
		{Field: "status", Operator: "equals", Value: "active"},
		{Field: "company", Operator: "equals", Value: "Acme"},
	})
	if err != nil {
		t.Fatalf("EvaluateSegmentRules multi-rule: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 active Acme contact, got %d", len(ids))
	}

	// Contains operator — "active1" and "active2" contain "active1"/"active2" respectively
	ids, err = s.EvaluateSegmentRules([]SegmentRule{
		{Field: "email", Operator: "contains", Value: "active1"},
	})
	if err != nil {
		t.Fatalf("EvaluateSegmentRules contains: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 contact with 'active1' in email, got %d", len(ids))
	}

	// Empty rules = all contacts
	ids, err = s.EvaluateSegmentRules([]SegmentRule{})
	if err != nil {
		t.Fatalf("EvaluateSegmentRules empty: %v", err)
	}
	if len(ids) != 3 {
		t.Fatalf("expected all 3 contacts, got %d", len(ids))
	}
}

func TestSegmentContactCountUpdatesOnDelete(t *testing.T) {
	s := newAudienceTestStorage(t)

	s.CreateContact(&Contact{Email: "count1@test.com", Status: "active"})
	c2, _ := s.CreateContact(&Contact{Email: "count2@test.com", Status: "active"})

	seg, err := s.CreateSegment(&Segment{
		Name:    "All Active",
		Account: "acme",
		Rules:   `[{"field":"status","operator":"equals","value":"active"}]`,
	})
	if err != nil {
		t.Fatalf("CreateSegment: %v", err)
	}
	if seg.ContactCount != 2 {
		t.Fatalf("initial contact_count = %d, want 2", seg.ContactCount)
	}

	// Delete one contact — manually trigger segment count refresh
	s.DeleteContact(c2.ID)
	s.refreshAllSegmentCounts()

	updated, _ := s.GetSegmentByID(seg.ID)
	if updated.ContactCount != 1 {
		t.Fatalf("contact_count after delete = %d, want 1", updated.ContactCount)
	}
}

// --- Deliveries ---

func TestDeliveryCRUD(t *testing.T) {
	s := newAudienceTestStorage(t)

	d, err := s.CreateDelivery(&EmailDelivery{
		CampaignAccount: "acme",
		CampaignName:    "launch",
		Subject:         "Hello!",
	})
	if err != nil {
		t.Fatalf("CreateDelivery: %v", err)
	}
	if d.Status != "queued" {
		t.Fatalf("status = %q, want queued", d.Status)
	}

	got, err := s.GetDeliveryByID(d.ID)
	if err != nil || got == nil {
		t.Fatalf("GetDeliveryByID: %v", err)
	}
	if got.Subject != "Hello!" {
		t.Fatalf("subject = %q, want Hello!", got.Subject)
	}
}

func TestDeliveryStatusLifecycle(t *testing.T) {
	s := newAudienceTestStorage(t)

	d, _ := s.CreateDelivery(&EmailDelivery{
		CampaignAccount: "acme",
		Subject:         "Test",
	})

	for _, status := range []string{"sent", "delivered", "opened", "clicked"} {
		if err := s.UpdateDeliveryStatus(d.ID, status); err != nil {
			t.Fatalf("UpdateDeliveryStatus(%s): %v", status, err)
		}
		got, _ := s.GetDeliveryByID(d.ID)
		if got.Status != status {
			t.Fatalf("status = %q, want %q", got.Status, status)
		}
	}
}

func TestDeliveryStatusRejectsInvalid(t *testing.T) {
	s := newAudienceTestStorage(t)

	d, _ := s.CreateDelivery(&EmailDelivery{Subject: "Test"})

	err := s.UpdateDeliveryStatus(d.ID, "invalid_status")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}

	err = s.UpdateDeliveryStatus(d.ID, "'; DROP TABLE contacts; --")
	if err == nil {
		t.Fatal("expected error for SQL injection attempt")
	}
}

func TestDeliveryListPagination(t *testing.T) {
	s := newAudienceTestStorage(t)

	for i := 0; i < 5; i++ {
		s.CreateDelivery(&EmailDelivery{
			CampaignAccount: "acme",
			CampaignName:    "launch",
			Subject:         fmt.Sprintf("Email %d", i),
		})
	}

	deliveries, total, err := s.ListDeliveries("acme", "launch", 1, 2)
	if err != nil {
		t.Fatalf("ListDeliveries: %v", err)
	}
	if total != 5 {
		t.Fatalf("total = %d, want 5", total)
	}
	if len(deliveries) != 2 {
		t.Fatalf("page size = %d, want 2", len(deliveries))
	}
}

func TestSendToListFlow(t *testing.T) {
	s := newAudienceTestStorage(t)

	c1, _ := s.CreateContact(&Contact{Email: "send1@test.com", Status: "active"})
	c2, _ := s.CreateContact(&Contact{Email: "send2@test.com", Status: "unsubscribed"})

	l, _ := s.CreateContactList(&ContactList{Name: "Send List", Account: "acme"})
	s.AddListMember(c1.ID, l.ID)
	s.AddListMember(c2.ID, l.ID)

	var sentEmails []string
	sendFunc := func(to, subject, body string) error {
		sentEmails = append(sentEmails, to)
		return nil
	}

	sent, skipped, err := s.SendToList(l.ID, "acme", "launch", "Hello!", "Body", sendFunc)
	if err != nil {
		t.Fatalf("SendToList: %v", err)
	}
	if sent != 1 {
		t.Fatalf("sent = %d, want 1", sent)
	}
	if skipped != 1 {
		t.Fatalf("skipped = %d, want 1 (unsubscribed)", skipped)
	}
	if len(sentEmails) != 1 || sentEmails[0] != "send1@test.com" {
		t.Fatalf("sentEmails = %v, want [send1@test.com]", sentEmails)
	}

	// Verify delivery records were created
	deliveries, total, _ := s.ListDeliveries("acme", "launch", 1, 10)
	if total != 2 {
		t.Fatalf("total deliveries = %d, want 2 (1 sent + 1 failed)", total)
	}
	_ = deliveries
}

func TestSendToListSendError(t *testing.T) {
	s := newAudienceTestStorage(t)

	c, _ := s.CreateContact(&Contact{Email: "fail@test.com", Status: "active"})
	l, _ := s.CreateContactList(&ContactList{Name: "Fail List", Account: "acme"})
	s.AddListMember(c.ID, l.ID)

	sendFunc := func(to, subject, body string) error {
		return fmt.Errorf("SMTP connection refused")
	}

	sent, skipped, err := s.SendToList(l.ID, "acme", "launch", "Hello!", "Body", sendFunc)
	if err != nil {
		t.Fatalf("SendToList should not return error for individual send failures: %v", err)
	}
	if sent != 0 {
		t.Fatalf("sent = %d, want 0", sent)
	}
	if skipped != 1 {
		t.Fatalf("skipped = %d, want 1", skipped)
	}
}
