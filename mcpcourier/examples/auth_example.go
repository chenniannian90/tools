package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/chenniannian90/tools/confmcp"
	"github.com/chenniannian90/tools/mcpcourier"
)

func main() {
	fmt.Println("=== mcpcourier API Key 认证示例 ===\n")

	// 示例 1: 简单列表验证
	simpleListExample()

	// 示例 2: 从环境变量读取
	envVarExample()

	// 示例 3: 数据库验证
	databaseValidatorExample()

	// 示例 4: 自定义中间件
	middlewareExample()

	// 示例 5: 组合使用
	combinedExample()

	// 阻塞主线程
	select {}
}

// 示例 1: 简单列表验证
func simpleListExample() {
	fmt.Println("1. 简单列表验证")
	fmt.Println("   启动服务器在 :3001")
	fmt.Println("   测试: curl -H 'X-API-Key: key-1' http://localhost:3001/mcp")

	// 定义验证器
	validator := func(apiKey string) (bool, error) {
		validKeys := map[string]bool{
			"key-1": true,
			"key-2": true,
		}
		valid := validKeys[apiKey]
		fmt.Printf("   [验证] API Key '%s' -> %v\n", apiKey, valid)
		return valid, nil
	}

	go func() {
		task := mcpcourier.NewTask(
			mcpcourier.WithTool(&confmcp.Tool{
				Name:        "echo",
				Description: "回显消息",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type": "string",
						},
					},
				},
				Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return fmt.Sprintf("Echo: %v", args["message"]), nil
				},
			}),
			mcpcourier.WithAPIKeyValidator(validator),
		)

		task.server.Protocol = "http"
		task.server.Port = 3001
		task.Run(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("   ✓ 服务器已启动\n")
}

// 示例 2: 从环境变量读取
func envVarExample() {
	fmt.Println("2. 从环境变量读取")
	fmt.Println("   设置环境变量: export MCP_API_KEY=my-env-key")
	fmt.Println("   启动服务器在 :3002")

	// 定义验证器
	validator := func(apiKey string) (bool, error) {
		envKey := "my-env-key" // 模拟从环境变量读取
		valid := apiKey == envKey
		fmt.Printf("   [验证] API Key '%s' vs ENV '%s' -> %v\n", apiKey, envKey, valid)
		return valid, nil
	}

	go func() {
		task := mcpcourier.NewTask(
			mcpcourier.WithTool(&confmcp.Tool{
				Name:        "ping",
				Description: "Ping 服务",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
				Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return "pong", nil
				},
			}),
			mcpcourier.WithAPIKeyValidator(validator),
		)

		task.server.Protocol = "http"
		task.server.Port = 3002
		task.Run(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("   ✓ 服务器已启动\n")
}

// 示例 3: 数据库验证
func databaseValidatorExample() {
	fmt.Println("3. 数据库验证")
	fmt.Println("   启动服务器在 :3003")
	fmt.Println("   测试: curl -H 'X-API-Key: user-123' http://localhost:3003/mcp")

	// 模拟数据库验证器
	dbValidator := func(apiKey string) (bool, error) {
		// 模拟数据库查询
		validKeys := map[string]bool{
			"user-123":  true,
			"admin-456": true,
		}
		valid := validKeys[apiKey]
		fmt.Printf("   [数据库验证] API Key '%s' -> %v\n", apiKey, valid)
		return valid, nil
	}

	go func() {
		task := mcpcourier.NewTask(
			mcpcourier.WithTool(&confmcp.Tool{
				Name:        "get_user",
				Description: "获取用户信息",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
				Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return map[string]interface{}{
						"user_id": 123,
						"name":    "Test User",
					}, nil
				},
			}),
			mcpcourier.WithAPIKeyValidator(dbValidator),
		)

		task.server.Protocol = "http"
		task.server.Port = 3003
		task.Run(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("   ✓ 服务器已启动\n")
}

// 示例 4: 自定义中间件
func middlewareExample() {
	fmt.Println("4. 自定义中间件（日志 + CORS）")
	fmt.Println("   启动服务器在 :3004")

	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log.Printf("[中间件] %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
			log.Printf("[中间件] 完成 in %v", time.Since(start))
		})
	}

	go func() {
		task := mcpcourier.NewTask(
			mcpcourier.WithTool(&confmcp.Tool{
				Name:        "time",
				Description: "获取当前时间",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
				Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return map[string]interface{}{
						"time": time.Now().Format(time.RFC3339),
					}, nil
				},
			}),
			mcpcourier.WithMiddleware(loggingMiddleware),
		)

		task.server.Protocol = "http"
		task.server.Port = 3004
		task.Run(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("   ✓ 服务器已启动\n")
}

// 示例 5: 组合使用（API Key + 自定义中间件）
func combinedExample() {
	fmt.Println("5. 组合使用（API Key + 自定义中间件）")
	fmt.Println("   启动服务器在 :3005")
	fmt.Println("   测试: curl -H 'X-API-Key: combined-key' http://localhost:3005/mcp")

	// API Key 验证器
	apiKeyValidator := func(apiKey string) (bool, error) {
		validKeys := map[string]bool{
			"combined-key": true,
		}
		valid := validKeys[apiKey]
		fmt.Printf("   [验证] API Key '%s' -> %v\n", apiKey, valid)
		return valid, nil
	}

	// 请求 ID 中间件
	requestIDMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = fmt.Sprintf("req-%d", time.Now().UnixNano())
			}
			w.Header().Set("X-Request-ID", requestID)
			log.Printf("[请求 ID: %s] %s %s", requestID, r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}

	go func() {
		task := mcpcourier.NewTask(
			mcpcourier.WithTool(&confmcp.Tool{
				Name:        "status",
				Description: "获取状态",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
				Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return map[string]interface{}{
						"status":  "running",
						"version": "1.0.0",
						"time":    time.Now().Format(time.RFC3339),
					}, nil
				},
			}),
			mcpcourier.WithAPIKeyValidator(apiKeyValidator),
			mcpcourier.WithMiddleware(requestIDMiddleware),
		)

		task.server.Protocol = "http"
		task.server.Port = 3005
		task.Run(context.Background())
	}()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("   ✓ 服务器已启动\n")
}
