# mcpcourier

MCP (Model Context Protocol) 服务器任务封装，用于集成到任务系统中。

## 概述

`mcpcourier` 提供了基于任务的 `confmcp.Server` 封装，遵循与 `courier` 包相同的设计模式。它允许将 MCP 服务器轻松集成到基于任务的工作流中。

## 使用方法

```go
package main

import (
	"context"
	"github.com/chenniannian90/tools/confmcp"
	"github.com/chenniannian90/tools/mcpcourier"
)

func main() {
	// 定义你的工具
	tools := []*confmcp.Tool{
		{
			Name:        "echo",
			Description: "回显输入内容",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "需要回显的文本",
					},
				},
				"required": []string{"text"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return args["text"], nil
			},
		},
	}

	// 创建并运行任务
	task := mcpcourier.NewTask(
		mcpcourier.WithTools(tools),
	)

	task.Run(context.Background())
}
```

## 配置选项

### 基础选项
- `WithServer(server *confmcp.Server)` - 使用自定义服务器实例
- `WithTools(tools []*confmcp.Tool)` - 一次性设置多个工具
- `WithTool(tool *confmcp.Tool)` - 添加单个工具

### API Key 认证
- `WithAPIKeyValidator(validator APIKeyValidator)` - 使用自定义 API Key 验证器

### 中间件选项
- `WithMiddleware(middleware func(http.Handler) http.Handler)` - 添加自定义 HTTP 中间件

## API Key 认证示例

### 简单列表验证

```go
validator := func(apiKey string) (bool, error) {
    validKeys := []string{"secret-key-1", "secret-key-2"}
    for _, key := range validKeys {
        if apiKey == key {
            return true, nil
        }
    }
    return false, nil
}

task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
)
task.Run(ctx)
```

**测试：**
```bash
curl -H "X-API-Key: secret-key-1" http://localhost:3000/mcp
```

### 从环境变量读取

```go
validator := func(apiKey string) (bool, error) {
    envKey := os.Getenv("MCP_API_KEY")
    return apiKey == envKey, nil
}

task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
)
task.Run(ctx)
```

### 数据库验证（推荐生产环境）

```go
validator := func(apiKey string) (bool, error) {
    user, err := db.ValidateAPIKey(apiKey)
    if err != nil {
        return false, err
    }
    return user != nil && !user.IsDisabled, nil
}

task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
)
task.Run(ctx)
```

## 自定义中间件示例

### 日志中间件

```go
loggingMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithMiddleware(loggingMiddleware),
)
task.Run(ctx)
```

### 组合使用（API Key + 中间件）

```go
validator := func(apiKey string) (bool, error) {
    return apiKey == "secret-key", nil
}

task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
    mcpcourier.WithMiddleware(loggingMiddleware),
    mcpcourier.WithMiddleware(corsMiddleware),
)
task.Run(ctx)
```

## 任务集成

`Task` 结构体实现了简单的任务系统接口：

```go
type Task struct {
    server       *confmcp.Server
    tools        []*confmcp.Tool
    middlewares  []func(http.Handler) http.Handler
    apiKeyConfig *confmcp.APIKeyConfig
}

func (t *Task) Run(ctx context.Context)
```

这使得它可以轻松集成到任务运行器和工作流系统中。

## 完整示例

查看 [examples/auth_example.go](examples/auth_example.go) 获取更多示例：
- 简单的 API Key 列表验证
- 从环境变量读取
- 自定义验证器（数据库验证）
- 自定义中间件
- 组合使用

## 详细文档

- [AUTH_OPTIONS.md](AUTH_OPTIONS.md) - API Key 认证和中间件的完整文档
- [confmcp 中间件文档](../confmcp/middleware.md) - 底层中间件实现
