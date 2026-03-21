package confmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

// Client represents an MCP client that connects to MCP servers
type Client struct {
	config      *MCP
	transport   Transport
	requestID   int64
	mu          sync.Mutex
	initialized bool
}

// Transport defines the interface for MCP transport layers
type Transport interface {
	Send(ctx context.Context, msg *JSONRPCMessage) (*JSONRPCMessage, error)
	Close() error
}

// StdioTransport implements stdio-based transport
type StdioTransport struct {
	encoder *json.Encoder
	decoder *json.Decoder
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		encoder: json.NewEncoder(os.Stdout),
		decoder: json.NewDecoder(os.Stdin),
	}
}

// Send sends a message and waits for response
func (t *StdioTransport) Send(ctx context.Context, msg *JSONRPCMessage) (*JSONRPCMessage, error) {
	if err := t.encoder.Encode(msg); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	var response JSONRPCMessage
	if err := t.decoder.Decode(&response); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("connection closed")
		}
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return &response, nil
}

// Close closes the transport
func (t *StdioTransport) Close() error {
	return nil
}

// NewClient creates a new MCP client
func NewClient(config *MCP) *Client {
	if config == nil {
		config = &MCP{}
	}

	return &Client{
		config: config,
	}
}

// Connect establishes connection to MCP server
func (c *Client) Connect(ctx context.Context) error {
	c.config.Init()

	switch c.config.Protocol {
	case "stdio":
		c.transport = NewStdioTransport()
	case "sse":
		return fmt.Errorf("SSE protocol not yet implemented")
	default:
		return fmt.Errorf("unsupported protocol: %s", c.config.Protocol)
	}

	// Initialize session
	if err := c.initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	c.initialized = true
	logrus.Info("MCP client connected and initialized")
	return nil
}

// initialize sends initialize request to server
func (c *Client) initialize(ctx context.Context) error {
	req := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
			},
			"clientInfo": map[string]interface{}{
				"name":    "confmcp-client",
				"version": "1.0.0",
			},
		},
	}

	resp, err := c.transport.Send(ctx, req)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("initialize failed: %s", resp.Error.Message)
	}

	logrus.Infof("Initialized with server: %+v", resp.Result)
	return nil
}

// nextRequestID generates next request ID
func (c *Client) nextRequestID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requestID++
	return c.requestID
}

// Close closes the client connection
func (c *Client) Close() error {
	if c.transport != nil {
		return c.transport.Close()
	}
	return nil
}

// ListTools lists available tools from server
func (c *Client) ListTools(ctx context.Context) ([]*Tool, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	req := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "tools/list",
	}

	resp, err := c.transport.Send(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to list tools: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	toolsInterface, ok := result["tools"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid tools format")
	}

	var tools []*Tool
	for _, toolInterface := range toolsInterface {
		toolMap, ok := toolInterface.(map[string]interface{})
		if !ok {
			continue
		}

		tool := &Tool{
			Name:        getString(toolMap, "name"),
			Description: getString(toolMap, "description"),
			InputSchema: getMap(toolMap, "inputSchema"),
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

// CallTool calls a tool on the server
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	req := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": args,
		},
	}

	resp, err := c.transport.Send(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tool call failed: %s", resp.Error.Message)
	}

	return resp.Result, nil
}

// ListResources lists available resources from server
func (c *Client) ListResources(ctx context.Context) ([]*Resource, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	req := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "resources/list",
	}

	resp, err := c.transport.Send(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to list resources: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	resourcesInterface, ok := result["resources"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid resources format")
	}

	var resources []*Resource
	for _, resInterface := range resourcesInterface {
		resMap, ok := resInterface.(map[string]interface{})
		if !ok {
			continue
		}

		resource := &Resource{
			URI:         getString(resMap, "uri"),
			Name:        getString(resMap, "name"),
			Description: getString(resMap, "description"),
			MimeType:    getString(resMap, "mimeType"),
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// ReadResource reads a resource from server
func (c *Client) ReadResource(ctx context.Context, uri string) (*ResourceContent, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	req := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "resources/read",
		Params: map[string]interface{}{
			"uri": uri,
		},
	}

	resp, err := c.transport.Send(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to read resource: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	contentsInterface, ok := result["contents"].([]interface{})
	if !ok || len(contentsInterface) == 0 {
		return nil, fmt.Errorf("invalid contents format")
	}

	contentMap, ok := contentsInterface[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid content format")
	}

	content := &ResourceContent{
		URI:      getString(contentMap, "uri"),
		MimeType: getString(contentMap, "mimeType"),
		Text:     getString(contentMap, "text"),
	}

	return content, nil
}

// ListPrompts lists available prompts from server
func (c *Client) ListPrompts(ctx context.Context) ([]*Prompt, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client not initialized")
	}

	req := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "prompts/list",
	}

	resp, err := c.transport.Send(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("failed to list prompts: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	promptsInterface, ok := result["prompts"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid prompts format")
	}

	var prompts []*Prompt
	for _, promptInterface := range promptsInterface {
		promptMap, ok := promptInterface.(map[string]interface{})
		if !ok {
			continue
		}

		prompt := &Prompt{
			Name:        getString(promptMap, "name"),
			Description: getString(promptMap, "description"),
		}

		if argsInterface, ok := promptMap["arguments"].([]interface{}); ok {
			for _, argInterface := range argsInterface {
				argMap, ok := argInterface.(map[string]interface{})
				if !ok {
					continue
				}

				arg := PromptArgument{
					Name:        getString(argMap, "name"),
					Description: getString(argMap, "description"),
				}
				if required, ok := argMap["required"].(bool); ok {
					arg.Required = required
				}
				prompt.Arguments = append(prompt.Arguments, arg)
			}
		}

		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// GetPrompt gets a prompt from server
func (c *Client) GetPrompt(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	if !c.initialized {
		return "", fmt.Errorf("client not initialized")
	}

	params := map[string]interface{}{
		"name": name,
	}
	if args != nil {
		params["arguments"] = args
	}

	req := &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "prompts/get",
		Params:  params,
	}

	resp, err := c.transport.Send(ctx, req)
	if err != nil {
		return "", err
	}

	if resp.Error != nil {
		return "", fmt.Errorf("failed to get prompt: %s", resp.Error.Message)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	messagesInterface, ok := result["messages"].([]interface{})
	if !ok || len(messagesInterface) == 0 {
		return "", fmt.Errorf("invalid messages format")
	}

	messageMap, ok := messagesInterface[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid message format")
	}

	content := getString(messageMap, "content")
	return content, nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if val, ok := m[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}
