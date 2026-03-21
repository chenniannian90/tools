package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/chenniannian90/tools/confmcp"
)

// 文件系统 MCP 服务器示例
// 展示如何实现 Resources（资源）功能，提供文件系统访问

func main() {
	// 获取当前工作目录作为基础目录
	baseDir, _ := os.Getwd()
	if baseDir == "" {
		baseDir = "."
	}

	config := &confmcp.MCP{
		Name:          "filesystem-server",
		Protocol:      "stdio",
	}

	server := confmcp.NewServer(config)

	// 注册文件系统工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "list_files",
		Description: "列出指定目录的文件和子目录",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "目录路径（相对于工作目录）",
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			path := "."
			if p, ok := args["path"].(string); ok && p != "" {
				path = p
			}

			fullPath := filepath.Join(baseDir, path)
			files, err := ioutil.ReadDir(fullPath)
			if err != nil {
				return nil, fmt.Errorf("无法读取目录 %s: %w", path, err)
			}

			result := fmt.Sprintf("目录 %s 的内容:\n", path)
			for _, file := range files {
				if file.IsDir() {
					result += fmt.Sprintf("📁 %s/\n", file.Name())
				} else {
					result += fmt.Sprintf("📄 %s (%d bytes)\n", file.Name(), file.Size())
				}
			}
			return result, nil
		},
	})

	// 注册读取文件工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "read_file",
		Description: "读取文件内容",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "文件路径（相对于工作目录）",
				},
			},
			"required": []string{"path"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			path, ok := args["path"].(string)
			if !ok || path == "" {
				return nil, fmt.Errorf("必须提供文件路径")
			}

			fullPath := filepath.Join(baseDir, path)
			content, err := ioutil.ReadFile(fullPath)
			if err != nil {
				return nil, fmt.Errorf("无法读取文件 %s: %w", path, err)
			}

			return fmt.Sprintf("文件: %s\n内容:\n%s", path, string(content)), nil
		},
	})

	// 注册写入文件工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "write_file",
		Description: "写入内容到文件",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "文件路径（相对于工作目录）",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "要写入的内容",
				},
			},
			"required": []string{"path", "content"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			path, ok := args["path"].(string)
			if !ok || path == "" {
				return nil, fmt.Errorf("必须提供文件路径")
			}

			content, ok := args["content"].(string)
			if !ok {
				return nil, fmt.Errorf("content 必须是字符串")
			}

			fullPath := filepath.Join(baseDir, path)

			// 确保目录存在
			dir := filepath.Dir(fullPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("无法创建目录: %w", err)
			}

			if err := ioutil.WriteFile(fullPath, []byte(content), 0644); err != nil {
				return nil, fmt.Errorf("无法写入文件 %s: %w", path, err)
			}

			return fmt.Sprintf("成功写入文件: %s (%d bytes)", path, len(content)), nil
		},
	})

	// 注册创建目录工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "create_directory",
		Description: "创建新目录",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "目录路径（相对于工作目录）",
				},
			},
			"required": []string{"path"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			path, ok := args["path"].(string)
			if !ok || path == "" {
				return nil, fmt.Errorf("必须提供目录路径")
			}

			fullPath := filepath.Join(baseDir, path)
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return nil, fmt.Errorf("无法创建目录 %s: %w", path, err)
			}

			return fmt.Sprintf("成功创建目录: %s", path), nil
		},
	})

	// 注册获取文件信息工具
	server.RegisterTool(&confmcp.Tool{
		Name:        "file_info",
		Description: "获取文件或目录的详细信息",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "文件或目录路径",
				},
			},
			"required": []string{"path"},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			path, ok := args["path"].(string)
			if !ok || path == "" {
				return nil, fmt.Errorf("必须提供路径")
			}

			fullPath := filepath.Join(baseDir, path)
			info, err := os.Stat(fullPath)
			if err != nil {
				return nil, fmt.Errorf("无法获取文件信息 %s: %w", path, err)
			}

			result := fmt.Sprintf("文件信息: %s\n", path)
			result += fmt.Sprintf("  大小: %d bytes\n", info.Size())
			result += fmt.Sprintf("  模式: %s\n", info.Mode())
			result += fmt.Sprintf("  修改时间: %s\n", info.ModTime())
			if info.IsDir() {
				result += "  类型: 目录\n"
			} else {
				result += "  类型: 文件\n"
			}

			return result, nil
		},
	})

	// 注册一些示例资源
	server.RegisterResource(&confmcp.Resource{
		URI:         "file:///README",
		Name:        "README 文件",
		Description: "项目的 README 文档",
		MimeType:    "text/plain",
		Handler: func(ctx context.Context, uri string) (*confmcp.ResourceContent, error) {
			readmePath := filepath.Join(baseDir, "README.md")
			content, err := ioutil.ReadFile(readmePath)
			if err != nil {
				return nil, fmt.Errorf("无法读取 README: %w", err)
			}
			return &confmcp.ResourceContent{
				URI:      uri,
				MimeType: "text/markdown",
				Text:     string(content),
			}, nil
		},
	})

	server.RegisterResource(&confmcp.Resource{
		URI:         "file:///config",
		Name:        "配置文件",
		Description: "MCP 服务器配置",
		MimeType:    "text/plain",
		Handler: func(ctx context.Context, uri string) (*confmcp.ResourceContent, error) {
			configContent := fmt.Sprintf("服务器名称: %s\n", config.Name)
			configContent += fmt.Sprintf("版本: %s\n", config.ServerVersion)
			configContent += fmt.Sprintf("工作目录: %s\n", baseDir)
			configContent += fmt.Sprintf("日志级别: %s\n", config.LogLevel)
			return &confmcp.ResourceContent{
				URI:      uri,
				MimeType: "text/plain",
				Text:     configContent,
			}, nil
		},
	})

	fmt.Fprintf(os.Stderr, "启动文件系统 MCP 服务器...\n")
	fmt.Fprintf(os.Stderr, "工作目录: %s\n", baseDir)
	fmt.Fprintf(os.Stderr, "可用工具: list_files, read_file, write_file, create_directory, file_info\n")
	fmt.Fprintf(os.Stderr, "可用资源: file:///README, file:///config\n")

	if err := server.Start(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "服务器错误: %v\n", err)
		os.Exit(1)
	}
}
