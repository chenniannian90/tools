package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/chenniannian90/tools/confmcp"
)

// 计算器 MCP 服务器示例
// 展示如何实现带有参数验证和错误处理的工具

func main() {
	server := confmcp.NewServer()
	server.Name = "calculator-server"
	server.Protocol = "stdio"

	// 加法工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "add",
		Description: "计算两个数的和",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{
					"type":        "number",
					"description": "第一个数",
				},
				"b": map[string]interface{}{
					"type":        "number",
					"description": "第二个数",
				},
			},
			"required": []string{"a", "b"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			a, b, err := getTwoNumbers(args)
			if err != nil {
				return nil, err
			}
			result := a + b
			return fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result), nil
		},
	})

	// 减法工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "subtract",
		Description: "计算两个数的差",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{
					"type":        "number",
					"description": "被减数",
				},
				"b": map[string]interface{}{
					"type":        "number",
					"description": "减数",
				},
			},
			"required": []string{"a", "b"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			a, b, err := getTwoNumbers(args)
			if err != nil {
				return nil, err
			}
			result := a - b
			return fmt.Sprintf("%.2f - %.2f = %.2f", a, b, result), nil
		},
	})

	// 乘法工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "multiply",
		Description: "计算两个数的积",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{
					"type":        "number",
					"description": "第一个数",
				},
				"b": map[string]interface{}{
					"type":        "number",
					"description": "第二个数",
				},
			},
			"required": []string{"a", "b"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			a, b, err := getTwoNumbers(args)
			if err != nil {
				return nil, err
			}
			result := a * b
			return fmt.Sprintf("%.2f × %.2f = %.2f", a, b, result), nil
		},
	})

	// 除法工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "divide",
		Description: "计算两个数的商",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{
					"type":        "number",
					"description": "被除数",
				},
				"b": map[string]interface{}{
					"type":        "number",
					"description": "除数",
				},
			},
			"required": []string{"a", "b"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			a, b, err := getTwoNumbers(args)
			if err != nil {
				return nil, err
			}
			if b == 0 {
				return nil, fmt.Errorf("除数不能为零")
			}
			result := a / b
			return fmt.Sprintf("%.2f ÷ %.2f = %.2f", a, b, result), nil
		},
	})

	// 平方工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "square",
		Description: "计算一个数的平方",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"x": map[string]interface{}{
					"type":        "number",
					"description": "要计算平方的数",
				},
			},
			"required": []string{"x"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			x, err := getNumber(args, "x")
			if err != nil {
				return nil, err
			}
			result := x * x
			return fmt.Sprintf("%.2f² = %.2f", x, result), nil
		},
	})

	// 开平方工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "sqrt",
		Description: "计算一个数的平方根",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"x": map[string]interface{}{
					"type":        "number",
					"description": "要计算平方根的数",
				},
			},
			"required": []string{"x"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			x, err := getNumber(args, "x")
			if err != nil {
				return nil, err
			}
			if x < 0 {
				return nil, fmt.Errorf("不能计算负数的平方根")
			}
			result := sqrt(x)
			return fmt.Sprintf("√%.2f = %.4f", x, result), nil
		},
	})

	fmt.Fprintf(os.Stderr, "启动计算器 MCP 服务器...\n")
	fmt.Fprintf(os.Stderr, "可用工具: add, subtract, multiply, divide, square, sqrt\n")

	if err := server.Start(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "服务器错误: %v\n", err)
		os.Exit(1)
	}
}

// 辅助函数：从参数中获取两个数字
func getTwoNumbers(args map[string]interface{}) (float64, float64, error) {
	a, err := getNumber(args, "a")
	if err != nil {
		return 0, 0, fmt.Errorf("参数 a 错误: %w", err)
	}

	b, err := getNumber(args, "b")
	if err != nil {
		return 0, 0, fmt.Errorf("参数 b 错误: %w", err)
	}

	return a, b, nil
}

// 辅助函数：从参数中获取数字
func getNumber(args map[string]interface{}, key string) (float64, error) {
	value, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("缺少参数: %s", key)
	}

	var num float64
	switch v := value.(type) {
	case float64:
		num = v
	case int:
		num = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("参数 %s 不是有效的数字: %s", key, v)
		}
		num = parsed
	default:
		return 0, fmt.Errorf("参数 %s 类型错误", key)
	}

	return num, nil
}

// 简单的平方根计算（牛顿迭代法）
func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 1000; i++ {
		nextZ := 0.5 * (z + x/z)
		if abs(z-nextZ) < 1e-10 {
			break
		}
		z = nextZ
	}
	return z
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
