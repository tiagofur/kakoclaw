package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sipeed/makoclaw/pkg/storage"
)

// StandingOrderTool manages persistent per-user instructions for the agent.
// It implements Tool and UserAwareTool (not WorkspaceTool — DB only).
type StandingOrderTool struct {
	storage *storage.Storage
	userID  int64
}

// NewStandingOrderTool creates a new StandingOrderTool backed by the given storage.
func NewStandingOrderTool(s *storage.Storage) (*StandingOrderTool, error) {
	if s == nil {
		return nil, fmt.Errorf("storage is nil")
	}
	return &StandingOrderTool{
		storage: s,
		userID:  1, // Default user for backward compatibility
	}, nil
}

// SetUserID implements the UserAwareTool interface.
func (t *StandingOrderTool) SetUserID(userID int64) {
	if userID > 0 {
		t.userID = userID
	}
}

func (t *StandingOrderTool) Name() string { return "standing_orders" }

func (t *StandingOrderTool) Description() string {
	return "Manage standing orders: persistent instructions that are injected into every agent session. Actions: list, create, update, delete, toggle."
}

func (t *StandingOrderTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type": "string",
				"enum": []string{"list", "create", "update", "delete", "toggle"},
			},
			"id":      map[string]interface{}{"type": "integer"},
			"content": map[string]interface{}{"type": "string"},
			"label":   map[string]interface{}{"type": "string"},
			"active":  map[string]interface{}{"type": "boolean"},
		},
		"required": []string{"action"},
	}
}

func (t *StandingOrderTool) Execute(_ context.Context, args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)

	switch action {
	case "list":
		orders, err := t.storage.ListStandingOrdersForUser(t.userID, false)
		if err != nil {
			return "", err
		}

		type orderView struct {
			ID        int64  `json:"id"`
			Content   string `json:"content"`
			Label     string `json:"label"`
			Active    bool   `json:"active"`
			CreatedAt string `json:"created_at"`
		}
		views := make([]orderView, 0, len(orders))
		for _, o := range orders {
			views = append(views, orderView{
				ID:        o.ID,
				Content:   o.Content,
				Label:     o.Label,
				Active:    o.Active,
				CreatedAt: o.CreatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}
		b, _ := json.Marshal(views)
		return string(b), nil

	case "create":
		content, ok := args["content"].(string)
		if !ok || content == "" {
			return "error: missing required parameter 'content'", nil
		}
		label, _ := args["label"].(string)

		id, err := t.storage.CreateStandingOrderForUser(t.userID, content, label)
		if err != nil {
			return fmt.Sprintf("error: %s", err.Error()), nil
		}
		return fmt.Sprintf("Standing order added (id=%d).", id), nil

	case "update":
		id, err := parseID(args)
		if err != nil {
			return "error: missing required parameter 'id'", nil
		}

		// Load current values so caller can update only some fields
		orders, listErr := t.storage.ListStandingOrdersForUser(t.userID, false)
		if listErr != nil {
			return "", listErr
		}

		var current *storage.StandingOrder
		for i := range orders {
			if orders[i].ID == id {
				current = &orders[i]
				break
			}
		}
		if current == nil {
			return fmt.Sprintf("error: standing order %d not found", id), nil
		}

		newContent := current.Content
		newLabel := current.Label

		if c, ok := args["content"].(string); ok {
			newContent = c
		}
		if l, ok := args["label"].(string); ok {
			newLabel = l
		}

		if err := t.storage.UpdateStandingOrderForUser(t.userID, id, newContent, newLabel); err != nil {
			return fmt.Sprintf("error: %s", err.Error()), nil
		}

		// Handle active toggle through update (spec says update takes optional `active`)
		if activeVal, ok := args["active"]; ok {
			if active, ok := activeVal.(bool); ok {
				// Only toggle if it differs from current
				if active != current.Active {
					if err := t.storage.ToggleStandingOrderForUser(t.userID, id); err != nil {
						return fmt.Sprintf("error toggling active: %s", err.Error()), nil
					}
				}
			}
		}

		return "Updated.", nil

	case "delete":
		id, err := parseID(args)
		if err != nil {
			return "error: missing required parameter 'id'", nil
		}
		if err := t.storage.DeleteStandingOrderForUser(t.userID, id); err != nil {
			return "", err
		}
		return "Removed.", nil

	case "toggle":
		id, err := parseID(args)
		if err != nil {
			return "error: missing required parameter 'id'", nil
		}
		if err := t.storage.ToggleStandingOrderForUser(t.userID, id); err != nil {
			return "", err
		}
		return "Toggled.", nil

	default:
		if action == "" {
			return "error: missing required parameter 'action'", nil
		}
		return fmt.Sprintf("error: unknown action '%s'", action), nil
	}
}

// parseID extracts an int64 id from the args map (handles float64 and string).
func parseID(args map[string]interface{}) (int64, error) {
	switch v := args["id"].(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid id: %w", err)
		}
		return id, nil
	default:
		return 0, fmt.Errorf("id not provided")
	}
}
