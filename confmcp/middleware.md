# confmcp API Key 认证中间件

灵活的 API Key 认证中间件，支持简单列表验证和自定义验证逻辑。

## 特性

- ✅ **简单易用** - 开箱即用的 API Key 列表验证
- ✅ **自定义验证** - 支持业务方实现自己的验证逻辑
- ✅ **环��变量支持** - 从环境变量读取 API Key
- ✅ **多种认证方式** - 支持 X-API-Key header 和 Bearer token
- ✅ **可选认证** - 支持可选的 API Key 验证
- ✅ **CORS 支持** - 自动处理 CORS 预检请求
- ✅ **灵活配置** - 支持禁用认证（开发/测试环境）

## 安装

```bash
go get github.com/chenniannian90/tools/confmcp
```

## 快速开始

### 方式 1: 简单的 API Key 列表验证

```go
package main

import (
    "net/http"
    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    mux := http.NewServeMux()

    // 使用 APIKeyAuthFunc
    apiKeys := []string{"secret-key-1", "secret-key-2"}
    mux.HandleFunc("/api/data", confmcp.APIKeyAuthFunc(apiKeys, func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"status": "success"}`))
    }))

    http.ListenAndServe(":8080", mux)
}
```

测试:
```bash
curl -H "X-API-Key: secret-key-1" http://localhost:8080/api/data
```

### 方式 2: 使用自定义验证器（推荐）

```go
package main

import (
    "net/http"
    "github.com/chenniannian90/tools/confmcp"
)

// 实现你的验证��辑
func myAPIKeyValidator(apiKey string) (bool, error) {
    // 从数据库验证
    user, err := db.ValidateAPIKey(apiKey)
    if err != nil {
        return false, err
    }
    return user != nil, nil
}

func main() {
    mux := http.NewServeMux()

    // 使用自定义验证器
    mux.Handle("/api/data", confmcp.APIKeyAuthWithValidator(myAPIKeyValidator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"status": "success"}`))
    })))

    http.ListenAndServe(":8080", mux)
}
```

### 方式 3: 从环境变量读取

```bash
# 设置环境变量
export MCP_API_KEY=your-secret-key
```

```go
mux.Handle("/api", confmcp.APIKeyAuthFromEnv("MCP_API_KEY")(http.HandlerFunc(myHandler)))
```

## 自定义验证器示例

### 1. 数据库验证

```go
type DatabaseValidator struct {
    db *Database
}

func (v *DatabaseValidator) Validate(apiKey string) (bool, error) {
    user, err := v.db.ValidateAPIKey(apiKey)
    if err != nil {
        return false, err
    }
    return user != nil && !user.Disabled, nil
}

// 使用
validator := &DatabaseValidator{db: myDB}
mux.Handle("/api", confmcp.APIKeyAuthWithValidator(validator.Validate)(myHandler))
```

### 2. 远程认证服务

```go
func RemoteAuthValidator(apiKey string) (bool, error) {
    resp, err := http.Post(
        "https://auth.example.com/validate",
        "application/json",
        strings.NewReader(`{"key":"`+apiKey+`"}`),
    )
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()

    return resp.StatusCode == 200, nil
}

// 使用
mux.Handle("/api", confmcp.APIKeyAuthWithValidator(RemoteAuthValidator)(myHandler))
```

### 3. JWT Token 验证

```go
import "github.com/golang-jwt/jwt/v5"

func JWTValidator(secret string) func(string) (bool, error) {
    return func(token string) (bool, error) {
        parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        })
        return parsed.Valid, err
    }
}

// 使用
mux.Handle("/api", confmcp.APIKeyAuthWithValidator(JWTValidator("my-secret"))(myHandler))
```

### 4. 带权限分级的验证

```go
type PermissionValidator struct {
    keys map[string][]string
}

func (v *PermissionValidator) Validate(apiKey string) (bool, error) {
    permissions, ok := v.keys[apiKey]
    if !ok {
        return false, nil
    }

    // 存储权限信息到请求上下文（可选）
    // ctx := context.WithValue(r.Context(), "permissions", permissions)

    return true, nil
}

// 使用
validator := &PermissionValidator{
    keys: map[string][]string{
        "admin-key": {"read", "write", "delete"},
        "user-key":  {"read"},
    },
}
mux.Handle("/api", confmcp.APIKeyAuthWithValidator(validator.Validate)(myHandler))
```

### 5. 带速率限制的验证

```go
type RateLimitValidator struct {
    requests map[string][]time.Time
    limit    int
    mu       sync.Mutex
}

func (v *RateLimitValidator) Validate(apiKey string) (bool, error) {
    v.mu.Lock()
    defer v.mu.Unlock()

    now := time.Now()
    cutoff := now.Add(-time.Minute)

    // 清理旧记录
    if timestamps, ok := v.requests[apiKey]; ok {
        valid := make([]time.Time, 0)
        for _, ts := range timestamps {
            if ts.After(cutoff) {
                valid = append(valid, ts)
            }
        }
        v.requests[apiKey] = valid
    }

    // 检查限制
    if len(v.requests[apiKey]) >= v.limit {
        return false, fmt.Errorf("rate limit exceeded")
    }

    // 记录请求
    v.requests[apiKey] = append(v.requests[apiKey], now)
    return true, nil
}

// 使用
validator := &RateLimitValidator{
    requests: make(map[string][]time.Time),
    limit:    100, // 每分钟100个请求
}
mux.Handle("/api", confmcp.APIKeyAuthWithValidator(validator.Validate)(myHandler))
```

## API 参考

### APIKeyAuth(config APIKeyConfig) func(http.Handler) http.Handler

创建一个配置完整的 API Key 认证中间件。

```go
config := confmcp.APIKeyConfig{
    APIKeys:     []string{"key1", "key2"},
    EnvVar:      "MY_API_KEY",  // 可选
    DisableAuth: false,          // 可选
    Validator:   myValidator,    // 可选，自定义验证器
}

mux.Handle("/api", confmcp.APIKeyAuth(config)(myHandler))
```

### APIKeyAuthFunc(apiKeys []string, handler http.HandlerFunc) http.HandlerFunc

简化的函数式 API，适合快速使用。

```go
mux.HandleFunc("/api", confmcp.APIKeyAuthFunc([]string{"key1"}, myHandler))
```

### APIKeyAuthWithValidator(validator APIKeyValidator) func(http.Handler) http.Handler

使用自定义验证器创建中间件（推荐）。

```go
mux.Handle("/api", confmcp.APIKeyAuthWithValidator(myValidator)(myHandler))
```

### APIKeyAuthWithValidatorFunc(validator APIKeyValidator, handler http.HandlerFunc) http.HandlerFunc

自定义验证器的函数式版本。

```go
mux.HandleFunc("/api", confmcp.APIKeyAuthWithValidatorFunc(myValidator, myHandler))
```

### APIKeyAuthFromEnv(envVar string) func(http.Handler) http.Handler

从环境变量读取 API Key。

```go
mux.Handle("/api", confmcp.APIKeyAuthFromEnv("MY_API_KEY")(myHandler))
```

### OptionalAPIKey(config APIKeyConfig) func(http.Handler) http.Handler

可选的 API Key 验证（如果提供了则验证，否则跳过）。

```go
mux.Handle("/api", confmcp.OptionalAPIKey(config)(myHandler))
```

### APIKeyFromBearer(apiKeys []string, handler http.HandlerFunc) http.HandlerFunc

使用 Bearer token 格式的认证。

```go
mux.HandleFunc("/api", confmcp.APIKeyFromBearer([]string{"token1"}, myHandler))
```

## 与 confmcp Server 集成

```go
package main

import (
    "context"
    "github.com/chenniannian90/tools/confmcp"
)

func main() {
    // 创建 server
    server := confmcp.NewServer()
    server.Name = "My MCP Server"
    server.Protocol = "http"
    server.Port = 3000

    // 配置 API Keys（简单方式）
    server.SetAPIKeys([]string{"mcp-key-1", "mcp-key-2"})

    // 或者使用自定义验证器
    // 在 mcpHandler 中配置

    // 注册工具
    server.RegisterTool(&confmcp.Tool{
        Name:        "my_tool",
        Description: "My protected tool",
        Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            return "Tool result", nil
        },
    })

    // 启动
    server.Start(context.Background())
}
```

## 测试

运行测试:

```bash
go test -v ./...
```

测试特定的认证场景:

```bash
# 测试基础认证
go test -v -run TestAPIKeyAuth

# 测试自定义验证器
go test -v -run TestAPIKeyAuthWithValidator
```

## 安全建议

1. **生产环境**:
   - 使用 HTTPS
   - 启用 API Key 认证
   - 定期轮换 API Keys
   - 使用强随机 API Keys

2. **API Key 存储**:
   - 不要在代码中硬编码 API Keys
   - 使用环境变量或密钥管理服务
   - 考虑使用哈希存储（如果支持）

3. **验证逻辑**:
   - 实现速率限制
   - 记录失败的验证尝试
   - 考虑 IP 白名单
   - 实现 API Key 过期机制

4. **开发环境**:
   - 使用 `DisableAuth: true` 或简单 API Keys
   - 不要在生产代码中禁用认证

## 常见问题

### Q: 如何禁用认证？
```go
config := confmcp.APIKeyConfig{
    DisableAuth: true,
}
mux.Handle("/api", confmcp.APIKeyAuth(config)(myHandler))
```

### Q: 如何支持多个 API Key？
```go
apiKeys := []string{"key1", "key2", "key3"}
mux.HandleFunc("/api", confmcp.APIKeyAuthFunc(apiKeys, myHandler))
```

### Q: 如何从数据库验证？
```go
validator := func(apiKey string) (bool, error) {
    return db.ValidateAPIKey(apiKey)
}
mux.Handle("/api", confmcp.APIKeyAuthWithValidator(validator)(myHandler))
```

### Q: 如何实现速率限制？
参见上面的 "带速率限制的验证" 示例。

## 许可证

本项目采用 MIT 许可证。
