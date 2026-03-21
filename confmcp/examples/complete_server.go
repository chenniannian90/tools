package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/chenniannian90/tools/confmcp"
)

// 完整的 MCP 服务器示例
// 展示 Tools、Resources、Prompts 三种功能的综合使用

func main() {
	config := &confmcp.MCP{
		Name:          "complete-mcp-server",
		Protocol:      "stdio",
	}

	server := confmcp.NewServer(config)

	// ==================== 注册 Tools ====================

	// 1. 随机数生成工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "random_number",
		Description: "生成指定范围的随机数",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"min": map[string]interface{}{
					"type":        "number",
					"description": "最小值",
				},
				"max": map[string]interface{}{
					"type":        "number",
					"description": "最大值",
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			min := getFloatArg(args, "min", 0)
			max := getFloatArg(args, "max", 100)
			num := min + rand.Float64()*(max-min)
			return fmt.Sprintf("随机数: %.2f (范围: %.2f - %.2f)", num, min, max), nil
		},
	})

	// 2. 时间戳工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "current_time",
		Description: "获取当前时间",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"format": map[string]interface{}{
					"type":        "string",
					"description": "时间格式 (ansi, iso, unix)",
					"enum":        []string{"ansi", "iso", "unix"},
					"default":     "iso",
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			format := getStringArg(args, "format", "iso")
			now := time.Now()

			switch format {
			case "ansi":
				return fmt.Sprintf("当前时间: %s", now.Format("2006-01-02 15:04:05")), nil
			case "iso":
				return fmt.Sprintf("当前时间: %s", now.Format(time.RFC3339)), nil
			case "unix":
				return fmt.Sprintf("Unix 时间戳: %d", now.Unix()), nil
			default:
				return fmt.Sprintf("当前时间: %s", now.Format(time.RFC3339)), nil
			}
		},
	})

	// 3. 倒计时工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "countdown",
		Description: "执行倒计时",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"seconds": map[string]interface{}{
					"type":        "number",
					"description": "倒计时秒数",
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			seconds := getIntArg(args, "seconds", 5)
			result := fmt.Sprintf("倒计时 %d 秒:\n", seconds)
			for i := seconds; i > 0; i-- {
				result += fmt.Sprintf("%d...\n", i)
			}
			result += "时间到！🎉"
			return result, nil
		},
	})

	// 4. 字符统计工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "count_chars",
		Description: "统计字符串的字符数、单词数和行数",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "要统计的文本",
				},
			},
			"required": []string{"text"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			text := getStringArg(args, "text", "")
			if text == "" {
				return nil, fmt.Errorf("text 不能为空")
			}

			charCount := len([]rune(text))
			wordCount := len(strings.Fields(text))
			lineCount := 1
			for _, c := range text {
				if c == '\n' {
					lineCount++
				}
			}

			result := fmt.Sprintf("文本统计:\n")
			result += fmt.Sprintf("  字符数: %d\n", charCount)
			result += fmt.Sprintf("  单词数: %d\n", wordCount)
			result += fmt.Sprintf("  行数: %d\n", lineCount)

			return result, nil
		},
	})

	// ==================== 注册 Resources ====================

	// 1. 服务器状态资源
	server.RegisterResource(&confmcp.Resource{
		URI:         "status://server",
		Name:        "服务器状态",
		Description: "MCP 服务器的当前状态信息",
		MimeType:    "text/plain",
		Handler: func(ctx context.Context, uri string) (*confmcp.ResourceContent, error) {
			status := fmt.Sprintf("MCP 服务器状态\n")
			status += fmt.Sprintf("================\n")
			status += fmt.Sprintf("名称: %s\n", config.Name)
			status += fmt.Sprintf("版本: %s\n", config.ServerVersion)
			status += fmt.Sprintf("协议: %s\n", config.Protocol)
			status += fmt.Sprintf("运行时间: %s\n", time.Since(time.Now()).String())
			status += fmt.Sprintf("当前时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

			return &confmcp.ResourceContent{
				URI:      uri,
				MimeType: "text/plain",
				Text:     status,
			}, nil
		},
	})

	// 2. 系统信息资源
	server.RegisterResource(&confmcp.Resource{
		URI:         "info://system",
		Name:        "系统信息",
		Description: "系统环境信息",
		MimeType:    "application/json",
		Handler: func(ctx context.Context, uri string) (*confmcp.ResourceContent, error) {
			info := map[string]interface{}{
				"hostname":     "localhost",
				"os":          "darwin",
				"arch":        "amd64",
				"go_version":  "1.20",
				"server_time": time.Now().Format(time.RFC3339),
				"timezone":    "Asia/Shanghai",
			}

			// 简单的 JSON 序列化
			jsonText := fmt.Sprintf("{\n")
			jsonText += fmt.Sprintf("  \"hostname\": \"%s\",\n", info["hostname"])
			jsonText += fmt.Sprintf("  \"os\": \"%s\",\n", info["os"])
			jsonText += fmt.Sprintf("  \"arch\": \"%s\",\n", info["arch"])
			jsonText += fmt.Sprintf("  \"go_version\": \"%s\",\n", info["go_version"])
			jsonText += fmt.Sprintf("  \"server_time\": \"%s\",\n", info["server_time"])
			jsonText += fmt.Sprintf("  \"timezone\": \"%s\"\n", info["timezone"])
			jsonText += fmt.Sprintf("}")

			return &confmcp.ResourceContent{
				URI:      uri,
				MimeType: "application/json",
				Text:     jsonText,
			}, nil
		},
	})

	// 3. 动态时间资源
	server.RegisterResource(&confmcp.Resource{
		URI:         "time://current",
		Name:        "当前时间",
		Description: "实时的当前时间",
		MimeType:    "text/plain",
		Handler: func(ctx context.Context, uri string) (*confmcp.ResourceContent, error) {
			now := time.Now()
			text := fmt.Sprintf("当前时间信息\n")
			text += fmt.Sprintf("============\n")
			text += fmt.Sprintf("日期: %s\n", now.Format("2006-01-02"))
			text += fmt.Sprintf("时间: %s\n", now.Format("15:04:05"))
			text += fmt.Sprintf("星期: %s\n", []string{"日", "一", "二", "三", "四", "五", "六"}[now.Weekday()])
			text += fmt.Sprintf("Unix 时间戳: %d\n", now.Unix())
			text += fmt.Sprintf("ISO 8601: %s\n", now.Format(time.RFC3339))

			return &confmcp.ResourceContent{
				URI:      uri,
				MimeType: "text/plain",
				Text:     text,
			}, nil
		},
	})

	// ==================== 注册 Prompts ====================

	// 1. 代码审查提示
	server.RegisterPrompt(&confmcp.Prompt{
		Name:        "code_review",
		Description: "生成代码审查提示",
		Arguments: []confmcp.PromptArgument{
			{
				Name:        "language",
				Description: "编程语言",
				Required:    false,
			},
			{
				Name:        "focus",
				Description: "审查重点 (performance, security, style)",
				Required:    false,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			language := getStringArg(args, "language", "通用代码")
			focus := getStringArg(args, "focus", "全面审查")

			prompt := fmt.Sprintf("请对以下 %s 进行代码审查，重点检查 %s：\n\n", language, focus)
			prompt += "## 审查要点\n\n"
			prompt += "1. **代码质量**：代码是否清晰、易读、符合最佳实践\n"
			prompt += "2. **潜在问题**：是否存在 bug、边界情况或错误处理\n"
			prompt += "3. **性能优化**：是否有性能瓶颈或可优化的地方\n"
			prompt += "4. **安全性**：是否存在安全漏洞或风险\n"
			prompt += "5. **可维护性**：代码是否易于维护和扩展\n\n"
			prompt += "请提供具体的改进建议和代码示例。"

			return prompt, nil
		},
	})

	// 2. 文档生成提示
	server.RegisterPrompt(&confmcp.Prompt{
		Name:        "generate_docs",
		Description: "生成代码文档提示",
		Arguments: []confmcp.PromptArgument{
			{
				Name:        "style",
				Description: "文档风格 (javadoc, godoc, jsdoc)",
				Required:    false,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			style := getStringArg(args, "style", "标准文档")

			prompt := fmt.Sprintf("请为以下代码生成 %s 风格的文档：\n\n", style)
			prompt += "## 文档要求\n\n"
			prompt += "1. **功能描述**：清晰描述函数/方法的功能\n"
			prompt += "2. **参数说明**：列出所有参数及其类型和用途\n"
			prompt += "3. **返回值**：说明返回值的类型和含义\n"
			prompt += "4. **使用示例**：提供简单的使用示例\n"
			prompt += "5. **注意事项**：说明重要的使用注意点\n\n"
			prompt += "请确保文档准确、完整、易于理解。"

			return prompt, nil
		},
	})

	// 3. 代码解释提示
	server.RegisterPrompt(&confmcp.Prompt{
		Name:        "explain_code",
		Description: "生成代码解释提示",
		Arguments: []confmcp.PromptArgument{
			{
				Name:        "detail_level",
				Description: "详细程度 (brief, detailed)",
				Required:    false,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			level := getStringArg(args, "detail_level", "detailed")

			prompt := "请解释以下代码的工作原理"
			if level == "brief" {
				prompt += "（简洁版）"
			} else {
				prompt += "（详细版）"
			}
			prompt += "：\n\n"

			if level == "detailed" {
				prompt += "## 解释要求\n\n"
				prompt += "1. **整体功能**：代码的主要功能是什么\n"
				prompt += "2. **逐行分析**：详细解释关键代码行\n"
				prompt += "3. **数据流程**：说明数据的流转过程\n"
				prompt += "4. **设计模式**：指出使用的设计模式（如果有）\n"
				prompt += "5. **优缺点**：分析实现的优缺点\n"
			} else {
				prompt += "## 解释要求\n\n"
				prompt += "1. **核心功能**：简要说明代码做什么\n"
				prompt += "2. **关键点**：指出2-3个关键实现点\n"
			}

			return prompt, nil
		},
	})

	// 4. 重构建议提示
	server.RegisterPrompt(&confmcp.Prompt{
		Name:        "refactor_suggestion",
		Description: "生成代码重构建议",
		Arguments: []confmcp.PromptArgument{
			{
				Name:        "goal",
				Description: "重构目标 (readability, performance, maintainability)",
				Required:    false,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
			goal := getStringArg(args, "goal", "全面提升")

			prompt := fmt.Sprintf("请对以下代码提供重构建议，目标：%s\n\n", goal)
			prompt += "## 建议要求\n\n"
			prompt += "1. **问题识别**：指出当前代码的问题\n"
			prompt += "2. **改进方案**：提供具体的重构方案\n"
			prompt += "3. **重构代码**：给出重构后的代码示例\n"
			prompt += "4. **对比说明**：说明重构前后的差异和改进\n"
			prompt += "5. **风险评估**：指出重构可能带来的风险\n\n"
			prompt += "请确保重构建议实用、可操作。"

			return prompt, nil
		},
	})

	// 打印服务器信息
	fmt.Fprintf(os.Stderr, "╔══════════════════════════════════════════╗\n")
	fmt.Fprintf(os.Stderr, "║   完整的 MCP 服务器                       ║\n")
	fmt.Fprintf(os.Stderr, "╚══════════════════════════════════════════╝\n")
	fmt.Fprintf(os.Stderr, "服务器名称: %s\n", config.Name)
	fmt.Fprintf(os.Stderr, "版本: %s\n", config.ServerVersion)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "📦 工具 (Tools):\n")
	fmt.Fprintf(os.Stderr, "   - random_number: 生成随机数\n")
	fmt.Fprintf(os.Stderr, "   - current_time: 获取当前时间\n")
	fmt.Fprintf(os.Stderr, "   - countdown: 倒计时\n")
	fmt.Fprintf(os.Stderr, "   - count_chars: 字符统计\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "📄 资源 (Resources):\n")
	fmt.Fprintf(os.Stderr, "   - status://server: 服务器状态\n")
	fmt.Fprintf(os.Stderr, "   - info://system: 系统信息\n")
	fmt.Fprintf(os.Stderr, "   - time://current: 当前时间\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "💬 提示 (Prompts):\n")
	fmt.Fprintf(os.Stderr, "   - code_review: 代码审查提示\n")
	fmt.Fprintf(os.Stderr, "   - generate_docs: 文档生成提示\n")
	fmt.Fprintf(os.Stderr, "   - explain_code: 代码解释提示\n")
	fmt.Fprintf(os.Stderr, "   - refactor_suggestion: 重构建议\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "服务器已启动，等待连接...\n")

	if err := server.Start(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "服务器错误: %v\n", err)
		os.Exit(1)
	}
}

// 辅助函数
func getFloatArg(args map[string]interface{}, key string, defaultValue float64) float64 {
	val, ok := args[key]
	if !ok {
		return defaultValue
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return defaultValue
	}
}

func getIntArg(args map[string]interface{}, key string, defaultValue int) int {
	val, ok := args[key]
	if !ok {
		return defaultValue
	}
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return defaultValue
	}
}

func getStringArg(args map[string]interface{}, key, defaultValue string) string {
	val, ok := args[key]
	if !ok {
		return defaultValue
	}
	if str, ok := val.(string); ok {
		return str
	}
	return defaultValue
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
