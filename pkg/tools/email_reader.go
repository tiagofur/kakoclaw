package tools

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	gomessage "github.com/emersion/go-message"
	gomail "github.com/emersion/go-message/mail"
	"golang.org/x/net/html"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
)

// ReadEmailTool allows the agent to read emails from the user's inbox.
type ReadEmailTool struct {
	cfg config.EmailChannelConfig
}

// NewReadEmailTool creates a new email reader tool.
func NewReadEmailTool(cfg config.EmailChannelConfig) *ReadEmailTool {
	return &ReadEmailTool{cfg: cfg}
}

func (t *ReadEmailTool) Name() string {
	return "read_email"
}

func (t *ReadEmailTool) Description() string {
	return "Read recent emails from the user's inbox. Can fetch unread emails or search by sender. Returns subject, sender, date, and body for each email."
}

func (t *ReadEmailTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"count": map[string]interface{}{
				"type":        "integer",
				"description": "Number of recent emails to fetch (default: 5, max: 20)",
			},
			"unread_only": map[string]interface{}{
				"type":        "boolean",
				"description": "Only fetch unread/unseen emails (default: true)",
			},
			"from": map[string]interface{}{
				"type":        "string",
				"description": "Filter by sender email address (optional)",
			},
		},
	}
}

func (t *ReadEmailTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if t.cfg.IMAPHost == "" {
		return "", fmt.Errorf("email channel not configured: missing IMAP host")
	}
	if t.cfg.Password == "" {
		return "", fmt.Errorf("email channel not configured: missing password")
	}

	count := 5
	if c, ok := args["count"].(float64); ok && c > 0 {
		count = int(c)
	}
	if count > 20 {
		count = 20
	}

	unreadOnly := true
	if u, ok := args["unread_only"].(bool); ok {
		unreadOnly = u
	}

	fromFilter := ""
	if f, ok := args["from"].(string); ok {
		fromFilter = strings.TrimSpace(f)
	}

	// Connect to IMAP
	imapPort := t.cfg.IMAPPort
	if imapPort <= 0 {
		imapPort = 993
	}
	addr := fmt.Sprintf("%s:%d", t.cfg.IMAPHost, imapPort)

	options := &imapclient.Options{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: t.cfg.InsecureSkipVerify,
		},
	}

	client, err := imapclient.DialTLS(addr, options)
	if err != nil {
		return "", fmt.Errorf("IMAP connect failed (%s): %w", addr, err)
	}
	defer client.Close()
	defer func() { _ = client.Logout().Wait() }()

	if err := client.Login(t.cfg.Username, t.cfg.Password).Wait(); err != nil {
		return "", fmt.Errorf("IMAP login failed: %w", err)
	}

	mailbox := t.cfg.Mailbox
	if mailbox == "" {
		mailbox = "INBOX"
	}

	selectData, err := client.Select(mailbox, nil).Wait()
	if err != nil {
		return "", fmt.Errorf("IMAP SELECT %q failed: %w", mailbox, err)
	}

	// Build search criteria
	criteria := &imap.SearchCriteria{}
	if unreadOnly {
		criteria.NotFlag = []imap.Flag{imap.FlagSeen}
	}

	searchData, err := client.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return "", fmt.Errorf("IMAP SEARCH failed: %w", err)
	}

	allUIDs := searchData.AllUIDs()
	if len(allUIDs) == 0 {
		status := "unread"
		if !unreadOnly {
			status = ""
		}
		return fmt.Sprintf("No %s emails found in %s (total messages: %d)", status, mailbox, selectData.NumMessages), nil
	}

	// Take only the last N UIDs (most recent)
	startIdx := 0
	if len(allUIDs) > count {
		startIdx = len(allUIDs) - count
	}
	recentUIDs := allUIDs[startIdx:]

	// Fetch messages
	uidSet := imap.UIDSetNum(recentUIDs...)
	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		UID:      true,
		Flags:    true,
		BodySection: []*imap.FetchItemBodySection{
			{Peek: true},
		},
	}

	fetchCmd := client.Fetch(uidSet, fetchOptions)
	defer fetchCmd.Close()

	type emailSummary struct {
		UID     uint32 `json:"uid"`
		From    string `json:"from"`
		Subject string `json:"subject"`
		Date    string `json:"date"`
		Body    string `json:"body"`
		Unread  bool   `json:"unread"`
	}

	var emails []emailSummary

	for {
		msgData := fetchCmd.Next()
		if msgData == nil {
			break
		}

		var uid imap.UID
		var envelope *imap.Envelope
		var bodyBytes []byte
		var flags []imap.Flag

		for {
			item := msgData.Next()
			if item == nil {
				break
			}
			switch data := item.(type) {
			case imapclient.FetchItemDataUID:
				uid = data.UID
			case imapclient.FetchItemDataEnvelope:
				envelope = data.Envelope
			case imapclient.FetchItemDataBodySection:
				// MUST read the literal immediately: data.Literal is backed by
				// the live IMAP TCP stream and will be discarded when Next() is
				// called for the following fetch item.
				rawBody, readErr := io.ReadAll(io.LimitReader(data.Literal, 10*1024*1024))
				if readErr == nil {
					bodyBytes = rawBody
				}
			case imapclient.FetchItemDataFlags:
				flags = data.Flags
			}
		}

		if envelope == nil {
			continue
		}

		senderAddr := ""
		if len(envelope.From) > 0 {
			senderAddr = envelope.From[0].Addr()
		}

		// Apply sender filter
		if fromFilter != "" && !strings.EqualFold(senderAddr, fromFilter) {
			continue
		}

		// Extract body
		body := ""
		if len(bodyBytes) > 0 {
			body, _ = extractEmailPlainText(bytes.NewReader(bodyBytes))
			// Truncate long bodies
			if len(body) > 2000 {
				body = body[:2000] + "\n...[truncated]"
			}
		}

		isUnread := true
		for _, f := range flags {
			if f == imap.FlagSeen {
				isUnread = false
				break
			}
		}

		dateStr := ""
		if envelope.Date != (time.Time{}) {
			dateStr = envelope.Date.Format("2006-01-02 15:04")
		}

		emails = append(emails, emailSummary{
			UID:     uint32(uid),
			From:    senderAddr,
			Subject: envelope.Subject,
			Date:    dateStr,
			Body:    strings.TrimSpace(body),
			Unread:  isUnread,
		})
	}

	if len(emails) == 0 {
		msg := "No emails found"
		if fromFilter != "" {
			msg += fmt.Sprintf(" from %s", fromFilter)
		}
		return msg, nil
	}

	result, _ := json.MarshalIndent(map[string]interface{}{
		"emails":      emails,
		"count":       len(emails),
		"total_in_search": len(allUIDs),
		"mailbox":     mailbox,
	}, "", "  ")

	logger.InfoCF("tool", "Read emails from inbox", map[string]interface{}{
		"count":   len(emails),
		"mailbox": mailbox,
	})

	return string(result), nil
}

// extractEmailPlainText extracts plain text from an email body (MIME parsing).
func extractEmailPlainText(r io.Reader) (string, error) {
	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	if len(bodyBytes) == 0 {
		return "", nil
	}

	// First pass: mail-aware parser handles real-world MIME trees and transfer encodings.
	if plain, ok := extractEmailWithMailReader(bodyBytes); ok {
		return plain, nil
	}

	entity, err := gomessage.Read(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return extractBodyFromRawRFC822(bodyBytes), nil
	}

	// Check if it's multipart
	if mr := entity.MultipartReader(); mr != nil {
		return extractFromEmailMultipart(mr)
	}

	// Single-part message
	contentType, _, _ := entity.Header.ContentType()
	bodyContent, err := io.ReadAll(entity.Body)
	if err != nil {
		return extractBodyFromRawRFC822(bodyBytes), nil
	}

	if strings.HasPrefix(contentType, "text/html") {
		return emailHtmlToPlaintext(string(bodyContent)), nil
	}
	plain := strings.TrimSpace(string(bodyContent))
	if plain == "" {
		return extractBodyFromRawRFC822(bodyBytes), nil
	}
	return plain, nil
}

func extractEmailWithMailReader(bodyBytes []byte) (string, bool) {
	mr, err := gomail.CreateReader(bytes.NewReader(bodyBytes))
	if err != nil {
		return "", false
	}

	var plainText, htmlText string
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch h := part.Header.(type) {
		case *gomail.InlineHeader:
			contentType, _, _ := h.ContentType()
			contentType = strings.ToLower(strings.TrimSpace(contentType))
			content, readErr := io.ReadAll(part.Body)
			if readErr != nil {
				continue
			}
			if strings.HasPrefix(contentType, "text/plain") && plainText == "" {
				plainText = strings.TrimSpace(string(content))
			}
			if strings.HasPrefix(contentType, "text/html") && htmlText == "" {
				htmlText = strings.TrimSpace(string(content))
			}
		}
	}

	if plainText != "" {
		return plainText, true
	}
	if htmlText != "" {
		plain := strings.TrimSpace(emailHtmlToPlaintext(htmlText))
		if plain != "" {
			return plain, true
		}
	}
	return "", false
}

// extractFromEmailMultipart walks multipart MIME parts looking for text content.
// Handles nested multipart structures (multipart/mixed > multipart/alternative > text/plain).
func extractFromEmailMultipart(mr gomessage.MultipartReader) (string, error) {
	var plainText, htmlText string

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		contentType, _, _ := part.Header.ContentType()
		contentType = strings.ToLower(strings.TrimSpace(contentType))

		// Recurse into nested multipart
		if nestedMR := part.MultipartReader(); nestedMR != nil {
			nested, err := extractFromEmailMultipart(nestedMR)
			if err == nil && nested != "" {
				if plainText == "" {
					plainText = nested
				}
			}
			continue
		}

		content, err := io.ReadAll(part.Body)
		if err != nil {
			continue
		}

		switch {
		case strings.HasPrefix(contentType, "text/plain"):
			if plainText == "" {
				plainText = string(content)
			}
		case strings.HasPrefix(contentType, "text/html"):
			if htmlText == "" {
				htmlText = string(content)
			}
		}
	}

	if plainText != "" {
		return plainText, nil
	}
	if htmlText != "" {
		return emailHtmlToPlaintext(htmlText), nil
	}
	return "", nil
}

func extractBodyFromRawRFC822(bodyBytes []byte) string {
	raw := string(bodyBytes)
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	parts := strings.SplitN(raw, "\n\n", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(raw)
}

// emailHtmlToPlaintext converts HTML to plaintext.
func emailHtmlToPlaintext(htmlContent string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))
	var sb strings.Builder
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return strings.TrimSpace(sb.String())
		case html.TextToken:
			text := strings.TrimSpace(tokenizer.Token().Data)
			if text != "" {
				sb.WriteString(text)
				sb.WriteString(" ")
			}
		case html.StartTagToken:
			tn, _ := tokenizer.TagName()
			tag := string(tn)
			if tag == "br" || tag == "p" || tag == "div" || tag == "li" || tag == "tr" {
				sb.WriteString("\n")
			}
		}
	}
}
