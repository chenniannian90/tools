package courier

import (
	"github.com/chenniannian90/tools/svcutil/confhttp"
	"github.com/go-courier/courier"
)

type Option func(*Task)

func WithServer(server *confhttp.Server) Option {
	return func(t *Task) {
		t.server = server
	}
}

func WithRouter(router *courier.Router) Option {
	return func(t *Task) {
		t.router = router
	}
}
