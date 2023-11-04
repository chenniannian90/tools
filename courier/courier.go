package courier

import (
	"context"
	"github.com/chenniannian90/tools/svcutil/confhttp"

	"github.com/go-courier/courier"
)

func Run(router *courier.Router) {
	_ = router
}

type Task struct {
	server *confhttp.Server
	router *courier.Router
}

func NewTask(opts ...Option) *Task {
	t := &Task{}
	for _, opt := range opts {
		opt(t)
	}

	return t
}

func (t *Task) Run(_ context.Context) {
	_ = t.server.Serve(t.router)
}
