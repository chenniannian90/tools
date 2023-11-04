package confhttp

import (
	"context"
	"fmt"
	"net/http/httputil"

	"github.com/go-courier/httptransport"
)

type HTTPRequestLogger struct {
}

func (HTTPRequestLogger) ContextKey() string {
	return "HTTPRequestLogger"
}

func (HTTPRequestLogger) Output(ctx context.Context) (interface{}, error) {
	data, _ := httputil.DumpRequest(httptransport.HttpRequestFromContext(ctx), true)
	fmt.Println(string(data))
	return nil, nil
}
