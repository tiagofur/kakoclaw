package storage

import (
	"path/filepath"
	"strings"
	"testing"
)

func newUserTestStorage(t *testing.T) *Storage {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "user_test.db")
	s, err := NewUserStorage(dbPath)
	if err != nil {
		t.Fatalf("NewUserStorage: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// ---------------------------------------------------------------------------
// Basic create + list
// ---------------------------------------------------------------------------

func TestCreateAndListStandingOrders(t *testing.T) {
	s := newUserTestStorage(t)

	id, err := s.CreateStandingOrder("Always respond in English.", "language")
	if err != nil {
		t.Fatalf("CreateStandingOrder: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	orders, err := s.ListStandingOrders(true)
	if err != nil {
		t.Fatalf("ListStandingOrders: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}

	o := orders[0]
	if o.Content != "Always respond in English." {
		t.Errorf("unexpected content: %q", o.Content)
	}
	if o.Label != "language" {
		t.Errorf("unexpected label: %q", o.Label)
	}
	if !o.Active {
		t.Error("expected order to be active by default")
	}
}

// ---------------------------------------------------------------------------
// 200-char content limit
// ---------------------------------------------------------------------------

func TestCreateStandingOrder_ContentTooLong(t *testing.T) {
	s := newUserTestStorage(t)

	longContent := strings.Repeat("x", maxStandingOrderContent+1)
	_, err := s.CreateStandingOrder(longContent, "")
	if err == nil {
		t.Fatal("expected error for content exceeding limit")
	}
}

func TestUpdateStandingOrder_ContentTooLong(t *testing.T) {
	s := newUserTestStorage(t)

	id, err := s.CreateStandingOrder("short content", "")
	if err != nil {
		t.Fatalf("CreateStandingOrder: %v", err)
	}

	longContent := strings.Repeat("x", maxStandingOrderContent+1)
	err = s.UpdateStandingOrder(id, longContent, "")
	if err == nil {
		t.Fatal("expected error for content exceeding limit")
	}
}

// ---------------------------------------------------------------------------
// 20-order active limit
// ---------------------------------------------------------------------------

func TestCreateStandingOrder_MaxActiveOrders(t *testing.T) {
	s := newUserTestStorage(t)

	for i := 0; i < maxStandingOrders; i++ {
		_, err := s.CreateStandingOrder("order content", "")
		if err != nil {
			t.Fatalf("CreateStandingOrder %d: %v", i, err)
		}
	}

	// The 21st should fail
	_, err := s.CreateStandingOrder("one too many", "")
	if err == nil {
		t.Fatal("expected error when exceeding max active orders")
	}
}

// ---------------------------------------------------------------------------
// 4000-char total limit
// ---------------------------------------------------------------------------

func TestCreateStandingOrder_TotalCharLimit(t *testing.T) {
	s := newUserTestStorage(t)

	// Each order = 200 chars → 20 orders = 4000 chars exactly at limit.
	// Add 19 orders of 200 chars (3800 chars) — should succeed.
	content := strings.Repeat("a", maxStandingOrderContent)
	for i := 0; i < 19; i++ {
		_, err := s.CreateStandingOrder(content, "")
		if err != nil {
			t.Fatalf("CreateStandingOrder %d: %v", i, err)
		}
	}

	// 20th order: 1 char pushes us from 3800 → 3801. Should still succeed (≤ 4000).
	_, err := s.CreateStandingOrder("x", "")
	if err != nil {
		t.Fatalf("CreateStandingOrder within total limit: %v", err)
	}

	// Now total is 3801 chars (19×200 + 1). Try adding 200 more → 4001 > 4000 → should fail.
	_, err = s.CreateStandingOrder(content, "")
	if err == nil {
		t.Fatal("expected error when total chars exceed limit")
	}
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdateStandingOrder(t *testing.T) {
	s := newUserTestStorage(t)

	id, err := s.CreateStandingOrder("original", "old-label")
	if err != nil {
		t.Fatalf("CreateStandingOrder: %v", err)
	}

	if err := s.UpdateStandingOrder(id, "updated content", "new-label"); err != nil {
		t.Fatalf("UpdateStandingOrder: %v", err)
	}

	orders, err := s.ListStandingOrders(false)
	if err != nil {
		t.Fatalf("ListStandingOrders: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	if orders[0].Content != "updated content" {
		t.Errorf("unexpected content after update: %q", orders[0].Content)
	}
	if orders[0].Label != "new-label" {
		t.Errorf("unexpected label after update: %q", orders[0].Label)
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDeleteStandingOrder(t *testing.T) {
	s := newUserTestStorage(t)

	id, err := s.CreateStandingOrder("to be deleted", "")
	if err != nil {
		t.Fatalf("CreateStandingOrder: %v", err)
	}

	if err := s.DeleteStandingOrder(id); err != nil {
		t.Fatalf("DeleteStandingOrder: %v", err)
	}

	orders, err := s.ListStandingOrders(false)
	if err != nil {
		t.Fatalf("ListStandingOrders: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0 orders after delete, got %d", len(orders))
	}
}

// ---------------------------------------------------------------------------
// Toggle
// ---------------------------------------------------------------------------

func TestToggleStandingOrder(t *testing.T) {
	s := newUserTestStorage(t)

	id, err := s.CreateStandingOrder("toggle me", "")
	if err != nil {
		t.Fatalf("CreateStandingOrder: %v", err)
	}

	// Active by default → toggle off
	if err := s.ToggleStandingOrder(id); err != nil {
		t.Fatalf("ToggleStandingOrder (disable): %v", err)
	}

	active, err := s.ListStandingOrders(true)
	if err != nil {
		t.Fatalf("ListStandingOrders (active): %v", err)
	}
	if len(active) != 0 {
		t.Errorf("expected 0 active orders after toggle off, got %d", len(active))
	}

	all, err := s.ListStandingOrders(false)
	if err != nil {
		t.Fatalf("ListStandingOrders (all): %v", err)
	}
	if len(all) != 1 {
		t.Errorf("expected 1 total order, got %d", len(all))
	}
	if all[0].Active {
		t.Error("expected order to be inactive after toggle")
	}

	// Toggle back on
	if err := s.ToggleStandingOrder(id); err != nil {
		t.Fatalf("ToggleStandingOrder (re-enable): %v", err)
	}

	active, err = s.ListStandingOrders(true)
	if err != nil {
		t.Fatalf("ListStandingOrders (active again): %v", err)
	}
	if len(active) != 1 {
		t.Errorf("expected 1 active order after re-enable, got %d", len(active))
	}
}

// ---------------------------------------------------------------------------
// Multi-user isolation (shared DB with user_id)
// ---------------------------------------------------------------------------

func TestStandingOrdersMultiUserIsolation(t *testing.T) {
	s := newTestStorage(t) // shared DB (isUserDB = false)

	var user1 int64 = 10
	var user2 int64 = 20

	_, err := s.CreateStandingOrderForUser(user1, "user1 order", "")
	if err != nil {
		t.Fatalf("CreateStandingOrderForUser user1: %v", err)
	}

	_, err = s.CreateStandingOrderForUser(user2, "user2 order", "")
	if err != nil {
		t.Fatalf("CreateStandingOrderForUser user2: %v", err)
	}

	orders1, err := s.ListStandingOrdersForUser(user1, false)
	if err != nil {
		t.Fatalf("ListStandingOrdersForUser user1: %v", err)
	}
	if len(orders1) != 1 {
		t.Fatalf("expected 1 order for user1, got %d", len(orders1))
	}
	if orders1[0].Content != "user1 order" {
		t.Errorf("unexpected content for user1: %q", orders1[0].Content)
	}

	orders2, err := s.ListStandingOrdersForUser(user2, false)
	if err != nil {
		t.Fatalf("ListStandingOrdersForUser user2: %v", err)
	}
	if len(orders2) != 1 {
		t.Fatalf("expected 1 order for user2, got %d", len(orders2))
	}
	if orders2[0].Content != "user2 order" {
		t.Errorf("unexpected content for user2: %q", orders2[0].Content)
	}
}

// ---------------------------------------------------------------------------
// ActiveOnly filter
// ---------------------------------------------------------------------------

func TestListStandingOrders_ActiveOnlyFilter(t *testing.T) {
	s := newUserTestStorage(t)

	id1, _ := s.CreateStandingOrder("active order", "")
	id2, _ := s.CreateStandingOrder("to be toggled off", "")
	_ = id1

	if err := s.ToggleStandingOrder(id2); err != nil {
		t.Fatalf("ToggleStandingOrder: %v", err)
	}

	active, err := s.ListStandingOrders(true)
	if err != nil {
		t.Fatalf("ListStandingOrders(true): %v", err)
	}
	if len(active) != 1 {
		t.Errorf("expected 1 active order, got %d", len(active))
	}

	all, err := s.ListStandingOrders(false)
	if err != nil {
		t.Fatalf("ListStandingOrders(false): %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 total orders, got %d", len(all))
	}
}
