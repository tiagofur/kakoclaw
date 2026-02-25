package web

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/storage"
	"golang.org/x/crypto/bcrypt"
)

type jwtClaims struct {
	Sub  string `json:"sub"`            // username
	UUID string `json:"uuid,omitempty"` // user UUID for per-user DB lookup
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

type authManager struct {
	store     *storage.CentralStorage
	jwtExpiry time.Duration
}

func newAuthManager(store *storage.CentralStorage, cfgUsername, cfgPassword, cfgExpiry string) (*authManager, error) {
	if store == nil {
		return nil, errors.New("central storage is required for authManager")
	}

	mgr := &authManager{
		store: store,
	}

	if cfgExpiry == "" {
		cfgExpiry = "24h"
	}
	expiry, err := time.ParseDuration(cfgExpiry)
	if err != nil || expiry <= 0 {
		expiry = 24 * time.Hour
	}
	mgr.jwtExpiry = expiry

	// Ensure JWT secret exists in DB
	secret, err := mgr.store.GetSetting("jwt_secret_b64")
	if err != nil {
		return nil, err
	}
	if secret == "" {
		newSecret := make([]byte, 32)
		if _, err := rand.Read(newSecret); err != nil {
			return nil, err
		}
		secret = base64.RawURLEncoding.EncodeToString(newSecret)
		if err := mgr.store.SetSetting("jwt_secret_b64", secret); err != nil {
			return nil, err
		}
	}

	// Ensure there is at least one admin user
	adminCount, err := mgr.store.CountAdmins()
	if err != nil {
		return nil, err
	}
	if adminCount == 0 {
		username := strings.TrimSpace(cfgUsername)
		if username == "" {
			username = "admin"
		}
		password := strings.TrimSpace(cfgPassword)
		if password == "" {
			return nil, errors.New("web admin bootstrap requires web.password when no admin user exists")
		}

		user, err := mgr.store.GetUserByUsername(username)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				if _, createErr := mgr.store.CreateUser(username, password, "admin"); createErr != nil {
					return nil, createErr
				}
			} else {
				return nil, err
			}
		} else if user.Role != "admin" {
			if err := mgr.store.UpdateUserRole(user.ID, "admin"); err != nil {
				return nil, err
			}
		}
	} else if adminCount > 0 {
		// Sync admin password from config if provided and user matches
		username := strings.TrimSpace(cfgUsername)
		if username == "" {
			username = "admin"
		}
		password := strings.TrimSpace(cfgPassword)
		if password != "" {
			user, err := mgr.store.GetUserByUsername(username)
			if err == nil && user.Role == "admin" {
				// Update password to match config to ensure .env is source of truth
				if err := mgr.store.UpdateUserPassword(user.ID, password); err != nil {
					// Don't fail the whole startup if sync fails, just log it
					fmt.Printf("Warning: Failed to sync admin password for %s: %v\n", username, err)
				}
			}
		}
	}

	return mgr, nil
}

func (m *authManager) login(username, password string) (string, error) {
	user, err := m.store.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			user, err = m.store.GetUserByEmail(username)
			if err != nil {
				if errors.Is(err, storage.ErrUserNotFound) {
					return "", errors.New("invalid credentials")
				}
				return "", err
			}
		} else {
			return "", err
		}
	}

	// Check if user is blocked before verifying password to avoid timing attacks
	if user.Blocked {
		return "", fmt.Errorf("usuario bloqueado. Motivo: %s. Contacte soporte", user.BlockedReason)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return m.signToken(user.Username, user.UUID, user.Role)
}

func (m *authManager) signToken(username, userUUID, role string) (string, error) {
	headerRaw := `{"alg":"HS256","typ":"JWT"}`
	claims := jwtClaims{
		Sub:  username,
		UUID: userUUID,
		Role: role,
		Exp:  time.Now().UTC().Add(m.jwtExpiry).Unix(),
	}
	payloadRaw, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(headerRaw))
	payload := base64.RawURLEncoding.EncodeToString(payloadRaw)
	signingInput := header + "." + payload

	secretB64, err := m.store.GetSetting("jwt_secret_b64")
	if err != nil {
		return "", err
	}
	secret, err := base64.RawURLEncoding.DecodeString(secretB64)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + sig, nil
}

func (m *authManager) verifyToken(token string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}
	signingInput := parts[0] + "." + parts[1]

	secretB64, err := m.store.GetSetting("jwt_secret_b64")
	if err != nil {
		return nil, err
	}
	secret, err := base64.RawURLEncoding.DecodeString(secretB64)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(expectedSig)) != 1 {
		return nil, errors.New("invalid token signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	if claims.Sub == "" || claims.Exp <= 0 || time.Now().UTC().Unix() >= claims.Exp {
		return nil, errors.New("token expired")
	}
	// Fallback for older tokens lacking role
	if claims.Role == "" {
		claims.Role = "user"
	}
	return &claims, nil
}

func (m *authManager) changePassword(username, oldPassword, newPassword string) error {
	user, err := m.store.GetUserByUsername(username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid credentials")
	}
	if len(strings.TrimSpace(newPassword)) < 8 {
		return errors.New("new password must be at least 8 characters")
	}

	if err := m.store.UpdateUserPassword(user.ID, newPassword); err != nil {
		return err
	}

	// Rotate JWT secret to invalidate all existing tokens
	newSecret := make([]byte, 32)
	if _, err := rand.Read(newSecret); err != nil {
		return err
	}
	return m.store.SetSetting("jwt_secret_b64", base64.RawURLEncoding.EncodeToString(newSecret))
}

