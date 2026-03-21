package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/chenniannian90/tools/confmcp"
)

// HTTP MCP 服务器示例
// 展示如何使用 HTTP 传输层实现 MCP 服务器

func main() {
	// 创建配置
	config := &confmcp.MCP{
		Name:     "http-mcp-server",
		Protocol: "http",
		Port:     3000,
	}

	// 创建服务器
	server := confmcp.NewServer(config)

	// 注册工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "ping",
		Description: "健康检查",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"status": "ok",
				"time":   time.Now().Format(time.RFC3339),
				"server": config.Name,
			}, nil
		},
	})

	server.RegisterTool(&confmcp.Tool{
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
	})

	server.RegisterTool(&confmcp.Tool{
		Name:        "calculate",
		Description: "数学计算",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"operation": map[string]interface{}{
					"type": "string",
					"enum": []interface{}{"add", "subtract", "multiply", "divide"},
				},
				"a": map[string]interface{}{
					"type": "number",
				},
				"b": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []string{"operation", "a", "b"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			op := args["operation"].(string)
			a, _ := args["a"].(float64)
			b, _ := args["b"].(float64)

			var result float64
			switch op {
			case "add":
				result = a + b
			case "subtract":
				result = a - b
			case "multiply":
				result = a * b
			case "divide":
				if b == 0 {
					return nil, fmt.Errorf("除数不能为零")
				}
				result = a / b
			}

			return fmt.Sprintf("%.2f %s %.2f = %.2f", a, op, b, result), nil
		},
	})

	// 注册资源
	server.RegisterResource(&confmcp.Resource{
		URI:         "info://server",
		Name:        "服务器信息",
		Description: "MCP 服务器信息",
		MimeType:    "application/json",
		Handler: func(ctx context.Context, uri string) (*confmcp.ResourceContent, error) {
			info := map[string]interface{}{
				"name":      config.Name,
				"version":   "1.0.0",
				"protocol":  "http",
				"uptime":    time.Since(time.Now()).String(),
				"timestamp": time.Now().Format(time.RFC3339),
			}
			jsonData, _ := json.MarshalIndent(info, "", "  ")
			return &confmcp.ResourceContent{
				URI:  uri,
				Text: string(jsonData),
			}, nil
		},
	})

	// 创建 Handler 适配器
	handler := &MCPHandler{Server: server}

	// 创建 HTTP 复用器
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
			sendError(w, -32700, "Parse error")
			return
		}

		response := handler.HandleMessage(r.Context(), &request)
		json.NewEncoder(w).Encode(response)
	})

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := config.LivenessCheck()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"server":  config.Name,
			"time":    time.Now().Format(time.RFC3339),
			"checks":  status,
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
		html := generateWebUI()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.Port)

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(os.Stderr, "║           🌐 HTTP MCP 服务器启动成功                          ║\n")
	fmt.Fprintf(os.Stderr, "╚══════════════════════════════════════════════════════════════╝\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "🌐 服务地址: http://localhost%s\n", addr)
	fmt.Fprintf(os.Stderr, "📊 MCP 接口: http://localhost%s/mcp\n", addr)
	fmt.Fprintf(os.Stderr, "💚 健康检查: http://localhost%s/health\n", addr)
	fmt.Fprintf(os.Stderr, "🛠️  工具列表: http://localhost%s/tools\n", addr)
	fmt.Fprintf(os.Stderr, "🎨 Web 界面: http://localhost%s/\n", addr)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "已注册工具: %d\n", server.GetTools().Count())
	fmt.Fprintf(os.Stderr, "已注册资源: %d\n", server.GetResources().Count())
	fmt.Fprintf(os.Stderr, "已注册提示: %d\n", server.GetPrompts().Count())
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "按 Ctrl+C 停止服务器\n")
	fmt.Fprintf(os.Stderr, "\n")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "服务器错误: %v\n", err)
		os.Exit(1)
	}
}

// MCPHandler 处理 MCP 请求
type MCPHandler struct {
	Server *confmcp.Server
}

// HandleMessage 处理 JSON-RPC 消息
func (h *MCPHandler) HandleMessage(ctx context.Context, message *confmcp.JSONRPCMessage) *confmcp.JSONRPCMessage {
	return h.Server.HandleRequest(ctx, *message)
}

// sendError 发送错误响应
func sendError(w http.ResponseWriter, code int, message string) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	json.NewEncoder(w).Encode(response)
}

// generateWebUI 生成 Web 界面
func generateWebUI() string {
	return `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HTTP MCP 服务器</title>
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
        <h1>🌐 HTTP MCP 服务器</h1>

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
                        <h3>ping</h3>
                        <p>健康检查</p>
                    </div>
                    <div class="tool-item">
                        <h3>echo</h3>
                        <p>回显消息</p>
                    </div>
                    <div class="tool-item">
                        <h3>calculate</h3>
                        <p>数学计算</p>
                    </div>
                </div>
            </div>
        </div>

        <div class="card">
            <div class="section">
                <h2>🧪 测试工具</h2>
                <p style="margin-bottom: 10px;">输入 JSON-RPC 请求：</p>
                <textarea id="request" rows="10">{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "echo",
    "arguments": {
      "message": "Hello from HTTP MCP!"
    }
  }
}</textarea>
                <br><br>
                <button class="btn" onclick="sendRequest()">发送请求</button>
                <div id="result"></div>
            </div>
        </div>

        <div class="card">
            <div class="section">
                <h2>📊 使用示例</h2>

                <h3 style="color: #667eea; margin: 10px 0;">Ping 服务器</h3>
                <div class="endpoint">
                    <code>curl http://localhost:3000/health</code>
                </div>

                <h3 style="color: #667eea; margin: 10px 0;">调用 echo 工具</h3>
                <div class="endpoint">
                    <code>curl -X POST http://localhost:3000/mcp -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello"}}}'</code>
                </div>

                <h3 style="color: #667eea; margin: 10px 0;">调用 calculate 工具</h3>
                <div class="endpoint">
                    <code>curl -X POST http://localhost:3000/mcp -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"calculate","arguments":{"operation":"add","a":10,"b":20}}}'</code>
                </div>
            </div>
        </div>
    </div>

    <script>
        async function sendRequest() {
            const requestText = document.getElementById('request').value;
            const resultDiv = document.getElementById('result');

            try {
                resultDiv.textContent = '发送中...';

                const response = await fetch('/mcp', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: requestText
                });

                const result = await response.json();
                resultDiv.textContent = JSON.stringify(result, null, 2);
            } catch (error) {
                resultDiv.textContent = '错误: ' + error.message;
            }
        }
    </script>
</body>
</html>
	`
}
