package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/chenniannian90/tools/confmcp"
)

// 示例 1: 数据库验证器
// 模拟从数据库查询验证 API Key
type DatabaseValidator struct {
	// 模拟数据库存储
	validKeys map[string]*UserInfo
}

type UserInfo struct {
	Username    string
	Permissions []string
	RateLimit   int
}

func NewDatabaseValidator() *DatabaseValidator {
	return &DatabaseValidator{
		validKeys: map[string]*UserInfo{
			"key-admin-123": {
				Username:    "admin",
				Permissions: []string{"read", "write", "delete"},
				RateLimit:   1000,
			},
			"key-user-456": {
				Username:    "user",
				Permissions: []string{"read"},
				RateLimit:   100,
			},
		},
	}
}

func (v *DatabaseValidator) Validate(apiKey string) (bool, error) {
	// 模拟数据库查询延迟
	time.Sleep(10 * time.Millisecond)

	user, exists := v.validKeys[apiKey]
	if !exists {
		return false, nil
	}

	// 这里可以添加更多验证逻辑，例如：
	// - 检查用户是否被禁用
	// - 检查 API Key 是否过期
	// - 检查 IP 白名单
	// - 记录验证日志

	fmt.Printf("✓ User %s authenticated with rate limit %d\n", user.Username, user.RateLimit)
	return true, nil
}

// 示例 2: 远程认证服务验证器
// 调用远程 REST API 进行验证
type RemoteAuthService struct {
	baseURL    string
	timeout    time.Duration
	cache      map[string]bool
	cacheTTL   time.Duration
	lastUpdate time.Time
}

func NewRemoteAuthService(baseURL string) *RemoteAuthService {
	return &RemoteAuthService{
		baseURL:  baseURL,
		timeout:  5 * time.Second,
		cache:    make(map[string]bool),
		cacheTTL: 5 * time.Minute,
	}
}

func (s *RemoteAuthService) Validate(apiKey string) (bool, error) {
	// 检查缓存
	if time.Since(s.lastUpdate) < s.cacheTTL {
		if valid, ok := s.cache[apiKey]; ok {
			return valid, nil
		}
	}

	// 模拟 HTTP 请求到认证服务
	// 实际使用时应该使用 http.Client 调用真实的服务
	// resp, err := http.Get(fmt.Sprintf("%s/validate?key=%s", s.baseURL, apiKey))

	// 这里只是示例，简化处理
	valid := len(apiKey) > 10 && apiKey[:4] == "remote"

	// 更新缓存
	s.cache[apiKey] = valid
	if len(s.cache) > 1000 {
		s.cache = make(map[string]bool) // 防止内存泄漏
	}

	return valid, nil
}

// 示例 3: JWT Token 验证器
// 验证 JWT Token 并提取用户信息
type JWTValidator struct {
	secretKey string
	issuer    string
}

func NewJWTValidator(secret, issuer string) *JWTValidator {
	return &JWTValidator{
		secretKey: secret,
		issuer:    issuer,
	}
}

func (v *JWTValidator) Validate(token string) (bool, error) {
	// 实际使用时应该使用 github.com/golang-jwt/jwt 库
	// 这里只是示例逻辑

	// 简化的 JWT 验证逻辑
	if len(token) < 20 {
		return false, nil
	}

	// 模拟验证签名、过期时间等
	// parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
	//     return []byte(v.secretKey), nil
	// })

	// 假设所有以 "jwt-" 开头的 token 都是有效的
	if len(token) > 4 && token[:4] == "jwt-" {
		return true, nil
	}

	return false, nil
}

// 示例 4: 带权限分级的验证器
// 不仅验证 API Key，还提取权限信息
type PermissionValidator struct {
	apiKeys map[string][]string // key -> permissions
}

func NewPermissionValidator() *PermissionValidator {
	return &PermissionValidator{
		apiKeys: map[string][]string{
			"key-read":  {"read"},
			"key-write": {"read", "write"},
			"key-admin": {"read", "write", "delete", "admin"},
		},
	}
}

func (v *PermissionValidator) Validate(apiKey string) (bool, error) {
	_, exists := v.apiKeys[apiKey]
	return exists, nil
}

func (v *PermissionValidator) GetPermissions(apiKey string) []string {
	if perms, ok := v.apiKeys[apiKey]; ok {
		return perms
	}
	return []string{}
}

// 示例 5: 带速率限制的验证器
type RateLimitValidator struct {
	requests map[string][]time.Time // key -> request timestamps
	limit    int                    // 每分钟请求数
	window   time.Duration
}

func NewRateLimitValidator(limit int) *RateLimitValidator {
	return &RateLimitValidator{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   time.Minute,
	}
}

func (v *RateLimitValidator) Validate(apiKey string) (bool, error) {
	now := time.Now()
	cutoff := now.Add(-v.window)

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

	// 检查是否超过限制
	if len(v.requests[apiKey]) >= v.limit {
		return false, fmt.Errorf("rate limit exceeded")
	}

	// 记录本次请求
	v.requests[apiKey] = append(v.requests[apiKey], now)

	return true, nil
}

func main() {
	fmt.Println("=== 自定义 API Key 验证示例 ===\n")

	// 示例 1: 使用数据库验证器
	databaseAuthExample()

	// 示例 2: 使用远程认证服务
	remoteAuthExample()

	// 示例 3: 使用 JWT 验证器
	jwtAuthExample()

	// 示例 4: 集成到 confmcp Server
	serverIntegrationExample()

	// 示例 5: 带速率限制的验证器
	rateLimitAuthExample()
}

func databaseAuthExample() {
	fmt.Println("1. 数据库验证器示例:")
	fmt.Println("   验证逻辑: 从数据库查询 API Key 对应的用户信息")

	dbValidator := NewDatabaseValidator()

	mux := http.NewServeMux()

	// 使用自定义验证器创建中间件
	mux.Handle("/api/data", confmcp.APIKeyAuthWithValidator(dbValidator.Validate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "success", "data": "from database"}`))
	})))

	fmt.Println("   ✓ 中间件已配置，使用数据库验证")
	fmt.Println()
}

func remoteAuthExample() {
	fmt.Println("2. 远程认证服务示例:")
	fmt.Println("   验证逻辑: 调用远程 REST API 进行验证")

	remoteAuth := NewRemoteAuthService("https://auth.example.com")

	mux := http.NewServeMux()

	// 使用远程认证验证器
	mux.HandleFunc("/api/remote", confmcp.APIKeyAuthWithValidatorFunc(remoteAuth.Validate, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "success", "auth": "remote service"}`))
	}))

	fmt.Println("   ✓ 中间件已配置，使用远程认证服务")
	fmt.Println()
}

func jwtAuthExample() {
	fmt.Println("3. JWT Token 验证器示例:")
	fmt.Println("   验证逻辑: 验证 JWT Token 签名和过期时间")

	jwtValidator := NewJWTValidator("my-secret-key", "my-app")

	mux := http.NewServeMux()

	mux.Handle("/api/jwt", confmcp.APIKeyAuthWithValidator(jwtValidator.Validate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "success", "auth": "JWT token"}`))
	})))

	fmt.Println("   ✓ 中间件已配置，使用 JWT 验证")
	fmt.Println()
}

func serverIntegrationExample() {
	fmt.Println("4. confmcp Server 集成示例:")

	// 创建自定义验证器
	permValidator := NewPermissionValidator()

	// 创建 MCP server
	server := confmcp.NewServer()
	server.Name = "Custom Auth MCP Server"
	server.Protocol = "http"
	server.Port = 3000

	// 注册工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "protected_tool",
		Description: "A tool protected by custom auth",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "Tool executed successfully!", nil
		},
	})

	// 为 MCP 端点配置自定义认证
	mux := http.NewServeMux()
	mux.Handle("/mcp", confmcp.APIKeyAuthWithValidator(permValidator.Validate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 这里调用 server 的处理逻辑
		w.Write([]byte(`{"jsonrpc":"2.0","result":"MCP endpoint with custom auth"}`))
	})))

	fmt.Println("   ✓ MCP Server 已配置自定义验证器")
	fmt.Println()
}

func rateLimitAuthExample() {
	fmt.Println("5. 速率限制验证器示例:")
	fmt.Println("   验证逻辑: 每个 API Key 每分钟最多 10 个请求")

	rateLimitValidator := NewRateLimitValidator(10)

	mux := http.NewServeMux()

	mux.Handle("/api/ratelimited", confmcp.APIKeyAuthWithValidator(rateLimitValidator.Validate)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "success", "message": "Rate limited endpoint"}`))
	})))

	fmt.Println("   ✓ 中间件已配置，使用速率限制")
	fmt.Println()
}

// 运行示例:
// go run examples/custom_auth_example.go
//
// 测试命令:
// curl -H "X-API-Key: key-admin-123" http://localhost:8080/api/data
// curl -H "X-API-Key: remote-123456" http://localhost:8080/api/remote
// curl -H "X-API-Key: jwt-abc123def456" http://localhost:8080/api/jwt
