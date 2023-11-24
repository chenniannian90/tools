package confhttp

import (
	"compress/gzip"
	"context"
	"net/http"
	"strconv"

	"github.com/chenniannian90/tools/svcutil/conflogger"

	"github.com/chenniannian90/tools/svcutil/confhttp/middlewares"
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport"
	_ "github.com/go-courier/httptransport/validator/strfmt"
	"github.com/go-courier/x/ptr"
	"go.opentelemetry.io/otel"
)

var customMiddlewares []httptransport.HttpMiddleware

func WithCustomMiddleware(hms ...httptransport.HttpMiddleware) {
	customMiddlewares = append(customMiddlewares, hms...)
}

type Server struct {
	Port            int                                       `env:",opt,expose"`
	OpenAPISpec     string                                    `env:",opt,copy"`
	Healthy         string                                    `env:",opt,healthCheck"`
	Debug           *bool                                     `env:""`
	ht              *httptransport.HttpTransport              `env:"-"`
	contextInjector func(ctx context.Context) context.Context `env:"-"`
}

func (s Server) WithContextInjector(contextInjector func(ctx context.Context) context.Context) *Server {
	s.contextInjector = contextInjector
	return &s
}

func (s *Server) LivenessCheck() map[string]string {
	statuses := map[string]string{}

	if s.ht != nil {
		statuses[s.ht.ServiceMeta.String()] = "ok"
	}

	return statuses
}

func (s *Server) SetDefaults() {
	if s.Port == 0 {
		s.Port = 80
	}

	if s.OpenAPISpec == "" {
		s.OpenAPISpec = "./openapi.json"
	}

	if s.Debug == nil {
		s.Debug = ptr.Bool(true)
	}

	if s.Healthy == "" {
		s.Healthy = "http://:" + strconv.FormatInt(int64(s.Port), 10) + "/"
	}
}

func (s *Server) Serve(router *courier.Router) error {
	ht := httptransport.NewHttpTransport()

	ht.Port = s.Port

	ht.SetDefaults()

	tracer := otel.Tracer("Server")

	ht.Middlewares = []httptransport.HttpMiddleware{
		defaultCompress,
	}
	ht.Middlewares = append(ht.Middlewares, []httptransport.HttpMiddleware{
		middlewares.RecoverHandler(),
		middlewares.DefaultCORS(),
		middlewares.HealthCheckHandler(),
		middlewares.PProfHandler(*s.Debug),
		TraceLogHandler(tracer),
		NewContextInjectorMiddleware(s.contextInjector),
	}...)
	ht.Middlewares = append(ht.Middlewares, customMiddlewares...)

	s.ht = ht

	ctx, _ := conflogger.NewContextAndLogger(context.Background(), "Serve")

	return ht.ServeContext(ctx, router)
}

func defaultCompress(h http.Handler) http.Handler {
	return middlewares.CompressHandlerLevel(h, gzip.DefaultCompression)
}
