# confmcp 使用示例

本目录包含了 confmcp 库的完整使用示例。

## 示例列表

### 1. 简单的 Echo 服务器 (simple_echo_server.go)

最简单的 MCP 服务器示例，实现一个 echo 工具。

### 2. 文件系统工具服务器 (filesystem_server.go)

实现文件系统操作的服务器，包括读取、写入、列表等功能。

### 3. HTTP 客户端工具 (http_client_server.go)

提供 HTTP 请求功能的 MCP 服务器。

### 4. 计算器服务器 (calculator_server.go)

提供数学计算功能的服务器。

### 5. 多功能综合服务器 (complete_server.go)

展示 tools、resources、prompts 三种功能综合使用的完整示例。

### 6. 客户端示例 (client_example.go)

展示如何使用 MCP 客户端连接到服务器。

### 7. 环境变量配置示例 (env_config_example.go)

展示如何通过环境变量配置 MCP。

## 运行示例

每个示例都可以直接运行：

```bash
# 运行简单 echo 服务器
go run examples/simple_echo_server.go

# 运行文件系统服务器
go run examples/filesystem_server.go

# 运行完整示例
go run examples/complete_server.go
```

## 测试 MCP 服务器

你可以使用 MCP 客户端测试这些服务器，或者通过 stdin/stdout 发送 JSON-RPC 消息：

```bash
# 发送初始化请求
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | go run examples/simple_echo_server.go

# 列出工具
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | go run examples/simple_echo_server.go

# 调用工具
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"message":"Hello MCP!"}}}' | go run examples/simple_echo_server.go
```

## 学习路径

建议按以下顺序学习：

1. 先看 `simple_echo_server.go` 了解基本结构
2. 看 `env_config_example.go` 了解配置方式
3. 看 `calculator_server.go` 了解参数验证
4. 看 `filesystem_server.go` 了解资源管理
5. 看 `http_client_server.go` 了解复杂工具
6. 最后看 `complete_server.go` 了解完整功能
7. 看 `client_example.go` 了解客户端使用
