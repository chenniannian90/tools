package confmcp

import (
	"context"
	"testing"
)

func TestNewResourceRegistry(t *testing.T) {
	registry := NewResourceRegistry()

	if registry == nil {
		t.Fatal("expected registry to be created")
	}

	if registry.Count() != 0 {
		t.Errorf("expected empty registry, got %d resources", registry.Count())
	}
}

func TestResourceRegistryRegister(t *testing.T) {
	registry := NewResourceRegistry()

	resource := &Resource{
		URI:         "test://example",
		Name:        "Test Resource",
		Description: "A test resource",
		MimeType:    "text/plain",
		Handler: func(ctx context.Context, uri string) (*ResourceContent, error) {
			return &ResourceContent{
				URI:  uri,
				Text: "test content",
			}, nil
		},
	}

	err := registry.Register(resource)
	if err != nil {
		t.Fatalf("failed to register resource: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("expected 1 resource, got %d", registry.Count())
	}

	if !registry.Has("test://example") {
		t.Error("expected resource to be registered")
	}
}

func TestResourceRegistryRegisterErrors(t *testing.T) {
	registry := NewResourceRegistry()

	// Test empty URI
	t.Run("empty URI", func(t *testing.T) {
		err := registry.Register(&Resource{
			URI: "",
			Handler: func(ctx context.Context, uri string) (*ResourceContent, error) {
				return nil, nil
			},
		})
		if err == nil {
			t.Error("expected error for empty URI")
		}
	})

	// Test nil handler
	t.Run("nil handler", func(t *testing.T) {
		err := registry.Register(&Resource{
			URI: "test://example",
		})
		if err == nil {
			t.Error("expected error for nil handler")
		}
	})
}

func TestResourceRegistryGet(t *testing.T) {
	registry := NewResourceRegistry()

	resource := &Resource{
		URI:         "test://example",
		Name:        "Test Resource",
		Description: "A test resource",
		Handler: func(ctx context.Context, uri string) (*ResourceContent, error) {
			return &ResourceContent{
				URI:  uri,
				Text: "test content",
			}, nil
		},
	}

	registry.Register(resource)

	// Test existing resource
	retrieved, ok := registry.Get("test://example")
	if !ok {
		t.Fatal("failed to retrieve registered resource")
	}
	if retrieved.Name != "Test Resource" {
		t.Errorf("expected resource name to be 'Test Resource', got %s", retrieved.Name)
	}

	// Test non-existing resource
	_, ok = registry.Get("non-existing")
	if ok {
		t.Error("expected false for non-existing resource")
	}
}

func TestResourceRegistryRead(t *testing.T) {
	registry := NewResourceRegistry()

	expectedContent := "test content"

	resource := &Resource{
		URI:         "test://example",
		Name:        "Test Resource",
		Description: "A test resource",
		Handler: func(ctx context.Context, uri string) (*ResourceContent, error) {
			return &ResourceContent{
				URI:  uri,
				Text: expectedContent,
			}, nil
		},
	}

	registry.Register(resource)

	// Test successful read
	content, err := registry.Read(context.Background(), "test://example")
	if err != nil {
		t.Fatalf("failed to read resource: %v", err)
	}
	if content.Text != expectedContent {
		t.Errorf("expected content to be '%s', got %s", expectedContent, content.Text)
	}

	// Test non-existing resource
	_, err = registry.Read(context.Background(), "non-existing")
	if err == nil {
		t.Error("expected error for non-existing resource")
	}
}

func TestResourceRegistryReadWithMimeType(t *testing.T) {
	registry := NewResourceRegistry()

	resource := &Resource{
		URI:         "test://json",
		Name:        "JSON Resource",
		Description: "A JSON test resource",
		MimeType:    "application/json",
		Handler: func(ctx context.Context, uri string) (*ResourceContent, error) {
			return &ResourceContent{
				URI:      uri,
				MimeType: "application/json",
				Text:     `{"key": "value"}`,
			}, nil
		},
	}

	registry.Register(resource)

	content, err := registry.Read(context.Background(), "test://json")
	if err != nil {
		t.Fatalf("failed to read resource: %v", err)
	}

	if content.MimeType != "application/json" {
		t.Errorf("expected mime type to be 'application/json', got %s", content.MimeType)
	}
}

func TestResourceRegistryList(t *testing.T) {
	registry := NewResourceRegistry()

	// Register multiple resources
	resources := []*Resource{
		{
			URI:     "test://res1",
			Name:    "Resource 1",
			Handler: func(ctx context.Context, uri string) (*ResourceContent, error) { return nil, nil },
		},
		{
			URI:     "test://res2",
			Name:    "Resource 2",
			Handler: func(ctx context.Context, uri string) (*ResourceContent, error) { return nil, nil },
		},
		{
			URI:     "test://res3",
			Name:    "Resource 3",
			Handler: func(ctx context.Context, uri string) (*ResourceContent, error) { return nil, nil },
		},
	}

	for _, resource := range resources {
		registry.Register(resource)
	}

	list := registry.List()
	if len(list) != 3 {
		t.Errorf("expected 3 resources, got %d", len(list))
	}
}

func TestResourceRegistryUnregister(t *testing.T) {
	registry := NewResourceRegistry()

	resource := &Resource{
		URI:     "test://example",
		Name:    "Test Resource",
		Handler: func(ctx context.Context, uri string) (*ResourceContent, error) { return nil, nil },
	}

	registry.Register(resource)

	if registry.Count() != 1 {
		t.Errorf("expected 1 resource before unregister, got %d", registry.Count())
	}

	registry.Unregister("test://example")

	if registry.Count() != 0 {
		t.Errorf("expected 0 resources after unregister, got %d", registry.Count())
	}

	if registry.Has("test://example") {
		t.Error("expected resource to be unregistered")
	}
}

func TestResourceRegistryClear(t *testing.T) {
	registry := NewResourceRegistry()

	// Register multiple resources
	for i := 0; i < 5; i++ {
		registry.Register(&Resource{
			URI:     string(rune('a' + i)),
			Name:    "Resource",
			Handler: func(ctx context.Context, uri string) (*ResourceContent, error) { return nil, nil },
		})
	}

	if registry.Count() != 5 {
		t.Errorf("expected 5 resources before clear, got %d", registry.Count())
	}

	registry.Clear()

	if registry.Count() != 0 {
		t.Errorf("expected 0 resources after clear, got %d", registry.Count())
	}
}

func TestResourceHandlerError(t *testing.T) {
	registry := NewResourceRegistry()

	expectedError := "failed to read resource"

	resource := &Resource{
		URI:  "test://error",
		Name: "Error Resource",
		Handler: func(ctx context.Context, uri string) (*ResourceContent, error) {
			return nil, &TestError{Message: expectedError}
		},
	}

	registry.Register(resource)

	_, err := registry.Read(context.Background(), "test://error")
	if err == nil {
		t.Fatal("expected error from resource read")
	}

	if err.Error() != expectedError {
		t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestResourceContentBlob(t *testing.T) {
	registry := NewResourceRegistry()

	blobContent := []byte{0x01, 0x02, 0x03, 0x04}

	resource := &Resource{
		URI:  "test://binary",
		Name: "Binary Resource",
		Handler: func(ctx context.Context, uri string) (*ResourceContent, error) {
			return &ResourceContent{
				URI:  uri,
				Blob: blobContent,
			}, nil
		},
	}

	registry.Register(resource)

	content, err := registry.Read(context.Background(), "test://binary")
	if err != nil {
		t.Fatalf("failed to read resource: %v", err)
	}

	if len(content.Blob) != len(blobContent) {
		t.Errorf("expected blob length %d, got %d", len(blobContent), len(content.Blob))
	}
}
