package confmcp

import (
	"context"
	"testing"
)

func TestNewToolRegistry(t *testing.T) {
	registry := NewToolRegistry()

	if registry == nil {
		t.Fatal("expected registry to be created")
	}

	if registry.Count() != 0 {
		t.Errorf("expected empty registry, got %d tools", registry.Count())
	}
}

func TestToolRegistryRegister(t *testing.T) {
	registry := NewToolRegistry()

	tool := &Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	}

	err := registry.Register(tool)
	if err != nil {
		t.Fatalf("failed to register tool: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("expected 1 tool, got %d", registry.Count())
	}

	if !registry.Has("test-tool") {
		t.Error("expected tool to be registered")
	}
}

func TestToolRegistryRegisterErrors(t *testing.T) {
	registry := NewToolRegistry()

	// Test empty name
	t.Run("empty name", func(t *testing.T) {
		err := registry.Register(&Tool{
			Name: "",
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return nil, nil
			},
		})
		if err == nil {
			t.Error("expected error for empty name")
		}
	})

	// Test nil handler
	t.Run("nil handler", func(t *testing.T) {
		err := registry.Register(&Tool{
			Name: "test",
		})
		if err == nil {
			t.Error("expected error for nil handler")
		}
	})
}

func TestToolRegistryGet(t *testing.T) {
	registry := NewToolRegistry()

	tool := &Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "test result", nil
		},
	}

	registry.Register(tool)

	// Test existing tool
	retrieved, ok := registry.Get("test-tool")
	if !ok {
		t.Fatal("failed to retrieve registered tool")
	}
	if retrieved.Name != "test-tool" {
		t.Errorf("expected tool name to be test-tool, got %s", retrieved.Name)
	}

	// Test non-existing tool
	_, ok = registry.Get("non-existing")
	if ok {
		t.Error("expected false for non-existing tool")
	}
}

func TestToolRegistryExecute(t *testing.T) {
	registry := NewToolRegistry()

	expectedResult := "test result"

	tool := &Tool{
		Name:        "test-tool",
		Description: "A test tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return expectedResult, nil
		},
	}

	registry.Register(tool)

	// Test successful execution
	result, err := registry.Execute(context.Background(), "test-tool", nil)
	if err != nil {
		t.Fatalf("failed to execute tool: %v", err)
	}
	if result != expectedResult {
		t.Errorf("expected result to be '%s', got %v", expectedResult, result)
	}

	// Test non-existing tool
	_, err = registry.Execute(context.Background(), "non-existing", nil)
	if err == nil {
		t.Error("expected error for non-existing tool")
	}
}

func TestToolRegistryExecuteWithArgs(t *testing.T) {
	registry := NewToolRegistry()

	tool := &Tool{
		Name: "echo-tool",
		Description: "Echoes back the message",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type": "string",
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return args["message"], nil
		},
	}

	registry.Register(tool)

	testMessage := "Hello, MCP!"
	result, err := registry.Execute(context.Background(), "echo-tool", map[string]interface{}{
		"message": testMessage,
	})

	if err != nil {
		t.Fatalf("failed to execute tool: %v", err)
	}

	if result != testMessage {
		t.Errorf("expected result to be '%s', got %v", testMessage, result)
	}
}

func TestToolRegistryList(t *testing.T) {
	registry := NewToolRegistry()

	// Register multiple tools
	tools := []*Tool{
		{
			Name: "tool1",
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		{
			Name: "tool2",
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return nil, nil
			},
		},
		{
			Name: "tool3",
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return nil, nil
			},
		},
	}

	for _, tool := range tools {
		registry.Register(tool)
	}

	list := registry.List()
	if len(list) != 3 {
		t.Errorf("expected 3 tools, got %d", len(list))
	}
}

func TestToolRegistryUnregister(t *testing.T) {
	registry := NewToolRegistry()

	tool := &Tool{
		Name: "test-tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}

	registry.Register(tool)

	if registry.Count() != 1 {
		t.Errorf("expected 1 tool before unregister, got %d", registry.Count())
	}

	registry.Unregister("test-tool")

	if registry.Count() != 0 {
		t.Errorf("expected 0 tools after unregister, got %d", registry.Count())
	}

	if registry.Has("test-tool") {
		t.Error("expected tool to be unregistered")
	}
}

func TestToolRegistryClear(t *testing.T) {
	registry := NewToolRegistry()

	// Register multiple tools
	for i := 0; i < 5; i++ {
		registry.Register(&Tool{
			Name:     string(rune('a' + i)),
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) { return nil, nil },
		})
	}

	if registry.Count() != 5 {
		t.Errorf("expected 5 tools before clear, got %d", registry.Count())
	}

	registry.Clear()

	if registry.Count() != 0 {
		t.Errorf("expected 0 tools after clear, got %d", registry.Count())
	}
}

func TestToolHandlerError(t *testing.T) {
	registry := NewToolRegistry()

	expectedError := "something went wrong"

	tool := &Tool{
		Name: "error-tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return nil, &TestError{Message: expectedError}
		},
	}

	registry.Register(tool)

	_, err := registry.Execute(context.Background(), "error-tool", nil)
	if err == nil {
		t.Fatal("expected error from tool execution")
	}

	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

// TestError is a test error type
type TestError struct {
	Message string
}

func (e *TestError) Error() string {
	return e.Message
}
