# mcpcourier API Key 认证

mcpcourier 提供了灵活的 `WithAPIKeyValidator` 选项来配置 API Key 认证，支持各种验证场景。

## WithAPIKeyValidator

使用自定义验证函数来验证 API Key。

```go
task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
)
task.Run(ctx)
```

## 使用示例

### 1. 简单列表验证

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

### 2. 从环境变量读取

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

### 3. 数据库验证（推荐生产环境）

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

### 4. 远程认证服务

```go
validator := func(apiKey string) (bool, error) {
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

task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
)
task.Run(ctx)
```

### 5. JWT Token 验证

```go
import "github.com/golang-jwt/jwt/v5"

validator := func(token string) (bool, error) {
    parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
        return []byte("my-secret-key"), nil
    })
    if err != nil {
        return false, err
    }
    return parsed.Valid, nil
}

task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
)
task.Run(ctx)
```

### 6. 带权限分级

```go
type PermissionValidator struct {
    db *Database
}

func (v *PermissionValidator) Validate(apiKey string) (bool, error) {
    user, err := v.db.FindByAPIKey(apiKey)
    if err != nil {
        return false, err
    }
    if user == nil || user.IsDisabled {
        return false, nil
    }

    // 可以存储权限到 context 供后续使用
    // context.WithValue(ctx, "permissions", user.Permissions)

    return true, nil
}

validator := (&PermissionValidator{db: myDB}).Validate

task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
)
task.Run(ctx)
```

### 7. 带速率限制

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

    // 清理过期记录
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

validator := (&RateLimitValidator{
    requests: make(map[string][]time.Time),
    limit:    100,
}).Validate

task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
)
task.Run(ctx)
```

## WithMiddleware - 自定义中间件

除了 API Key 验证，还可以添加自定义 HTTP 中间件。

```go
loggingMiddleware := func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
    mcpcourier.WithMiddleware(loggingMiddleware),
)
task.Run(ctx)
```

## 组合使用

可以组合多个中间件：

```go
task := mcpcourier.NewTask(
    mcpcourier.WithTool(myTool),
    mcpcourier.WithAPIKeyValidator(validator),
    mcpcourier.WithMiddleware(corsMiddleware),
    mcpcourier.WithMiddleware(loggingMiddleware),
    mcpcourier.WithMiddleware(rateLimitMiddleware),
)
task.Run(ctx)
```

## 完整示例

```go
package main

import (
    "context"
    "log"
    "net/http"
    "github.com/chenniannian90/tools/confmcp"
    "github.com/chenniannian90/tools/mcpcourier"
)

func main() {
    // API Key 验证器
    validator := func(apiKey string) (bool, error) {
        user, err := db.ValidateAPIKey(apiKey)
        return user != nil && !user.IsDisabled, err
    }

    // 日志中间件
    logger := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
            next.ServeHTTP(w, r)
        })
    }

    task := mcpcourier.NewTask(
        mcpcourier.WithTool(&confmcp.Tool{
            Name:        "secure_tool",
            Description: "需要认证的工具",
            Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
                return "Protected data!", nil
            },
        }),
        mcpcourier.WithAPIKeyValidator(validator),
        mcpcourier.WithMiddleware(logger),
    )

    task.Run(context.Background())
}
```

## 测试

```bash
# 启动服务
go run examples/auth_example.go

# 测试 API Key 认证
curl -H "X-API-Key: key-1" http://localhost:3001/mcp

# 测试环境变量
export MCP_API_KEY=my-env-key
curl -H "X-API-Key: my-env-key" http://localhost:3002/mcp

# 测试数据库验证
curl -H "X-API-Key: user-123" http://localhost:3003/mcp
```

## 优势

- ✅ **灵活** - 支持任意验证逻辑
- ✅ **简单** - 只需实现一个函数
- ✅ **可扩展** - 支持数据库、远程服务、JWT 等
- ✅ **可组合** - 可以添加多个中间件

## 相关文档

- [confmcp 中间件文档](../confmcp/middleware.md) - 底层中间件实现
- [README.md](README.md) - 项目说明
- [示例代码](examples/auth_example.go) - 完整示例代码
