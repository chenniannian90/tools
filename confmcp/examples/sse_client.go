package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/chenniannian90/tools/confmcp"
)

// SSE MCP 客户端示例
// 展示如何连接到 SSE MCP 服务器并使用工具

func main() {
	// 创建客户端配置
	config := &confmcp.MCP{
		Name:     "sse-mcp-client",
		Protocol: "sse",
		Port:     3000,
	}

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           🌐 MCP SSE 客户端                                      ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 1. 测试 SSE 连接
	fmt.Println("📡 测试 SSE 连接...")
	if err := testSSEConnection(); err != nil {
		fmt.Printf("❌ SSE 连接失败: %v\n", err)
	} else {
		fmt.Println("✅ SSE 连接成功")
	}
	fmt.Println()

	// 2. 测试健康检查
	fmt.Println("💚 测试健康检查...")
	if err := testHealthCheck(); err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
	}
	fmt.Println()

	// 3. 测试 MCP 接口
	fmt.Println("📊 测试 MCP 接口...")
	testMCPInterface()
	fmt.Println()

	// 4. 监听 SSE 事件（持续运行）
	fmt.Println("🎧 监听 SSE 事件（按 Ctrl+C 退出）...")
	fmt.Println()

	if err := listenSSEEvents(); err != nil {
		fmt.Printf("❌ 监听失败: %v\n", err)
		os.Exit(1)
	}
}

// testSSEConnection 测试 SSE 连接
func testSSEConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3000/sse", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// 读取第一个事件
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	if n > 0 {
		fmt.Printf("   收到事件: %s", string(buf[:n]))
	}

	return nil
}

// testHealthCheck 测试健康检查
func testHealthCheck() error {
	resp, err := http.Get("http://localhost:3000/health")
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	fmt.Printf("   状态: %v\n", result["status"])
	fmt.Printf("   服务器: %v\n", result["server"])
	fmt.Printf("   检查: %v\n", result["checks"])

	return nil
}

// testMCPInterface 测试 MCP 接口
func testMCPInterface() error {
	// 调用 get_time 工具
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "get_time",
			"arguments": map[string]interface{}{},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:3000/mcp", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return err
	}

	if result, ok := response["result"].(map[string]interface{}); ok {
		if message, ok := result["message"].(string); ok {
			fmt.Printf("   工具调用结果: %s\n", message)
		}
	}

	return nil
}

// listenSSEEvents 监听 SSE 事件
func listenSSEEvents() error {
	resp, err := http.Get("http://localhost:3000/sse")
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	fmt.Println("   开始接收 SSE 事件...")
	fmt.Println()

	decoder := json.NewDecoder(resp.Body)
	eventCount := 0
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("   📊 已接收 %d 个事件\n", eventCount)
		default:
			// 读取 SSE 事件
			var event struct {
				Event string          `json:"event"`
				Data  json.RawMessage `json:"data"`
			}

			line, err := readLine(resp.Body)
			if err != nil {
				if err == io.EOF {
					fmt.Println("\n   ⛔ 服务器关闭连接")
					return nil
				}
				continue
			}

			if line == "" {
				continue // 空行
			}

			// 解析事件行
			if len(line) > 6 && line[:6] == "event:" {
				event.Event = line[7:]
			} else if len(line) > 5 && line[:5] == "data:" {
				event.Data = json.RawMessage(line[6:])
			} else if line == "" {
				// 空行表示事件结束
				if event.Event != "" {
					displaySSEEvent(event.Event, event.Data)
					eventCount++
				}
				event.Event = ""
				event.Data = nil
			}
		}
	}
}

// readLine 读取一行
func readLine(r io.Reader) (string, error) {
	var buf [1024]byte
	var line []byte

	for {
		n, err := r.Read(buf[:1])
		if err != nil {
			if len(line) == 0 {
				return "", err
			}
			return string(line), nil
		}

		if buf[0] == '\n' {
			return string(line), nil
		}

		if buf[0] != '\r' {
			line = append(line, buf[0])
		}
	}
}

// displaySSEEvent 显示 SSE 事件
func displaySSEEvent(event string, data json.RawMessage) {
	var dataMap map[string]interface{}
	json.Unmarshal(data, &dataMap)

	switch event {
	case "connected":
		fmt.Printf("   ✅ %s\n", dataMap["message"])
	case "heartbeat":
		fmt.Printf("   💓 心跳: %s\n", dataMap["time"])
	default:
		fmt.Printf("   📨 事件: %s - %v\n", event, string(data))
	}
}
