# confmcp API 参考文档

本文档提供 confmcp 库的完整 API 参考。

## 目录

- [MCP 配置](#mcp-配置)
- [Server 服务器](#server-服务器)
- [Client 客户端](#client-客户端)
- [Tool 工具](#tool-工具)
- [Resource 资源](#resource-资源)
- [Prompt 提示](#prompt-提示)
- [类型定义](#类型定义)

---

## MCP 配置

MCP 是核心配置结构，参考 confhttp 设计。

### 结构体

```go
type MCP struct {
    // 端口配置，使用 opt,expose 标签
    Port int `env:",opt,expose"`

    // 协议类型：stdio 或 sse
    Protocol string `env:""`

    // 服务器名称
    Name string `env:""`

    // 内部字段
    Initialized bool   `env:"-"`
    retry        *Retry `env:"-"`
}
```

### 方法

#### SetDefaults()

设置默认配置值。

```go
func (m *MCP) SetDefaults()
```

**默认值：**
- `Protocol`: "stdio"
- `Port`: 3000 (仅当 Protocol 为 "sse" 时)
- `retry.Repeats`: 3
- `retry.Interval`: 10秒

#### Init()

初始化 MCP。

```go
func (m *MCP) Init()
```

多次调用不会产生副作用。

#### GetAddress()

获取服务地址。

```go
func (m *MCP) GetAddress() string
```

**返回值：**
- stdio 协议: "stdio"
- SSE 协议: ":3000" 或配置的端口

#### GetServerInfo()

获取服务器信息。

```go
func (m *MCP) GetServerInfo() ServerInfo
```

**返回值：** 包含名称、版本、协议、地址和能力的服务器信息。

#### LivenessCheck()

健康检查。

```go
func (m *MCP) LivenessCheck() map[string]string
```

**返回值：** 地址到状态的映射。

**示例：**
```go
// 已初始化
{"stdio": "ok"}
// 未初始化
{"stdio": "not initialized"}
```

#### SetRetryConfig()

设置重试配置。

```go
func (m *MCP) SetRetryConfig(repeats int, interval time.Duration)
```

**参数：**
- `repeats`: 重试次数
- `interval`: 重试间隔

---

## Server 服务器

MCP 服务器实现。

### 结构体

```go
type Server struct {
    // 私有字段
}
```

### 构造函数

#### NewServer()

创建新的 MCP 服务器。

```go
func NewServer(config *MCP) *Server
```

**参数：**
- `config`: MCP 配置

**返回值：** 新的服务器实例

### 方法

#### Start()

启动服务器。

```go
func (s *Server) Start(ctx context.Context) error
```

**参数：**
- `ctx`: 上下文

**返回值：** 错误信息

**阻塞操作**，直到服务器停止或出错。

#### Stop()

停止服务器。

```go
func (s *Server) Stop()
```

#### RegisterTool()

注册工具。

```go
func (s *Server) RegisterTool(tool *Tool) error
```

**参数：**
- `tool`: 工具定义

**返回值：** 错误信息

#### RegisterResource()

注册资源。

```go
func (s *Server) RegisterResource(resource *Resource) error
```

**参数：**
- `resource`: 资源定义

**返回值：** 错误信息

#### RegisterPrompt()

注册提示。

```go
func (s *Server) RegisterPrompt(prompt *Prompt) error
```

**参数：**
- `prompt`: 提示定义

**返回值：** 错误信息

#### GetTools()

获取工具注册表。

```go
func (s *Server) GetTools() *ToolRegistry
```

#### GetResources()

获取资源注册表。

```go
func (s *Server) GetResources() *ResourceRegistry
```

#### GetPrompts()

获取提示注册表。

```go
func (s *Server) GetPrompts() *PromptRegistry
```

---

## Client 客户端

MCP 客户端实现。

### 结构体

```go
type Client struct {
    // 私有字段
}
```

### 构造函数

#### NewClient()

创建新的 MCP 客户端。

```go
func NewClient(config *MCP) *Client
```

**参数：**
- `config`: MCP 配置

**返回值：** 新的客户端实例

### 方法

#### Connect()

连接到服务器。

```go
func (c *Client) Connect(ctx context.Context) error
```

**参数：**
- `ctx`: 上下文

**返回值：** 错误信息

#### Close()

关闭连接。

```go
func (c *Client) Close() error
```

#### ListTools()

列出可用工具。

```go
func (c *Client) ListTools(ctx context.Context) ([]*Tool, error)
```

**返回值：** 工具列表和错误

#### CallTool()

调用工具。

```go
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)
```

**参数：**
- `name`: 工具名称
- `args`: 工具参数

**返回值：** 工具执行结果和错误

#### ListResources()

列出可用资源。

```go
func (c *Client) ListResources(ctx context.Context) ([]*Resource, error)
```

#### ReadResource()

读取资源。

```go
func (c *Client) ReadResource(ctx context.Context, uri string) (*ResourceContent, error)
```

**参数：**
- `uri`: 资源 URI

**返回值：** 资源内容和错误

#### ListPrompts()

列出可用提示。

```go
func (c *Client) ListPrompts(ctx context.Context) ([]*Prompt, error)
```

#### GetPrompt()

获取提示。

```go
func (c *Client) GetPrompt(ctx context.Context, name string, args map[string]interface{}) (string, error)
```

**参数：**
- `name`: 提示名称
- `args`: 提示参数

**返回值：** 生成的提示文本和错误

---

## Tool 工具

工具是可被调用的函数。

### 结构体

```go
type Tool struct {
    Name        string                 // 工具名称（必填）
    Description string                 // 工具描述
    InputSchema map[string]interface{} // 输入参数的 JSON Schema
    Handler     ToolHandler            // 处理函数
}

type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)
```

### ToolRegistry 工具注册表

#### 方法

```go
// 注册工具
func (r *ToolRegistry) Register(tool *Tool) error

// 注销工具
func (r *ToolRegistry) Unregister(name string)

// 获取工具
func (r *ToolRegistry) Get(name string) (*Tool, bool)

// 列出所有工具
func (r *ToolRegistry) List() []*Tool

// 执行工具
func (r *ToolRegistry) Execute(ctx context.Context, name string, args map[string]interface{}) (interface{}, error)

// 工具数量
func (r *ToolRegistry) Count() int

// 清空工具
func (r *ToolRegistry) Clear()

// 检查工具是否存在
func (r *ToolRegistry) Has(name string) bool
```

---

## Resource 资源

资源代表服务器提供的数据。

### 结构体

```go
type Resource struct {
    URI         string           // 资源 URI（必填）
    Name        string           // 资源名称
    Description string           // 资源描述
    MimeType    string           // MIME 类型
    Handler     ResourceHandler  // 处理函数
}

type ResourceHandler func(ctx context.Context, uri string) (*ResourceContent, error)

type ResourceContent struct {
    URI      string // 资源 URI
    MimeType string // MIME 类型
    Text     string // 文本内容
    Blob     []byte // 二进制内容
}
```

### ResourceRegistry 资源注册表

#### 方法

```go
// 注册资源
func (r *ResourceRegistry) Register(resource *Resource) error

// 注销资源
func (r *ResourceRegistry) Unregister(uri string)

// 获取资源
func (r *ResourceRegistry) Get(uri string) (*Resource, bool)

// 列出所有资源
func (r *ResourceRegistry) List() []*Resource

// 读取资源
func (r *ResourceRegistry) Read(ctx context.Context, uri string) (*ResourceContent, error)

// 资源数量
func (r *ResourceRegistry) Count() int

// 清空资源
func (r *ResourceRegistry) Clear()

// 检查资源是否存在
func (r *ResourceRegistry) Has(uri string) bool
```

---

## Prompt 提示

提示是预定义的提示模板。

### 结构体

```go
type Prompt struct {
    Name        string           // 提示名称（必填）
    Description string           // 提示描述
    Arguments   []PromptArgument // 参数定义
    Handler     PromptHandler    // 处理函数
}

type PromptArgument struct {
    Name        string // 参数名称
    Description string // 参数描述
    Required    bool   // 是否必填
}

type PromptHandler func(ctx context.Context, args map[string]interface{}) (string, error)
```

### PromptRegistry 提示注册表

#### 方法

```go
// 注册提示
func (r *PromptRegistry) Register(prompt *Prompt) error

// 注销提示
func (r *PromptRegistry) Unregister(name string)

// 获取提示
func (r *PromptRegistry) Get(name string) (*Prompt, bool)

// 列出所有提示
func (r *PromptRegistry) List() []*Prompt

// 生成提示
func (r *PromptRegistry) Generate(ctx context.Context, name string, args map[string]interface{}) (string, error)

// 提示数量
func (r *PromptRegistry) Count() int

// 清空提示
func (r *PromptRegistry) Clear()

// 检查提示是否存在
func (r *PromptRegistry) Has(name string) bool
```

---

## 类型定义

### ServerInfo

服务器信息。

```go
type ServerInfo struct {
    Name         string       // 服务器名称
    Version      string       // 服务器版本
    Protocol     string       // 协议类型
    Address      string       // 服务地址
    Capabilities Capabilities // 能力列表
}
```

### Capabilities

服务器能力。

```go
type Capabilities struct {
    Tools     bool // 支持工具
    Resources bool // 支持资源
    Prompts   bool // 支持提示
    Roots     bool // 支持根列表
    Sampling  bool // 支持采样
}
```

### Retry

重试配置。

```go
type Retry struct {
    Repeats  int              // 重试次数
    Interval envconf.Duration // 重试间隔
}

func (r Retry) Do(exec func() error) error
```

### JSONRPCMessage

JSON-RPC 2.0 消息。

```go
type JSONRPCMessage struct {
    JSONRPC string        // 协议版本 "2.0"
    ID      interface{}   // 请求 ID
    Method  string        // 方法名
    Params  interface{}   // 参数
    Result  interface{}   // 结果
    Error   *JSONRPCError // 错误
}

type JSONRPCError struct {
    Code    int    // 错误代码
    Message string // 错误消息
}
```

---

## 使用示例

### 完整的服务器示例

```go
package main

import (
    "context"
    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    // 创建配置
    config := &confmcp.MCP{
        Name:     "my-server",
        Protocol: "stdio",
    }

    // 创建服务器
    server := confmcp.NewServer(config)

    // 注册工具
    server.RegisterTool(&confmcp.Tool{
        Name:        "greet",
        Description: "向指定的人打招呼",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "name": map[string]interface{}{
                    "type": "string",
                },
            },
            "required": []string{"name"},
        },
        Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            name := args["name"].(string)
            return "Hello, " + name + "!", nil
        },
    })

    // 启动服务器
    server.Start(context.Background())
}
```

### 完整的客户端示例

```go
package main

import (
    "context"
    "fmt"
    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    // 创建客户端
    client := confmcp.NewClient(&confmcp.MCP{
        Protocol: "stdio",
    })

    // 连接
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        panic(err)
    }
    defer client.Close()

    // 调用工具
    result, err := client.CallTool(ctx, "greet", map[string]interface{}{
        "name": "World",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result) // 输出: Hello, World!
}
```

---

更多示例请查看 [examples/](examples/) 目录。
