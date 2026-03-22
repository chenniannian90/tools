package confmcp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAPIKeyAuth_Success(t *testing.T) {
	apiKeys := []string{"test-key-1", "test-key-2"}
	handler := APIKeyAuthFunc(apiKeys, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-API-Key", "test-key-1")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "success" {
		t.Errorf("Expected body 'success', got '%s'", w.Body.String())
	}
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	apiKeys := []string{"test-key-1"}
	handler := APIKeyAuthFunc(apiKeys, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	apiKeys := []string{"test-key-1"}
	handler := APIKeyAuthFunc(apiKeys, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAPIKeyAuth_NoKeysConfigured(t *testing.T) {
	// 如果没有配置任何 keys，仍然需要提供 API Key（安全默认值）
	// 要禁用认证，应使用 DisableAuth: true
	apiKeys := []string{}
	handler := APIKeyAuthFunc(apiKeys, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("allowed"))
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	// 当没有配置 keys 但也不禁用认证时，仍然要求提供 API Key
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 (empty key list still requires auth), got %d", w.Code)
	}
}

func TestAPIKeyAuth_DisableAuthExplicitly(t *testing.T) {
	// 正确的方式：显式禁用认证
	config := APIKeyConfig{
		APIKeys:     []string{},
		DisableAuth: true,
	}

	middleware := APIKeyAuth(config)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("allowed without auth"))
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (auth explicitly disabled), got %d", w.Code)
	}
}

func TestAPIKeyAuth_DisabledAuth(t *testing.T) {
	config := APIKeyConfig{
		APIKeys:     []string{"test-key"},
		DisableAuth: true,
	}

	middleware := APIKeyAuth(config)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("no auth"))
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	// 不提供 API Key
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (auth disabled), got %d", w.Code)
	}
}

func TestAPIKeyAuthFromEnv(t *testing.T) {
	// 设置环境变量
	os.Setenv("TEST_API_KEY", "env-key-123")
	defer os.Unsetenv("TEST_API_KEY")

	middleware := APIKeyAuthFromEnv("TEST_API_KEY")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("env auth"))
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-API-Key", "env-key-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestOptionalAPIKey_WithKey(t *testing.T) {
	config := APIKeyConfig{
		APIKeys: []string{"optional-key"},
	}

	middleware := OptionalAPIKey(config)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("authenticated"))
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-API-Key", "optional-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestOptionalAPIKey_WithoutKey(t *testing.T) {
	config := APIKeyConfig{
		APIKeys: []string{"optional-key"},
	}

	middleware := OptionalAPIKey(config)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("anonymous"))
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	// 不提供 API Key
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 (optional auth), got %d", w.Code)
	}
}

func TestAPIKeyFromBearer_Success(t *testing.T) {
	apiKeys := []string{"bearer-token-123"}
	handler := APIKeyFromBearer(apiKeys, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("bearer success"))
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer bearer-token-123")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAPIKeyFromBearer_MissingHeader(t *testing.T) {
	apiKeys := []string{"bearer-token-123"}
	handler := APIKeyFromBearer(apiKeys, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAPIKeyFromBearer_InvalidFormat(t *testing.T) {
	apiKeys := []string{"bearer-token-123"}
	handler := APIKeyFromBearer(apiKeys, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat bearer-token-123")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAPIKeyAuth_OptionsRequest(t *testing.T) {
	apiKeys := []string{"test-key"}
	handler := APIKeyAuthFunc(apiKeys, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
	}

	// 检查 CORS headers
	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	if corsHeader != "*" {
		t.Errorf("Expected CORS header '*', got '%s'", corsHeader)
	}
}

func TestAPIKeyAuthWithValidator_Success(t *testing.T) {
	// 自定义验证器：接受所有以 "custom-" 开头的 key
	validator := func(apiKey string) (bool, error) {
		return len(apiKey) > 7 && apiKey[:7] == "custom-", nil
	}

	middleware := APIKeyAuthWithValidator(validator)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("custom validated"))
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-API-Key", "custom-123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAPIKeyAuthWithValidator_Failure(t *testing.T) {
	// 自定义验证器：拒绝不以 "valid-" 开头的 key
	validator := func(apiKey string) (bool, error) {
		return len(apiKey) > 6 && apiKey[:6] == "valid-", nil
	}

	middleware := APIKeyAuthWithValidator(validator)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAPIKeyAuthWithValidator_Error(t *testing.T) {
	// 自定义验证器：返回错误
	validator := func(apiKey string) (bool, error) {
		return false, fmt.Errorf("database connection failed")
	}

	middleware := APIKeyAuthWithValidator(validator)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-API-Key", "any-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAPIKeyAuthWithValidatorFunc(t *testing.T) {
	// 简单验证器：key 长度必须大于 5
	validator := func(apiKey string) (bool, error) {
		return len(apiKey) > 5, nil
	}

	handler := APIKeyAuthWithValidatorFunc(validator, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("validated"))
	})

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-API-Key", "longkey")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 测试短 key
	req2 := httptest.NewRequest("POST", "/test", nil)
	req2.Header.Set("X-API-Key", "short")
	w2 := httptest.NewRecorder()

	handler(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for short key, got %d", w2.Code)
	}
}
