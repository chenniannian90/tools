package confmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-courier/envconf"
	"github.com/sirupsen/logrus"
)

// MCP 配置结构（用于客户端）
type MCP struct {
	Port int `env:",opt,expose"`

	// 协议类型：stdio、http 或 sse
	Protocol string `env:""`

	// 服务器名称
	Name string `env:""`

	// 内部字段
	Initialized bool   `env:"-"`
	retry       *Retry `env:"-"`
}

// SetDefaults 设置默认配置值
func (m *MCP) SetDefaults() {
	if m.Protocol == "" {
		m.Protocol = "stdio"
	}

	// 如果是 SSE 或 HTTP 协议，设置默认端口
	if (m.Protocol == "sse" || m.Protocol == "http") && m.Port == 0 {
		m.Port = 3000
	}

	// 初始化重试配置
	if m.retry == nil {
		m.retry = &Retry{}
		m.retry.SetDefaults()
	}
}

// Init 初始化 MCP
func (m *MCP) Init() {
	if !m.Initialized {
		m.SetDefaults()
		m.Initialized = true
	}
}

// GetAddress 获取服务地址
func (m *MCP) GetAddress() string {
	if m.Port > 0 {
		return ":" + strconv.Itoa(m.Port)
	}
	return "stdio"
}

// GetServerInfo 获取服务器信息
func (m *MCP) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:     m.Name,
		Version:  "1.0.0",
		Protocol: m.Protocol,
		Address:  m.GetAddress(),
		Capabilities: Capabilities{
			Tools: true,
		},
	}
}

// LivenessCheck 健康检查
func (m *MCP) LivenessCheck() map[string]string {
	status := map[string]string{}

	if m.Initialized {
		status[m.GetAddress()] = "ok"
	} else {
		status[m.GetAddress()] = "not initialized"
	}

	return status
}

// SetRetryConfig 设置重试配置
func (m *MCP) SetRetryConfig(repeats int, interval time.Duration) {
	if m.retry == nil {
		m.retry = &Retry{}
	}
	m.retry.Repeats = repeats
	m.retry.Interval = envconf.Duration(interval)
}

// connect 建立连接（带重试）
func (m *MCP) connect() error {
	if m.retry == nil {
		m.retry = &Retry{}
		m.retry.SetDefaults()
	}

	return m.retry.Do(func() error {
		switch m.Protocol {
		case "stdio":
			return nil // stdio 不需要连接
		case "sse":
			return fmt.Errorf("SSE 协议暂未实现")
		default:
			return fmt.Errorf("不支持的协议: %s", m.Protocol)
		}
	})
}

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
				"tools": map[string]interface{}{
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
