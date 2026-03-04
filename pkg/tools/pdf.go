package tools

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
)

// PDFTool extracts text from PDF files.
type PDFTool struct {
	workspace string
	restrict  bool
}

func NewPDFTool(workspace string, restrict bool) *PDFTool {
	return &PDFTool{workspace: workspace, restrict: restrict}
}

func (t *PDFTool) Name() string        { return "read_pdf" }
func (t *PDFTool) Description() string {
	return "Extract text content from a PDF file. Supports page ranges like '1-5', '3', or '1,3,5'."
}

func (t *PDFTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the PDF file",
			},
			"pages": map[string]interface{}{
				"type":        "string",
				"description": "Optional page range: '1-5', '3', '1,3,5'. Omit for all pages.",
			},
		},
		"required": []string{"file_path"},
	}
}

func (t *PDFTool) SetWorkspace(workspace string) {
	t.workspace = workspace
}

func (t *PDFTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	absPath, err := validatePath(filePath, t.workspace, t.restrict)
	if err != nil {
		return "", err
	}

	pagesArg, _ := args["pages"].(string)
	wantPages, err := parsePageRange(pagesArg)
	if err != nil {
		return "", fmt.Errorf("invalid page range %q: %w", pagesArg, err)
	}

	text, err := ExtractPDFText(absPath, wantPages)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("no readable text extracted from PDF (may be image-based or encrypted)")
	}

	return text, nil
}

// ExtractPDFText reads a PDF from disk and returns its text content.
// If wantPages is non-nil, only those 1-based page numbers are extracted.
func ExtractPDFText(filePath string, wantPages map[int]bool) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening PDF: %w", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	if totalPages == 0 {
		return "", fmt.Errorf("PDF has no pages")
	}

	var buf strings.Builder
	for i := 1; i <= totalPages; i++ {
		if wantPages != nil && !wantPages[i] {
			continue
		}

		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			buf.WriteString(fmt.Sprintf("\n--- Page %d (extraction error: %v) ---\n", i, err))
			continue
		}

		if strings.TrimSpace(text) != "" {
			buf.WriteString(fmt.Sprintf("\n--- Page %d ---\n", i))
			buf.WriteString(text)
		}
	}

	return strings.TrimSpace(buf.String()), nil
}

// ExtractPDFTextFromBytes reads a PDF from bytes and returns its text content.
func ExtractPDFTextFromBytes(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("reading PDF: %w", err)
	}

	totalPages := r.NumPage()
	if totalPages == 0 {
		return "", fmt.Errorf("PDF has no pages")
	}

	var buf strings.Builder
	for i := 1; i <= totalPages; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}

		if strings.TrimSpace(text) != "" {
			if buf.Len() > 0 {
				buf.WriteString("\n")
			}
			buf.WriteString(text)
		}
	}

	return strings.TrimSpace(buf.String()), nil
}

// parsePageRange parses page specifications like "1-5", "3", "1,3,5".
// Returns nil map if input is empty (meaning all pages).
func parsePageRange(spec string) (map[int]bool, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, nil
	}

	pages := make(map[int]bool)
	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid page number %q", bounds[0])
			}
			end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid page number %q", bounds[1])
			}
			if start < 1 || end < start {
				return nil, fmt.Errorf("invalid range %d-%d", start, end)
			}
			for i := start; i <= end; i++ {
				pages[i] = true
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid page number %q", part)
			}
			if n < 1 {
				return nil, fmt.Errorf("page number must be >= 1")
			}
			pages[n] = true
		}
	}

	return pages, nil
}
