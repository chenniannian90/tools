package confmcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

// HTTPTransport 实现 HTTP 传输层
type HTTPTransport struct {
	server      *http.Server
	mux         *http.ServeMux
	handler     MCPHandler
	connections map[string]context.CancelFunc
	connMutex   sync.RWMutex
}

// MCPHandler 处理 MCP 请求的接口
type MCPHandler interface {
	HandleMessage(ctx context.Context, message *JSONRPCMessage) *JSONRPCMessage
}

// NewHTTPTransport 创建 HTTP 传输层
func NewHTTPTransport() *HTTPTransport {
	return &HTTPTransport{
		connections: make(map[string]context.CancelFunc),
	}
}

// Start 启动 HTTP 服务器
func (t *HTTPTransport) Start(addr string, handler MCPHandler) error {
	t.mux = http.NewServeMux()
	t.handler = handler

	// 注册 MCP 端点
	t.mux.HandleFunc("/mcp", t.handleMCPRequest)

	t.server = &http.Server{
		Addr:    addr,
		Handler: t.mux,
	}

	logrus.Infof("HTTP 服务器启动: http://%s", addr)
	return t.server.ListenAndServe()
}

// Shutdown 关闭 HTTP 服务器
func (t *HTTPTransport) Shutdown(ctx context.Context) error {
	logrus.Info("HTTP 服务器关闭中...")
	return t.server.Shutdown(ctx)
}

// handleMCPRequest 处理 MCP 请求
func (t *HTTPTransport) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var request JSONRPCMessage
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logrus.Errorf("解码请求失败: %v", err)
		sendJSONRPCError(w, -32700, "Parse error", nil)
		return
	}

	// 处理请求
	response := t.handler.HandleMessage(r.Context(), &request)

	// 发送响应
	json.NewEncoder(w).Encode(response)
}

// sendJSONRPCError 发送 JSON-RPC 错误响应
func sendJSONRPCError(w http.ResponseWriter, code int, message string, id interface{}) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// HTTPClient HTTP MCP 客户端
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	requestID  int64
	mu         sync.Mutex
}

// NewHTTPClient 创建 HTTP 客户端
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{},
	}
}

// Connect 连接到服务器
func (c *HTTPClient) Connect(ctx context.Context) error {
	// 测试连接
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("健康检查失败: %d", resp.StatusCode)
	}

	logrus.Infof("HTTP 客户端连接成功: %s", c.baseURL)
	return nil
}

// Send 发送 JSON-RPC 请求
func (c *HTTPClient) Send(ctx context.Context, message *JSONRPCMessage) (*JSONRPCMessage, error) {
	c.mu.Lock()
	if message.ID == nil {
		c.requestID++
		message.ID = c.requestID
	}
	c.mu.Unlock()

	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/mcp", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response JSONRPCMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Close 关闭客户端
func (c *HTTPClient) Close() error {
	// HTTP 客户端不需要显式关闭
	return nil
}

// StreamingHTTPClient 支持流式 HTTP 的客户端
type StreamingHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	requestID  int64
	mu         sync.Mutex
}

// NewStreamingHTTPClient 创建流式 HTTP 客户端
func NewStreamingHTTPClient(baseURL string) *StreamingHTTPClient {
	return &StreamingHTTPClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// ConnectWithEvents 连接并监听事件
func (c *StreamingHTTPClient) ConnectWithEvents(ctx context.Context, eventHandler func(event string, data []byte)) error {
	resp, err := c.httpClient.Get(c.baseURL + "/events")
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("连接失败: %d", resp.StatusCode)
	}

	// 读取 SSE 事件
	scanner := bufio.NewScanner(resp.Body)
	var currentEvent string
	var currentData []byte

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) > 6 && line[:6] == "event:" {
			currentEvent = line[7:]
		} else if len(line) > 5 && line[:5] == "data:" {
			currentData = []byte(line[6:])
		} else if line == "" {
			// 空行表示事件结束
			if currentEvent != "" && eventHandler != nil {
				eventHandler(currentEvent, currentData)
			}
			currentEvent = ""
			currentData = nil
		}
	}

	return scanner.Err()
}

// Send 发送 JSON-RPC 请求
func (c *StreamingHTTPClient) Send(ctx context.Context, message *JSONRPCMessage) (*JSONRPCMessage, error) {
	c.mu.Lock()
	if message.ID == nil {
		c.requestID++
		message.ID = c.requestID
	}
	c.mu.Unlock()

	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/mcp", bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response JSONRPCMessage
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// SendWithRetry 发送带重试的请求
func (c *StreamingHTTPClient) SendWithRetry(ctx context.Context, message *JSONRPCMessage, retries int) (*JSONRPCMessage, error) {
	var lastErr error

	for i := 0; i <= retries; i++ {
		response, err := c.Send(ctx, message)
		if err == nil {
			return response, nil
		}

		lastErr = err
		if i < retries {
			logrus.Warningf("请求失败，重试 [%d/%d]: %v", i+1, retries, err)
		}
	}

	return nil, lastErr
}

// HTTPServerConfig HTTP 服务器配置
type HTTPServerConfig struct {
	Addr            string
	ReadTimeout     int
	WriteTimeout    int
	MaxHeaderBytes  int
	EnableCORS      bool
	EnableAuth      bool
	APIKey          string
}

// SetDefaults 设置默认值
func (c *HTTPServerConfig) SetDefaults() {
	if c.Addr == "" {
		c.Addr = ":3000"
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 10
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 10
	}
	if c.MaxHeaderBytes == 0 {
		c.MaxHeaderBytes = 1 << 20 // 1MB
	}
}

// StartHTTPServer 启动 HTTP 服务器（便捷函数）
func StartHTTPServer(config *MCP, httpConfig *HTTPServerConfig) error {
	if httpConfig == nil {
		httpConfig = &HTTPServerConfig{}
	}
	httpConfig.SetDefaults()

	transport := NewHTTPTransport()

	// 创建适配器
	handler := &httpHandlerAdapter{config: config}

	// 配置服务器
	if config.Port > 0 {
		httpConfig.Addr = fmt.Sprintf(":%d", config.Port)
	}

	return transport.Start(httpConfig.Addr, handler)
}

// httpHandlerAdapter 适配器，将 Server 转换为 MCPHandler
type httpHandlerAdapter struct {
	config *MCP
}

// HandleMessage 实现 MCPHandler 接口
func (a *httpHandlerAdapter) HandleMessage(ctx context.Context, message *JSONRPCMessage) *JSONRPCMessage {
	server := NewServer(a.config)

	// 注册默认工具（如果需要）
	if server.GetTools().Count() == 0 {
		// 可以在这里添加默认工具
	}

	// 处理请求
	return server.HandleRequest(ctx, *message)
}
