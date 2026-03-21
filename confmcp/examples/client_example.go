package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chenniannian90/tools/confmcp"
)

// MCP 客户端使用示例
// 展示如何连接到 MCP 服务器并调用工具

func main() {
	// 创建客户端配置
	config := &confmcp.MCP{
		Name:     "mcp-client",
		Protocol: "stdio",
	}

	// 创建客户端
	client := confmcp.NewClient(config)

	// 连接到服务器
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "连接失败: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Fprintf(os.Stderr, "成功连接到 MCP 服务器\n\n")

	// 列出可用的工具
	fmt.Fprintf(os.Stderr, "=== 获取可用工具列表 ===\n")
	tools, err := client.ListTools(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取工具列表失败: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "找到 %d 个工具:\n", len(tools))
		for _, tool := range tools {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", tool.Name, tool.Description)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// 调用 echo 工具
	fmt.Fprintf(os.Stderr, "=== 调用 echo 工具 ===\n")
	result, err := client.CallTool(ctx, "echo", map[string]interface{}{
		"message": "Hello from MCP Client!",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "调用工具失败: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "结果: %v\n\n", result)
	}

	// 列出可用的资源
	fmt.Fprintf(os.Stderr, "=== 获取可用资源列表 ===\n")
	resources, err := client.ListResources(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取资源列表失败: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "找到 %d 个资源:\n", len(resources))
		for _, resource := range resources {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", resource.URI, resource.Name)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// 读取资源
	if len(resources) > 0 {
		fmt.Fprintf(os.Stderr, "=== 读取资源 ===\n")
		content, err := client.ReadResource(ctx, resources[0].URI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "读取资源失败: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "资源内容:\n%s\n\n", content.Text)
		}
	}

	// 列出可用的提示
	fmt.Fprintf(os.Stderr, "=== 获取可用提示列表 ===\n")
	prompts, err := client.ListPrompts(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取提示列表失败: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "找到 %d 个提示:\n", len(prompts))
		for _, prompt := range prompts {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", prompt.Name, prompt.Description)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// 获取提示
	if len(prompts) > 0 {
		fmt.Fprintf(os.Stderr, "=== 获取提示 ===\n")
		promptText, err := client.GetPrompt(ctx, prompts[0].Name, map[string]interface{}{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取提示失败: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "提示内容:\n%s\n", promptText)
		}
	}
}
