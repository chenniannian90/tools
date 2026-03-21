# Serve 方法使用指南

`Server.Serve()` 是一个便捷方法，用于快速启动 HTTP MCP 服务器。

## 🚀 快速开始

最简单的 HTTP MCP 服务器只需几行代码：

```go
package main

import (
    "context"
    "time"

    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    config := &confmcp.MCP{
        Name:     "my-server",
        Protocol: "http",
        Port:     3000,
    }

    server := confmcp.NewServer(config)

    // 一键启动 HTTP 服务器！
    err := server.Serve([]*confmcp.Tool{
        {
            Name:        "ping",
            Description: "健康检查",
            InputSchema: map[string]interface{}{
                "type": "object",
            },
            Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
                return map[string]interface{}{
                    "status": "ok",
                    "time":   time.Now().Format(time.RFC3339),
                }, nil
            },
        },
    })

    if err != nil {
        panic(err)
    }
}
```

## ✨ 特性

`Serve` 方法自动为你：

1. ✅ **注册所有工具** - 自动注册传入的 Tool 数组
2. ✅ **创建 HTTP 服务器** - 自动设置 http.ServeMux
3. ✅ **配置所有端点**:
   - `/mcp` - MCP JSON-RPC 接口
   - `/health` - 健康检查
   - `/tools` - 工具列表
   - `/resources` - 资源列表
   - `/prompts` - 提示列表
4. ✅ **CORS 支持** - 自动添加跨域支持
5. ✅ **错误处理** - 统一的错误响应格式
6. ✅ **启动服务器** - 开始监听配置的端口

## 📡 可用的 HTTP 端点

### 1. `/mcp` - MCP JSON-RPC 接口

处理所有 MCP 协议请求：

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

### 2. `/health` - 健康检查

返回服务器健康状态：

```bash
curl http://localhost:3000/health
```

响应：
```json
{
  "status": "healthy",
  "server": "my-server",
  "checks": {
    ":3000": "ok"
  }
}
```

### 3. `/tools` - 工具列表

返回所有已注册的工具：

```bash
curl http://localhost:3000/tools
```

响应：
```json
{
  "tools": [
    {
      "name": "ping",
      "description": "健康检查",
      "inputSchema": {...}
    }
  ]
}
```

### 4. `/resources` - 资源列表

返回所有已注册的资源（如果有的话）：

```bash
curl http://localhost:3000/resources
```

### 5. `/prompts` - 提示列表

返回所有已注册的提示（如果有的话）：

```bash
curl http://localhost:3000/prompts
```

## 🎯 使用场景

### 场景 1: 快速原型开发

当你需要快速创建一个 MCP 服务器时：

```go
server := confmcp.NewServer(&confmcp.MCP{
    Name:     "quick-server",
    Protocol: "http",
    Port:     3000,
})

server.Serve([]*confmcp.Tool{
    // 直接在这里定义工具
    { /* tool 1 */ },
    { /* tool 2 */ },
    { /* tool 3 */ },
})
```

### 场景 2: 微服务架构

每个服务只需要一个 HTTP 端点：

```go
// user_service.go
userServer := confmcp.NewServer(&confmcp.MCP{
    Name: "user-service",
    Port: 3001,
})

userServer.Serve([]*confmcp.Tool{
    getUserTool,
    listUsersTool,
    createUserTool,
})

// order_service.go
orderServer := confmcp.NewServer(&confmcp.MCP{
    Name: "order-service",
    Port: 3002,
})

orderServer.Serve([]*confmcp.Tool{
    getOrderTool,
    listOrdersTool,
    createOrderTool,
})
```

### 场景 3: 与前端集成

前端可以直接通过 HTTP 调用工具：

```javascript
// 前端代码
async function callTool(name, args) {
    const response = await fetch('http://localhost:3000/mcp', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            jsonrpc: '2.0',
            id: 1,
            method: 'tools/call',
            params: { name, arguments: args }
        })
    });

    return await response.json();
}

// 使用
const result = await callTool('ping', {});
```

## 🔧 高级用法

### 动态工具注册

```go
// 先创建服务器
server := confmcp.NewServer(config)

// 动态添加工具
tools := []*confmcp.Tool{}
for _, endpoint := range endpoints {
    tools = append(tools, &confmcp.Tool{
        Name:        endpoint.Name,
        Description: endpoint.Description,
        Handler:     createHandler(endpoint),
    })
}

// 然后启动
server.Serve(tools)
```

### 结合资源使用

```go
server := confmcp.NewServer(config)

// 使用 Serve 启动 HTTP 服务器
go server.Serve(tools)

// 在另一个 goroutine 中注册资源
server.RegisterResource(&confmcp.Resource{
    URI:  "config://app",
    Name: "应用配置",
    Handler: func(ctx context.Context, uri string) (*confmcp.ResourceContent, error) {
        return &confmcp.ResourceContent{
            URI:  uri,
            Text: configJSON,
        }, nil
    },
})
```

## 📊 对比传统方式

### 传统方式（需要更多代码）

```go
server := confmcp.NewServer(config)

// 手动注册工具
server.RegisterTool(tool1)
server.RegisterTool(tool2)
server.RegisterTool(tool3)

// 手动创建 HTTP 服务器
mux := http.NewServeMux()

mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
    // 手动处理 MCP 请求...
})

mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    // 手动处理健康检查...
})

// 手动启动
http.ListenAndServe(":3000", mux)
```

### 使用 Serve 方法（简洁）

```go
server := confmcp.NewServer(config)

// 一行代码完成所有事情！
server.Serve([]*confmcp.Tool{tool1, tool2, tool3})
```

## 💡 最佳实践

1. **错误处理**: 始终检查 Serve 返回的错误
   ```go
   if err := server.Serve(tools); err != nil {
       log.Fatal(err)
   }
   ```

2. **优雅关闭**: 使用 context 实现优雅关闭
   ```go
   ctx, cancel := context.WithCancel(context.Background())
   defer cancel()

   // 在信号处理中调用 cancel()
   ```

3. **端口配置**: 通过环境变量配置端口
   ```go
   config := &confmcp.MCP{
       Port: 3000,  // 可通过环境变量 MCP_PORT 覆盖
   }
   ```

4. **工具分组**: 相关工具放在一起管理
   ```go
   userTools := []*confmcp.Tool{getUser, listUsers, createUser}
   orderTools := []*confmcp.Tool{getOrder, listOrders}

   // 可以分别启动不同的服务器
   userServer.Serve(userTools)
   orderServer.Serve(orderTools)
   ```

## 🎨 示例

完整示例请参考：
- [examples/serve_example.go](examples/serve_example.go) - Serve 方法使用示例
- [examples/http_server.go](examples/http_server.go) - 完整的 HTTP 服务器
- [HTTP.md](HTTP.md) - HTTP 传输层详细文档

## 🚦 注意事项

1. Serve 方法会阻塞当前 goroutine，直到服务器停止
2. 如果需要在后台运行，使用 `go server.Serve(tools)`
3. 默认端口为 3000，可通过配置更改
4. CORS 已默认启用，允许所有来源访问
5. 所有端点都返回 JSON 格式响应

## 🔗 相关文档

- [HTTP 传输层](HTTP.md) - HTTP 协议详细说明
- [README.md](README.md) - 项目主文档
- [API.md](API.md) - 完整 API 参考
