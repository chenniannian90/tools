package mcpcourier

import (
	"github.com/chenniannian90/tools/confmcp"
)

type Option func(*Task)

func WithServer(server *confmcp.Server) Option {
	return func(t *Task) {
		t.server = server
	}
}

func WithTools(tools []*confmcp.Tool) Option {
	return func(t *Task) {
		t.tools = tools
	}
}

func WithTool(tool *confmcp.Tool) Option {
	return func(t *Task) {
		t.tools = append(t.tools, tool)
	}
}
