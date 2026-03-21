package confmcp

import (
	"context"
	"testing"
)

func TestNewPromptRegistry(t *testing.T) {
	registry := NewPromptRegistry()

	if registry == nil {
		t.Fatal("expected registry to be created")
	}

	if registry.Count() != 0 {
		t.Errorf("expected empty registry, got %d prompts", registry.Count())
	}
}

func TestPromptRegistryRegister(t *testing.T) {
	registry := NewPromptRegistry()

	prompt := &Prompt{
		Name:        "test-prompt",
		Description: "A test prompt",
		Arguments: []PromptArgument{
			{
				Name:        "topic",
				Description: "The topic to write about",
				Required:    true,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "test prompt", nil
		},
	}

	err := registry.Register(prompt)
	if err != nil {
		t.Fatalf("failed to register prompt: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("expected 1 prompt, got %d", registry.Count())
	}

	if !registry.Has("test-prompt") {
		t.Error("expected prompt to be registered")
	}
}

func TestPromptRegistryRegisterErrors(t *testing.T) {
	registry := NewPromptRegistry()

	// Test empty name
	t.Run("empty name", func(t *testing.T) {
		err := registry.Register(&Prompt{
			Name: "",
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				return "", nil
			},
		})
		if err == nil {
			t.Error("expected error for empty name")
		}
	})

	// Test nil handler
	t.Run("nil handler", func(t *testing.T) {
		err := registry.Register(&Prompt{
			Name: "test",
		})
		if err == nil {
			t.Error("expected error for nil handler")
		}
	})
}

func TestPromptRegistryGet(t *testing.T) {
	registry := NewPromptRegistry()

	prompt := &Prompt{
		Name:        "test-prompt",
		Description: "A test prompt",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "test prompt", nil
		},
	}

	registry.Register(prompt)

	// Test existing prompt
	retrieved, ok := registry.Get("test-prompt")
	if !ok {
		t.Fatal("failed to retrieve registered prompt")
	}
	if retrieved.Name != "test-prompt" {
		t.Errorf("expected prompt name to be 'test-prompt', got %s", retrieved.Name)
	}

	// Test non-existing prompt
	_, ok = registry.Get("non-existing")
	if ok {
		t.Error("expected false for non-existing prompt")
	}
}

func TestPromptRegistryGenerate(t *testing.T) {
	registry := NewPromptRegistry()

	expectedPrompt := "Write a story about"

	prompt := &Prompt{
		Name:        "story-prompt",
		Description: "Generate a story prompt",
		Arguments: []PromptArgument{
			{
				Name:        "topic",
				Description: "The story topic",
				Required:    true,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			topic := args["topic"].(string)
			return expectedPrompt + " " + topic, nil
		},
	}

	registry.Register(prompt)

	// Test successful generation
	result, err := registry.Generate(context.Background(), "story-prompt", map[string]interface{}{
		"topic": "space exploration",
	})
	if err != nil {
		t.Fatalf("failed to generate prompt: %v", err)
	}

	expectedResult := expectedPrompt + " space exploration"
	if result != expectedResult {
		t.Errorf("expected prompt to be '%s', got %s", expectedResult, result)
	}
}

func TestPromptRegistryGenerateMissingRequiredArg(t *testing.T) {
	registry := NewPromptRegistry()

	prompt := &Prompt{
		Name:        "greeting-prompt",
		Description: "Generate a greeting",
		Arguments: []PromptArgument{
			{
				Name:        "name",
				Description: "Name to greet",
				Required:    true,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			name, ok := args["name"]
			if !ok {
				return "", &TestError{Message: "name is required"}
			}
			return "Hello, " + name.(string) + "!", nil
		},
	}

	registry.Register(prompt)

	// Test missing required argument
	_, err := registry.Generate(context.Background(), "greeting-prompt", map[string]interface{}{})
	if err == nil {
		t.Error("expected error for missing required argument")
	}
}

func TestPromptRegistryList(t *testing.T) {
	registry := NewPromptRegistry()

	// Register multiple prompts
	prompts := []*Prompt{
		{
			Name: "prompt1",
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				return "", nil
			},
		},
		{
			Name: "prompt2",
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				return "", nil
			},
		},
		{
			Name: "prompt3",
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				return "", nil
			},
		},
	}

	for _, prompt := range prompts {
		registry.Register(prompt)
	}

	list := registry.List()
	if len(list) != 3 {
		t.Errorf("expected 3 prompts, got %d", len(list))
	}
}

func TestPromptRegistryUnregister(t *testing.T) {
	registry := NewPromptRegistry()

	prompt := &Prompt{
		Name: "test-prompt",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "", nil
		},
	}

	registry.Register(prompt)

	if registry.Count() != 1 {
		t.Errorf("expected 1 prompt before unregister, got %d", registry.Count())
	}

	registry.Unregister("test-prompt")

	if registry.Count() != 0 {
		t.Errorf("expected 0 prompts after unregister, got %d", registry.Count())
	}

	if registry.Has("test-prompt") {
		t.Error("expected prompt to be unregistered")
	}
}

func TestPromptRegistryClear(t *testing.T) {
	registry := NewPromptRegistry()

	// Register multiple prompts
	for i := 0; i < 5; i++ {
		registry.Register(&Prompt{
			Name:     string(rune('a' + i)),
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) { return "", nil },
		})
	}

	if registry.Count() != 5 {
		t.Errorf("expected 5 prompts before clear, got %d", registry.Count())
	}

	registry.Clear()

	if registry.Count() != 0 {
		t.Errorf("expected 0 prompts after clear, got %d", registry.Count())
	}
}

func TestPromptHandlerError(t *testing.T) {
	registry := NewPromptRegistry()

	expectedError := "failed to generate prompt"

	prompt := &Prompt{
		Name: "error-prompt",
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			return "", &TestError{Message: expectedError}
		},
	}

	registry.Register(prompt)

	_, err := registry.Generate(context.Background(), "error-prompt", nil)
	if err == nil {
		t.Fatal("expected error from prompt generation")
	}

	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestPromptArgument(t *testing.T) {
	arg := PromptArgument{
		Name:        "test-arg",
		Description: "A test argument",
		Required:    true,
	}

	if arg.Name != "test-arg" {
		t.Errorf("expected name to be 'test-arg', got %s", arg.Name)
	}

	if arg.Description != "A test argument" {
		t.Errorf("expected description to be 'A test argument', got %s", arg.Description)
	}

	if !arg.Required {
		t.Error("expected argument to be required")
	}
}

func TestPromptWithOptionalArguments(t *testing.T) {
	registry := NewPromptRegistry()

	prompt := &Prompt{
		Name:        "flexible-prompt",
		Description: "A prompt with optional arguments",
		Arguments: []PromptArgument{
			{
				Name:        "required",
				Description: "Required argument",
				Required:    true,
			},
			{
				Name:        "optional",
				Description: "Optional argument",
				Required:    false,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			result := "Required: " + args["required"].(string)
			if optional, ok := args["optional"]; ok {
				result += ", Optional: " + optional.(string)
			}
			return result, nil
		},
	}

	registry.Register(prompt)

	// Test with only required argument
	result, err := registry.Generate(context.Background(), "flexible-prompt", map[string]interface{}{
		"required": "value1",
	})
	if err != nil {
		t.Fatalf("failed to generate prompt: %v", err)
	}

	if result != "Required: value1" {
		t.Errorf("expected 'Required: value1', got %s", result)
	}

	// Test with both arguments
	result, err = registry.Generate(context.Background(), "flexible-prompt", map[string]interface{}{
		"required": "value1",
		"optional": "value2",
	})
	if err != nil {
		t.Fatalf("failed to generate prompt: %v", err)
	}

	if result != "Required: value1, Optional: value2" {
		t.Errorf("expected 'Required: value1, Optional: value2', got %s", result)
	}
}
