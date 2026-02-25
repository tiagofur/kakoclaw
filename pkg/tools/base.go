package tools

import "context"

type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

// ContextualTool is an optional interface that tools can implement
// to receive the current message context (channel, chatID)
type ContextualTool interface {
	Tool
	SetContext(channel, chatID string)
}

// WorkspaceTool is an optional interface for tools that depend on a workspace path.
type WorkspaceTool interface {
	Tool
	SetWorkspace(workspace string)
}

// UserAwareTool is an optional interface for tools that need to filter data by user.
type UserAwareTool interface {
	Tool
	SetUserID(userID int64)
}

// UserConfigTool is for tools that need access to user config files.
// Unlike UserAwareTool (which only provides userID for data filtering),
// this interface provides the userUUID needed to locate config files.
type UserConfigTool interface {
	Tool
	SetUserContext(userID int64, userUUID string)
}

func ToolToSchema(tool Tool) map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"parameters":  tool.Parameters(),
		},
	}
}
