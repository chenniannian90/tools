package mcpcourier

import (
	"context"

	"github.com/chenniannian90/tools/confmcp"
)

func Run(server *confmcp.Server) {
	_ = server
}

type Task struct {
	server *confmcp.Server
	tools  []*confmcp.Tool
}

func NewTask(opts ...Option) *Task {
	t := &Task{
		server: confmcp.NewServer(),
	}
	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *Task) Run(ctx context.Context) {
	// Register all tools
	for _, tool := range t.tools {
		if err := t.server.RegisterTool(tool); err != nil {
			panic(err)
		}
	}

	// Start server
	if err := t.server.Start(ctx); err != nil {
		panic(err)
	}
}
