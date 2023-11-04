package conflogger

import (
	"context"
	"fmt"
	"os"

	"github.com/go-courier/metax"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type OutputType string

var (
	OutputAlways    OutputType = "Always"
	OutputOnFailure OutputType = "OnFailure"
	OutputNever     OutputType = "Never"
)

type FormatType string

var (
	FormatTEXT FormatType = "text"
	FormatJSON FormatType = "json"
)

func OutputFilter(outputType OutputType) SpanMapper {
	return func(data *sdktrace.SpanSnapshot) *sdktrace.SpanSnapshot {
		if outputType == OutputNever {
			return nil
		}
		if outputType == OutputOnFailure {
			if data.StatusCode == codes.Ok {
				return nil
			}
		}
		return data
	}
}

type SpanMapper = func(data *sdktrace.SpanSnapshot) *sdktrace.SpanSnapshot

func WithSpanMapExporter(mappers ...SpanMapper) func(spanExporter sdktrace.SpanExporter) sdktrace.SpanExporter {
	return func(spanExporter sdktrace.SpanExporter) sdktrace.SpanExporter {
		return &spanMapExporter{
			mappers:      mappers,
			SpanExporter: spanExporter,
		}
	}
}

type spanMapExporter struct {
	mappers []SpanMapper
	sdktrace.SpanExporter
}

func (e *spanMapExporter) ExportSpans(ctx context.Context, spanData []*sdktrace.SpanSnapshot) error {
	finalSpanSnapshot := make([]*sdktrace.SpanSnapshot, 0)

	mappers := e.mappers

	for i := range spanData {
		data := spanData[i]

		for _, m := range mappers {
			data = m(data)
		}

		if data != nil {
			finalSpanSnapshot = append(finalSpanSnapshot, data)
		}
	}

	if len(finalSpanSnapshot) == 0 {
		return nil
	}

	return e.SpanExporter.ExportSpans(ctx, finalSpanSnapshot)
}

func StdoutSpanExporter(formatType FormatType) sdktrace.SpanExporter {
	if formatType == FormatJSON {
		return &stdoutSpanExporter{formatter: &logrus.JSONFormatter{}}
	}
	return &stdoutSpanExporter{formatter: &logrus.TextFormatter{}}
}

type stdoutSpanExporter struct {
	formatter logrus.Formatter
}

func (e *stdoutSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}

// ExportSpan writes a SpanSnapshot in json format to stdout.
func (e *stdoutSpanExporter) ExportSpans(ctx context.Context, spanData []*sdktrace.SpanSnapshot) error {
	for i := range spanData {
		data := spanData[i]

		for _, event := range data.MessageEvents {
			if event.Name == "" || event.Name[0] != '@' {
				continue
			}

			lv, err := logrus.ParseLevel(event.Name[1:])
			if err != nil {
				continue
			}

			entry := logrus.NewEntry(logrus.StandardLogger())
			entry.Level = lv
			entry.Time = event.Time
			entry.Data = logrus.Fields{}

			for _, kv := range event.Attributes {
				k := string(kv.Key)

				switch k {
				case "message":
					entry.Message = kv.Value.AsString()
				default:
					entry.Data[k] = kv.Value.AsInterface()
				}
			}

			for _, kv := range data.Attributes {
				k := string(kv.Key)
				if k == "meta" {
					meta := metax.ParseMeta(kv.Value.AsString())
					for k := range meta {
						if k == "_id" {
							continue
						}
						entry.Data[k] = meta[k]
					}
					continue
				}
				entry.Data[k] = kv.Value.AsInterface()
			}

			entry.Data["span"] = data.Name
			entry.Data["traceID"] = data.SpanContext.TraceID()

			if data.SpanContext.HasSpanID() {
				entry.Data["spanID"] = data.SpanContext.SpanID()
			}

			if data.Parent.IsValid() {
				entry.Data["parentSpanID"] = data.Parent.SpanID()
			}

			log(e.formatter, entry)
		}
	}

	return nil
}

func log(formatter logrus.Formatter, e *logrus.Entry) {
	data, err := formatter.Format(e)
	if err != nil {
		fmt.Println(err)
	}
	_, _ = os.Stdout.Write(data)
}
