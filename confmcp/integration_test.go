package confmcp

import (
	"context"
	"testing"
	"time"
)

// TestServerInitialization 测试服务器初始化流程
func TestServerInitialization(t *testing.T) {
	server := NewServer()
	server.Name = "test-server"
	server.Protocol = "stdio"

	if server == nil {
		t.Fatal("服务器创建失败")
	}

	// 注册测试工具
	err := server.RegisterTool(&Tool{
		Name:        "test",
		Description: "测试工具",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	})

	if err != nil {
		t.Fatalf("工具注册失败: %v", err)
	}

	// 检查工具是否注册成功
	if !server.GetTools().Has("test") {
		t.Error("工具未成功注册")
	}
}

// TestServerWithMultipleTools 测试服务器处理多个工具
func TestServerWithMultipleTools(t *testing.T) {
	server := NewServer()
	server.Name = "multi-tool-server"
	server.Protocol = "stdio"

	// 注册多个工具
	tools := []*Tool{
		{
			Name: "tool1",
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return "result1", nil
			},
		},
		{
			Name: "tool2",
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return "result2", nil
			},
		},
		{
			Name: "tool3",
			Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
				return "result3", nil
			},
		},
	}

	for _, tool := range tools {
		if err := server.RegisterTool(tool); err != nil {
			t.Fatalf("注册工具 %s 失败: %v", tool.Name, err)
		}
	}

	// 验证所有工具都已注册
	if server.GetTools().Count() != 3 {
		t.Errorf("期望 3 个工具，得到 %d", server.GetTools().Count())
	}

	// 测试所有工具的执行
	ctx := context.Background()
	for i, toolName := range []string{"tool1", "tool2", "tool3"} {
		result, err := server.GetTools().Execute(ctx, toolName, nil)
		if err != nil {
			t.Errorf("执行工具 %s 失败: %v", toolName, err)
		}
		expected := []string{"result1", "result2", "result3"}[i]
		if result != expected {
			t.Errorf("工具 %s 期望结果 %s，得到 %v", toolName, expected, result)
		}
	}
}

// TestConcurrentToolExecution 测试并发工具执行
func TestConcurrentToolExecution(t *testing.T) {
	server := NewServer()
	server.Name = "concurrent-server"
	server.Protocol = "stdio"

	// 注册一个耗时工具
	server.RegisterTool(&Tool{
		Name: "slow_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			return "done", nil
		},
	})

	ctx := context.Background()

	// 并发执行多个调用
	const concurrency = 10
	results := make(chan string, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			result, err := server.GetTools().Execute(ctx, "slow_tool", nil)
			if err != nil {
				errors <- err
			} else {
				results <- result.(string)
			}
		}()
	}

	// 收集结果
	successCount := 0
	errorCount := 0
	for i := 0; i < concurrency; i++ {
		select {
		case <-results:
			successCount++
		case <-errors:
			errorCount++
		case <-time.After(5 * time.Second):
			t.Fatal("超时等待结果")
		}
	}

	if successCount != concurrency {
		t.Errorf("期望 %d 个成功，得到 %d 个成功，%d 个错误", concurrency, successCount, errorCount)
	}
}

// TestServerLifecycle 测试服务器生命周期
func TestServerLifecycle(t *testing.T) {
	server := NewServer()
	server.Name = "lifecycle-server"
	server.Protocol = "stdio"

	// 测试初始状态
	if server.GetTools().Count() != 0 {
		t.Error("新服务器应该没有工具")
	}

	// 添加工具
	server.RegisterTool(&Tool{
		Name: "tool1",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	})

	if server.GetTools().Count() != 1 {
		t.Error("工具注册后应该有1个工具")
	}

	// 清空工具
	server.GetTools().Clear()

	if server.GetTools().Count() != 0 {
		t.Error("清空后应该没有工具")
	}
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	server := NewServer()
	server.Name = "error-server"
	server.Protocol = "stdio"

	// 注册一个会返回错误的工具
	server.RegisterTool(&Tool{
		Name: "error_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return nil, &TestError{Message: "测试错误"}
		},
	})

	ctx := context.Background()

	// 测试错误处理
	_, err := server.GetTools().Execute(ctx, "error_tool", nil)
	if err == nil {
		t.Error("期望返回错误")
	}

	if err.Error() != "测试错误" {
		t.Errorf("错误消息不匹配: %v", err)
	}

	// 测试调用不存在的工具
	_, err = server.GetTools().Execute(ctx, "non_existent", nil)
	if err == nil {
		t.Error("调用不存在的工具应该返回错误")
	}
}

// TestServerInfo 测试服务器信息
func TestServerInfo(t *testing.T) {
	config := &MCP{
		Name:     "test-server",
		Protocol: "stdio",
	}
	config.Init()

	info := config.GetServerInfo()

	if info.Name != "test-server" {
		t.Errorf("期望名称 test-server，得到 %s", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("期望版本 1.0.0，得到 %s", info.Version)
	}
	if info.Protocol != "stdio" {
		t.Errorf("期望协议 stdio，得到 %s", info.Protocol)
	}

	// 验证 capabilities 只有 tools
	if !info.Capabilities.Tools {
		t.Error("期望 Tools capability 为 true")
	}
}

// BenchmarkToolExecution 基准测试工具执行
func BenchmarkToolExecution(b *testing.B) {
	server := NewServer()
	server.Name = "benchmark-server"
	server.Protocol = "stdio"

	server.RegisterTool(&Tool{
		Name: "bench_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "result", nil
		},
	})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.GetTools().Execute(ctx, "bench_tool", nil)
	}
}

// BenchmarkConcurrentToolExecution 基准测试并发工具执行
func BenchmarkConcurrentToolExecution(b *testing.B) {
	server := NewServer()
	server.Name = "concurrent-bench-server"
	server.Protocol = "stdio"

	server.RegisterTool(&Tool{
		Name: "fast_tool",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return "result", nil
		},
	})

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			server.GetTools().Execute(ctx, "fast_tool", nil)
		}
	})
}

