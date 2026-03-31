package tools

import (
	"context"
	"fmt"

	"github.com/sipeed/makoclaw/pkg/channels"
)

// InteractiveMessageTool sends interactive messages (buttons, reactions) to channels.
type InteractiveMessageTool struct {
	action   channels.ChannelAction
	enabled  bool
	endpoint string
	channel  string
	chatID   string
}

// NewInteractiveMessageTool creates a new InteractiveMessageTool.
// Pass enabled=false or empty endpoint to disable the tool.
func NewInteractiveMessageTool(action channels.ChannelAction, enabled bool, endpoint string) *InteractiveMessageTool {
	return &InteractiveMessageTool{
		action:   action,
		enabled:  enabled,
		endpoint: endpoint,
	}
}

func (t *InteractiveMessageTool) Name() string {
	return "interactive_message"
}

func (t *InteractiveMessageTool) Description() string {
	return "Send an interactive message with buttons or approval actions to a channel. Use for approval gates, confirmations, and user choices."
}

func (t *InteractiveMessageTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "The message text to display",
			},
			"actions": map[string]interface{}{
				"type":        "array",
				"description": "List of action buttons",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"label": map[string]interface{}{
							"type":        "string",
							"description": "Button label text",
						},
						"value": map[string]interface{}{
							"type":        "string",
							"description": "Value returned when button is clicked",
						},
						"style": map[string]interface{}{
							"type":        "string",
							"description": "Button style: primary or danger",
						},
					},
				},
			},
			"channel": map[string]interface{}{
				"type":        "string",
				"description": "Target channel (discord, slack, etc.)",
			},
			"chat_id": map[string]interface{}{
				"type":        "string",
				"description": "Target chat/channel ID",
			},
		},
		"required": []string{"message"},
	}
}

func (t *InteractiveMessageTool) SetContext(channel, chatID string) {
	t.channel = channel
	t.chatID = chatID
}

func (t *InteractiveMessageTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if !t.enabled {
		return "", fmt.Errorf("channel actions are disabled. Set channel_actions.enabled = true to use this feature.")
	}

	if t.endpoint == "" {
		return "", fmt.Errorf("channel actions require a public HTTPS interactions_endpoint_url")
	}

	message, _ := args["message"].(string)
	if message == "" {
		return "", fmt.Errorf("message is required")
	}

	channel, _ := args["channel"].(string)
	chatID, _ := args["chat_id"].(string)

	if channel == "" {
		channel = t.channel
	}
	if chatID == "" {
		chatID = t.chatID
	}

	if t.action == nil {
		return "", fmt.Errorf("no channel action handler available for channel %q", channel)
	}

	// Parse actions from args
	var actionConfigs []channels.ActionConfig
	if rawActions, ok := args["actions"].([]interface{}); ok {
		for _, ra := range rawActions {
			if actionMap, ok := ra.(map[string]interface{}); ok {
				actionConfigs = append(actionConfigs, channels.ActionConfig{
					Label: stringOrDefault(actionMap, "label", ""),
					Value: stringOrDefault(actionMap, "value", ""),
					Style: stringOrDefault(actionMap, "style", "primary"),
				})
			}
		}
	}

	req := channels.InteractiveMessageRequest{
		Message: message,
		Actions: actionConfigs,
		ChatID:  chatID,
	}

	messageID, err := t.action.SendInteractiveMessage(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to send interactive message: %w", err)
	}

	return fmt.Sprintf("Interactive message sent (ID: %s). Waiting for user response.", messageID), nil
}

func stringOrDefault(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultVal
}
