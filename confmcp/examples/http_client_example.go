package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chenniannian90/tools/confmcp"
)

// HTTP MCP 客户端示例
// 展示如何使用 HTTP 传输层连接到 MCP 服务器

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           🌐 HTTP MCP 客户端                                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	baseURL := "http://localhost:3000"

	// 1. 测试健康检查
	fmt.Println("💚 测试健康检查...")
	if err := testHealthCheck(baseURL); err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		fmt.Println("\n请确保服务器正在运行：")
		fmt.Println("  go run http_server.go")
		os.Exit(1)
	}
	fmt.Println("✅ 健康检查通过")
	fmt.Println()

	// 2. 测试 ping 工具
	fmt.Println("🏓 测试 ping 工具...")
	if err := callTool(baseURL, "ping", nil); err != nil {
		fmt.Printf("❌ 调用失败: %v\n", err)
	} else {
		fmt.Println("✅ ping 工具调用成功")
	}
	fmt.Println()

	// 3. 测试 echo 工具
	fmt.Println("📢 测试 echo 工具...")
	args := map[string]interface{}{
		"message": "Hello from HTTP client!",
	}
	if err := callTool(baseURL, "echo", args); err != nil {
		fmt.Printf("❌ 调用失败: %v\n", err)
	} else {
		fmt.Println("✅ echo 工具调用成功")
	}
	fmt.Println()

	// 4. 测试 calculate 工具
	fmt.Println("🧮 测试 calculate 工具...")
	calcArgs := map[string]interface{}{
		"operation": "multiply",
		"a":          float64(15),
		"b":          float64(3),
	}
	if err := callTool(baseURL, "calculate", calcArgs); err != nil {
		fmt.Printf("❌ 调用失败: %v\n", err)
	} else {
		fmt.Println("✅ calculate 工具调用成功")
	}
	fmt.Println()

	// 5. 获取工具列表
	fmt.Println("📋 获取工具列表...")
	if err := listTools(baseURL); err != nil {
		fmt.Printf("❌ 获取失败: %v\n", err)
	}
	fmt.Println()

	fmt.Println("✅ 所有测试完成！")
}

// testHealthCheck 测试健康检查
func testHealthCheck(baseURL string) error {
	client := confmcp.NewHTTPClient(baseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	fmt.Printf("   状态: 可用\n")
	return nil
}

// callTool 调用工具
func callTool(baseURL, toolName string, args map[string]interface{}) error {
	client := confmcp.NewHTTPClient(baseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 构造 JSON-RPC 请求
	request := &confmcp.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("RPC 错误: %s", response.Error.Message)
	}

	if result, ok := response.Result.(map[string]interface{}); ok {
		if content, ok := result["content"].(string); ok {
			fmt.Printf("   结果: %s\n", content)
		}
	}

	return nil
}

// listTools 列出工具
func listTools(baseURL string) error {
	client := confmcp.NewHTTPClient(baseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 构造工具列表请求
	request := &confmcp.JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
		Params:  map[string]interface{}{},
	}

	response, err := client.Send(ctx, request)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return fmt.Errorf("RPC 错误: %s", response.Error.Message)
	}

	if result, ok := response.Result.(map[string]interface{}); ok {
		if tools, ok := result["tools"].([]interface{}); ok {
			fmt.Printf("   找到 %d 个工具:\n", len(tools))
			for _, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					name := toolMap["name"]
					desc := toolMap["description"]
					fmt.Printf("   • %s: %s\n", name, desc)
				}
			}
		}
	}

	return nil
}
