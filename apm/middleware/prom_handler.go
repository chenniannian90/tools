package middleware

import (
	"github.com/chenniannian90/tools/apm/constants"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func PromHandler(enabled bool, path string) func(handler http.Handler) http.Handler {
	if path == "" {
		path = constants.DefaultMetricsPath
	}
	return func(handler http.Handler) http.Handler {
		return &promHandler{
			enabled:     enabled,
			path:        path,
			nextHandler: handler,
		}
	}
}

type promHandler struct {
	enabled     bool
	path        string
	nextHandler http.Handler
}

func (h *promHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.enabled && req.URL.Path == h.path {
		// 避免重复压缩(也可以先判断rw的 header 是否已经存在了压缩头
		promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer, promhttp.HandlerFor(prometheus.DefaultGatherer,
				promhttp.HandlerOpts{DisableCompression: true}),
		).ServeHTTP(rw, req)
		return
	}
	h.nextHandler.ServeHTTP(rw, req)
}
