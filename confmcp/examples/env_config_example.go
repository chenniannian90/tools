package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chenniannian90/tools/confmcp"
)

// 环境变量配置示例
// 展示如何通过环境变量配置 MCP 服务器

func main() {
	// 演示：设置环境变量（实际使用时应该在外部设置）
	// export MCP_NAME="my-server"
	// export MCP_PROTOCOL="stdio"
	// export MCP_PORT=3000
	// export MCP_LOG_LEVEL="debug"
	// export MCP_SERVER_VERSION="2.0.0"
	// export MCP_RETRY_REPEATS=5

	// 打印当前环境变量配置
	fmt.Fprintf(os.Stderr, "=== 环境变量配置示例 ===\n\n")
	fmt.Fprintf(os.Stderr, "可以通过以下环境变量配置 MCP 服务器:\n\n")
	fmt.Fprintf(os.Stderr, "  MCP_NAME           - 服务器名称 (默认: empty)\n")
	fmt.Fprintf(os.Stderr, "  MCP_PROTOCOL       - 协议类型 (默认: stdio)\n")
	fmt.Fprintf(os.Stderr, "  MCP_HOST           - 主机地址 (默认: empty)\n")
	fmt.Fprintf(os.Stderr, "  MCP_PORT           - 端口号 (默认: 3000)\n")
	fmt.Fprintf(os.Stderr, "  MCP_LOG_LEVEL      - 日志级别 (默认: info)\n")
	fmt.Fprintf(os.Stderr, "  MCP_SERVER_VERSION - 服务器版本 (默认: 1.0.0)\n")
	fmt.Fprintf(os.Stderr, "  MCP_RETRY_REPEATS  - 重试次数 (默认: 3)\n")
	fmt.Fprintf(os.Stderr, "\n示例:\n")
	fmt.Fprintf(os.Stderr, "  export MCP_NAME=my-mcp-server\n")
	fmt.Fprintf(os.Stderr, "  export MCP_LOG_LEVEL=debug\n")
	fmt.Fprintf(os.Stderr, "  go run main.go\n\n")

	// 创建配置（将从环境变量读取）
	config := &confmcp.MCP{}

	// 显示默认值
	fmt.Fprintf(os.Stderr, "=== 默认配置值 ===\n")
	config.SetDefaults()
	fmt.Fprintf(os.Stderr, "  Name: %s\n", config.Name)
	fmt.Fprintf(os.Stderr, "  Protocol: %s\n", config.Protocol)
	fmt.Fprintf(os.Stderr, "  Port: %d\n", config.Port)
	fmt.Fprintf(os.Stderr, "  Retry.Repeats: %d\n", config.retry.Repeats)
	fmt.Fprintf(os.Stderr, "\n")

	// 创建服务器
	server := confmcp.NewServer(config)

	// 注册一个简单的配置查看工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "get_config",
		Description: "查看当前服务器配置",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			result := fmt.Sprintf("服务器配置信息:\n\n")
			result += fmt.Sprintf("名称: %s\n", config.Name)
			result += fmt.Sprintf("协议: %s\n", config.Protocol)
			result += fmt.Sprintf("端口: %d\n", config.Port)
			result += fmt.Sprintf("日志级别: %s\n", config.LogLevel)
			result += fmt.Sprintf("版本: %s\n", config.ServerVersion)

			if config.Host != "" {
				result += fmt.Sprintf("主机: %s\n", config.Host)
			}

			result += fmt.Sprintf("\n能力:\n")
			result += fmt.Sprintf("  - 工具: %v\n", config.Capabilities.Tools)
			result += fmt.Sprintf("  - 资源: %v\n", config.Capabilities.Resources)
			result += fmt.Sprintf("  - 提示: %v\n", config.Capabilities.Prompts)

			return result, nil
		},
	})

	// 注册环境变量查看工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "get_env",
		Description: "查看指定环境变量的值",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"key": map[string]interface{}{
					"type":        "string",
					"description": "环境变量名称",
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			key, ok := args["key"].(string)
			if !ok || key == "" {
				// 列出所有 MCP 相关的环境变量
				result := "MCP 相关环境变量:\n\n"
				envVars := []string{
					"MCP_NAME", "MCP_PROTOCOL", "MCP_HOST",
					"MCP_PORT", "MCP_LOG_LEVEL", "MCP_SERVER_VERSION",
				}
				for _, envVar := range envVars {
					value := os.Getenv(envVar)
					if value == "" {
						result += fmt.Sprintf("%s=(未设置)\n", envVar)
					} else {
						result += fmt.Sprintf("%s=%s\n", envVar, value)
					}
				}
				return result, nil
			}

			value := os.Getenv(key)
			if value == "" {
				return fmt.Sprintf("环境变量 %s 未设置", key), nil
			}
			return fmt.Sprintf("%s=%s", key, value), nil
		},
	})

	// 注册设置环境变量工具（仅用于演示）
	server.RegisterTool(&confmcp.Tool{
		Name:        "set_env",
		Description: "设置环境变量（仅当前进程有效）",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"key": map[string]interface{}{
					"type":        "string",
					"description": "环境变量名称",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "环境变量值",
				},
			},
			"required": []string{"key", "value"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			key, _ := args["key"].(string)
			value, _ := args["value"].(string)

			os.Setenv(key, value)
			return fmt.Sprintf("已设置环境变量: %s=%s", key, value), nil
		},
	})

	fmt.Fprintf(os.Stderr, "启动环境变量配置示例服务器...\n")
	fmt.Fprintf(os.Stderr, "可用工具: get_config, get_env, set_env\n\n")
	fmt.Fprintf(os.Stderr, "提示: 使用 tools/call 调用这些工具来查看和修改配置\n\n")

	if err := server.Start(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "服务器错误: %v\n", err)
		os.Exit(1)
	}
}
