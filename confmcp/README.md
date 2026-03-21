# confmcp

[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

confmcp 是一个用 Go 实现的 **MCP (Model Context Protocol)** 服务器框架，专注于 **Tools（工具）** 功能，遵循与 `confmysql`、`confredis` 和 `confhttp` 相同的设计模式。

## ✨ 特性

- 🛠️ **工具 (Tools)** - 定义和执行可调用的工具
- 🔄 **重试机制** - 内置连接重试和错误恢复
- 🌐 **多传输层支持**:
  - **stdio** - 标准输入输出（默认）
  - **HTTP** - HTTP RESTful API
  - **SSE** - Server-Sent Events（支持服务器推送）
- 📝 **环境变量配置** - 使用 `envconf` 进行灵活的配置管理
- 🎯 **简洁设计** - 只关注核心功能，易于使用和维护

## 📦 安装

```bash
go get github.com/chenniannian90/tools/confmcp
```

## 🚀 快速开始

### 最简单的 stdio 服务器

```go
package main

import (
    "context"
    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    // 创建服务器
    server := confmcp.NewServer()
    server.Name = "my-server"

    // 注册工具
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
        },
        Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            return args["message"], nil
        },
    })

    // 启动服务器（stdio 模式）
    server.Start(context.Background())
}
```

### HTTP 服务器（推荐）

使用 HTTP 协议更简单，只需一行配置：

```go
package main

import (
    "context"
    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    // 创建服务器
    server := confmcp.NewServer()
    server.Name = "my-http-server"
    server.Protocol = "http"
    server.Port = 3000  // 可选，默认 3000

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
                "message": "Server is running!",
            }, nil
        },
    })

    // 启动服务器
    server.Start(context.Background())
}
```

**可用的 HTTP 端点：**
- `http://localhost:3000/mcp` - MCP JSON-RPC 接口
- `http://localhost:3000/health` - 健康检查
- `http://localhost:3000/tools` - 工具列表

**测试：**
```bash
# 健康检查
curl http://localhost:3000/health

# 调用工具
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"ping","arguments":{}}}'
```

### 使用 Serve 方法（最简洁）

如果你不想手动注册工具，可以使用 `Serve` 方法一次性传入所有工具：

```go
server := confmcp.NewServer()
server.Name = "my-server"
server.Protocol = "http"
server.Port = 3000

// 一行代码启动！自动注册所有工具并创建 HTTP 端点
server.Serve([]*confmcp.Tool{
    {
        Name:        "ping",
        Description: "健康检查",
        InputSchema: map[string]interface{}{"type": "object"},
        Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            return "pong", nil
        },
    },
})
```

### SSE 服务器（支持实时推送）

SSE (Server-Sent Events) 协议支持服务器向客户端主动推送事件：

```go
package main

import (
    "context"
    "time"
    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    server := confmcp.NewServer()
    server.Name = "sse-server"
    server.Protocol = "sse"
    server.Port = 3001

    // 注册工具
    server.RegisterTool(&confmcp.Tool{
        Name:        "get_time",
        Description: "获取当前时间",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "timezone": map[string]interface{}{
                    "type": "string",
                    "default": "UTC",
                },
            },
        },
        Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            return map[string]interface{}{
                "time": time.Now().Format(time.RFC3339),
            }, nil
        },
    })

    // 启动服务器
    server.Start(context.Background())
}
```

**SSE 端点：**
- `http://localhost:3001/sse` - SSE 事件流（实时推送）
- `http://localhost:3001/mcp` - MCP JSON-RPC 接口
- `http://localhost:3001/health` - 健康检查
- `http://localhost:3001/tools` - 工具列表

**测试 SSE：**
```bash
# 监听 SSE 事件流
curl -N http://localhost:3001/sse

# 在另一个终端调用工具
curl -X POST http://localhost:3001/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_time","arguments":{}}}'
```

### 使用客户端

```go
// 创建客户端
client := confmcp.NewClient(&confmcp.MCP{
    Protocol: "stdio",
})

// 连接
client.Connect(context.Background())

// 列出工具
tools, _ := client.ListTools(context.Background())

// 调用工具
result, _ := client.CallTool(context.Background(), "echo", map[string]interface{}{
    "message": "Hello MCP!",
})
```

## 📚 核心概念

### Server 配置

```go
server := confmcp.NewServer()
server.Name = "my-server"      // 服务器名称
server.Protocol = "stdio"      // 协议: "stdio", "http", "sse"
server.Port = 3000            // 端口（HTTP/SSE 时使用）
```

**设计特点（参考 confhttp）：**
- ✅ 只配置 `Port`，不需要 `Host`
- ✅ 通过 `GetAddress()` 获取服务地址
- ✅ `LivenessCheck()` 返回健康状态
- ✅ 简洁的配置结构

### 工具 (Tools)

工具是可被调用的函数：

```go
server.RegisterTool(&confmcp.Tool{
    Name:        "calculate",
    Description: "执行数学计算",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "expression": map[string]interface{}{
                "type": "string",
            },
        },
        "required": []string{"expression"},
    },
    Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
        expr := args["expression"].(string)
        // 执行计算逻辑
        return result, nil
    },
})
```

**工具方法：**
- `RegisterTool(tool)` - 注册工具
- `GetTools()` - 获取工具注册表
- `GetTools().List()` - 列出所有工具
- `GetTools().Execute(ctx, name, args)` - 执行工具

## 🧪 测试

运行测试：

```bash
# 运行所有测试
go test -v .

# 运行特定测试
go test -v -run TestServer

# 测试覆盖率
go test -cover .
```

## 📝 项目结构

```
confmcp/
├── server.go           # MCP 服务器实现（stdio/http/sse）
├── client.go           # MCP 客户端实现
├── tools.go            # 工具注册表
├── types.go            # 核��类型定义
├── retry.go            # 重试逻辑
├── go.mod              # Go 模块定义
├── README.md           # 本文件
└── examples/           # 示例代码
    ├── serve_example.go       # HTTP 服务器示例
    ├── calculator_server.go   # 计算器工具示例
    ├── simple_echo_server.go  # Echo 服务器示例
    └── sse_example.go         # SSE 服务器示例
```

## 🎯 API 参考

### Server

```go
// 创建服务器
server := confmcp.NewServer()

// 配置
server.Name = "my-server"
server.Protocol = "stdio"  // "stdio", "http", "sse"
server.Port = 3000

// 工具操作
server.RegisterTool(tool)
server.GetTools()

// 启动/停止
server.Start(ctx)
server.Stop()

// 便捷方法
server.Serve(tools)  // 一次性启动 HTTP 服务器
```

### Client

```go
// 创建客户端
client := confmcp.NewClient(config)

// 连接
client.Connect(ctx)

// 工具操作
client.ListTools(ctx)
client.CallTool(ctx, name, args)

// 关闭
client.Close()
```

### Tool

```go
type Tool struct {
    Name        string                 // 工具名称
    Description string                 // 工具描述
    InputSchema map[string]interface{} // 输入参数 JSON Schema
    Handler     ToolHandler            // 处理函数
}

type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)
```

## 🔗 相关链接

- [MCP 协议规范](https://modelcontextprotocol.io/)
- [confhttp](../svcutil/confhttp/) - HTTP 配置库
- [confmysql](../confmysql/) - MySQL 配置库
- [confredis](../confredis/) - Redis 配置库

## 📄 许可证

MIT License
