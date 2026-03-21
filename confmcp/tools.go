package confmcp

import (
	"context"
	"fmt"
)

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     ToolHandler            `json:"-"`
}

// ToolHandler handles tool execution
type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

// ToolRegistry manages MCP tools
type ToolRegistry struct {
	tools map[string]*Tool
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]*Tool),
	}
}

// Register adds a tool to the registry
func (r *ToolRegistry) Register(tool *Tool) error {
	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if tool.Handler == nil {
		return fmt.Errorf("tool handler cannot be nil")
	}

	r.tools[tool.Name] = tool
	return nil
}

// Unregister removes a tool from the registry
func (r *ToolRegistry) Unregister(name string) {
	delete(r.tools, name)
}

// Get retrieves a tool by name
func (r *ToolRegistry) Get(name string) (*Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tools
func (r *ToolRegistry) List() []*Tool {
	tools := make([]*Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Execute runs a tool with given arguments
func (r *ToolRegistry) Execute(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Handler(ctx, args)
}

// Count returns the number of registered tools
func (r *ToolRegistry) Count() int {
	return len(r.tools)
}

// Clear removes all tools from the registry
func (r *ToolRegistry) Clear() {
	r.tools = make(map[string]*Tool)
}

// Has checks if a tool is registered
func (r *ToolRegistry) Has(name string) bool {
	_, ok := r.tools[name]
	return ok
}
