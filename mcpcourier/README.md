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

- `WithServer(server *confmcp.Server)` - 使用自定义服务器实例
- `WithTools(tools []*confmcp.Tool)` - 一次性设置多个工具
- `WithTool(tool *confmcp.Tool)` - 添加单个工具

## 任务集成

`Task` 结构体实现了简单的任务系统接口：

```go
type Task struct {
    server *confmcp.Server
    tools  []*confmcp.Tool
}

func (t *Task) Run(ctx context.Context)
```

这使得它可以轻松集成到任务运行器和工作流系统中。
