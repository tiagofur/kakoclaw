package devmemory

import (
	"context"
	"fmt"
	"strings"
)

// Inject searches for relevant memories and formats them as a markdown block.
// Format is suitable for appending to system prompts before calling the LLM.
func (s *Store) Inject(ctx context.Context, query string, limit int) (string, error) {
	memories, err := s.Search(ctx, query, limit)
	if err != nil {
		return "", err
	}

	if len(memories) == 0 {
		return "", nil
	}

	var b strings.Builder
	b.WriteString("\n## 🧠 Relevant Developer Workspace Memories\n\n")
	b.WriteString("The following past context might be relevant to your current task:\n\n")

	for _, m := range memories {
		b.WriteString(fmt.Sprintf("- [%s] %s\n", m.CreatedAt.Format("2006-01-02"), m.Content))
	}

	b.WriteString("\nUse this context to inform your code generation and decision making.\n")

	return b.String(), nil
}
