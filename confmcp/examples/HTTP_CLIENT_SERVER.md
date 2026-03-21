# HTTP 客户端 MCP 服务器

这是一个提供 HTTP 请求工具的 MCP 服务器示例，通过 HTTP 协议提供强大的 HTTP 客户端功能。

## 🚀 快速启动

```bash
cd examples
go run http_client_server.go
```

服务器将在 http://localhost:3001 启动

## 🌐 访问服务

启动后访问：
- **Web 界面**: http://localhost:3001/
- **MCP 接口**: http://localhost:3001/mcp
- **健康检查**: http://localhost:3001/health
- **工具列表**: http://localhost:3001/tools

## 🛠️  可用工具

### 1. http_get
发送 HTTP GET 请求

**参数：**
- `url` (必需): 请求的 URL
- `headers` (可选): 请求头对象

**示例：**
```bash
curl -X POST http://localhost:3001/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "http_get",
      "arguments": {
        "url": "https://api.github.com"
      }
    }
  }'
```

### 2. http_post_json
发送 HTTP POST 请求（JSON 格式）

**参数：**
- `url` (必需): 请求的 URL
- `data` (必需): 要发送的 JSON 数据对象
- `headers` (可选): 额外的请求头

**示例：**
```bash
curl -X POST http://localhost:3001/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "http_post_json",
      "arguments": {
        "url": "https://httpbin.org/post",
        "data": {
          "message": "Hello from HTTP Client Server!",
          "timestamp": "2024-01-01T00:00:00Z"
        }
      }
    }
  }'
```

### 3. check_url
检查 URL 是否可访问

**参数：**
- `url` (必需): 要检查的 URL

**示例：**
```bash
curl -X POST http://localhost:3001/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "check_url",
      "arguments": {
        "url": "https://github.com"
      }
    }
  }'
```

**返回示例：**
```
URL: https://github.com
状态: 200 OK
响应时间: 45ms
内容类型: text/html; charset=utf-8
内容长度:  bytes
```

### 4. get_ip_info
获取当前 IP 地址信息

**参数：**
- `provider` (可选): API 提供商，可选值为 "ipify" 或 "ipapi"，默认为 "ipify"

**示例：**
```bash
curl -X POST http://localhost:3001/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "get_ip_info",
      "arguments": {
        "provider": "ipapi"
      }
    }
  }'
```

**使用 ipify（默认）：**
```bash
curl -X POST http://localhost:3001/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "get_ip_info",
      "arguments": {}
    }
  }'
```

## 🎨 Web 界面使用

1. 在浏览器中打开 http://localhost:3001/
2. 查看 API 端点和可用工具
3. 点击"获取 IP 信息"按钮测试工具
4. 查看底部的 curl 命令示例

## 📊 响应格式

所有工具返回字符串格式的响应：

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": "HTTP GET https://api.github.com\n状态码: 200 OK\n响应头:\n  Content-Type: application/json; charset=utf-8\n\n响应内容:\n{...}"
  }
}
```

## 🔧 使用场景

这个服务器适合以下场景：

1. **HTTP 代理**: 从受限环境发起 HTTP 请求
2. **API 测试**: 测试和调试外部 API
3. **数据获取**: 获取远程数据而不需要外部依赖
4. **URL 监控**: 检查 URL 可访问性和响应时间
5. **IP 查询**: 获取服务器公网 IP 信息

## 💡 提示

- HTTP 客户端默认超时时间为 30 秒
- 支持自定义请求头
- 自动处理 JSON 序列化/反序列化
- 完整的错误处理和状态报告

## 🔗 相关资源

- [HTTP 传输层文档](../HTTP.md)
- [MCP 协议规范](https://modelcontextprotocol.io/)
- [HTTP 服务器示例](http_server.go)
