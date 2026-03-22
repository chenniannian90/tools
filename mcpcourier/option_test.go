package mcpcourier

import (
	"net/http"
	"testing"

	"github.com/chenniannian90/tools/confmcp"
)

func TestWithAPIKeyValidator(t *testing.T) {
	validator := func(apiKey string) (bool, error) {
		return apiKey == "valid-key", nil
	}

	task := NewTask(
		WithAPIKeyValidator(validator),
	)

	if task.apiKeyConfig == nil {
		t.Fatal("apiKeyConfig should not be nil")
	}

	if task.apiKeyConfig.Validator == nil {
		t.Fatal("Validator should not be nil")
	}
}

func TestWithAPIKeyValidator_SimpleList(t *testing.T) {
	// 简单列表验证示例
	validKeys := map[string]bool{
		"key1": true,
		"key2": true,
	}

	validator := func(apiKey string) (bool, error) {
		return validKeys[apiKey], nil
	}

	task := NewTask(
		WithAPIKeyValidator(validator),
	)

	if task.apiKeyConfig == nil {
		t.Fatal("apiKeyConfig should not be nil")
	}

	// 验证可以正常调用
	valid, err := task.apiKeyConfig.Validator("key1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !valid {
		t.Error("Expected key1 to be valid")
	}

	invalid, err := task.apiKeyConfig.Validator("invalid-key")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if invalid {
		t.Error("Expected invalid-key to be invalid")
	}
}

func TestWithMiddleware(t *testing.T) {
	task := NewTask(
		WithMiddleware(func(next http.Handler) http.Handler {
			return next
		}),
	)

	if len(task.middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(task.middlewares))
	}
}

func TestWithMultipleMiddleware(t *testing.T) {
	task := NewTask(
		WithMiddleware(func(next http.Handler) http.Handler {
			return next
		}),
		WithMiddleware(func(next http.Handler) http.Handler {
			return next
		}),
	)

	if len(task.middlewares) != 2 {
		t.Errorf("Expected 2 middlewares, got %d", len(task.middlewares))
	}
}

func TestCombinedOptions(t *testing.T) {
	validator := func(apiKey string) (bool, error) {
		return apiKey == "test-key", nil
	}

	task := NewTask(
		WithTool(&confmcp.Tool{
			Name: "test",
		}),
		WithMiddleware(func(next http.Handler) http.Handler {
			return next
		}),
		WithAPIKeyValidator(validator),
	)

	if len(task.tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(task.tools))
	}

	if len(task.middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(task.middlewares))
	}

	if task.apiKeyConfig == nil {
		t.Fatal("apiKeyConfig should not be nil")
	}

	if task.apiKeyConfig.Validator == nil {
		t.Fatal("Validator should not be nil")
	}
}

func TestNewTaskDefaults(t *testing.T) {
	task := NewTask()

	if task.server == nil {
		t.Error("Expected server to be initialized")
	}

	if task.middlewares == nil {
		t.Error("Expected middlewares to be initialized")
	}

	if len(task.middlewares) != 0 {
		t.Errorf("Expected 0 middlewares by default, got %d", len(task.middlewares))
	}

	if task.apiKeyConfig != nil {
		t.Error("Expected apiKeyConfig to be nil by default")
	}
}
