package confmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

// Server represents an MCP server instance
type Server struct {
	config    *MCP
	tools     *ToolRegistry
	resources *ResourceRegistry
	prompts   *PromptRegistry
	running   bool
}

// NewServer creates a new MCP server
func NewServer(config *MCP) *Server {
	if config == nil {
		config = &MCP{}
	}

	return &Server{
		config:    config,
		tools:     NewToolRegistry(),
		resources: NewResourceRegistry(),
		prompts:   NewPromptRegistry(),
	}
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	s.config.Init()

	logrus.Infof("Starting MCP server: %s", s.config.Name)

	switch s.config.Protocol {
	case "stdio":
		return s.startStdio(ctx)
	case "sse":
		return fmt.Errorf("SSE protocol not yet implemented")
	default:
		return fmt.Errorf("unsupported protocol: %s", s.config.Protocol)
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
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := s.config.LivenessCheck()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"server":  s.config.Name,
			"time":    r.Context().Value("time"),
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

	// 资源列表端点
	mux.HandleFunc("/resources", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resources := s.GetResources().List()

		resourceList := make([]map[string]interface{}, 0, len(resources))
		for _, resource := range resources {
			resourceList = append(resourceList, map[string]interface{}{
				"uri":         resource.URI,
				"name":        resource.Name,
				"description": resource.Description,
				"mimeType":    resource.MimeType,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"resources": resourceList,
		})
	})

	// 提示列表端点
	mux.HandleFunc("/prompts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		prompts := s.GetPrompts().List()

		promptList := make([]map[string]interface{}, 0, len(prompts))
		for _, prompt := range prompts {
			promptList = append(promptList, map[string]interface{}{
				"name":        prompt.Name,
				"description": prompt.Description,
				"arguments":   prompt.Arguments,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"prompts": promptList,
		})
	})

	// 确定监听地址
	addr := s.config.GetAddress()
	if addr == "stdio" {
		addr = ":3000" // 默认 HTTP 端口
	}

	// 打印启动信息
	logrus.Infof("Starting HTTP MCP server on %s", addr)
	logrus.Infof("MCP endpoint: http://%s/mcp", addr)
	logrus.Infof("Health check: http://%s/health", addr)
	logrus.Infof("Tools list: http://%s/tools", addr)
	logrus.Infof("Resources list: http://%s/resources", addr)
	logrus.Infof("Prompts list: http://%s/prompts", addr)
	logrus.Infof("Registered %d tools", len(tools))

	// 启动 HTTP 服务器
	if err := http.ListenAndServe(addr, mux); err != nil {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
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
	case "resources/list":
		return s.handleResourcesList(request)
	case "resources/read":
		return s.handleResourcesRead(ctx, request)
	case "prompts/list":
		return s.handlePromptsList(request)
	case "prompts/get":
		return s.handlePromptsGet(ctx, request)
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
				"name":    s.config.Name,
				"version": "1.0.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
				"resources": map[string]interface{}{
					"subscribe":   true,
					"listChanged": true,
				},
				"prompts": map[string]interface{}{
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

// handleResourcesList handles resources/list request
func (s *Server) handleResourcesList(request JSONRPCMessage) *JSONRPCMessage {
	resources := s.resources.List()

	// Convert to serializable format
	resourcesList := make([]map[string]interface{}, 0, len(resources))
	for _, resource := range resources {
		resourceMap := map[string]interface{}{
			"uri":         resource.URI,
			"name":        resource.Name,
			"description": resource.Description,
		}
		if resource.MimeType != "" {
			resourceMap["mimeType"] = resource.MimeType
		}
		resourcesList = append(resourcesList, resourceMap)
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"resources": resourcesList,
		},
	}
}

// handleResourcesRead handles resources/read request
func (s *Server) handleResourcesRead(ctx context.Context, request JSONRPCMessage) *JSONRPCMessage {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return errorResponse(request.ID, -32602, "Invalid params")
	}

	uri, _ := params["uri"].(string)
	logrus.Infof("Reading resource: %s", uri)

	content, err := s.resources.Read(ctx, uri)
	if err != nil {
		logrus.Errorf("Resource read failed: %v", err)
		return errorResponse(request.ID, -32000, err.Error())
	}

	// Convert to serializable format
	contentMap := map[string]interface{}{
		"uri": content.URI,
	}

	if content.Text != "" {
		contentMap["text"] = content.Text
	}

	if content.MimeType != "" {
		contentMap["mimeType"] = content.MimeType
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"contents": []interface{}{contentMap},
		},
	}
}

// handlePromptsList handles prompts/list request
func (s *Server) handlePromptsList(request JSONRPCMessage) *JSONRPCMessage {
	prompts := s.prompts.List()

	// Convert to serializable format
	promptsList := make([]map[string]interface{}, 0, len(prompts))
	for _, prompt := range prompts {
		promptMap := map[string]interface{}{
			"name":        prompt.Name,
			"description": prompt.Description,
		}

		if len(prompt.Arguments) > 0 {
			args := make([]map[string]interface{}, 0, len(prompt.Arguments))
			for _, arg := range prompt.Arguments {
				argMap := map[string]interface{}{
					"name":        arg.Name,
					"description": arg.Description,
					"required":    arg.Required,
				}
				args = append(args, argMap)
			}
			promptMap["arguments"] = args
		}

		promptsList = append(promptsList, promptMap)
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"prompts": promptsList,
		},
	}
}

// handlePromptsGet handles prompts/get request
func (s *Server) handlePromptsGet(ctx context.Context, request JSONRPCMessage) *JSONRPCMessage {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		return errorResponse(request.ID, -32602, "Invalid params")
	}

	name, _ := params["name"].(string)
	var args map[string]interface{}
	if argsInterface, ok := params["arguments"]; ok {
		args, _ = argsInterface.(map[string]interface{})
	}

	logrus.Infof("Getting prompt: %s with args: %+v", name, args)

	result, err := s.prompts.Generate(ctx, name, args)
	if err != nil {
		logrus.Errorf("Prompt get failed: %v", err)
		return errorResponse(request.ID, -32000, err.Error())
	}

	return &JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": result,
				},
			},
		},
	}
}

// RegisterTool registers a tool with the server
func (s *Server) RegisterTool(tool *Tool) error {
	return s.tools.Register(tool)
}

// RegisterResource registers a resource with the server
func (s *Server) RegisterResource(resource *Resource) error {
	return s.resources.Register(resource)
}

// RegisterPrompt registers a prompt with the server
func (s *Server) RegisterPrompt(prompt *Prompt) error {
	return s.prompts.Register(prompt)
}

// GetTools returns the tool registry
func (s *Server) GetTools() *ToolRegistry {
	return s.tools
}

// GetResources returns the resource registry
func (s *Server) GetResources() *ResourceRegistry {
	return s.resources
}

// GetPrompts returns the prompt registry
func (s *Server) GetPrompts() *PromptRegistry {
	return s.prompts
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
