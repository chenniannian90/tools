# HTTP 传输层文档

本目录包含使用 HTTP 协议的 MCP 服务器和客户端示例。

## 📡 HTTP 传输层

HTTP 传输层提供基于 RESTful API 的 MCP 通信方式，适合 Web 应用集成和远程访问。

### 核心组件

#### 1. HTTPTransport

HTTP 服务器传输层实现：

```go
transport := confmcp.NewHTTPTransport()

handler := &MyHandler{} // 实现 MCPHandler 接口
err := transport.Start(":3000", handler)
```

#### 2. HTTPClient

HTTP 客户端实现：

```go
client := confmcp.NewHTTPClient("http://localhost:3000")

// 连接到服务器
ctx := context.Background()
err := client.Connect(ctx)

// 发送请求
request := &confmcp.JSONRPCMessage{
    JSONRPC: "2.0",
    ID:      1,
    Method:  "tools/call",
    Params: map[string]interface{}{
        "name": "echo",
        "arguments": map[string]interface{}{
            "message": "Hello!",
        },
    },
}

response, err := client.Send(ctx, request)
```

#### 3. StreamingHTTPClient

支持 SSE 的流式 HTTP 客户端：

```go
client := confmcp.NewStreamingHTTPClient("http://localhost:3000")

// 监听 SSE 事件
err := client.ConnectWithEvents(ctx, func(event string, data []byte) {
    fmt.Printf("Event: %s, Data: %s\n", event, data)
})

// 带重试的请求
response, err := client.SendWithRetry(ctx, request, 3)
```

---

## 🌐 HTTP 服务器示例 (http_server.go)

### 功能特性

1. **MCP JSON-RPC 接口**
   - `/mcp` 端点处理 MCP 请求
   - 支持工具调用
   - 支持资源读取
   - 支持提示生成

2. **健康检查**
   - `/health` 端点
   - 服务状态监控

3. **工具列表**
   - `/tools` 端点
   - 获取所有可用工具

4. **Web 界面**
   - 内置交互式 Web 控制台
   - 测试 MCP 工具
   - 实时结果显示

### 已注册工具

- **ping**: 健康检查
- **echo**: 回显消息
- **calculate**: 数学计算 (加、减、乘、除)

### 已注册资源

- **info://server**: 服务器信息

### 运行方式

```bash
cd examples
go run http_server.go
```

### 访问服务

启动后访问：
- **Web 界面**: http://localhost:3000/
- **MCP 接口**: http://localhost:3000/mcp
- **健康检查**: http://localhost:3000/health
- **工具列表**: http://localhost:3000/tools

### 使用 Web 界面

1. 在浏览器打开 http://localhost:3000/
2. 查看 API 端点和可用工具
3. 在测试区域输入 JSON-RPC 请求
4. 点击"发送请求"查看结果

### API 调用示例

**调用 ping 工具：**
```bash
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "ping",
      "arguments": {}
    }
  }'
```

**调用 echo 工具：**
```bash
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "echo",
      "arguments": {
        "message": "Hello from HTTP!"
      }
    }
  }'
```

**调用 calculate 工具：**
```bash
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "calculate",
      "arguments": {
        "operation": "multiply",
        "a": 15,
        "b": 3
      }
    }
  }'
```

**健康检查：**
```bash
curl http://localhost:3000/health
```

**获取工具列表：**
```bash
curl http://localhost:3000/tools
```

---

## 🌐 HTTP 客户端示例 (http_client_example.go)

展示了如何使用 HTTP 客户端连接到 MCP 服务器：

1. ✅ 健康检查
2. ✅ MCP 工具调用
3. ✅ 工具列表获取
4. ✅ 错误处理

### 运行方式

先启动服务器：
```bash
go run http_server.go
```

然后在另一个终端运行客户端：
```bash
go run http_client_example.go
```

---

## 🔧 HTTP vs stdio vs SSE 对比

| 特性 | stdio | HTTP | SSE |
|------|-------|------|-----|
| **传输方式** | 标准输入/输出 | HTTP RESTful | HTTP Server-Sent Events |
| **使用场景** | 命令行工具、本地进程 | Web 应用、API 集成 | Web 应用、实时推送 |
| **实时通信** | 单向请求-响应 | 单向请求-响应 | 服务端推送事件 |
| **部署方式** | 进程间通信 | HTTP 服务 | HTTP 服务 |
| **浏览器支持** | 不支持 | 原生支持 | 原生支持 |
| **连接管理** | 进程生命周期 | 无状态 | 持久连接 |
| **复杂度** | 低 | 低 | 中 |

### 选择建议

- **使用 stdio**：
  - CLI 工具集成
  - 本地进程通信
  - 管道处理
  - 不需要网络功能

- **使用 HTTP**：
  - RESTful API 集成
  - 无状态请求-响应
  - 标准的 Web 服务
  - 不需要服务端推送

- **使用 SSE**：
  - 需要服务端推送
  - 实时事件通知
  - Web 应用集成
  - 需要持久连接

---

## 🧪 测试 HTTP 功能

### 使用 curl 测试

```bash
# 健康检查
curl http://localhost:3000/health

# 调用工具
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"ping","arguments":{}}}'

# 获取工具列表
curl http://localhost:3000/tools
```

### 使用 JavaScript 测试

```javascript
// 调用 MCP 工具
async function callTool(name, args) {
    const response = await fetch('http://localhost:3000/mcp', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            jsonrpc: '2.0',
            id: 1,
            method: 'tools/call',
            params: {
                name: name,
                arguments: args
            }
        })
    });

    return await response.json();
}

// 使用示例
callTool('echo', { message: 'Hello!' })
    .then(result => console.log(result));
```

### 使用 Python 测试

```python
import requests
import json

def call_tool(name, args):
    url = 'http://localhost:3000/mcp'
    payload = {
        'jsonrpc': '2.0',
        'id': 1,
        'method': 'tools/call',
        'params': {
            'name': name,
            'arguments': args
        }
    }
    response = requests.post(url, json=payload)
    return response.json()

# 使用示例
result = call_tool('calculate', {
    'operation': 'add',
    'a': 10,
    'b': 20
})
print(result)
```

---

## 🚀 高级用法

### 自定义 HTTP 配置

```go
config := &confmcp.HTTPServerConfig{
    Addr:           ":3000",
    ReadTimeout:    10,
    WriteTimeout:   10,
    MaxHeaderBytes: 1 << 20,
    EnableCORS:     true,
    EnableAuth:     false,
}

err := confmcp.StartHTTPServer(mcpConfig, config)
```

### 添加认证

```go
// 在 HTTP 处理器中添加认证
mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
    // 检查 API Key
    apiKey := r.Header.Get("X-API-Key")
    if apiKey != "your-secret-key" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // ... MCP 处理
})
```

### 添加中间件

```go
// 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
        next.ServeHTTP(w, r)
    })
}

// 使用中间件
mux := http.NewServeMux()
handler := loggingMiddleware(mux)
```

---

## 📚 相关资源

- [MCP 协议规范](https://modelcontextprotocol.io/)
- [HTTP 传输层实现](http.go)
- [HTTP 服务器示例](examples/http_server.go)
- [HTTP 客户端示例](examples/http_client_example.go)

---

## 🎯 最佳实践

1. **错误处理**: 总是检查错误并返回适当的 HTTP 状态码
2. **超时设置**: 为所有 HTTP 请求设置合理的超时时间
3. **CORS 配置**: 根据需要配置跨域资源共享
4. **日志记录**: 记录所有请求和响应以便调试
5. **健康检查**: 实现 `/health` 端点用于监控
6. **API 版本**: 考虑在路径中包含版本号（如 `/v1/mcp`）
7. **安全**: 使用 HTTPS 和认证保护生产环境
8. **速率限制**: 实现速率限制防止滥用

---

## 💡 注意事项

- HTTP 传输层是无状态的，每个请求都是独立的
- 对于需要状态管理的场景，考虑使用 SSE 或其他持久连接方式
- 生产环境建议使用反向代理（如 Nginx）
- 注意 CORS 配置，特别是跨域访问时
- HTTP 客户端默认无重试机制，可根据需要实现
