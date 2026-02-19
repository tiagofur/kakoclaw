package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/sipeed/kakoclaw/pkg/storage"
)

// KnowledgeTool allows the agent to search the knowledge base (RAG).
type KnowledgeTool struct {
	store *storage.Storage
}

func NewKnowledgeTool(store *storage.Storage) *KnowledgeTool {
	return &KnowledgeTool{store: store}
}

func (t *KnowledgeTool) Name() string {
	return "query_knowledge"
}

func (t *KnowledgeTool) Description() string {
	return "Search the local knowledge base for relevant information. The knowledge base contains user-uploaded documents (PDF, TXT, Markdown). Use this when the user asks about topics that might be covered by their documents."
}

func (t *KnowledgeTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query (keywords or natural language)",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return (1-20, default 5)",
				"minimum":     1.0,
				"maximum":     20.0,
			},
		},
		"required": []string{"query"},
	}
}

func (t *KnowledgeTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, ok := args["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return "", fmt.Errorf("query is required")
	}

	limit := 5
	if l, ok := args["limit"].(float64); ok && int(l) > 0 && int(l) <= 20 {
		limit = int(l)
	}

	results, err := t.store.SearchKnowledge(query, limit)
	if err != nil {
		// FTS5 MATCH can fail on invalid syntax — return a user-friendly message
		if strings.Contains(err.Error(), "fts5") || strings.Contains(err.Error(), "MATCH") {
			return fmt.Sprintf("Knowledge base search error (try simpler keywords): %v", err), nil
		}
		return "", fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return "No relevant documents found in the knowledge base for that query.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d relevant chunks from the knowledge base:\n\n", len(results)))
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("--- Result %d (from: %s) ---\n", i+1, r.DocumentName))
		sb.WriteString(r.Content)
		sb.WriteString("\n\n")
	}

	return sb.String(), nil
}
