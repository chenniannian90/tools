package confmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-courier/envconf"
	"github.com/sirupsen/logrus"
)

// Server represents an MCP server instance (tools only)
type Server struct {
	// 公开配置字段（类似 http.Server）
	Name     string // 服务器名称
	Protocol string // 协议类型: "stdio", "http", "sse"
	Port     int    // 端口（HTTP/SSE 时使用）

	// 内部字段
	tools   *ToolRegistry
	running bool
	retry   *Retry
	Initialized bool
}

// NewServer creates a new MCP server with default configuration
func NewServer() *Server {
	return &Server{
		Protocol: "stdio", // 默认协议
		tools:    NewToolRegistry(),
	}
}

// SetDefaults 设置默认配置值
func (s *Server) SetDefaults() {
	if s.Protocol == "" {
		s.Protocol = "stdio"
	}

	// 如果是 SSE 或 HTTP 协议，设置默认端口
	if (s.Protocol == "sse" || s.Protocol == "http") && s.Port == 0 {
		s.Port = 3000
	}

	// 初始化重试配置
	if s.retry == nil {
		s.retry = &Retry{}
		s.retry.SetDefaults()
	}
}

// Init 初始化 Server
func (s *Server) Init() {
	if !s.Initialized {
		s.SetDefaults()
		s.Initialized = true
	}
}

// GetAddress 获取服务地址
func (s *Server) GetAddress() string {
	if s.Port > 0 {
		return fmt.Sprintf(":%d", s.Port)
	}
	return "stdio"
}

// GetServerInfo 获取服务器信息
func (s *Server) GetServerInfo() ServerInfo {
	s.Init()

	return ServerInfo{
		Name:     s.Name,
		Version:  "1.0.0",
		Protocol: s.Protocol,
		Address:  s.GetAddress(),
		Capabilities: Capabilities{
			Tools: true,
		},
	}
}

// LivenessCheck 健康检查
func (s *Server) LivenessCheck() map[string]string {
	s.Init()

	status := make(map[string]string)
	addr := s.GetAddress()

	if addr == "stdio" {
		status["stdio"] = "ok"
	} else {
		status[addr] = "ok"
	}

	return status
}

// SetRetryConfig 设置重试配置
func (s *Server) SetRetryConfig(repeats int, interval time.Duration) {
	s.retry = &Retry{
		Repeats:  repeats,
		Interval: envconf.Duration(interval),
	}
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	s.Init()

	logrus.Infof("Starting MCP server: %s", s.Name)

	switch s.Protocol {
	case "stdio":
		return s.startStdio(ctx)
	case "http":
		return s.startHTTP(ctx)
	case "sse":
		return s.startSSE(ctx)
	default:
		return fmt.Errorf("unsupported protocol: %s", s.Protocol)
	}
}

// Serve starts an HTTP server with the provided tools
// This is a convenience method for quickly setting up an HTTP MCP server
func (s *Server) Serve(tools []*Tool) error {
	// 自动注册所有工具
	for _, tool := range tools {
		if err := s.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name, err)
		}
	}

	// 创建 HTTP 服务器
	mux := http.NewServeMux()

	// MCP 端点
	mux.HandleFunc("/mcp", s.mcpHandler())

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := s.LivenessCheck()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"server":  s.Name,
			"checks":  status,
		})
	})

	// 工具列表端点
	mux.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tools := s.GetTools().List()

		toolList := make([]map[string]interface{}, 0, len(tools))
		for _, tool := range tools {
			toolList = append(toolList, map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": tool.InputSchema,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"tools": toolList,
		})
	})

	// 确定监听地址
	addr := s.GetAddress()
	if addr == "stdio" {
		addr = ":3000" // 默认 HTTP 端口
	}

	// 打印启动信息
	logrus.Infof("Starting HTTP MCP server on %s", addr)
	logrus.Infof("MCP endpoint: http://%s/mcp", addr)
	logrus.Infof("Health check: http://%s/health", addr)
	logrus.Infof("Tools list: http://%s/tools", addr)
	logrus.Infof("Registered %d tools", len(tools))

	// 启动 HTTP 服务器
	if err := http.ListenAndServe(addr, mux); err != nil {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// mcpHandler creates the MCP HTTP handler
func (s *Server) mcpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request JSONRPCMessage
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			logrus.Errorf("Failed to decode request: %v", err)
			s.sendJSONRPCError(w, -32700, "Parse error", nil)
			return
		}

		response := s.HandleRequest(r.Context(), request)
		json.NewEncoder(w).Encode(response)
	}
}

// sendJSONRPCError sends a JSON-RPC error response
func (s *Server) sendJSONRPCError(w http.ResponseWriter, code int, message string, id interface{}) {
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

// startHTTP starts HTTP-based server
func (s *Server) startHTTP(ctx context.Context) error {
	mux := http.NewServeMux()

	// MCP 端点
	mux.HandleFunc("/mcp", s.mcpHandler())

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := s.LivenessCheck()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"server":  s.Name,
			"checks":  status,
		})
	})

	// 工具列表端点
	mux.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tools := s.GetTools().List()

		toolList := make([]map[string]interface{}, 0, len(tools))
		for _, tool := range tools {
			toolList = append(toolList, map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": tool.InputSchema,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"tools": toolList,
		})
	})

	// 确定监听地址
	addr := s.GetAddress()
	if addr == "stdio" {
		addr = ":3000"
	}

	// 打印启动信息
	logrus.Infof("Starting HTTP MCP server on %s", addr)
	logrus.Infof("MCP endpoint: http://%s/mcp", addr)
	logrus.Infof("Health check: http://%s/health", addr)
	logrus.Infof("Tools list: http://%s/tools", addr)
	logrus.Infof("Registered %d tools", s.GetTools().Count())

	s.running = true

	// 启动 HTTP 服务器
	if err := http.ListenAndServe(addr, mux); err != nil {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// startSSE starts SSE-based server
func (s *Server) startSSE(ctx context.Context) error {
	mux := http.NewServeMux()

	// SSE 端点 - 用于服务器推送事件
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		// 设置 SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// 发送连接成功消息
		fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
		w.(http.Flusher).Flush()

		// 保持连接活跃
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				logrus.Info("SSE client disconnected")
				return
			case <-ticker.C:
				// 发送心跳保持连接
				fmt.Fprintf(w, ": heartbeat\n\n")
				w.(http.Flusher).Flush()
			case <-ctx.Done():
				return
			}
		}
	})

	// MCP 端点 - 用于接收 JSON-RPC 请求
	mux.HandleFunc("/mcp", s.mcpHandler())

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := s.LivenessCheck()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"server":  s.Name,
			"checks":  status,
		})
	})

	// 工具列表端点
	mux.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tools := s.GetTools().List()

		toolList := make([]map[string]interface{}, 0, len(tools))
		for _, tool := range tools {
			toolList = append(toolList, map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": tool.InputSchema,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"tools": toolList,
		})
	})

	// 确定监听地址
	addr := s.GetAddress()
	if addr == "stdio" {
		addr = ":3000"
	}

	// 打印启动信息
	logrus.Infof("Starting SSE MCP server on %s", addr)
	logrus.Infof("SSE endpoint: http://%s/sse", addr)
	logrus.Infof("MCP endpoint: http://%s/mcp", addr)
	logrus.Infof("Health check: http://%s/health", addr)
	logrus.Infof("Tools list: http://%s/tools", addr)
	logrus.Infof("Registered %d tools", s.GetTools().Count())

	s.running = true

	// 启动 HTTP 服务器
	if err := http.ListenAndServe(addr, mux); err != nil {
		return fmt.Errorf("SSE server error: %w", err)
	}

	return nil
}

// startStdio starts stdio-based server
func (s *Server) startStdio(ctx context.Context) error {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	s.running = true
	logrus.Info("MCP server listening on stdio")

	for s.running {
		var request JSONRPCMessage
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				logrus.Info("Connection closed by client")
				return nil
			}
			logrus.Errorf("Failed to decode request: %v", err)
			continue
		}

		response := s.HandleRequest(ctx, request)
		if response != nil {
			if err := encoder.Encode(response); err != nil {
				logrus.Errorf("Failed to encode response: %v", err)
			}
		}
	}

	return nil
}

// Stop stops the server
func (s *Server) Stop() {
	s.running = false
	logrus.Info("MCP server stopped")
}

// HandleRequest handles incoming JSON-RPC requests
func (s *Server) HandleRequest(ctx context.Context, request JSONRPCMessage) *JSONRPCMessage {
	logrus.Debugf("Received request: %s", request.Method)

	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "initialized":
		return nil // Notification, no response needed
	case "shutdown":
		return s.handleShutdown(request)
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolsCall(ctx, request)
	default:
		return &JSONRPCMessage{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", request.Method),
			},
		}
	}
}

// handleInitialize handles initialize request
func (s *Server) handleInitialize(request JSONRPCMessage) *JSONRPCMessage {
	logrus.Info("Handling initialize request")

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]interface{}{
				"name":    s.Name,
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
		},
	}
}

// handleShutdown handles shutdown request
func (s *Server) handleShutdown(request JSONRPCMessage) *JSONRPCMessage {
	logrus.Info("Handling shutdown request")
	s.running = false

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  map[string]interface{}{},
	}
}

// handleToolsList handles tools/list request
func (s *Server) handleToolsList(request JSONRPCMessage) *JSONRPCMessage {
	tools := s.tools.List()

	// Convert to serializable format
	toolsList := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		toolMap := map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
		toolsList = append(toolsList, toolMap)
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"tools": toolsList,
		},
	}
}

// handleToolsCall handles tools/call request
func (s *Server) handleToolsCall(ctx context.Context, request JSONRPCMessage) *JSONRPCMessage {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return errorResponse(request.ID, -32602, "Invalid params")
	}

	name, _ := params["name"].(string)
	args, _ := params["arguments"].(map[string]interface{})

	logrus.Infof("Calling tool: %s with args: %+v", name, args)

	result, err := s.tools.Execute(ctx, name, args)
	if err != nil {
		logrus.Errorf("Tool execution failed: %v", err)
		return errorResponse(request.ID, -32000, err.Error())
	}

	// Format result as content
	var content interface{}
	if strResult, ok := result.(string); ok {
		content = []map[string]interface{}{
			{
				"type": "text",
				"text": strResult,
			},
		}
	} else {
		// Convert to JSON
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return errorResponse(request.ID, -32000, "Failed to serialize result")
		}
		content = []map[string]interface{}{
			{
				"type": "text",
				"text": string(jsonBytes),
			},
		}
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"content": content,
		},
	}
}

// RegisterTool registers a tool with the server
func (s *Server) RegisterTool(tool *Tool) error {
	return s.tools.Register(tool)
}

// GetTools returns the tool registry
func (s *Server) GetTools() *ToolRegistry {
	return s.tools
}

// errorResponse creates an error response
func errorResponse(id interface{}, code int, message string) *JSONRPCMessage {
	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
}
