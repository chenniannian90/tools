package confmcp

import (
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// APIKeyValidator 自定义 API Key 验证函数接口
// 业务方可以实现自己的验证逻辑，例如：
// - 从数据库查询
// - 调用远程认证服务
// - 实现 JWT 验证
// - 实现 API Key 的权限分级
type APIKeyValidator func(apiKey string) (valid bool, err error)

// APIKeyConfig API Key 认证配置
type APIKeyConfig struct {
	// APIKeys 有效的 API Key 列表（简单验证方式）
	APIKeys []string
	// EnvVar 从指定环境变量读取 API Key（可选，默认为 "MCP_API_KEY"）
	EnvVar string
	// DisableAuth 是否禁用认证（用于开发/测试环境）
	DisableAuth bool
	// Validator 自定义验证函数（优先级高于 APIKeys）
	// 如果设置了 Validator，将忽略 APIKeys 字段
	Validator APIKeyValidator
}

// APIKeyAuth 创建 API Key 认证中间件
// 用法:
//
//	mux.HandleFunc("/mcp", APIKeyAuth(config)(myHandler))
func APIKeyAuth(config APIKeyConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置 CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			// 检查是否禁用认证
			if config.DisableAuth {
				logrus.Debugf("API Key auth disabled for %s", r.URL.Path)
				next.ServeHTTP(w, r)
				return
			}

			// 获取 API Key
			apiKey := r.Header.Get("X-API-Key")

			// 从环境变量读取（如果配置了）
			if config.EnvVar != "" {
				if envKey := os.Getenv(config.EnvVar); envKey != "" {
					config.APIKeys = append(config.APIKeys, envKey)
				}
			} else if envKey := os.Getenv("MCP_API_KEY"); envKey != "" {
				// 默认从 MCP_API_KEY 读取
				config.APIKeys = append(config.APIKeys, envKey)
			}

			// 检查是否提供了 API Key
			if apiKey == "" {
				logrus.Warnf("Missing X-API-Key header from %s", r.RemoteAddr)
				respondUnauthorized(w, "Missing X-API-Key header")
				return
			}

			// 验证 API Key
			var isValid bool
			var err error

			// 优先使用自定义验证器
			if config.Validator != nil {
				isValid, err = config.Validator(apiKey)
				if err != nil {
					logrus.Errorf("API Key validation error: %v", err)
					respondUnauthorized(w, "Authentication error")
					return
				}
			} else {
				// 使用默认的列表验证
				if !isValidAPIKey(apiKey, config.APIKeys) {
					logrus.Warnf("Invalid X-API-Key from %s", r.RemoteAddr)
					respondUnauthorized(w, "Invalid API Key")
					return
				}
				isValid = true
			}

			if !isValid {
				logrus.Warnf("Invalid X-API-Key from %s", r.RemoteAddr)
				respondUnauthorized(w, "Invalid API Key")
				return
			}

			// 认证通过，继续处理请求
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyAuthFunc 函数式 API Key 认证中间件
// 用法:
//
//	mux.HandleFunc("/mcp", APIKeyAuthFunc([]string{"key1", "key2"}, myHandler))
func APIKeyAuthFunc(apiKeys []string, handler http.HandlerFunc) http.HandlerFunc {
	config := APIKeyConfig{
		APIKeys: apiKeys,
	}
	middleware := APIKeyAuth(config)
	wrappedHandler := middleware(handler)
	return wrappedHandler.ServeHTTP
}

// APIKeyAuthFromEnv 从环境变量创建 API Key 认证中间件
// 用法:
//
//	mux.HandleFunc("/mcp", APIKeyAuthFromEnv("MY_API_KEY")(myHandler))
func APIKeyAuthFromEnv(envVar string) func(http.Handler) http.Handler {
	config := APIKeyConfig{
		EnvVar: envVar,
	}
	return APIKeyAuth(config)
}

// isValidAPIKey 验证 API Key 是否有效
func isValidAPIKey(apiKey string, validAPIKeys []string) bool {
	// 如果没有配置任何 API Keys，禁用认证
	if len(validAPIKeys) == 0 {
		return true
	}

	// 验证 API Key 是否在有效列表中
	for _, validKey := range validAPIKeys {
		if apiKey == validKey {
			return true
		}
	}

	return false
}

// respondUnauthorized 返回 401 未授权响应
func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32001,"message":"` + message + `"}}`))
}

// RequireAPIKey 简单的 API Key 验证中间件（兼容性函数）
// 用法:
//
//	mux.HandleFunc("/mcp", RequireAPIKey([]string{"key1", "key2"}, myHandler))
func RequireAPIKey(apiKeys []string, handler http.HandlerFunc) http.HandlerFunc {
	return APIKeyAuthFunc(apiKeys, handler)
}

// OptionalAPIKey 可选的 API Key 验证中间件
// 如果提供了 API Key 则验证，否则跳过验证
func OptionalAPIKey(config APIKeyConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")

			// 如果没有提供 API Key，直接通过
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			// 如果提供了 API Key，则进行验证
			APIKeyAuth(config)(next).ServeHTTP(w, r)
		})
	}
}

// APIKeyFromBearer 从 Bearer Token 中提取 API Key
// 用法:
//
//	mux.HandleFunc("/mcp", APIKeyFromBearer([]string{"key1"}, myHandler))
func APIKeyFromBearer(apiKeys []string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			respondUnauthorized(w, "Missing Authorization header")
			return
		}

		// 提取 Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			respondUnauthorized(w, "Invalid Authorization header format")
			return
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")

		// 验证 API Key
		if !isValidAPIKey(apiKey, apiKeys) {
			respondUnauthorized(w, "Invalid API Key")
			return
		}

		// 验证通过，调用原 handler
		handler(w, r)
	}
}

// APIKeyAuthWithValidator 使用自定义验证器创建 API Key 认证中间件
// 这是业务方实现自定义验证逻辑的推荐方式
//
// 用法示例:
//
//	validator := func(apiKey string) (bool, error) {
//	    // 从数据库验证
//	    user, err := db.ValidateAPIKey(apiKey)
//	    if err != nil {
//	        return false, err
//	    }
//	    return user != nil, nil
//	}
//	mux.Handle("/api", APIKeyAuthWithValidator(validator)(myHandler))
func APIKeyAuthWithValidator(validator APIKeyValidator) func(http.Handler) http.Handler {
	config := APIKeyConfig{
		Validator: validator,
	}
	return APIKeyAuth(config)
}

// APIKeyAuthWithValidatorFunc 函数式版本，更简洁
//
// 用法示例:
//
//	mux.HandleFunc("/api", APIKeyAuthWithValidatorFunc(func(apiKey string) (bool, error) {
//	    return myAuthService.Validate(apiKey), nil
//	}, myHandler))
func APIKeyAuthWithValidatorFunc(validator APIKeyValidator, handler http.HandlerFunc) http.HandlerFunc {
	middleware := APIKeyAuthWithValidator(validator)
	wrappedHandler := middleware(handler)
	return wrappedHandler.ServeHTTP
}