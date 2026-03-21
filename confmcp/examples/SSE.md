# SSE (Server-Sent Events) 示例

本目录包含使用 SSE 协议的 MCP 服务器和客户端示例。

## 📡 SSE 服务器示例 (sse_server.go)

实现了基于 SSE 的 MCP 服务器，包含：

### 功能特性

1. **SSE 实时通信**
   - `/sse` 端点提供 SSE 实时事件流
   - 自动心跳机制
   - 连接状态管理

2. **MCP JSON-RPC 接口**
   - `/mcp` 端点处理 MCP 请求
   - 支持工具调用
   - 支持资源读取
   - 支持提示生成

3. **Web 界面**
   - 内置 Web 控制台
   - SSE 连接测试
   - 实时事件显示

4. **健康检查**
   - `/health` 端点
   - 服务状态监控

### 运行方式

```bash
go run sse_server.go
```

### 访问服务

启动后访问：
- **Web 界面**: http://localhost:3000/
- **SSE 端点**: http://localhost:3000/sse
- **健康检查**: http://localhost:3000/health
- **MCP 接口**: http://localhost:3000/mcp

### 使用 Web 界面

1. 在浏览器打开 http://localhost:3000/
2. 点击"连接 SSE"按钮
3. 查看实时事件流
4. 点击"断开连接"停止接收

### API 调用示例

**调用 get_time 工具：**
```bash
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "get_time",
      "arguments": {}
    }
  }'
```

**调用 add 工具：**
```bash
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "add",
      "arguments": {
        "a": 10,
        "b": 20
      }
    }
  }'
```

**健康检查：**
```bash
curl http://localhost:3000/health
```

---

## 🌐 SSE 客户端示例 (sse_client.go)

实现了 SSE MCP 客户端，展示如何：

1. **连接到 SSE 服务器**
2. **监听实时事件**
3. **调用 MCP 工具**
4. **健康检查**

### 运行方式

先启动服务器：
```bash
go run sse_server.go
```

然后在另一个终端运行客户端：
```bash
go run sse_client.go
```

### 客户端功能

- ✅ SSE 连接测试
- ✅ 健康检查
- ✅ MCP 工具调用
- ✅ 实时事件监听
- ✅ 自动重连机制

---

## 🔧 SSE vs stdio 对比

| 特性 | stdio | SSE |
|------|-------|-----|
| **传输方式** | 标准输入/输出 | HTTP Server-Sent Events |
| **使用场景** | 命令行工具、本地进程 | Web 应用、远程服务 |
| **实时通信** | 单向请求-响应 | 服务端推送事件 |
| **部署方式** | 进程间通信 | HTTP 服务 |
| **浏览器支持** | 不支持 | 原生支持 |
| **连接管理** | 进程生命周期 | 持久连接 |

### 选择建议

- **使用 stdio**：
  - CLI 工具集成
  - 本地进程通信
  - 管道处理
  - 不需要 Web 界面

- **使用 SSE**：
  - Web 应用集成
  - 需要服务端推送
  - 远程服务访问
  - 需要 Web 界面

---

## 📝 SSE 协议格式

SSE 是一种服务器推送技术，使用简单文本格式：

```
event: connected
data: {"message": "连接成功", "time": "2024-01-01T00:00:00Z"}

event: heartbeat
data: {"time": "2024-01-01T00:00:10Z"}

event: custom_event
data: {"key": "value"}
```

### 关键点

1. **Content-Type**: `text/event-stream`
2. **事件格式**: `event: <name>\ndata: <json>\n\n`
3. **心跳**: 定期发送保持连接
4. **重连**: 浏览器自动重连

---

## 🧪 测试 SSE 功能

### 使用 curl 测试

```bash
# 连接 SSE 端点（会持续接收事件）
curl -N http://localhost:3000/sse
```

### 使用 JavaScript 测试

```javascript
const eventSource = new EventSource('http://localhost:3000/sse');

eventSource.addEventListener('connected', (e) => {
  const data = JSON.parse(e.data);
  console.log('连接成功:', data);
});

eventSource.addEventListener('heartbeat', (e) => {
  const data = JSON.parse(e.data);
  console.log('心跳:', data.time);
});

eventSource.onerror = (e) => {
  console.error('连接错误:', e);
};
```

### 使用 Python 测试

```python
import requests
import json

def test_sse():
    r = requests.get('http://localhost:3000/sse', stream=True)
    r.headers.set('Accept', 'text/event-stream')

    for line in r.iter_lines():
        if line:
            print(f'Event: {line}')

if __name__ == '__main__':
    test_sse()
```

---

## 🚀 高级用法

### 自定义 SSE 事件

```go
// 在 sse_server.go 中添加自定义事件
sendSSEEvent(w, flusher, "custom_event", map[string]interface{}{
    "message": "自定义事件数据",
    "timestamp": time.Now().Unix(),
})
```

### 客户端过滤事件

```javascript
// 只接收特定事件
eventSource.addEventListener('heartbeat', (e) => {
    // 只处理心跳事件
});

// 或者处理所有事件
eventSource.onmessage = (e) => {
    const data = JSON.parse(e.data);
    console.log('收到事件:', data);
};
```

### 添加认证

```go
// 在 SSE 端点添加认证
mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
    // 检查 API Key
    apiKey := r.Header.Get("X-API-Key")
    if apiKey != "your-secret-key" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // ... SSE 处理
})
```

---

## 📚 相关资源

- [MDN - Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events)
- [MCP 协议规范](https://modelcontextprotocol.io/)
- [EventSource API](https://developer.mozilla.org/en-US/docs/Web/API/EventSource)

---

## 🎯 最佳实践

1. **心跳机制**: 每 10-30 秒发送心跳保持连接
2. **重连策略**: 客户端自动重连，指数退避
3. **错误处理**: 优雅处理连接断开
4. **资源清理**: 断开时正确关闭连接
5. **CORS**: 配置跨域支持
6. **缓冲**: 使用 Flush 确保实时推送

---

## 💡 注意事项

- SSE 是单向的（服务器到客户端）
- 如果需要双向通信，考虑 WebSocket
- 生产环境建议使用反向代理（如 Nginx）
- 注意连接数限制和资源管理
