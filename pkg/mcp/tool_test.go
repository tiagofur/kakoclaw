package mcp

import (
	"testing"
)

func TestMCPTool_Name(t *testing.T) {
	client := &Client{name: "myserver"}
	info := MCPToolInfo{
		Name:        "mytool",
		Description: "A test tool",
	}
	tool := NewMCPTool(client, info)

	got := tool.Name()
	want := "mcp_myserver_mytool"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestMCPTool_Name_VariousPrefixes(t *testing.T) {
	tests := []struct {
		serverName string
		toolName   string
		want       string
	}{
		{"server1", "tool1", "mcp_server1_tool1"},
		{"my-server", "my-tool", "mcp_my-server_my-tool"},
		{"", "tool", "mcp__tool"},
		{"srv", "", "mcp_srv_"},
	}

	for _, tt := range tests {
		t.Run(tt.serverName+"_"+tt.toolName, func(t *testing.T) {
			client := &Client{name: tt.serverName}
			info := MCPToolInfo{Name: tt.toolName}
			tool := NewMCPTool(client, info)

			got := tool.Name()
			if got != tt.want {
				t.Errorf("Name() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMCPTool_Description(t *testing.T) {
	client := &Client{name: "myserver"}
	info := MCPToolInfo{
		Name:        "mytool",
		Description: "Does something useful",
	}
	tool := NewMCPTool(client, info)

	got := tool.Description()
	want := "[MCP:myserver] Does something useful"
	if got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestMCPTool_Description_Empty(t *testing.T) {
	client := &Client{name: "myserver"}
	info := MCPToolInfo{
		Name:        "mytool",
		Description: "",
	}
	tool := NewMCPTool(client, info)

	got := tool.Description()
	want := `[MCP:myserver] Tool from MCP server "myserver"`
	if got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}
}

func TestMCPTool_Parameters(t *testing.T) {
	client := &Client{name: "myserver"}
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path",
			},
		},
		"required": []interface{}{"path"},
	}
	info := MCPToolInfo{
		Name:        "mytool",
		InputSchema: schema,
	}
	tool := NewMCPTool(client, info)

	params := tool.Parameters()
	if params == nil {
		t.Fatal("Parameters() returned nil, expected schema map")
	}

	if params["type"] != "object" {
		t.Errorf("Parameters()[\"type\"] = %v, want \"object\"", params["type"])
	}

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Parameters()[\"properties\"] is not a map")
	}

	if _, ok := props["path"]; !ok {
		t.Error("Parameters() properties missing \"path\" key")
	}
}

func TestMCPTool_Parameters_Empty(t *testing.T) {
	client := &Client{name: "myserver"}
	info := MCPToolInfo{
		Name:        "mytool",
		InputSchema: nil,
	}
	tool := NewMCPTool(client, info)

	params := tool.Parameters()
	if params == nil {
		t.Fatal("Parameters() returned nil, expected fallback schema")
	}

	// When InputSchema is nil, the fallback is an empty object schema
	if params["type"] != "object" {
		t.Errorf("Parameters()[\"type\"] = %v, want \"object\"", params["type"])
	}

	props, ok := params["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Parameters()[\"properties\"] is not a map")
	}

	if len(props) != 0 {
		t.Errorf("Expected empty properties map, got %d entries", len(props))
	}
}
