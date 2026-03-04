package web

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/storage"
	"github.com/sipeed/makoclaw/pkg/tools"
)

// ==================== PROMPTS (F7 - Prompt Templates Library) ====================

// handlePrompts handles GET (list) and POST (create) for prompt templates
func (s *Server) handlePrompts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or storage unavailable", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		prompts, err := store.ListPrompts()
		if err != nil {
			logger.ErrorCF("web", "failed to list prompts", map[string]interface{}{"error": err.Error()})
			http.Error(w, "failed to list prompts", http.StatusInternalServerError)
			return
		}
		if prompts == nil {
			prompts = []storage.Prompt{}
		}
		writeJSONResponse(w, map[string]interface{}{"prompts": prompts})

	case http.MethodPost:
		var req struct {
			Title       string `json:"title"`
			Content     string `json:"content"`
			Description string `json:"description"`
			Tags        string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
			http.Error(w, "title and content are required", http.StatusBadRequest)
			return
		}
		if len(req.Title) > 255 {
			http.Error(w, "title too long (max 255 chars)", http.StatusBadRequest)
			return
		}
		if len(req.Description) > 2000 {
			http.Error(w, "description too long (max 2000 chars)", http.StatusBadRequest)
			return
		}
		if len(req.Content) > 100000 {
			http.Error(w, "content too long (max 100000 chars)", http.StatusBadRequest)
			return
		}
		p, err := store.CreatePrompt(req.Title, req.Content, req.Description, req.Tags)
		if err != nil {
			logger.ErrorCF("web", "failed to create prompt", map[string]interface{}{"error": err.Error()})
			http.Error(w, "failed to create prompt", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		writeJSONResponse(w, p)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePromptAction handles PUT and DELETE for /api/v1/prompts/{id}
func (s *Server) handlePromptAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	store, _, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized or storage unavailable", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/prompts/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid prompt ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		var req struct {
			Title       string `json:"title"`
			Content     string `json:"content"`
			Description string `json:"description"`
			Tags        string `json:"tags"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if len(req.Title) > 255 {
			http.Error(w, "title too long (max 255 chars)", http.StatusBadRequest)
			return
		}
		if len(req.Description) > 2000 {
			http.Error(w, "description too long (max 2000 chars)", http.StatusBadRequest)
			return
		}
		if len(req.Content) > 100000 {
			http.Error(w, "content too long (max 100000 chars)", http.StatusBadRequest)
			return
		}
		if err := store.UpdatePrompt(id, req.Title, req.Content, req.Description, req.Tags); err != nil {
			logger.ErrorCF("web", "failed to update prompt", map[string]interface{}{"id": id, "error": err.Error()})
			http.Error(w, "failed to update prompt", http.StatusInternalServerError)
			return
		}
		p, _ := store.GetPrompt(id)
		writeJSONResponse(w, p)

	case http.MethodDelete:
		if err := store.DeletePrompt(id); err != nil {
			logger.ErrorCF("web", "failed to delete prompt", map[string]interface{}{"id": id, "error": err.Error()})
			http.Error(w, "failed to delete prompt", http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, map[string]string{"status": "deleted"})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ==================== CHAT FILE ATTACHMENTS (F9 - File Upload in Chat) ====================

// handleChatAttachment handles POST /api/v1/chat/attachments
// Accepts multipart/form-data with a "file" field.
// Extracts text content from the file and returns it as a string to be injected into the chat.
func (s *Server) handleChatAttachment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to 10MB
	const maxUploadSize = 10 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large or invalid form (max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing 'file' field in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	name := header.Filename
	size := header.Size

	// Read file content
	data, err := io.ReadAll(io.LimitReader(file, maxUploadSize))
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	extractedText, mimeType, err := extractTextFromFile(data, ext, header.Filename)
	if err != nil {
		logger.ErrorCF("web", "unsupported file type", map[string]interface{}{"filename": header.Filename, "ext": ext, "error": err.Error()})
		http.Error(w, "unsupported file type", http.StatusBadRequest)
		return
	}

	// Truncate if very long to avoid overwhelming the prompt
	const maxChars = 50000
	truncated := false
	if len(extractedText) > maxChars {
		extractedText = extractedText[:maxChars]
		truncated = true
	}

	writeJSONResponse(w, map[string]interface{}{
		"name":      name,
		"size":      size,
		"mime_type": mimeType,
		"content":   extractedText,
		"truncated": truncated,
	})
}

// extractTextFromFile extracts plain text from common file formats
func extractTextFromFile(data []byte, ext, filename string) (string, string, error) {
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	switch ext {
	case ".txt", ".md", ".markdown", ".log", ".env", ".sh", ".bash", ".zsh":
		return sanitizeText(string(data)), mimeType, nil

	case ".json":
		// Pretty-print JSON for readability
		var v interface{}
		if err := json.Unmarshal(data, &v); err == nil {
			pretty, err := json.MarshalIndent(v, "", "  ")
			if err == nil {
				return string(pretty), "application/json", nil
			}
		}
		return sanitizeText(string(data)), "application/json", nil

	case ".csv":
		return sanitizeText(string(data)), "text/csv", nil

	case ".html", ".htm":
		// Strip HTML tags
		text := stripHTMLTags(string(data))
		return sanitizeText(text), "text/html", nil

	case ".xml", ".svg":
		return sanitizeText(string(data)), "text/xml", nil

	case ".yaml", ".yml":
		return sanitizeText(string(data)), "text/yaml", nil

	case ".py", ".go", ".js", ".ts", ".java", ".c", ".cpp", ".h", ".cs", ".rb", ".rs", ".php", ".swift", ".kt":
		return sanitizeText(string(data)), "text/plain", nil

	case ".pdf":
		text, err := tools.ExtractPDFTextFromBytes(data)
		if err != nil {
			return "", "", fmt.Errorf("PDF extraction failed: %w", err)
		}
		if text == "" {
			return "", "", fmt.Errorf("no readable text found in PDF (may be image-based or encrypted)")
		}
		return text, "application/pdf", nil

	default:
		// Try as UTF-8 text if it looks like text
		if isLikelyText(data) {
			return sanitizeText(string(data)), "text/plain", nil
		}
		return "", "", fmt.Errorf("binary file format '%s' not supported for text extraction", ext)
	}
}

func sanitizeText(s string) string {
	// Remove null bytes and normalize line endings
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSpace(s)
	return s
}

func stripHTMLTags(s string) string {
	inTag := false
	var buf strings.Builder
	for _, r := range s {
		if r == '<' {
			inTag = true
			buf.WriteRune(' ')
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}


func isLikelyText(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	// Check first 512 bytes for null bytes or high proportion of non-text chars
	sample := data
	if len(sample) > 512 {
		sample = sample[:512]
	}
	nonText := 0
	for _, b := range sample {
		if b == 0 || (b < 32 && b != '\n' && b != '\r' && b != '\t') {
			nonText++
		}
	}
	return float64(nonText)/float64(len(sample)) < 0.1
}
