package confmcp

import (
	"testing"
	"time"
)

func TestMCPSetDefaults(t *testing.T) {
	m := &MCP{}
	m.SetDefaults()

	if m.Protocol != "stdio" {
		t.Errorf("期望协议为 stdio，得到 %s", m.Protocol)
	}
}

func TestMCPInit(t *testing.T) {
	m := &MCP{Name: "test-server"}
	m.Init()

	if !m.Initialized {
		t.Error("期望 MCP 已初始化")
	}

	// 多次初始化不会出错
	m.Init()
	if !m.Initialized {
		t.Error("期望 MCP 保持初始化状态")
	}
}

func TestMCPOperator(t *testing.T) {
	m := &MCP{Name: "test-server"}
	m.Init()

	// 测试 GetServerInfo
	info := m.GetServerInfo()
	if info.Name != "test-server" {
		t.Errorf("期望服务器名称为 test-server，得到 %s", info.Name)
	}

	// 测试 LivenessCheck
	status := m.LivenessCheck()
	if len(status) == 0 {
		t.Error("期望健康检查返回状态")
	}

	addr := m.GetAddress()
	if status[addr] != "ok" {
		t.Errorf("期望状态为 ok，得到 %s", status[addr])
	}
}

func TestMCPGetAddress(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		port     int
		expected string
	}{
		{
			name:     "stdio 协议",
			protocol: "stdio",
			expected: "stdio",
		},
		{
			name:     "sse 协议有端口",
			protocol: "sse",
			port:     3000,
			expected: ":3000",
		},
		{
			name:     "sse 协议默认端口",
			protocol: "sse",
			expected: ":3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MCP{
				Protocol: tt.protocol,
				Port:     tt.port,
			}
			m.SetDefaults()
			addr := m.GetAddress()
			if addr != tt.expected {
				t.Errorf("期望地址 %s，得到 %s", tt.expected, addr)
			}
		})
	}
}

func TestMCPSetRetryConfig(t *testing.T) {
	m := &MCP{}
	m.SetRetryConfig(5, 20*time.Second)

	if m.retry.Repeats != 5 {
		t.Errorf("期望 5 次重试，得到 %d", m.retry.Repeats)
	}

	expectedInterval := 20 * time.Second
	if time.Duration(m.retry.Interval) != expectedInterval {
		t.Errorf("期望间隔 %v，得到 %v", expectedInterval, m.retry.Interval)
	}
}

func TestCapabilities(t *testing.T) {
	cap := Capabilities{
		Tools:     true,
		Resources: true,
		Prompts:   true,
		Roots:     false,
		Sampling:  false,
	}

	if !cap.Tools {
		t.Error("期望工具已启用")
	}

	if !cap.Resources {
		t.Error("期望资源已启用")
	}

	if !cap.Prompts {
		t.Error("期望提示已启用")
	}

	if cap.Roots {
		t.Error("期望 roots 未启用")
	}

	if cap.Sampling {
		t.Error("期望 sampling 未启用")
	}
}

func TestMCPGetServerInfo(t *testing.T) {
	m := &MCP{
		Name:     "test-server",
		Protocol: "stdio",
	}
	m.Init()

	info := m.GetServerInfo()

	if info.Name != "test-server" {
		t.Errorf("期望名称为 test-server，得到 %s", info.Name)
	}

	if info.Version != "1.0.0" {
		t.Errorf("期望版本为 1.0.0，得到 %s", info.Version)
	}

	if info.Protocol != "stdio" {
		t.Errorf("期望协议为 stdio，得到 %s", info.Protocol)
	}

	if info.Address != "stdio" {
		t.Errorf("期望地址为 stdio，得到 %s", info.Address)
	}

	// 检查默认能力
	if !info.Capabilities.Tools {
		t.Error("期望默认启用工具能力")
	}
	if !info.Capabilities.Resources {
		t.Error("期望默认启用资源能力")
	}
	if !info.Capabilities.Prompts {
		t.Error("期望默认启用提示能力")
	}
}

func TestMCPPortConfiguration(t *testing.T) {
	t.Run("SSE 协议自动设置端口", func(t *testing.T) {
		m := &MCP{
			Protocol: "sse",
		}
		m.SetDefaults()

		if m.Port != 3000 {
			t.Errorf("期望 SSE 协议默认端口为 3000，得到 %d", m.Port)
		}
	})

	t.Run("stdio 协议不设置端口", func(t *testing.T) {
		m := &MCP{
			Protocol: "stdio",
		}
		m.SetDefaults()

		if m.Port != 0 {
			t.Errorf("期望 stdio 协议端口为 0，得到 %d", m.Port)
		}
	})

	t.Run("显式设置端口不被覆盖", func(t *testing.T) {
		m := &MCP{
			Protocol: "sse",
			Port:     8080,
		}
		m.SetDefaults()

		if m.Port != 8080 {
			t.Errorf("期望端口保持 8080，得到 %d", m.Port)
		}
	})
}
