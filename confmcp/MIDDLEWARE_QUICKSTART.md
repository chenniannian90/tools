# API Key 认证中间件 - 快速开始

## 为业务方提供的自定义验证接口

### 核心接口

```go
// APIKeyValidator 自定义验证函数类型
type APIKeyValidator func(apiKey string) (valid bool, err error)
```

### 三种使用方式

#### 1. 简单列表验证（快速开始）

```go
mux.HandleFunc("/api", confmcp.APIKeyAuthFunc(
    []string{"key1", "key2"},
    myHandler,
))
```

#### 2. 配置对象（灵活配置）

```go
config := confmcp.APIKeyConfig{
    APIKeys:     []string{"key1"},
    EnvVar:      "MY_API_KEY",
    DisableAuth: false,
}

mux.Handle("/api", confmcp.APIKeyAuth(config)(myHandler))
```

#### 3. 自定义验证器（推荐用于生产环境）

```go
// 实现你的验证逻辑
myValidator := func(apiKey string) (bool, error) {
    // 从数据库查询
    user := db.FindByAPIKey(apiKey)
    return user != nil, nil
}

mux.Handle("/api", confmcp.APIKeyAuthWithValidator(myValidator)(myHandler))
```

## 自定义验证器示例

### 数据库验证
```go
func validateFromDB(apiKey string) (bool, error) {
    user, err := db.QueryUserByAPIKey(apiKey)
    if err != nil {
        return false, err
    }
    return user != nil && !user.IsDisabled, nil
}
```

### 远程认证服务
```go
func validateFromRemote(apiKey string) (bool, error) {
    resp, err := http.Post("https://auth.example.com/validate", "application/json",
        strings.NewReader(`{"key":"`+apiKey+`"}`))
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    return resp.StatusCode == 200, nil
}
```

### JWT Token
```go
func validateJWT(token string) (bool, error) {
    parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
        return []byte("my-secret"), nil
    })
    return parsed.Valid, err
}
```

### 带权限分级
```go
func validateWithPermissions(apiKey string) (bool, error) {
    user := db.FindByAPIKey(apiKey)
    if user == nil {
        return false, nil
    }
    // 可以存储权限到 context
    // context.WithValue(r.Context(), "permissions", user.Permissions)
    return true, nil
}
```

### 带速率限制
```go
type RateLimitValidator struct {
    requests map[string][]time.Time
    limit    int
}

func (v *RateLimitValidator) Validate(apiKey string) (bool, error) {
    now := time.Now()
    // 清理过期请求
    // 检查限制
    // 记录请求
    return len(v.requests[apiKey]) < v.limit, nil
}
```

## 测试你的验证器

```bash
# 运行所有测试
go test -v

# 测试自定义验证器
go test -v -run TestAPIKeyAuthWithValidator
```

## 文件说明

- `middleware.go` - 中间件实现
- `middleware_test.go` - 测试用例
- `middleware.md` - 完整文档
- `examples/middleware_example.go` - 基础示例
- `examples/custom_auth_example.go` - 自定义验证器示例

## 快速测试

```bash
# 终端 1: 启动服务
go run examples/middleware_example.go

# 终端 2: 测试请求
curl -H "X-API-Key: secret-key-1" http://localhost:8080/api/data
```
