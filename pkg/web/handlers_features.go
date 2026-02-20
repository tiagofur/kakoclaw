package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sipeed/kakoclaw/pkg/storage"
)

// ==================== PROMPTS (F7 - Prompt Templates Library) ====================

// handlePrompts handles GET (list) and POST (create) for prompt templates
func (s *Server) handlePrompts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.store == nil {
		http.Error(w, "storage not available", http.StatusServiceUnavailable)
		return
	}

	switch r.Method {
	case http.MethodGet:
		prompts, err := s.store.ListPrompts()
		if err != nil {
			http.Error(w, "failed to list prompts: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if prompts == nil {
			prompts = []storage.Prompt{}
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"prompts": prompts})

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
		p, err := s.store.CreatePrompt(req.Title, req.Content, req.Description, req.Tags)
		if err != nil {
			http.Error(w, "failed to create prompt: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(p)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePromptAction handles PUT and DELETE for /api/v1/prompts/{id}
func (s *Server) handlePromptAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.store == nil {
		http.Error(w, "storage not available", http.StatusServiceUnavailable)
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
		if err := s.store.UpdatePrompt(id, req.Title, req.Content, req.Description, req.Tags); err != nil {
			http.Error(w, "failed to update prompt: "+err.Error(), http.StatusInternalServerError)
			return
		}
		p, _ := s.store.GetPrompt(id)
		_ = json.NewEncoder(w).Encode(p)

	case http.MethodDelete:
		if err := s.store.DeletePrompt(id); err != nil {
			http.Error(w, "failed to delete prompt: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})

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
		http.Error(w, "unsupported file type: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Truncate if very long to avoid overwhelming the prompt
	const maxChars = 50000
	truncated := false
	if len(extractedText) > maxChars {
		extractedText = extractedText[:maxChars]
		truncated = true
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
		// Basic PDF text extraction: look for text between BT/ET markers
		text := extractPDFText(data)
		if text == "" {
			return "", "", fmt.Errorf("could not extract text from PDF (binary or encrypted)")
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

// extractPDFText does basic extraction by scanning for readable text between BT/ET markers
func extractPDFText(data []byte) string {
	var buf strings.Builder
	content := string(data)
	
	// Look for strings inside parentheses in PDF streams (very basic heuristic)
	i := 0
	for i < len(content) {
		if i+1 < len(content) && content[i] == 'B' && content[i+1] == 'T' {
			// Inside a text block, find parenthesized strings
			end := strings.Index(content[i:], "ET")
			if end < 0 {
				break
			}
			block := content[i : i+end]
			for j := 0; j < len(block); j++ {
				if block[j] == '(' {
					start := j + 1
					for k := start; k < len(block); k++ {
						if block[k] == ')' && (k == 0 || block[k-1] != '\\') {
							text := block[start:k]
							// Filter out non-printable chars
							clean := strings.Map(func(r rune) rune {
								if r >= 32 && r < 127 {
									return r
								}
								return -1
							}, text)
							if len(clean) > 1 {
								buf.WriteString(clean)
								buf.WriteRune(' ')
							}
							j = k
							break
						}
					}
				}
			}
			buf.WriteRune('\n')
			i += end + 2
		} else {
			i++
		}
	}
	
	return strings.TrimSpace(buf.String())
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

// Avoid unused import errors
var _ = bytes.NewBuffer
