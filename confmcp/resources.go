package confmcp

import (
	"context"
	"fmt"
)

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType,omitempty"`
	Handler     ResourceHandler
}

// ResourceHandler handles resource reading
type ResourceHandler func(ctx context.Context, uri string) (*ResourceContent, error)

// ResourceContent represents resource content
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     []byte `json:"blob,omitempty"`
}

// ResourceRegistry manages MCP resources
type ResourceRegistry struct {
	resources map[string]*Resource
}

// NewResourceRegistry creates a new resource registry
func NewResourceRegistry() *ResourceRegistry {
	return &ResourceRegistry{
		resources: make(map[string]*Resource),
	}
}

// Register adds a resource to the registry
func (r *ResourceRegistry) Register(resource *Resource) error {
	if resource.URI == "" {
		return fmt.Errorf("resource URI cannot be empty")
	}
	if resource.Handler == nil {
		return fmt.Errorf("resource handler cannot be nil")
	}

	r.resources[resource.URI] = resource
	return nil
}

// Unregister removes a resource from the registry
func (r *ResourceRegistry) Unregister(uri string) {
	delete(r.resources, uri)
}

// Get retrieves a resource by URI
func (r *ResourceRegistry) Get(uri string) (*Resource, bool) {
	resource, ok := r.resources[uri]
	return resource, ok
}

// List returns all registered resources
func (r *ResourceRegistry) List() []*Resource {
	resources := make([]*Resource, 0, len(r.resources))
	for _, resource := range r.resources {
		resources = append(resources, resource)
	}
	return resources
}

// Read reads a resource by URI
func (r *ResourceRegistry) Read(ctx context.Context, uri string) (*ResourceContent, error) {
	resource, ok := r.Get(uri)
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", uri)
	}

	return resource.Handler(ctx, uri)
}

// Count returns the number of registered resources
func (r *ResourceRegistry) Count() int {
	return len(r.resources)
}

// Clear removes all resources from the registry
func (r *ResourceRegistry) Clear() {
	r.resources = make(map[string]*Resource)
}

// Has checks if a resource is registered
func (r *ResourceRegistry) Has(uri string) bool {
	_, ok := r.resources[uri]
	return ok
}
