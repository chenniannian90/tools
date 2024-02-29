package confhttp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chenniannian90/tools/svcutil/conflogger"
	"github.com/go-courier/httptransport/httpx"
	"github.com/go-courier/logr"
	"github.com/go-courier/metax"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func NewLogRoundTripper() func(roundTripper http.RoundTripper) http.RoundTripper {
	return func(roundTripper http.RoundTripper) http.RoundTripper {
		return &LogRoundTripper{
			nextRoundTripper: roundTripper,
		}
	}
}

type LogRoundTripper struct {
	nextRoundTripper http.RoundTripper
}

func (rt *LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startedAt := time.Now()

	ctx := req.Context()
	// inject b3 form context
	(&b3.B3{InjectEncoding: b3.B3SingleHeader}).Inject(ctx, propagation.HeaderCarrier(req.Header))

	ctx, log := logr.Start(ctx, "Request")
	defer log.End()

	resp, err := rt.nextRoundTripper.RoundTrip(req.WithContext(ctx))

	level, _ := logr.ParseLevel(strings.ToLower(req.Header.Get("x-log-level")))

	cost := time.Since(startedAt)
	logger := log.WithValues(
		"cost", fmt.Sprintf("%0.3fms", float64(cost/time.Millisecond)),
		"method", req.Method,
		"url", omitAuthorization(req.URL),
	)

	if err == nil {
		if level >= logr.InfoLevel {
			logger.Info("success")
		}
	} else {
		if level >= logr.WarnLevel {
			logger.Warn(errors.Wrap(err, "http request failed"))
		}
	}

	return resp, err
}

func TraceLogHandler(tracer trace.Tracer) func(handler http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			ctx = (&b3.B3{}).Extract(ctx, propagation.HeaderCarrier(req.Header))

			startAt := time.Now()

			ctx, span := tracer.Start(ctx, "UnknownOperation", trace.WithTimestamp(startAt))
			defer func() {
				span.End(trace.WithTimestamp(time.Now()))
			}()

			log := conflogger.SpanLogger(span)

			loggerRw := newLoggerResponseWriter(rw)

			// for async pick
			(&b3.B3{InjectEncoding: b3.B3SingleHeader}).Inject(ctx, propagation.HeaderCarrier(loggerRw.Header()))

			meta := metax.ParseMeta(loggerRw.Header().Get("X-Meta"))
			meta["_id"] = []string{span.SpanContext().TraceID().String()}
			meta["_host"] = []string{req.Host}
			meta["_client"] = []string{req.RemoteAddr}
			ctx = metax.ContextWithMeta(ctx, meta)
			ctx = logr.WithLogger(ctx, log)
			nextHandler.ServeHTTP(loggerRw, req.WithContext(ctx))

			operator := metax.ParseMeta(loggerRw.Header().Get("X-Meta")).Get("operator")
			if operator == "" {
				// https://github.com/go-courier/httptransport/commit/95ccb86d2f27a1b811f895fc2f0fce9a64dff1dd
				// 兼容 go-courier/httptransport v1.20.3
				operator = loggerRw.Header().Get("X-Meta")
			}
			if operator != "" {
				span.SetName(operator)
			}

			level, _ := logr.ParseLevel(strings.ToLower(req.Header.Get("x-log-level")))

			duration := time.Since(startAt)

			header := req.Header

			keyAndValues := []interface{}{
				"tag", "access",
				"remote_ip", httpx.ClientIP(req),
				"cost", fmt.Sprintf("%0.3fms", float64(duration/time.Millisecond)),
				"method", req.Method,
				"request_uri", omitAuthorization(req.URL),
				"user_agent", header.Get(httpx.HeaderUserAgent),
				"status", loggerRw.statusCode,
			}

			if loggerRw.err != nil {
				if loggerRw.statusCode >= http.StatusInternalServerError {
					if level >= logr.ErrorLevel {
						log.WithValues(keyAndValues...).Error(loggerRw.err)
					}
				} else {
					if level >= logr.WarnLevel {
						log.WithValues(keyAndValues...).Warn(loggerRw.err)
					}
				}
			} else {
				if level >= logr.InfoLevel {
					log.WithValues(keyAndValues...).Info("")
				}
			}
		})
	}
}

func newLoggerResponseWriter(rw http.ResponseWriter) *loggerResponseWriter {
	h, hok := rw.(http.Hijacker)
	if !hok {
		h = nil
	}

	f, fok := rw.(http.Flusher)
	if !fok {
		f = nil
	}

	return &loggerResponseWriter{
		ResponseWriter: rw,
		Hijacker:       h,
		Flusher:        f,
	}
}

type loggerResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
	http.Flusher

	headerWritten bool
	statusCode    int
	err           error
}

func (rw *loggerResponseWriter) WriteError(err error) {
	rw.err = err
}

func (rw *loggerResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *loggerResponseWriter) WriteHeader(statusCode int) {
	rw.writeHeader(statusCode)
}

func (rw *loggerResponseWriter) Write(data []byte) (int, error) {
	if rw.err == nil && rw.statusCode >= http.StatusBadRequest {
		rw.err = errors.New(string(data))
	}
	return rw.ResponseWriter.Write(data)
}

func (rw *loggerResponseWriter) writeHeader(statusCode int) {
	if !rw.headerWritten {
		rw.ResponseWriter.WriteHeader(statusCode)
		rw.statusCode = statusCode
		rw.headerWritten = true
	}
}

func omitAuthorization(u *url.URL) string {
	query := u.Query()
	query.Del("authorization")
	u.RawQuery = query.Encode()
	return u.String()
}
