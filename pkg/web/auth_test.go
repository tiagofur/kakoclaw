package web

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/storage"
)

func TestAuthManagerLoginVerifyAndChangePassword(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "test.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	defer store.Close()

	mgr, err := newAuthManager(store, "admin", "InitialPass123!", "1h")
	if err != nil {
		t.Fatalf("newAuthManager failed: %v", err)
	}

	token, err := mgr.login("admin", "InitialPass123!")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	claims, err := mgr.verifyToken(token)
	if err != nil {
		t.Fatalf("verifyToken failed: %v", err)
	}
	if claims.Sub != "admin" {
		t.Fatalf("expected sub admin, got %s", claims.Sub)
	}

	if err := mgr.changePassword("admin", "InitialPass123!", "NewPass12345!"); err != nil {
		t.Fatalf("changePassword failed: %v", err)
	}
	// Old token should be invalid after password change (JWT secret rotated)
	if _, err := mgr.verifyToken(token); err == nil {
		t.Fatal("old token should be invalid after password change")
	}
	if _, err := mgr.login("admin", "InitialPass123!"); err == nil {
		t.Fatal("old password should fail after change")
	}
	if _, err := mgr.login("admin", "NewPass12345!"); err != nil {
		t.Fatalf("new password should pass: %v", err)
	}
}

func TestAuthManagerTokenExpiry(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "test_expiry.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	defer store.Close()

	mgr, err := newAuthManager(store, "admin", "InitialPass123!", "1ms")
	if err != nil {
		t.Fatalf("newAuthManager failed: %v", err)
	}
	token, err := mgr.login("admin", "InitialPass123!")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if _, err := mgr.verifyToken(token); err == nil {
		t.Fatal("expected expired token to fail")
	}
}

