package middleware

import (
	"github.com/chenniannian90/tools/apm/common"
	"github.com/chenniannian90/tools/apm/constants"
	"github.com/chenniannian90/tools/apm/constants/enums"
	"github.com/chenniannian90/tools/apm/utils"
	"github.com/go-courier/metax"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

func newResponseWriter(rw http.ResponseWriter) *responseWriter {
	h, hok := rw.(http.Hijacker)
	if !hok {
		h = nil
	}

	f, fok := rw.(http.Flusher)
	if !fok {
		f = nil
	}
	return &responseWriter{
		ResponseWriter: rw,
		Hijacker:       h,
		Flusher:        f,
	}
}

type responseWriter struct {
	http.ResponseWriter
	http.Hijacker
	http.Flusher

	headerWritten bool
	statusCode    int
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.headerWritten {
		rw.ResponseWriter.WriteHeader(statusCode)
		rw.statusCode = statusCode
		rw.headerWritten = true
	}
}

type metricsHandler struct {
	nextHandler           http.Handler
	project               string
	requestTotalMetric    *prometheus.CounterVec
	requestDurationMetric prometheus.ObserverVec
}

func MetricsHandler(opts ...Option) func(handler http.Handler) http.Handler {
	m := &metricsHandler{}
	for _, opt := range opts {
		opt(m)
	}
	labels := map[string]string{
		constants.Project:         m.project,
		constants.Report:          string(enums.ReportTypeDestination),
		constants.Destination:     m.project,
		constants.DestinationType: string(enums.DestinationTypeCommon),
		constants.Protocol:        string(enums.ProtocolHttp),
	}
	m.requestTotalMetric = common.ApmRequestTotalMetric.MustCurryWith(labels)
	m.requestDurationMetric = common.ApmRequestDurationMetric.MustCurryWith(labels)

	return func(handler http.Handler) http.Handler {
		m.nextHandler = handler
		return m
	}
}

func (h *metricsHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	metricRw := newResponseWriter(rw)
	startAt := time.Now()
	h.nextHandler.ServeHTTP(metricRw, req)

	operator := metax.ParseMeta(metricRw.Header().Get("X-Meta")).Get("operator")
	if operator == "" {
		// go-courier/httptransport v1.20.3 调整了
		operator = metricRw.Header().Get("X-Meta")
		if operator == "" {
			return // 404 找不到接口
		}
	}
	source := req.Header.Get(constants.MetaSource)
	if len(source) == 0 {
		source = constants.Unknown
	}
	operator = utils.NormOperator(operator)
	statusCode := metricRw.statusCode
	statusCodeStr := strconv.Itoa(statusCode)
	labels := map[string]string{
		constants.OperatorID: operator,
		constants.StatusCode: statusCodeStr,
		constants.Source:     source,
	}
	duration := time.Since(startAt)
	h.requestTotalMetric.With(labels).Inc()
	h.requestDurationMetric.With(labels).Observe(duration.Seconds())
}
