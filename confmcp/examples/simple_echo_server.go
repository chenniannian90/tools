package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chenniannian90/tools/confmcp"
)

// 最简单的 MCP 服务器示例 - Echo 工具
// 这个示例展示如何创建一个基本的 MCP 服务器，提供一个简单的 echo 工具

func main() {
	server := confmcp.NewServer()
	server.Name = "echo-server"
	server.Protocol = "stdio"

	// 注册 echo 工具
	err := server.RegisterTool(&confmcp.Tool{
		Name:        "echo",
		Description: "回显输入的消息",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "要回显的消息",
				},
			},
			"required": []string{"message"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			message, ok := args["message"].(string)
			if !ok {
				return nil, fmt.Errorf("message 参数必须是字符串")
			}
			return fmt.Sprintf("Echo: %s", message), nil
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "注册工具失败: %v\n", err)
		os.Exit(1)
	}

	// 注册问候工具
	err = server.RegisterTool(&confmcp.Tool{
		Name:        "greet",
		Description: "向指定的人打招呼",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "要问候的人的名字",
				},
			},
			"required": []string{"name"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			name, ok := args["name"].(string)
			if !ok {
				return nil, fmt.Errorf("name 参数必须是字符串")
			}
			return fmt.Sprintf("你好, %s! 欢迎使用 MCP 服务器!", name), nil
		},
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "注册工具失败: %v\n", err)
		os.Exit(1)
	}

	// 启动服务器（阻塞运行）
	fmt.Fprintf(os.Stderr, "启动 Echo MCP 服务器...\n")
	if err := server.Start(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "服务器错误: %v\n", err)
		os.Exit(1)
	}
}
