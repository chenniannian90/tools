package confmcp

import (
	"context"
	"fmt"
)

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
	Handler     PromptHandler    `json:"-"`
}

// PromptArgument defines a prompt argument
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
}

// PromptHandler handles prompt generation
type PromptHandler func(ctx context.Context, args map[string]interface{}) (string, error)

// PromptRegistry manages MCP prompts
type PromptRegistry struct {
	prompts map[string]*Prompt
}

// NewPromptRegistry creates a new prompt registry
func NewPromptRegistry() *PromptRegistry {
	return &PromptRegistry{
		prompts: make(map[string]*Prompt),
	}
}

// Register adds a prompt to the registry
func (r *PromptRegistry) Register(prompt *Prompt) error {
	if prompt.Name == "" {
		return fmt.Errorf("prompt name cannot be empty")
	}
	if prompt.Handler == nil {
		return fmt.Errorf("prompt handler cannot be nil")
	}

	r.prompts[prompt.Name] = prompt
	return nil
}

// Unregister removes a prompt from the registry
func (r *PromptRegistry) Unregister(name string) {
	delete(r.prompts, name)
}

// Get retrieves a prompt by name
func (r *PromptRegistry) Get(name string) (*Prompt, bool) {
	prompt, ok := r.prompts[name]
	return prompt, ok
}

// List returns all registered prompts
func (r *PromptRegistry) List() []*Prompt {
	prompts := make([]*Prompt, 0, len(r.prompts))
	for _, prompt := range r.prompts {
		prompts = append(prompts, prompt)
	}
	return prompts
}

// Generate generates a prompt with given arguments
func (r *PromptRegistry) Generate(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	prompt, ok := r.Get(name)
	if !ok {
		return "", fmt.Errorf("prompt not found: %s", name)
	}

	return prompt.Handler(ctx, args)
}

// Count returns the number of registered prompts
func (r *PromptRegistry) Count() int {
	return len(r.prompts)
}

// Clear removes all prompts from the registry
func (r *PromptRegistry) Clear() {
	r.prompts = make(map[string]*Prompt)
}

// Has checks if a prompt is registered
func (r *PromptRegistry) Has(name string) bool {
	_, ok := r.prompts[name]
	return ok
}
