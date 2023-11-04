package conflogger

import (
	"context"
	"os"
	"time"

	"github.com/go-courier/httptransport/client"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"

	"github.com/go-courier/logr"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var zipkinCollectorEndpoint = "http://tempo-ingest.tempo-system.svc.cluster.local:9411"

func init() {
	if v := os.Getenv("ZIPKIN_COLLECTOR_ENDPOINT"); v != "" {
		zipkinCollectorEndpoint = v
	}
}

func InstallNewPipeline(projectName string, outputType OutputType, formatType FormatType) error {
	stdout := StdoutSpanExporter(formatType)

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(WithSpanMapExporter(OutputFilter(outputType))(stdout)),
	}

	z, err := zipkin.NewRawExporter(
		zipkinCollectorEndpoint+"/api/v2/spans",
		zipkin.WithClient(client.GetShortConnClient(1*time.Second)),
	)

	if err != nil {
		return err
	}

	opts = append(opts, sdktrace.WithBatcher(
		WithSpanMapExporter(OutputFilter(outputType), SpanOnlyFilter())(
			WithErrIgnoreExporter()(z),
		),
	))

	tp := sdktrace.NewTracerProvider(opts...)

	otel.SetTracerProvider(tp)

	return nil
}

// Deprecated: Use NewContextAndLogger instead.
func NewContextWithLogger(name string) (context.Context, logr.Logger) {
	ctx, span := otel.Tracer(name).Start(context.Background(), name, trace.WithTimestamp(time.Now()))
	log := SpanLogger(span)
	return logr.WithLogger(ctx, log), log
}

func NewContextAndLogger(ctx context.Context, name string) (context.Context, logr.Logger) {
	ctx, span := otel.Tracer(name).Start(ctx, name, trace.WithTimestamp(time.Now()))
	log := SpanLogger(span)
	return logr.WithLogger(ctx, log), log
}

func SpanOnlyFilter() SpanMapper {
	return func(data *sdktrace.SpanSnapshot) *sdktrace.SpanSnapshot {
		if data == nil {
			return nil
		}

		d := &sdktrace.SpanSnapshot{}
		d.SpanContext = data.SpanContext
		d.Parent = data.Parent
		d.SpanKind = data.SpanKind
		d.Name = data.Name
		d.StartTime = data.StartTime
		d.EndTime = data.EndTime
		d.Attributes = data.Attributes
		d.Links = data.Links
		d.StatusCode = data.StatusCode
		d.StatusMessage = data.StatusMessage
		d.DroppedAttributeCount = data.DroppedAttributeCount
		d.DroppedMessageEventCount = data.DroppedMessageEventCount
		d.DroppedLinkCount = data.DroppedLinkCount
		d.ChildSpanCount = data.ChildSpanCount
		d.Resource = data.Resource
		d.InstrumentationLibrary = data.InstrumentationLibrary
		return d
	}
}
