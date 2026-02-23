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

func TestAuthManagerRequiresBootstrapPasswordWhenNoAdmins(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "test_bootstrap_missing_password.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	defer store.Close()

	_, err = newAuthManager(store, "admin", "", "1h")
	if err == nil {
		t.Fatal("expected error when no admin exists and bootstrap password is empty")
	}
}

func TestAuthManagerPromotesConfiguredUserWhenNoAdmins(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "test_promote_configured_user.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	defer store.Close()

	if _, err := store.CreateUser("alice", "AlicePass123!", "user"); err != nil {
		t.Fatalf("CreateUser alice failed: %v", err)
	}

	if _, err := newAuthManager(store, "alice", "AnyBootstrapPassword123!", "1h"); err != nil {
		t.Fatalf("newAuthManager failed: %v", err)
	}

	alice, err := store.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("GetUserByUsername alice failed: %v", err)
	}
	if alice.Role != "admin" {
		t.Fatalf("expected alice role admin after recovery, got %q", alice.Role)
	}
}

func TestAuthManagerCreatesConfiguredAdminWhenNoAdmins(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "test_create_configured_admin.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	defer store.Close()

	if _, err := store.CreateUser("existing", "ExistingPass123!", "user"); err != nil {
		t.Fatalf("CreateUser existing failed: %v", err)
	}

	mgr, err := newAuthManager(store, "admin", "BootstrapPass123!", "1h")
	if err != nil {
		t.Fatalf("newAuthManager failed: %v", err)
	}

	admin, err := store.GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("GetUserByUsername admin failed: %v", err)
	}
	if admin.Role != "admin" {
		t.Fatalf("expected admin role admin, got %q", admin.Role)
	}

	if _, err := mgr.login("admin", "BootstrapPass123!"); err != nil {
		t.Fatalf("bootstrap admin login failed: %v", err)
	}
}

func TestAuthManagerKeepsExistingAdminAndDoesNotPromoteConfiguredUser(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "test_keep_existing_admin.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	defer store.Close()

	if _, err := store.CreateUser("admin1", "Admin1Pass123!", "admin"); err != nil {
		t.Fatalf("CreateUser admin1 failed: %v", err)
	}
	if _, err := store.CreateUser("alice", "AlicePass123!", "user"); err != nil {
		t.Fatalf("CreateUser alice failed: %v", err)
	}

	if _, err := newAuthManager(store, "alice", "BootstrapPass123!", "1h"); err != nil {
		t.Fatalf("newAuthManager failed: %v", err)
	}

	alice, err := store.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("GetUserByUsername alice failed: %v", err)
	}
	if alice.Role != "user" {
		t.Fatalf("expected alice to remain user when admin already exists, got %q", alice.Role)
	}
}

func TestAuthManagerInvalidCredentialsErrorShape(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.New(config.StorageConfig{Path: filepath.Join(dir, "test_invalid_credentials_shape.db")})
	if err != nil {
		t.Fatalf("storage.New failed: %v", err)
	}
	defer store.Close()

	mgr, err := newAuthManager(store, "admin", "InitialPass123!", "1h")
	if err != nil {
		t.Fatalf("newAuthManager failed: %v", err)
	}

	_, err = mgr.login("admin", "wrong-pass")
	if err == nil {
		t.Fatal("expected login error for wrong password")
	}
	if err.Error() != "invalid credentials" {
		t.Fatalf("expected generic invalid credentials error, got %v", err)
	}
}
