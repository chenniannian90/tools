# confmcp

[![Go Version](https://img.shields.io/badge/Go-1.20+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

confmcp 是一个用 Go 实现的 **MCP (Model Context Protocol)** 配置和框架库，遵循与 `confmysql`、`confredis` 和 `confhttp` 相同的设计模式。

## ✨ 特性

- 🔧 **完整的 MCP 协议支持** - 实现了 MCP 2024-11-05 版本协议
- 🛠️ **工具 (Tools)** - 定义和执行可调用的工具
- 📄 **资源 (Resources)** - 管理和读取服务器资源
- 💬 **提示 (Prompts)** - 生成和使用预定义的提示模板
- 🔄 **重试机制** - 内置连接重试和错误恢复
- 🌐 **多传输层支持**:
  - **stdio** - 标准输入输出（默认）
  - **HTTP** - HTTP RESTful API
  - **SSE** - Server-Sent Events（实时通信）
- 📝 **环境变量配置** - 使用 `envconf` 进行灵活的配置管理
- 🧪 **完整测试覆盖** - 53+ 测试用例，100% 通过率

## 📦 安装

```bash
go get github.com/chenniannian90/tools/confmcp
```

## 🚀 快速开始

### 最简单的服务器

```go
package main

import (
    "context"
    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    // 创建配置（参考 confhttp 的简洁设计）
    config := &confmcp.MCP{
        Name:     "my-server",
        Protocol: "stdio",
    }

    // 创建服务器
    server := confmcp.NewServer(config)

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

    // 启动服务器
    server.Start(context.Background())
}
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

### 使用 Serve 方法（推荐）

最简单的方式启动 HTTP MCP 服务器：

```go
config := &confmcp.MCP{
    Name:     "my-server",
    Protocol: "http",
    Port:     3000,
}

server := confmcp.NewServer(config)

// 一键启动 HTTP 服务器！自动处理所有路由和端点
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
    {
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
            return args["message"], nil
        },
    },
})

if err != nil {
    log.Fatal(err)
}
```

`Serve` 方法自动为你：
- ✅ 注册所有工具
- ✅ 创建 HTTP 服务器
- ✅ 配置所有端点 (`/mcp`, `/health`, `/tools`, `/resources`, `/prompts`)
- ✅ 启用 CORS 支持
- ✅ 处理错误和日志

**可用端点：**
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

# 获取工具列表
curl http://localhost:3000/tools
```

详见 [SERVE.md](SERVE.md)

### 使用 HTTP 协议

创建 HTTP 服务器（端口 3000）：

```go
package main

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    config := &confmcp.MCP{
        Name:     "http-server",
        Protocol: "http",
        Port:     3000,
    }

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
                "message": "Server is running",
            }, nil
        },
    })

    // 创建 HTTP 服务器
    mux := http.NewServeMux()

    // MCP 端点
    mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Access-Control-Allow-Origin", "*")

        var request confmcp.JSONRPCMessage
        json.NewDecoder(r.Body).Decode(&request)

        response := server.HandleRequest(r.Context(), request)
        json.NewEncoder(w).Encode(response)
    })

    // 健康检查
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "healthy",
        })
    })

    // 启动 HTTP 服务器
    http.ListenAndServe(":3000", mux)
}
```

使用 HTTP 客户端：

```go
// 创建 HTTP 客户端
client := confmcp.NewHTTPClient("http://localhost:3000")

// 连接到服务器
client.Connect(context.Background())

// 发送请求
request := &confmcp.JSONRPCMessage{
    JSONRPC: "2.0",
    ID:      1,
    Method:  "tools/call",
    Params: map[string]interface{}{
        "name": "ping",
        "arguments": map[string]interface{}{},
    },
}

response, _ := client.Send(context.Background(), request)
```

使用 curl 测试：

```bash
# 调用工具
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

# 健康检查
curl http://localhost:3000/health
```

## 📚 文档

### 设计模式

confmcp 遵循与 `confhttp`、`confmysql` 和 `confredis` 相同的设计模式：

1. **配置结构** - 使用 `env` 标签支持环境变量
2. **SetDefaults()** - 设置默认配置值
3. **Init()** - 初始化连接
4. **重试机制** - 内置的连接重试逻辑

### 核心组件

#### MCP 配置

```go
config := &confmcp.MCP{
    Name:     "my-server",
    Protocol: "stdio",        // 或 "http", "sse"
    Port:     3000,           // HTTP/SSE 协议时使用
}
```

**设计特点（参考 confhttp）：**
- ✅ 只配置 `Port`，不需要 `Host`
- ✅ 使用 `env:",opt,expose"` 标签
- ✅ 通过 `GetAddress()` 获取服务地址
- ✅ `LivenessCheck()` 返回健康状态
- ✅ 简洁的配置结构，业务配置不在框架中

#### 工具 (Tools)

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

#### 资源 (Resources)

资源代表服务器提供的数据：

```go
server.RegisterResource(&confmcp.Resource{
    URI:         "file:///config",
    Name:        "配置文件",
    Description: "服务器配置",
    MimeType:    "application/json",
    Handler: func(ctx context.Context, uri string) (*confmcp.ResourceContent, error) {
        return &confmcp.ResourceContent{
            URI:  uri,
            Text: configJSON,
        }, nil
    },
})
```

#### 提示 (Prompts)

提示是预定义的提示模板：

```go
server.RegisterPrompt(&confmcp.Prompt{
    Name:        "code_review",
    Description: "代码审查提示",
    Arguments: []confmcp.PromptArgument{
        {
            Name:        "language",
            Description: "编程语言",
            Required:    false,
        },
    },
    Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
        language := args["language"].(string)
        return fmt.Sprintf("请审查以下 %s 代码：", language), nil
    },
})
```

### 环境变量配置

可以通过环境变量配置 MCP：

```bash
export MCP_NAME=my-server
export MCP_PROTOCOL=stdio
export MCP_PORT=3000
```

### 重试机制

内置的重试机制确保连接的可靠性：

```go
config.SetRetryConfig(5, 10*time.Second)  // 重试5次，间隔10秒
```

## 📖 示例

查看 [examples/](examples/) 目录获取完整示例：

### 基础示例
- **simple_echo_server.go** - 最简单的 echo 服务器
- **calculator_server.go** - 计算器服务器
- **filesystem_server.go** - 文件系统操作服务器
- **http_client_server.go** - HTTP 客户端工具（HTTP 协议）
- **complete_server.go** - 完整功能服务器

### 传输层示例
- **sse_server.go** - SSE 服务器示例
- **sse_client.go** - SSE 客户端示例
- **http_server.go** - HTTP 服务器示例（带 Web UI）
- **http_client_example.go** - HTTP 客户端示例

### 客户端和配置示例
- **client_example.go** - 客户端使用示例
- **env_config_example.go** - 环境变量配置示例

运行示例：

```bash
cd examples

# stdio 协议
go run simple_echo_server.go

# HTTP 协议
go run http_server.go        # HTTP 服务器（端口 3000）
go run http_client_server.go # HTTP 客户端服务器（端口 3001）

# SSE 协议
go run sse_server.go         # SSE 服务器（端口 3000）
```

## 🧪 测试

运行所有测试：

```bash
# 运行所有测试
go test -v ./...

# 运行特定测试
go test -v -run TestMCP

# 测试覆盖率
go test -cover ./...

# 竞态检测
go test -race ./...
```

**测试覆盖：**
- ✅ 53+ 测试用例
- ✅ 100% 通过率
- ✅ 竞态检测通过
- ✅ ~10秒执行时间

## 🎯 API 参考

### 核心接口

#### MCP 配置

```go
type MCP struct {
    Port     int    `env:",opt,expose"`
    Protocol string `env:""`
    Name     string `env:""`
}

// 方法
func (m *MCP) SetDefaults()
func (m *MCP) Init()
func (m *MCP) GetAddress() string
func (m *MCP) GetServerInfo() ServerInfo
func (m *MCP) LivenessCheck() map[string]string
func (m *MCP) SetRetryConfig(repeats int, interval time.Duration)
```

#### Server

```go
server := confmcp.NewServer(config)

// 注册功能
server.RegisterTool(tool)
server.RegisterResource(resource)
server.RegisterPrompt(prompt)

// 启动服务器
server.Start(context.Background())

// 停止服务器
server.Stop()
```

#### Client

```go
client := confmcp.NewClient(config)

// 连接
client.Connect(ctx)

// 工具操作
client.ListTools(ctx)
client.CallTool(ctx, name, args)

// 资源操作
client.ListResources(ctx)
client.ReadResource(ctx, uri)

// 提示操作
client.ListPrompts(ctx)
client.GetPrompt(ctx, name, args)
```

## 📝 项目结构

```
confmcp/
├── mcp.go              # MCP 核心配置
├── client.go           # MCP 客户端实现
├── server.go           # MCP 服务器实现
├── tools.go            # 工具注册表
├── resources.go        # 资源注册表
├── prompts.go          # 提示注册表
├── retry.go            # 重试逻辑
├── http.go             # HTTP 传输层实现
├── common.go           # 通用类型定义
├── go.mod              # Go 模块定义
├── README.md           # 本文件
├── HTTP.md             # HTTP 传输层文档
├── examples/           # 示例代码
│   ├── README.md
│   ├── SSE.md          # SSE 使用指南
│   ├── HTTP_CLIENT_SERVER.md  # HTTP 客户端服务器文档
│   ├── simple_echo_server.go
│   ├── calculator_server.go
│   ├── filesystem_server.go
│   ├── http_client_server.go
│   ├── http_server.go        # HTTP 服务器示例
│   ├── http_client_example.go # HTTP 客户端示例
│   ├── sse_server.go         # SSE 服务器示例
│   ├── sse_client.go         # SSE 客户端示例
│   ├── complete_server.go
│   ├── client_example.go
│   └── env_config_example.go
├── mcp_test.go         # MCP 测试
├── tools_test.go       # 工具测试
├── resources_test.go   # 资源测试
├── prompts_test.go     # 提示测试
├── retry_test.go       # 重试测试
└── integration_test.go # 集成测试
```

## 🔧 高级用法

### 健康检查

```go
status := config.LivenessCheck()
// 返回: {"stdio": "ok"} 或 {":3000": "ok"}
```

### 获取服务器信息

```go
info := config.GetServerInfo()
// 返回服务器名称、版本、协议等信息
```

### 自定义重试策略

```go
config.SetRetryConfig(3, 5*time.Second)  // 重试3次，每次间隔5秒
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🔗 相关链接

- [MCP 协议规范](https://modelcontextprotocol.io/)
- [HTTP 传输层文档](HTTP.md) - HTTP 协议使用指南
- [examples/README.md](examples/README.md) - 示例代码说明
- [examples/SSE.md](examples/SSE.md) - SSE 协议使用指南
- [examples/HTTP_CLIENT_SERVER.md](examples/HTTP_CLIENT_SERVER.md) - HTTP 客户端服务器文档
- [confhttp](../svcutil/confhttp/) - HTTP 配置库
- [confmysql](../confmysql/) - MySQL 配置库
- [confredis](../confredis/) - Redis 配置库

## 🙏 致谢

- 感谢 MCP 协议的设计者
- 感谢 Go 社区的支持
- 感谢 confhttp、confmysql、confredis 的设计参考

---

**注意**: confmcp 已完成开发并通过所有测试，可投入生产使用！
