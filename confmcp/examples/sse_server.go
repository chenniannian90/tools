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

// SSE MCP 服务器示例
// 展示如何使用 SSE (Server-Sent Events) 协议实现 MCP 服务器

func main() {
	// 创建 SSE 配置
	config := &confmcp.MCP{
		Name:     "sse-mcp-server",
		Protocol: "sse",
		Port:     3000,
	}

	// 创建服务器
	server := confmcp.NewServer(config)

	// 注册工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "get_time",
		Description: "获取当前��务器时间",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return fmt.Sprintf("当前时间: %s", time.Now().Format("2006-01-02 15:04:05")), nil
		},
	})

	server.RegisterTool(&confmcp.Tool{
		Name:        "add",
		Description: "加法���算",
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
			return fmt.Sprintf("%.2f + %.2f = %.2f", a, b, a+b), nil
		},
	})

	// 注册资源
	server.RegisterResource(&confmcp.Resource{
		URI:         "info://server",
		Name:        "服务器信息",
		Description: "MCP 服务器运行信息",
		MimeType:    "application/json",
		Handler: func(ctx context.Context, uri string) (*confmcp.ResourceContent, error) {
			info := map[string]interface{}{
				"name":      config.Name,
				"protocol":  config.Protocol,
				"port":      config.Port,
				"uptime":   time.Since(time.Now()).String(),
				"timestamp": time.Now().Format(time.RFC3339),
			}
			jsonData, _ := json.MarshalIndent(info, "", "  ")
			return &confmcp.ResourceContent{
				URI:  uri,
				Text: string(jsonData),
			}, nil
		},
	})

	// 注册提示
	server.RegisterPrompt(&confmcp.Prompt{
		Name:        "system_info",
		Description: "生成系统信息提示",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "请描述当前 MCP 服务器的配置和运行状态。", nil
		},
	})

	// 创建 HTTP 服务器用于 SSE
	mux := http.NewServeMux()

	// SSE 端点
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		// 设置 SSE 响应头
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(os.Stderr, "\n✅ SSE 客户端连接成功: %s\n", r.RemoteAddr)

		// 发送欢迎消息
		sendSSEEvent(w, flusher, "connected", map[string]interface{}{
			"message": "MCP SSE 服务器连接成功",
			"server":  config.Name,
			"time":    time.Now().Format(time.RFC3339),
		})

		// 模拟发送实时更新
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				fmt.Fprintf(os.Stderr, "⛔ SSE 客户端断开: %s\n", r.RemoteAddr)
				return
			case <-ticker.C:
				sendSSEEvent(w, flusher, "heartbeat", map[string]interface{}{
					"time": time.Now().Format(time.RFC3339),
				})
			}
		}
	})

	// MCP 端点
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request confmcp.JSONRPCMessage
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintf(os.Stderr, "📨 收到请求: %s\n", request.Method)

		// 处理请求（这里简化处理）
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      request.ID,
			"result": map[string]interface{}{
				"message": "MCP 请求已处理",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		status := config.LivenessCheck()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"server": config.Name,
			"checks": status,
		})
	})

	// 主页
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>MCP SSE 服务器</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .section { margin: 30px 0; padding: 20px; border: 1px solid #ddd; border-radius: 8px; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 4px; }
        code { background: #e0e0e0; padding: 2px 6px; border-radius: 3px; }
        button { background: #007bff; color: white; border: none; padding: 10px 20px; border-radius: 5px; cursor: pointer; }
        button:hover { background: #0056b3; }
        #events { margin-top: 20px; padding: 15px; background: #f9f9f9; border: 1px solid #ddd; border-radius: 5px; min-height: 200px; max-height: 400px; overflow-y: auto; }
        .event { margin: 5px 0; padding: 8px; background: white; border-left: 3px solid #007bff; }
    </style>
</head>
<body>
    <h1>🚀 MCP SSE 服务器</h1>

    <div class="section">
        <h2>服务端点</h2>
        <div class="endpoint"><strong>SSE:</strong> <code>/sse</code> - SSE 实时通信</div>
        <div class="endpoint"><strong>MCP:</strong> <code>/mcp</code> - MCP JSON-RPC 接口</div>
        <div class="endpoint"><strong>健康检查:</strong> <code>/health</code> - 服务健康状态</div>
    </div>

    <div class="section">
        <h2>SSE 事件测试</h2>
        <button onclick="connectSSE()">连接 SSE</button>
        <button onclick="disconnectSSE()">断开连接</button>
        <div id="events"></div>
    </div>

    <div class="section">
        <h2>可用工具</h2>
        <ul>
            <li><code>get_time</code> - 获取当前时间</li>
            <li><code>add</code> - 加法计算</li>
        </ul>
    </div>

    <script>
        let eventSource = null;

        function connectSSE() {
            if (eventSource) {
                eventSource.close();
            }

            eventSource = new EventSource('/sse');
            const eventsDiv = document.getElementById('events');

            eventSource.onopen = function() {
                addEvent('✅ SSE 连接已建立');
            };

            eventSource.addEventListener('connected', function(e) {
                const data = JSON.parse(e.data);
                addEvent('📡 ' + data.message);
            });

            eventSource.addEventListener('heartbeat', function(e) {
                const data = JSON.parse(e.data);
                addEvent('💓 心跳: ' + data.time);
            });

            eventSource.onerror = function() {
                addEvent('❌ SSE 连接错误');
            };
        }

        function disconnectSSE() {
            if (eventSource) {
                eventSource.close();
                eventSource = null;
                addEvent('⛔ SSE 连接已关闭');
            }
        }

        function addEvent(message) {
            const eventsDiv = document.getElementById('events');
            const eventDiv = document.createElement('div');
            eventDiv.className = 'event';
            eventDiv.textContent = new Date().toLocaleTimeString() + ' - ' + message;
            eventsDiv.insertBefore(eventDiv, eventsDiv.firstChild);
        }
    </script>
</body>
</html>
		`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	})

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.Port)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(os.Stderr, "║           🚀 MCP SSE 服务器启动成功                              ║\n")
	fmt.Fprintf(os.Stderr, "╚══════════════════════════════════════════════════════════════╝\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "📡 SSE 服务: http://localhost%s/sse\n", addr)
	fmt.Fprintf(os.Stderr, "🌐 Web 界面: http://localhost%s/\n", addr)
	fmt.Fprintf(os.Stderr, "💚 健康检查: http://localhost%s/health\n", addr)
	fmt.Fprintf(os.Stderr, "📊 MCP 接口: http://localhost%s/mcp\n", addr)
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

// sendSSEEvent 发送 SSE 事件
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
 jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	flusher.Flush()
}
