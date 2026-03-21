package confmcp

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// TestServeMethod 测试 Serve 方法的基本功能
func TestServeMethod(t *testing.T) {
	server := NewServer()
	server.Name = "test-serve"
	server.Protocol = "http"
	server.Port = 3999 // 使用非标准端口避免冲突

	// 在后台启动服务器
	go func() {
		tools := []*Tool{
			{
				Name:        "test_ping",
				Description: "测试 ping",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
				Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return map[string]interface{}{
						"status": "ok",
					}, nil
				},
			},
		}
		_ = server.Serve(tools)
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 测试健康检查端点
	resp, err := http.Get("http://localhost:3999/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// TestServeWithMultipleTools 测试使用 Serve 注册多个工具
func TestServeWithMultipleTools(t *testing.T) {
	server := NewServer()
	server.Name = "test-multi-tools"
	server.Protocol = "http"
	server.Port = 3998

	tools := []*Tool{
		{
			Name:        "tool1",
			Description: "第一个工具",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return "tool1", nil
			},
		},
		{
			Name:        "tool2",
			Description: "第二个工具",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return "tool2", nil
			},
		},
		{
			Name:        "tool3",
			Description: "第三个工具",
			InputSchema: map[string]interface{}{
				"type": "object",
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return "tool3", nil
			},
		},
	}

	go func() {
		_ = server.Serve(tools)
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 验证工具数量
	if server.GetTools().Count() != 3 {
		t.Errorf("Expected 3 tools, got %d", server.GetTools().Count())
	}

	// 验证工具已正确注册
	tool1, ok := server.GetTools().Get("tool1")
	if !ok {
		t.Error("Tool1 not registered")
	}
	if tool1.Description != "第一个工具" {
		t.Errorf("Expected description '第一个工具', got '%s'", tool1.Description)
	}
}
