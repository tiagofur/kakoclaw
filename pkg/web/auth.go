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
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type authState struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	JWTSecretB64 string `json:"jwt_secret_b64"`
}

type jwtClaims struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
}

type authManager struct {
	path      string
	jwtExpiry time.Duration
	state     authState
}

func newAuthManager(dataDir string, cfgUsername, cfgPassword, cfgExpiry string) (*authManager, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	mgr := &authManager{
		path: filepath.Join(dataDir, "web-auth.json"),
	}
	if cfgExpiry == "" {
		cfgExpiry = "24h"
	}
	expiry, err := time.ParseDuration(cfgExpiry)
	if err != nil || expiry <= 0 {
		expiry = 24 * time.Hour
	}
	mgr.jwtExpiry = expiry

	if _, err := os.Stat(mgr.path); errors.Is(err, os.ErrNotExist) {
		username := strings.TrimSpace(cfgUsername)
		if username == "" {
			username = "admin"
		}
		password := strings.TrimSpace(cfgPassword)
		if password == "" {
			password = randomPassword(20)
			fmt.Printf("🔐 Web temporary password for '%s': %s\n", username, password)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, err
		}
		mgr.state = authState{
			Username:     username,
			PasswordHash: string(hash),
			JWTSecretB64: base64.RawURLEncoding.EncodeToString(secret),
		}
		if err := mgr.save(); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	if err := mgr.load(); err != nil {
		return nil, err
	}
	return mgr, nil
}

func (m *authManager) load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}
	var st authState
	if err := json.Unmarshal(data, &st); err != nil {
		return err
	}
	if st.Username == "" || st.PasswordHash == "" || st.JWTSecretB64 == "" {
		return errors.New("invalid auth state")
	}
	m.state = st
	return nil
}

func (m *authManager) save() error {
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, data, 0600)
}

func (m *authManager) login(username, password string) (string, error) {
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(username)), []byte(m.state.Username)) != 1 {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(m.state.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return m.signToken(m.state.Username)
}

func (m *authManager) signToken(username string) (string, error) {
	headerRaw := `{"alg":"HS256","typ":"JWT"}`
	claims := jwtClaims{
		Sub: username,
		Exp: time.Now().UTC().Add(m.jwtExpiry).Unix(),
	}
	payloadRaw, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(headerRaw))
	payload := base64.RawURLEncoding.EncodeToString(payloadRaw)
	signingInput := header + "." + payload

	secret, err := base64.RawURLEncoding.DecodeString(m.state.JWTSecretB64)
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

	secret, err := base64.RawURLEncoding.DecodeString(m.state.JWTSecretB64)
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
	return &claims, nil
}

func (m *authManager) changePassword(oldPassword, newPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(m.state.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid credentials")
	}
	if len(strings.TrimSpace(newPassword)) < 10 {
		return errors.New("new password must be at least 10 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	m.state.PasswordHash = string(hash)
	return m.save()
}

func randomPassword(n int) string {
	if n < 8 {
		n = 8
	}
	const alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$%&*"
	buf := make([]byte, n)
	rnd := make([]byte, n)
	if _, err := rand.Read(rnd); err != nil {
		return "ChangeMeNow123!"
	}
	for i := range buf {
		buf[i] = alphabet[int(rnd[i])%len(alphabet)]
	}
	return string(buf)
}

