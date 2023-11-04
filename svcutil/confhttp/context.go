package confhttp

import (
	"context"
	"net/http"
)

func NewContextInjectorMiddleware(withContext func(ctx context.Context) context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if withContext != nil {
				req = req.WithContext(withContext(req.Context()))
			}
			next.ServeHTTP(rw, req)
		})
	}
}

type WithContext = func(ctx context.Context) context.Context

func WithContextCompose(withContexts ...WithContext) WithContext {
	return func(ctx context.Context) context.Context {
		for i := range withContexts {
			ctx = withContexts[i](ctx)
		}
		return ctx
	}
}
