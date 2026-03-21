package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chenniannian90/tools/confmcp"
)

// 使用 Serve 方法的简化示例
// 展示如何一行代码启动 HTTP MCP 服务器

func main() {
	// 创建配置
	config := &confmcp.MCP{
		Name:     "serve-example",
		Protocol: "http",
		Port:     3002,
	}

	// 创建服务器
	server := confmcp.NewServer(config)

	// 使用 Serve 方法一键启动 HTTP 服务器
	// 只需要传入工具数组，自动处理所有 HTTP 端点
	err := server.Serve([]*confmcp.Tool{
		{
			Name:        "ping",
			Description: "健康检查",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return map[string]interface{}{
					"status":  "ok",
					"time":    time.Now().Format(time.RFC3339),
					"server":  config.Name,
					"message": "Server is running healthy!",
				}, nil
			},
		},
		{
			Name:        "echo",
			Description: "回显消息",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"message"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				msg := args["message"].(string)
				return fmt.Sprintf("Echo: %s", msg), nil
			},
		},
		{
			Name:        "current_time",
			Description: "获取当前时间",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"timezone": map[string]interface{}{
						"type":        "string",
						"description": "时区（可选）",
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
		},
		{
			Name:        "add",
			Description: "加法运算",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{
						"type": "number",
					},
					"b": map[string]interface{}{
						"type": "number",
					},
				},
				"required": []string{"a", "b"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				a, _ := args["a"].(float64)
				b, _ := args["b"].(float64)
				return fmt.Sprintf("%d + %d = %d", int(a), int(b), int(a+b)), nil
			},
		},
	})

	if err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}
