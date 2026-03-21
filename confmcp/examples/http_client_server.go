package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/chenniannian90/tools/confmcp"
)

// HTTP 客户端 MCP 服务器示例
// 展示如何实现 HTTP 请求工具

func main() {
	config := &confmcp.MCP{
		Name:          "http-client-server",
		Protocol:      "stdio",
	}

	// 创建自定义 HTTP 客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	server := confmcp.NewServer(config)

	// GET 请求工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "http_get",
		Description: "发送 HTTP GET 请求",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "请求的 URL",
				},
				"headers": map[string]interface{}{
					"type":        "object",
					"description": "请求头（可选）",
				},
			},
			"required": []string{"url"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			url, ok := args["url"].(string)
			if !ok || url == "" {
				return nil, fmt.Errorf("必须提供 URL")
			}

			// 创建请求
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return nil, fmt.Errorf("创建请求失败: %w", err)
			}

			// 添加请求头
			if headers, ok := args["headers"].(map[string]interface{}); ok {
				for key, value := range headers {
					if valueStr, ok := value.(string); ok {
						req.Header.Set(key, valueStr)
					}
				}
			}

			// 发送请求
			resp, err := httpClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("请求失败: %w", err)
			}
			defer resp.Body.Close()

			// 读取响应
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("读取响应失败: %w", err)
			}

			// 格式化响应
			result := fmt.Sprintf("HTTP GET %s\n", url)
			result += fmt.Sprintf("状态码: %d %s\n", resp.StatusCode, resp.Status)
			result += fmt.Sprintf("响应头:\n")
			for key, values := range resp.Header {
				for _, value := range values {
					result += fmt.Sprintf("  %s: %s\n", key, value)
				}
			}
			result += fmt.Sprintf("\n响应内容:\n%s\n", string(body))

			return result, nil
		},
	})

	// POST JSON 请求工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "http_post_json",
		Description: "发送 HTTP POST 请求（JSON 格式）",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "请求的 URL",
				},
				"data": map[string]interface{}{
					"type":        "object",
					"description": "要发送的 JSON 数据",
				},
				"headers": map[string]interface{}{
					"type":        "object",
					"description": "额外的请求头（可选）",
				},
			},
			"required": []string{"url", "data"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			url, ok := args["url"].(string)
			if !ok || url == "" {
				return nil, fmt.Errorf("必须提供 URL")
			}

			data, ok := args["data"].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("data 必须是对象")
			}

			// 序列化 JSON
			jsonData, err := json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("序列化 JSON 失败: %w", err)
			}

			// 创建请求
			req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
			if err != nil {
				return nil, fmt.Errorf("创建请求失败: %w", err)
			}

			req.Header.Set("Content-Type", "application/json")

			// 添加额外请求头
			if headers, ok := args["headers"].(map[string]interface{}); ok {
				for key, value := range headers {
					if valueStr, ok := value.(string); ok {
						req.Header.Set(key, valueStr)
					}
				}
			}

			// 发送请求
			resp, err := httpClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("请求失败: %w", err)
			}
			defer resp.Body.Close()

			// 读取响应
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("读取响应失败: %w", err)
			}

			result := fmt.Sprintf("HTTP POST %s\n", url)
			result += fmt.Sprintf("请求体: %s\n", string(jsonData))
			result += fmt.Sprintf("状态码: %d %s\n", resp.StatusCode, resp.Status)
			result += fmt.Sprintf("\n响应内容:\n%s\n", string(body))

			return result, nil
		},
	})

	// 检查 URL 工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "check_url",
		Description: "检查 URL 是否可访问",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "要检查的 URL",
				},
			},
			"required": []string{"url"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			url, ok := args["url"].(string)
			if !ok || url == "" {
				return nil, fmt.Errorf("必须提供 URL")
			}

			start := time.Now()
			resp, err := httpClient.Head(url)
			if err != nil {
				return fmt.Sprintf("URL 不可访问: %v", err), nil
			}
			defer resp.Body.Close()

			duration := time.Since(start)

			result := fmt.Sprintf("URL: %s\n", url)
			result += fmt.Sprintf("状态: %s\n", resp.Status)
			result += fmt.Sprintf("响应时间: %v\n", duration)
			result += fmt.Sprintf("内容类型: %s\n", resp.Header.Get("Content-Type"))
			result += fmt.Sprintf("内容长度: %s bytes\n", resp.Header.Get("Content-Length"))

			return result, nil
		},
	})

	// IP 地址查询工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "get_ip_info",
		Description: "获取当前 IP 地址信息",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"provider": map[string]interface{}{
					"type":        "string",
					"description": "API 提供商 (ipify, ipapi)",
					"enum":        []string{"ipify", "ipapi"},
					"default":     "ipify",
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			provider := "ipify"
			if p, ok := args["provider"].(string); ok {
				provider = p
			}

			var url string
			switch provider {
			case "ipify":
				url = "https://api.ipify.org?format=json"
			case "ipapi":
				url = "https://ipapi.co/json/"
			default:
				return nil, fmt.Errorf("不支持的提供商: %s", provider)
			}

			resp, err := httpClient.Get(url)
			if err != nil {
				return nil, fmt.Errorf("请求失败: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("读取响应失败: %w", err)
			}

			return fmt.Sprintf("IP 信息 (使用 %s):\n%s", provider, string(body)), nil
		},
	})

	fmt.Fprintf(os.Stderr, "启动 HTTP 客户端 MCP 服务器...\n")
	fmt.Fprintf(os.Stderr, "可用工具: http_get, http_post_json, check_url, get_ip_info\n")

	if err := server.Start(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "服务器错误: %v\n", err)
		os.Exit(1)
	}
}
