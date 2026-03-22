package mcpcourier

import (
	"net/http"

	"github.com/chenniannian90/tools/confmcp"
)

type Option func(*Task)

func WithServer(server *confmcp.Server) Option {
	return func(t *Task) {
		t.server = server
	}
}

func WithTools(tools []*confmcp.Tool) Option {
	return func(t *Task) {
		t.tools = tools
	}
}

func WithTool(tool *confmcp.Tool) Option {
	return func(t *Task) {
		t.tools = append(t.tools, tool)
	}
}

// WithMiddleware 添加自定义 HTTP 中间件
// 用于在 HTTP/SSE 模式下包装 MCP 端点
// 用法:
//
//	task := mcpcourier.NewTask(
//	    mcpcourier.WithTool(myTool),
//	    mcpcourier.WithMiddleware(func(next http.Handler) http.Handler {
//	        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	            // 前置处理
//	            log.Printf("Request: %s", r.URL.Path)
//	            next.ServeHTTP(w, r)
//	            // 后置处理
//	        })
//	    }),
//	)
//	task.Run(ctx)
func WithMiddleware(middleware func(http.Handler) http.Handler) Option {
	return func(t *Task) {
		t.middlewares = append(t.middlewares, middleware)
	}
}

// WithAPIKeyValidator 使用自定义 API Key 验证器
// 支持从数据库、远程服务等验证
//
// 简单场景示例（列表验证）:
//
//	validator := func(apiKey string) (bool, error) {
//	    validKeys := []string{"key1", "key2"}
//	    for _, key := range validKeys {
//	        if apiKey == key {
//	            return true, nil
//	        }
//	    }
//	    return false, nil
//	}
//	task := mcpcourier.NewTask(
//	    mcpcourier.WithTool(myTool),
//	    mcpcourier.WithAPIKeyValidator(validator),
//	)
//	task.Run(ctx)
//
// 数据库验证示例:
//
//	validator := func(apiKey string) (bool, error) {
//	    user, err := db.ValidateAPIKey(apiKey)
//	    return user != nil, err
//	}
//	task := mcpcourier.NewTask(
//	    mcpcourier.WithTool(myTool),
//	    mcpcourier.WithAPIKeyValidator(validator),
//	)
//	task.Run(ctx)
//
// 从环境变量读取示例:
//
//	validator := func(apiKey string) (bool, error) {
//	    envKey := os.Getenv("MCP_API_KEY")
//	    return apiKey == envKey, nil
//	}
func WithAPIKeyValidator(validator confmcp.APIKeyValidator) Option {
	return func(t *Task) {
		t.apiKeyConfig = &confmcp.APIKeyConfig{
			Validator: validator,
		}
	}
}
