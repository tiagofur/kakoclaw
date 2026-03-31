package marketing

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/sipeed/makoclaw/pkg/config"
)

type SMTPSender struct {
	cfg     config.EmailToolsConfig
	baseURL string
}

func NewSender(cfg config.EmailToolsConfig, baseURL string) *SMTPSender {
	return &SMTPSender{cfg: cfg, baseURL: baseURL}
}

func (s *SMTPSender) Send(to, subject, htmlBody, textBody string, extraHeaders map[string]string) error {
	if s.cfg.Host == "" {
		return fmt.Errorf("SMTP not configured: host is empty")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%s\r\n",
		s.cfg.From, to, subject, writer.Boundary())
	for k, v := range extraHeaders {
		headers += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	headers += "\r\n"

	textPart, _ := writer.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/plain; charset=UTF-8"}})
	textPart.Write([]byte(textBody))

	htmlPart, _ := writer.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html; charset=UTF-8"}})
	htmlPart.Write([]byte(htmlBody))

	writer.Close()

	auth := smtp.PlainAuth("", s.cfg.Username, normalizePassword(s.cfg.Password), s.cfg.Host)
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(headers+buf.String()))
}

func (s *SMTPSender) GenerateUnsubscribeToken(contactID, campaignID int64, secret string) string {
	payload := fmt.Sprintf("%d:%d", contactID, campaignID)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	token := base64.URLEncoding.EncodeToString([]byte(payload + ":" + sig))
	return token
}

func (s *SMTPSender) VerifyUnsubscribeToken(token string, secret string) (contactID, campaignID int64, err error) {
	data, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid token encoding: %w", err)
	}
	parts := strings.SplitN(string(data), ":", 3)
	if len(parts) != 3 {
		return 0, 0, fmt.Errorf("invalid token format")
	}
	fmt.Sscanf(parts[0], "%d", &contactID)
	fmt.Sscanf(parts[1], "%d", &campaignID)
	payload := parts[0] + ":" + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return 0, 0, fmt.Errorf("invalid token signature")
	}
	return contactID, campaignID, nil
}

func AppendUnsubscribeFooter(htmlBody, textBody, unsubURL string) (string, string) {
	htmlFooter := fmt.Sprintf(`<br><br><hr style="border:none;border-top:1px solid #ddd;"><p style="font-size:12px;color:#888;"><a href="%s" style="color:#888;">Unsubscribe</a> from these emails.</p>`, unsubURL)
	textFooter := fmt.Sprintf("\n\n---\nUnsubscribe: %s", unsubURL)
	return htmlBody + htmlFooter, textBody + textFooter
}

func normalizePassword(pwd string) string {
	return strings.ReplaceAll(pwd, " ", "")
}
