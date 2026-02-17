package web

import (
	"testing"
	"time"
)

func TestAuthManagerLoginVerifyAndChangePassword(t *testing.T) {
	dir := t.TempDir()
	mgr, err := newAuthManager(dir, "admin", "InitialPass123!", "1h")
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

	if err := mgr.changePassword("InitialPass123!", "NewPass12345!"); err != nil {
		t.Fatalf("changePassword failed: %v", err)
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
	mgr, err := newAuthManager(dir, "admin", "InitialPass123!", "1ms")
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

