package storage_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/storage"
)

func newPairingTestStorage(t *testing.T) *storage.Storage {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := storage.NewUserStorage(path)
	if err != nil {
		t.Fatalf("NewUserStorage: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestPairingStore_InsertAndApprove(t *testing.T) {
	s := newPairingTestStorage(t)
	ps := s.Pairing()

	ok, err := ps.IsApproved("telegram", "u1")
	if err != nil {
		t.Fatalf("IsApproved: %v", err)
	}
	if ok {
		t.Error("expected IsApproved=false before any entry")
	}

	if err := ps.InsertPending("telegram", "u1", "abc123"); err != nil {
		t.Fatalf("InsertPending: %v", err)
	}

	pending, err := ps.HasPending("telegram", "u1")
	if err != nil {
		t.Fatalf("HasPending: %v", err)
	}
	if !pending {
		t.Error("expected HasPending=true after InsertPending")
	}

	senderID, err := ps.Approve("telegram", "abc123")
	if err != nil {
		t.Fatalf("Approve: %v", err)
	}
	if senderID != "u1" {
		t.Errorf("expected senderID=%q, got %q", "u1", senderID)
	}

	approved, err := ps.IsApproved("telegram", "u1")
	if err != nil {
		t.Fatalf("IsApproved after Approve: %v", err)
	}
	if !approved {
		t.Error("expected IsApproved=true after Approve")
	}
}

func TestPairingStore_DeduplicatePending(t *testing.T) {
	s := newPairingTestStorage(t)
	ps := s.Pairing()

	if err := ps.InsertPending("telegram", "u2", "code1"); err != nil {
		t.Fatalf("InsertPending(code1): %v", err)
	}
	// Second insert for same sender must not error and must not overwrite
	if err := ps.InsertPending("telegram", "u2", "code2"); err != nil {
		t.Fatalf("InsertPending(code2): %v", err)
	}

	// Original code still works
	_, err := ps.Approve("telegram", "code1")
	if err != nil {
		t.Fatalf("Approve(code1): %v", err)
	}
}

func TestPairingStore_UnknownCodeReturnsError(t *testing.T) {
	s := newPairingTestStorage(t)
	ps := s.Pairing()

	_, err := ps.Approve("telegram", "nope")
	if !errors.Is(err, storage.ErrCodeNotFound) {
		t.Errorf("expected ErrCodeNotFound, got %v", err)
	}
}
