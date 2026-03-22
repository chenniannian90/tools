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

## 🔐 API Key 认证

confmcp 提供了灵活的 API Key 认证中间件，支持简单列表验证和自���义验证逻辑。

### 方式 1: 使用 Server 内置认证（最简单）

```go
server := confmcp.NewServer()
server.Name = "my-secure-server"
server.Protocol = "http"
server.Port = 3000

// 设置 API Keys（三种方式）
// 方式1: 直接设置
server.SetAPIKeys([]string{"secret-key-1", "secret-key-2"})

// 方式2: 添加单个 key
server.AddAPIKey("secret-key-3")

// 方式3: 环境变量（自动读取 MCP_API_KEY��
// export MCP_API_KEY=your-secret-key

// 注册工具
server.RegisterTool(&confmcp.Tool{
    Name:        "protected_tool",
    Description: "需要认证的工具",
    Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
        return "Secret data!", nil
    },
})

// 启动服务器（自动应用 API Key 认证）
server.Start(context.Background())
```

**测试：**
```bash
# 成功调用（有 API Key）
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -H "X-API-Key: secret-key-1" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"protected_tool","arguments":{}}}'

# 失败调用（无 API Key）
curl -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"protected_tool","arguments":{}}}'
# 返回: {"jsonrpc":"2.0","error":{"code":-32001,"message":"Unauthorized: Missing X-API-Key header"}}
```

### 方式 2: 使用自定义验证器（推荐用于生产环境）

业务方可以实现自己的验证逻辑，例如从数据库查询、调用远程认证服务、JWT 验证等。

#### 2.1 数据库验证示例

```go
package main

import (
    "context"
    "net/http"
    "github.com/chenniannian90/tools/confmcp"
)

// 实现自定义验证器
type DatabaseValidator struct {
    db *Database
}

func (v *DatabaseValidator) Validate(apiKey string) (bool, error) {
    // 从数据库验证
    user, err := v.db.FindUserByAPIKey(apiKey)
    if err != nil {
        return false, err
    }
    return user != nil && !user.IsDisabled, nil
}

func main() {
    validator := &DatabaseValidator{db: myDB}

    // 创建 MCP server
    server := confmcp.NewServer()
    server.Name = "database-auth-server"
    server.Protocol = "http"
    server.Port = 3000

    server.RegisterTool(&confmcp.Tool{
        Name:        "secure_tool",
        Description: "需要数据库认证的工具",
        Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            return "Protected by database!", nil
        },
    })

    // 创建自定义 HTTP server 并应用中间件
    mux := http.NewServeMux()

    // 使用自定义验证器
    mux.Handle("/mcp", confmcp.APIKeyAuthWithValidator(validator.Validate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 处理 MCP 请求
        // 这里可以调用 server.HandleRequest()
    })))

    http.ListenAndServe(":3000", mux)
}
```

#### 2.2 远程认证服务示例

```go
// 从远程 REST API 验证
func RemoteAuthValidator(baseURL string) confmcp.APIKeyValidator {
    return func(apiKey string) (bool, error) {
        resp, err := http.Post(
            baseURL+"/validate",
            "application/json",
            strings.NewReader(`{"key":"`+apiKey+`"}`),
        )
        if err != nil {
            return false, err
        }
        defer resp.Body.Close()
        return resp.StatusCode == 200, nil
    }
}

// 使用
mux.Handle("/mcp", confmcp.APIKeyAuthWithValidator(
    RemoteAuthValidator("https://auth.example.com"),
)(http.HandlerFunc(myHandler)))
```

#### 2.3 JWT Token 验证示例

```go
import "github.com/golang-jwt/jwt/v5"

func JWTValidator(secret string) confmcp.APIKeyValidator {
    return func(token string) (bool, error) {
        parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        })
        if err != nil {
            return false, err
        }
        return parsed.Valid, nil
    }
}

// 使用
mux.Handle("/mcp", confmcp.APIKeyAuthWithValidator(
    JWTValidator("my-secret-key"),
)(http.HandlerFunc(myHandler)))
```

### 方式 3: 中间件直接使用（适用于自定义 HTTP server）

```go
import (
    "net/http"
    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    mux := http.NewServeMux()

    // 简单列表验证
    apiKeys := []string{"key1", "key2"}
    mux.HandleFunc("/api/data", confmcp.APIKeyAuthFunc(apiKeys, func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"status": "success"}`))
    }))

    // 从环境变量验证
    mux.Handle("/api/env", confmcp.APIKeyAuthFromEnv("MY_API_KEY")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"status": "authenticated from env"}`))
    })))

    // 可选认证（提供 key 则验证，否则允许访问）
    config := confmcp.APIKeyConfig{
        APIKeys: []string{"optional-key"},
    }
    mux.Handle("/api/public", confmcp.OptionalAPIKey(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"status": "optional auth"}`))
    })))

    http.ListenAndServe(":8080", mux)
}
```

### 认证中间件 API 参考

```go
// 自定义验证器类型
type APIKeyValidator func(apiKey string) (valid bool, err error)

// 配置对象
type APIKeyConfig struct {
    APIKeys     []string           // API Key 列表
    EnvVar      string             // 环境变量名
    DisableAuth bool               // 禁用认证（开发环境）
    Validator   APIKeyValidator    // 自定义验证器
}

// 主要函数
APIKeyAuth(config APIKeyConfig) func(http.Handler) http.Handler
APIKeyAuthFunc(apiKeys []string, handler http.HandlerFunc) http.HandlerFunc
APIKeyAuthWithValidator(validator APIKeyValidator) func(http.Handler) http.Handler
APIKeyAuthWithValidatorFunc(validator APIKeyValidator, handler http.HandlerFunc) http.HandlerFunc
APIKeyAuthFromEnv(envVar string) func(http.Handler) http.Handler
OptionalAPIKey(config APIKeyConfig) func(http.Handler) http.Handler
```

**查看详细文档：**
- [middleware.md](middleware.md) - 完整的中间件文档
- [MIDDLEWARE_QUICKSTART.md](MIDDLEWARE_QUICKSTART.md) - 快速开始指南
- [examples/custom_auth_example.go](examples/custom_auth_example.go) - 5种自定义验证器示例

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
