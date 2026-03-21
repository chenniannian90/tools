package confmcp

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-courier/envconf"
)

// MCP 配置结构，参考 confhttp 的设计
type MCP struct {
	// 端口配置，使用 opt,expose 标签
	Port int `env:",opt,expose"`

	// 协议类型：stdio 或 sse
	Protocol string `env:""`

	// 服务器名称
	Name string `env:""`

	// 内部字段
	Initialized bool   `env:"-"`
	retry       *Retry `env:"-"`
}

// Capabilities 定义 MCP 服务器能力
type Capabilities struct {
	Tools     bool `json:"tools"`
	Resources bool `json:"resources"`
	Prompts   bool `json:"prompts"`
	Roots     bool `json:"roots"`
	Sampling  bool `json:"sampling"`
}

// SetDefaults 设置默认配置值
func (m *MCP) SetDefaults() {
	if m.Protocol == "" {
		m.Protocol = "stdio"
	}

	// 如果是 SSE 或 HTTP 协议，设置默认端口
	if (m.Protocol == "sse" || m.Protocol == "http") && m.Port == 0 {
		m.Port = 3000
	}

	// 初始化重试配置
	if m.retry == nil {
		m.retry = &Retry{}
		m.retry.SetDefaults()
	}
}

// Init 初始化 MCP
func (m *MCP) Init() {
	if !m.Initialized {
		m.SetDefaults()
		m.Initialized = true
	}
}

// GetAddress 获取服务地址（类似 confhttp 的 LivenessCheck）
func (m *MCP) GetAddress() string {
	if m.Port > 0 {
		return ":" + strconv.Itoa(m.Port)
	}
	return "stdio"
}

// GetServerInfo 获取服务器信息
func (m *MCP) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:     m.Name,
		Version:  "1.0.0",
		Protocol: m.Protocol,
		Address:  m.GetAddress(),
		Capabilities: Capabilities{
			Tools:     true,
			Resources: true,
			Prompts:   true,
		},
	}
}

// LivenessCheck 健康检查（参考 confhttp）
func (m *MCP) LivenessCheck() map[string]string {
	status := map[string]string{}

	if m.Initialized {
		status[m.GetAddress()] = "ok"
	} else {
		status[m.GetAddress()] = "not initialized"
	}

	return status
}

// SetRetryConfig 设置重试配置
func (m *MCP) SetRetryConfig(repeats int, interval time.Duration) {
	if m.retry == nil {
		m.retry = &Retry{}
	}
	m.retry.Repeats = repeats
	m.retry.Interval = envconf.Duration(interval)
}

// connect 建立连接（带重试）
func (m *MCP) connect() error {
	if m.retry == nil {
		m.retry = &Retry{}
		m.retry.SetDefaults()
	}

	return m.retry.Do(func() error {
		switch m.Protocol {
		case "stdio":
			return nil // stdio 不需要连接
		case "sse":
			return fmt.Errorf("SSE 协议暂未实现")
		default:
			return fmt.Errorf("不支持的协议: %s", m.Protocol)
		}
	})
}

// ServerInfo 服务器信息
type ServerInfo struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Protocol     string       `json:"protocol"`
	Address      string       `json:"address"`
	Capabilities Capabilities `json:"capabilities"`
}
