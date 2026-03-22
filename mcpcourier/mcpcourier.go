package mcpcourier

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/chenniannian90/tools/confmcp"
)

func Run(server *confmcp.Server) {
	_ = server
}

type Task struct {
	server       *confmcp.Server
	tools        []*confmcp.Tool
	middlewares  []func(http.Handler) http.Handler
	apiKeyConfig *confmcp.APIKeyConfig
}

func NewTask(opts ...Option) *Task {
	t := &Task{
		server:      confmcp.NewServer(),
		middlewares: []func(http.Handler) http.Handler{},
	}
	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *Task) Run(ctx context.Context) {
	// Apply API Key configuration to server
	if t.apiKeyConfig != nil {
		// 如果配置了 API Keys，直接设置到 server
		if len(t.apiKeyConfig.APIKeys) > 0 && t.apiKeyConfig.Validator == nil {
			t.server.SetAPIKeys(t.apiKeyConfig.APIKeys)
		}
		// 注意：如果有自定义 Validator，需要在 HTTP handler 层面应用
		// 这将在后续的 HTTP 服务器启动时处理
	}

	// Register all tools
	for _, tool := range t.tools {
		if err := t.server.RegisterTool(tool); err != nil {
			panic(err)
		}
	}

	// 如果有自定义中间件或 API Key 验证器，使用自定义 HTTP server
	if len(t.middlewares) > 0 || (t.apiKeyConfig != nil && t.apiKeyConfig.Validator != nil) {
		t.RunWithCustomHandler(ctx)
	} else {
		// Start server with default configuration
		if err := t.server.Start(ctx); err != nil {
			panic(err)
		}
	}
}

// RunWithCustomHandler 使用自定义中间件启动 HTTP 服务器
func (t *Task) RunWithCustomHandler(ctx context.Context) {
	// 确保 server 配置已初始化
	t.server.Init()

	// 获取 MCP handler
	mcpHandler := t.createMCPHandler()

	// 应用所有中间件
	finalHandler := http.Handler(mcpHandler)
	for _, middleware := range t.middlewares {
		finalHandler = middleware(finalHandler)
	}

	// 如果有 API Key 验证器，应用到最后
	if t.apiKeyConfig != nil && t.apiKeyConfig.Validator != nil {
		apiKeyMiddleware := confmcp.APIKeyAuth(*t.apiKeyConfig)
		finalHandler = apiKeyMiddleware(finalHandler)
	}

	// 创建 HTTP mux
	mux := http.NewServeMux()

	// MCP 端点
	mux.Handle("/mcp", finalHandler)

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := t.server.LivenessCheck()
		w.Write([]byte(`{"status":"healthy","checks":` + mustMarshalJSON(status) + `}`))
	})

	// 工具列表端点
	mux.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tools := t.server.GetTools().List()
		toolList := make([]map[string]interface{}, 0, len(tools))
		for _, tool := range tools {
			toolList = append(toolList, map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": tool.InputSchema,
			})
		}
		w.Write([]byte(`{"tools":` + mustMarshalJSON(toolList) + `}`))
	})

	// 确定监听地址
	addr := t.server.GetAddress()
	if addr == "stdio" {
		addr = ":3000"
	}

	// 启动 HTTP 服务器
	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}
}

// createMCPHandler 创建 MCP 处理器
func (t *Task) createMCPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 解析 JSON-RPC 请求
		var request confmcp.JSONRPCMessage
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// 处理请求
		response := t.server.HandleRequest(r.Context(), request)
		json.NewEncoder(w).Encode(response)
	})
}

func mustMarshalJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
