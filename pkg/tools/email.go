package tools

import (
	"context"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/sipeed/kakoclaw/pkg/config"
	"github.com/sipeed/kakoclaw/pkg/logger"
)

type EmailTool struct {
	cfg config.EmailToolsConfig
}

func NewEmailTool(cfg config.EmailToolsConfig) *EmailTool {
	return &EmailTool{cfg: cfg}
}

func (t *EmailTool) Name() string {
	return "send_email_report"
}

func (t *EmailTool) Description() string {
	return "Send a report or content via email. Use this to send weekly summaries, task results, or important notifications to the user."
}

func (t *EmailTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"subject": map[string]interface{}{
				"type":        "string",
				"description": "Subject of the email",
			},
			"body": map[string]interface{}{
				"type":        "string",
				"description": "Body content of the email (Markdown supported)",
			},
			"to": map[string]interface{}{
				"type":        "string",
				"description": "Recipient email address (optional, defaults to configured default)",
			},
		},
		"required": []string{"subject", "body"},
	}
}

func (t *EmailTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if !t.cfg.Enabled {
		return "", fmt.Errorf("email tool is disabled in configuration")
	}

	subject, _ := args["subject"].(string)
	body, _ := args["body"].(string)
	to, _ := args["to"].(string)

	if subject == "" || body == "" {
		return "", fmt.Errorf("subject and body are required")
	}

	if to == "" {
		to = t.cfg.To
	}
	if to == "" {
		return "", fmt.Errorf("no recipient specified")
	}

	// Construct email
	from := t.cfg.From
	if from == "" {
		from = t.cfg.Username
	}
	envelopeFrom, fromHeader, err := parseFromAddress(from, t.cfg.Username)
	if err != nil {
		return "", err
	}
	password := normalizedSMTPPassword(t.cfg.Host, t.cfg.Password)

	msg := []byte("From: " + fromHeader + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	auth := smtp.PlainAuth("", t.cfg.Username, password, t.cfg.Host)

	addr := fmt.Sprintf("%s:%d", t.cfg.Host, t.cfg.Port)
	if err := smtp.SendMail(addr, auth, envelopeFrom, []string{to}, msg); err != nil {
		logger.ErrorCF("tools", "Failed to send email", map[string]interface{}{
			"error": err.Error(),
			"host":  t.cfg.Host,
			"port":  t.cfg.Port,
			"to":    to,
		})
		return "", fmt.Errorf("failed to send email via %s (from=%s, to=%s): %w", addr, envelopeFrom, to, err)
	}

	logger.InfoCF("tools", "Email sent successfully", map[string]interface{}{"to": to, "subject": subject})
	return "Email sent successfully to " + to, nil
}

func parseFromAddress(from, fallback string) (envelope, header string, err error) {
	trimmedFrom := strings.TrimSpace(from)
	if trimmedFrom == "" {
		trimmedFrom = strings.TrimSpace(fallback)
	}
	if trimmedFrom == "" {
		return "", "", fmt.Errorf("from address is required")
	}

	parsed, parseErr := mail.ParseAddress(trimmedFrom)
	if parseErr == nil && parsed != nil {
		return parsed.Address, parsed.String(), nil
	}
	if strings.Contains(trimmedFrom, "<") || strings.Contains(trimmedFrom, ">") {
		return "", "", fmt.Errorf("invalid from address %q", from)
	}
	return trimmedFrom, trimmedFrom, nil
}

func normalizedSMTPPassword(host, password string) string {
	if strings.EqualFold(strings.TrimSpace(host), "smtp.gmail.com") {
		return strings.ReplaceAll(password, " ", "")
	}
	return password
}
