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
// 展示如何实现一个提供 HTTP 请求工具的 MCP HTTP 服务器

func main() {
	config := &confmcp.MCP{
		Name:     "http-client-server",
		Protocol: "http",
		Port:     3001,
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

	// 创建 HTTP 服务器
	mux := http.NewServeMux()

	// MCP 端点
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method == http.MethodOptions {
			return
		}

		var request confmcp.JSONRPCMessage
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			sendErrorResponse(w, -32700, "Parse error")
			return
		}

		response := server.HandleRequest(r.Context(), request)
		json.NewEncoder(w).Encode(response)
	})

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := config.LivenessCheck()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"server": config.Name,
			"time":   time.Now().Format(time.RFC3339),
			"checks": status,
		})
	})

	// 工具列表
	mux.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tools := server.GetTools().List()

		toolList := make([]map[string]interface{}, 0, len(tools))
		for _, tool := range tools {
			toolList = append(toolList, map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": tool.InputSchema,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"tools": toolList,
		})
	})

	// Web 界面
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := generateHttpClientWebUI()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.Port)

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(os.Stderr, "║     🌐 HTTP 客户端 MCP 服务器启动成功                        ║\n")
	fmt.Fprintf(os.Stderr, "╚══════════════════════════════════════════════════════════════╝\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "🌐 服务地址: http://localhost%s\n", addr)
	fmt.Fprintf(os.Stderr, "📊 MCP 接口: http://localhost%s/mcp\n", addr)
	fmt.Fprintf(os.Stderr, "💚 健康检查: http://localhost%s/health\n", addr)
	fmt.Fprintf(os.Stderr, "🛠️  工具列表: http://localhost%s/tools\n", addr)
	fmt.Fprintf(os.Stderr, "🎨 Web 界面: http://localhost%s/\n", addr)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "已注册工具: %d\n", server.GetTools().Count())
	fmt.Fprintf(os.Stderr, "- http_get: 发送 HTTP GET 请求\n")
	fmt.Fprintf(os.Stderr, "- http_post_json: 发送 HTTP POST 请求（JSON）\n")
	fmt.Fprintf(os.Stderr, "- check_url: 检查 URL 可访问性\n")
	fmt.Fprintf(os.Stderr, "- get_ip_info: 获取 IP 地址信息\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "按 Ctrl+C 停止服务器\n")
	fmt.Fprintf(os.Stderr, "\n")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "服务器错误: %v\n", err)
		os.Exit(1)
	}
}

// sendErrorResponse 发送错误响应
func sendErrorResponse(w http.ResponseWriter, code int, message string) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// generateHttpClientWebUI 生成 Web 界面
func generateHttpClientWebUI() string {
	return `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HTTP 客户端 MCP 服务器</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        h1 {
            color: white;
            text-align: center;
            margin-bottom: 30px;
            font-size: 2.5em;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }
        .card {
            background: white;
            border-radius: 12px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        .section {
            margin-bottom: 30px;
        }
        .section h2 {
            color: #333;
            margin-bottom: 15px;
            font-size: 1.5em;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }
        .endpoint {
            background: #f7f7f7;
            padding: 15px;
            margin: 10px 0;
            border-radius: 6px;
            border-left: 4px solid #667eea;
        }
        .endpoint strong {
            color: #667eea;
        }
        code {
            background: #f0f0f0;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'Courier New', monospace;
        }
        .tool-list {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 15px;
        }
        .tool-item {
            background: #f9f9f9;
            padding: 15px;
            border-radius: 8px;
            border: 1px solid #e0e0e0;
        }
        .tool-item h3 {
            color: #667eea;
            margin-bottom: 8px;
        }
        .btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            font-size: 16px;
            cursor: pointer;
            transition: all 0.3s;
        }
        .btn:hover {
            background: #5568d3;
            transform: translateY(-2px);
        }
        textarea {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 6px;
            font-family: 'Courier New', monospace;
            resize: vertical;
        }
        #result {
            background: #f9f9f9;
            padding: 15px;
            border-radius: 6px;
            margin-top: 15px;
            min-height: 100px;
            white-space: pre-wrap;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌐 HTTP 客户端 MCP 服务器</h1>

        <div class="card">
            <div class="section">
                <h2>📡 服务端点</h2>
                <div class="endpoint">
                    <strong>MCP 接口:</strong> <code>/mcp</code>
                </div>
                <div class="endpoint">
                    <strong>健康检查:</strong> <code>/health</code>
                </div>
                <div class="endpoint">
                    <strong>工具列表:</strong> <code>/tools</code>
                </div>
            </div>
        </div>

        <div class="card">
            <div class="section">
                <h2>🛠️  可用工具</h2>
                <div class="tool-list">
                    <div class="tool-item">
                        <h3>http_get</h3>
                        <p>发送 HTTP GET 请求</p>
                    </div>
                    <div class="tool-item">
                        <h3>http_post_json</h3>
                        <p>发送 HTTP POST 请求（JSON）</p>
                    </div>
                    <div class="tool-item">
                        <h3>check_url</h3>
                        <p>检查 URL 可访问性</p>
                    </div>
                    <div class="tool-item">
                        <h3>get_ip_info</h3>
                        <p>获取 IP 地址信息</p>
                    </div>
                </div>
            </div>
        </div>

        <div class="card">
            <div class="section">
                <h2>🧪 测试工具</h2>
                <p style="margin-bottom: 10px;">调用 IP 查询工具：</p>
                <button class="btn" onclick="testGetIP()">获取 IP 信息</button>
                <div id="result"></div>
            </div>
        </div>

        <div class="card">
            <div class="section">
                <h2>📊 使用示例</h2>

                <h3 style="color: #667eea; margin: 10px 0;">获取 IP 信息</h3>
                <div class="endpoint">
                    <code>curl -X POST http://localhost:3001/mcp -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_ip_info","arguments":{"provider":"ipify"}}}'</code>
                </div>

                <h3 style="color: #667eea; margin: 10px 0;">检查 URL</h3>
                <div class="endpoint">
                    <code>curl -X POST http://localhost:3001/mcp -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"check_url","arguments":{"url":"https://github.com"}}}'</code>
                </div>

                <h3 style="color: #667eea; margin: 10px 0;">发送 GET 请求</h3>
                <div class="endpoint">
                    <code>curl -X POST http://localhost:3001/mcp -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"http_get","arguments":{"url":"https://api.github.com"}}}'</code>
                </div>
            </div>
        </div>
    </div>

    <script>
        async function testGetIP() {
            const resultDiv = document.getElementById('result');
            resultDiv.textContent = '获取中...';

            try {
                const response = await fetch('/mcp', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        jsonrpc: '2.0',
                        id: 1,
                        method: 'tools/call',
                        params: {
                            name: 'get_ip_info',
                            arguments: {
                                provider: 'ipify'
                            }
                        }
                    })
                });

                const result = await response.json();

                if (result.error) {
                    resultDiv.textContent = '错误: ' + result.error.message;
                } else {
                    // 提取并格式化结果
                    const content = result.result.content;
                    if (typeof content === 'string') {
                        resultDiv.textContent = content;
                    } else {
                        resultDiv.textContent = JSON.stringify(result, null, 2);
                    }
                }
            } catch (error) {
                resultDiv.textContent = '错误: ' + error.message;
            }
        }
    </script>
</body>
</html>
	`
}
