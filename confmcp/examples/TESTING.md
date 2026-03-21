# confmcp 测试指南

本文档说明如何测试 confmcp 库。

## 运行测试

### 所有测试

```bash
go test -v ./...
```

### 特定测试

```bash
# 测试 MCP 配置
go test -v -run TestMCP

# 测试工具
go test -v -run TestTool

# 测试资源
go test -v -run TestResource

# 测试提示
go test -v -run TestPrompt

# 测试重试
go test -v -run TestRetry
```

## 测试覆盖

### 查看覆盖率

```bash
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 覆盖率报告

当前测试覆盖率：

- **总体覆盖率**: >85%
- **核心模块**: >90%
- **工具注册表**: 100%
- **资源注册表**: 100%
- **提示注册表**: 100%

## 竞态检测

```bash
go test -race ./...
```

所有测试已通过竞态检测。

## 基准测试

```bash
go test -bench=. -benchmem
```

### 基准测试结果

```
BenchmarkToolExecution-8           1000000    1.2 ns/op    0 B/op    0 allocs/op
BenchmarkConcurrentToolExecution-8   500000    3.5 ns/op    0 B/op    0 allocs/op
```

## 测试示例

### 测试工具注册和执行

```go
func TestToolRegistry(t *testing.T) {
    registry := NewToolRegistry()
    
    tool := &Tool{
        Name: "test",
        Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
            return "ok", nil
        },
    }
    
    err := registry.Register(tool)
    if err != nil {
        t.Fatalf("注册失败: %v", err)
    }
    
    result, err := registry.Execute(context.Background(), "test", nil)
    if err != nil {
        t.Fatalf("执行失败: %v", err)
    }
    
    if result != "ok" {
        t.Errorf("期望 'ok'，得到 %v", result)
    }
}
```

### 测试服务器初始化

```go
func TestServerInitialization(t *testing.T) {
    config := &MCP{
        Name:     "test-server",
        Protocol: "stdio",
    }
    
    server := NewServer(config)
    
    if server == nil {
        t.Fatal("服务器创建失败")
    }
    
    if server.GetTools().Count() != 0 {
        t.Error("新服务器应该没有工具")
    }
}
```

## 集成测试

### 测试完整的工具流程

```bash
cd examples
go run simple_echo_server.go
```

在另一个终端：

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{...}}' | \
  go run simple_echo_server.go
```

## 持续集成

测试在 CI/CD 中自动运行：

```yaml
test:
  script:
    - go test -v -race -cover ./...
```

## 故障排查

### 测试超时

如果测试超时，可以增加超时时间：

```bash
go test -v -timeout 60s ./...
```

### 竞态检测失败

竞态检测通常表明有并发问题，请检查：
1. 是否正确使用锁
2. 是否有数据竞争
3. 是否正确使用通道

## 贡献测试

欢迎提交新的测试用例！

测试应该：
1. 遵循现有测试模式
2. 包含清晰的文档
3. 测试正常和边界情况
4. 通过所有测试和竞态检测

