package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/chenniannian90/tools/confmcp"
)

func main() {
	// 示例 1: 使用 APIKeyAuthFunc（最简单）
	simpleExample()

	// 示例 2: 使用 APIKeyAuth 中间件（更灵活）
	advancedExample()

	// 示例 3: 从环境变量读取 API Key
	envExample()

	// 示例 4: 集成到 confmcp Server
	serverExample()
}

// 示例 1: 简单的 API Key 认证
func simpleExample() {
	fmt.Println("=== 示例 1: 简单的 API Key 认证 ===")

	mux := http.NewServeMux()

	// 使用 APIKeyAuthFunc 包装 handler
	apiKeys := []string{"secret-key-1", "secret-key-2"}
	mux.HandleFunc("/api/data", confmcp.APIKeyAuthFunc(apiKeys, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "success", "data": "protected data"}`))
	}))

	fmt.Println("Server started on :8080")
	fmt.Println("Test with: curl -H 'X-API-Key: secret-key-1' http://localhost:8080/api/data")
}

// 示例 2: 使用完整的中间件（支持更多配置）
func advancedExample() {
	fmt.Println("\n=== 示例 2: 高级配置 ===")

	mux := http.NewServeMux()

	// 创建配置
	config := confmcp.APIKeyConfig{
		APIKeys: []string{"admin-key", "user-key"},
		// EnvVar: "CUSTOM_API_KEY", // 可选：从环境变量读取
		// DisableAuth: false,       // 可选：禁用认证（开发环境）
	}

	// 使用中间件包装 handler
	mux.Handle("/api/admin", confmcp.APIKeyAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "success", "role": "admin"}`))
	})))

	fmt.Println("Advanced server configured")
}

// 示例 3: 从环境变量读取 API Key
func envExample() {
	fmt.Println("\n=== 示例 3: 从环境变量读取 ===")

	mux := http.NewServeMux()

	// 使用 APIKeyAuthFromEnv 中间件
	// 设置环境变量: export MCP_API_KEY=your-secret-key
	mux.Handle("/api/env", confmcp.APIKeyAuthFromEnv("MCP_API_KEY")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "authenticated from env"}`))
	})))

	fmt.Println("Set MCP_API_KEY environment variable to test")
}

// 示例 4: 集成到 confmcp Server
func serverExample() {
	fmt.Println("\n=== 示例 4: confmcp Server 集成 ===")

	// 创建 MCP server
	server := confmcp.NewServer()
	server.Name = "My MCP Server"
	server.Protocol = "http"
	server.Port = 3000

	// 配置 API Keys（方式1: 直接设置）
	server.SetAPIKeys([]string{"mcp-key-1", "mcp-key-2"})

	// 方式2: 添加单个 key
	// server.AddAPIKey("mcp-key-3")

	// 方式3: 从环境变量读取（在 server.getValidAPIKeys() 中自动处理）
	// export MCP_API_KEY=your-key

	// 注册工具
	echoTool := &confmcp.Tool{
		Name:        "echo",
		Description: "Echo back the input",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "Message to echo back",
				},
			},
			"required": []string{"message"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			msg, _ := args["message"].(string)
			return fmt.Sprintf("Echo: %s", msg), nil
		},
	}

	server.RegisterTool(echoTool)

	// 启动 server
	fmt.Println("Starting MCP server with API Key authentication")
	fmt.Println("Test with: curl -X POST -H 'X-API-Key: mcp-key-1' -H 'Content-Type: application/json' -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}' http://localhost:3000/mcp")

	// 注意：在实际使用中，这里应该是:
	// if err := server.Start(context.Background()); err != nil {
	//     log.Fatal(err)
	// }
}

// 示例 5: 可选的 API Key 认证
func optionalAuthExample() {
	fmt.Println("\n=== 示例 5: 可选的 API Key 认证 ===")

	mux := http.NewServeMux()

	// OptionalAPIKey: 如果提供了 API Key 则验证，否则允许访问
	config := confmcp.APIKeyConfig{
		APIKeys: []string{"optional-key"},
	}

	mux.Handle("/api/public", confmcp.OptionalAPIKey(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			w.Write([]byte(`{"status": "authenticated", "level": "premium"}`))
		} else {
			w.Write([]byte(`{"status": "anonymous", "level": "basic"}`))
		}
	})))

	fmt.Println("Optional auth: access with or without API Key")
}

// 示例 6: 使用 Bearer Token
func bearerTokenExample() {
	fmt.Println("\n=== 示例 6: Bearer Token 认证 ===")

	mux := http.NewServeMux()

	apiKeys := []string{"bearer-token-1", "bearer-token-2"}

	// 使用 Bearer token 格式
	mux.HandleFunc("/api/bearer", confmcp.APIKeyFromBearer(apiKeys, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "authenticated with bearer token"}`))
	}))

	fmt.Println("Bearer token: curl -H 'Authorization: Bearer bearer-token-1' http://localhost:8080/api/bearer")
}
