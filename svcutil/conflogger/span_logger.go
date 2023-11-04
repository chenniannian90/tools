package conflogger

import (
	"context"
	"fmt"
	"time"

	"github.com/go-courier/logr"
	"github.com/go-courier/metax"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var l = logr.DebugLevel

func SetLevel(lvl logr.Level) {
	l = lvl
}

func SpanLogger(span trace.Span) logr.Logger {
	return &spanLogger{span: span}
}

type spanLogger struct {
	span       trace.Span
	attributes []attribute.KeyValue
}

func (t *spanLogger) Trace(msg string, args ...interface{}) {
	//TODO implement me
	panic("implement me")
}

func (t *spanLogger) Fatal(err error) {
	//TODO implement me
	panic("implement me")
}

func (t *spanLogger) Panic(err error) {
	//TODO implement me
	panic("implement me")
}

func (t *spanLogger) Start(ctx context.Context, name string, keyAndValues ...interface{}) (context.Context, logr.Logger) {
	span := trace.SpanFromContext(ctx)
	meta := metax.MetaFromContext(ctx)

	if len(meta) > 0 {
		keyAndValues = append(keyAndValues, "meta", meta)
	}

	ctx, span = span.Tracer().Start(
		ctx, name,
		trace.WithAttributes(attrsFromKeyAndValues(keyAndValues...)...),
		trace.WithTimestamp(time.Now()),
	)
	return ctx, &spanLogger{span: span}
}

func (t *spanLogger) End() {
	t.span.End(trace.WithTimestamp(time.Now()))
}

func (t *spanLogger) WithValues(keyAndValues ...interface{}) logr.Logger {
	return &spanLogger{span: t.span, attributes: append(t.attributes, attrsFromKeyAndValues(keyAndValues...)...)}
}

func (t *spanLogger) info(level logr.Level, msg fmt.Stringer) {
	if level > l {
		return
	}

	t.span.AddEvent(
		"@"+level.String(),
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(t.attributes...),
		trace.WithAttributes(
			attribute.Stringer("message", msg),
		),
	)
}

func (t *spanLogger) error(level logr.Level, err error) {
	if level > l {
		return
	}

	if t.span == nil || err == nil || !t.span.IsRecording() {
		return
	}

	attributes := append(t.attributes, attribute.String("message", err.Error()))

	if level <= logr.ErrorLevel {
		attributes = append(attributes, attribute.String("stack", fmt.Sprintf("%+v", err)))
	}

	t.span.SetStatus(codes.Error, "")
	t.span.AddEvent(
		"@"+level.String(),
		trace.WithTimestamp(time.Now()),
		trace.WithAttributes(attributes...),
	)
}

func (t *spanLogger) Debug(msgOrFormat string, args ...interface{}) {
	t.info(logr.DebugLevel, Sprintf(msgOrFormat, args...))
}

func (t *spanLogger) Info(msgOrFormat string, args ...interface{}) {
	t.info(logr.InfoLevel, Sprintf(msgOrFormat, args...))
}

func (t *spanLogger) Warn(err error) {
	t.error(logr.WarnLevel, err)
}

func (t *spanLogger) Error(err error) {
	t.error(logr.ErrorLevel, err)
}

func attrsFromKeyAndValues(keysAndValues ...interface{}) []attribute.KeyValue {
	n := len(keysAndValues)
	if n > 0 && n%2 == 0 {
		fields := make([]attribute.KeyValue, len(keysAndValues)/2)
		for i := range fields {
			k, v := keysAndValues[2*i], keysAndValues[2*i+1]

			if s, ok := k.(string); ok {
				fields[i] = attribute.Any(s, v)
			}
		}
		return fields
	}
	return nil
}

func Sprintf(format string, args ...interface{}) fmt.Stringer {
	return &printer{format: format, args: args}
}

type printer struct {
	format string
	args   []interface{}
}

func (p *printer) String() string {
	if len(p.args) == 0 {
		return p.format
	}
	return fmt.Sprintf(p.format, p.args...)
}
