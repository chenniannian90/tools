package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chenniannian90/tools/confmcp"
)

// SSE 服务器示例
// 展示如何使用 SSE 协议创建 MCP 服务器
// SSE 支持服务器向客户端主动推送事件

func main() {
	server := confmcp.NewServer()
	server.Name = "sse-example-server"
	server.Protocol = "sse"
	server.Port = 3001

	// 注册工具
	err := server.RegisterTool(&confmcp.Tool{
		Name:        "get_time",
		Description: "获取当前时间",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"timezone": map[string]interface{}{
					"type":        "string",
					"description": "时区，例如 UTC、Asia/Shanghai",
					"default":     "UTC",
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			timezone := "UTC"
			if tz, ok := args["timezone"].(string); ok {
				timezone = tz
			}

			return map[string]interface{}{
				"time":     time.Now().Format(time.RFC3339),
				"timezone": timezone,
				"unix":     time.Now().Unix(),
			}, nil
		},
	})

	if err != nil {
		fmt.Printf("注册工具失败: %v\n", err)
		return
	}

	// 注册计算工具
	err = server.RegisterTool(&confmcp.Tool{
		Name:        "calculate",
		Description: "执行简单的数学计算",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"expression": map[string]interface{}{
					"type":        "string",
					"description": "数学表达式，支持 +, -, *, /",
				},
			},
			"required": []string{"expression"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			expr, ok := args["expression"].(string)
			if !ok {
				return nil, fmt.Errorf("expression 必须是字符串")
			}

			// 简单的示例：只支持两个数的加法
			var a, b float64
			var op string
			n, _ := fmt.Sscanf(expr, "%f%c%f", &a, &op, &b)

			if n != 3 {
				return nil, fmt.Errorf("无效的表达式格式，请使用: 数字+运算符+数字")
			}

			var result float64
			switch op {
			case "+":
				result = a + b
			case "-":
				result = a - b
			case "*":
				result = a * b
			case "/":
				if b == 0 {
					return nil, fmt.Errorf("除数不能为零")
				}
				result = a / b
			default:
				return nil, fmt.Errorf("不支持的运算符: %s", op)
			}

			return map[string]interface{}{
				"expression": expr,
				"result":     result,
			}, nil
		},
	})

	if err != nil {
		fmt.Printf("注册工具失败: %v\n", err)
		return
	}

	// 注册事件推送工具
	err = server.RegisterTool(&confmcp.Tool{
		Name:        "notify",
		Description: "模拟服务器推送通知",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "通知消息",
				},
			},
			"required": []string{"message"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			message, ok := args["message"].(string)
			if !ok {
				return nil, fmt.Errorf("message 必须是字符串")
			}

			return map[string]interface{}{
				"type":      "notification",
				"message":   message,
				"timestamp": time.Now().Format(time.RFC3339),
				"status":    "sent",
			}, nil
		},
	})

	if err != nil {
		fmt.Printf("注册工具失败: %v\n", err)
		return
	}

	fmt.Println("启动 SSE MCP 服务器...")
	fmt.Printf("SSE 端点: http://localhost:%d/sse\n", server.Port)
	fmt.Printf("MCP 端点: http://localhost:%d/mcp\n", server.Port)
	fmt.Printf("健康检查: http://localhost:%d/health\n", server.Port)
	fmt.Printf("工具列表: http://localhost:%d/tools\n", server.Port)
	fmt.Println("\n可用工具:")
	fmt.Println("  - get_time: 获取当前时间")
	fmt.Println("  - calculate: 执行数学计算")
	fmt.Println("  - notify: 模拟通知推送")
	fmt.Println("\n使用 curl 测试:")
	fmt.Printf("  curl http://localhost:%d/health\n", server.Port)
	fmt.Printf("  curl -X POST http://localhost:%d/mcp -H 'Content-Type: application/json' -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"get_time\",\"arguments\":{}}}'\n", server.Port)
	fmt.Println("\n监听 SSE 事件流:")
	fmt.Printf("  curl -N http://localhost:%d/sse\n", server.Port)

	// 启动服务器
	if err := server.Start(context.Background()); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}
