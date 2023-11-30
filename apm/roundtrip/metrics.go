package roundtrip

import (
	"github.com/chenniannian90/tools/apm/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/go-courier/metax"

	"github.com/chenniannian90/tools/apm/common"
	"github.com/chenniannian90/tools/apm/constants"
	"github.com/chenniannian90/tools/apm/constants/enums"
	"github.com/prometheus/client_golang/prometheus"
)

type metricsRoundTripper struct {
	nextRoundTripper      http.RoundTripper
	project               string
	destination           string
	destinationType       enums.DestinationType
	requestTotalMetric    *prometheus.CounterVec
	requestDurationMetric prometheus.ObserverVec
	opFromResponse        bool
}

func NewMetricsRoundTripper(opts ...Option) func(roundTripper http.RoundTripper) http.RoundTripper {
	return func(roundTripper http.RoundTripper) http.RoundTripper {
		rt := &metricsRoundTripper{
			nextRoundTripper: roundTripper,
		}
		for _, opt := range opts {
			opt(rt)
		}
		labels := map[string]string{
			constants.Project:         rt.project,
			constants.Report:          string(enums.ReportTypeSource),
			constants.Destination:     rt.destination,
			constants.Source:          rt.project,
			constants.DestinationType: string(enums.DestinationTypeCommon),
			constants.Protocol:        string(enums.ProtocolHttp),
		}
		rt.requestTotalMetric = common.ApmRequestTotalMetric.MustCurryWith(labels)
		rt.requestDurationMetric = common.ApmRequestDurationMetric.MustCurryWith(labels)
		return rt
	}
}

func (rt *metricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startAt := time.Now()
	ctx := req.Context()
	// 往 header 中添加 source
	req.Header.Add(constants.MetaSource, rt.project)

	resp, err := rt.nextRoundTripper.RoundTrip(req.WithContext(ctx))
	statusCode := constants.ValidStatusCode
	if err == nil && resp != nil {
		statusCode = strconv.Itoa(resp.StatusCode)
	}
	meta := metax.MetaFromContext(ctx)
	operator := meta.Get("operationID")
	if len(operator) != 0 {
		operator = utils.NormOperator(operator)
	}
	if len(operator) == 0 {
		operator = req.Header.Get(constants.MetaOperator)
	}
	if len(operator) == 0 && rt.opFromResponse && resp != nil {
		operator = metax.ParseMeta(resp.Header.Get("X-Meta")).Get("operator")
		if operator == "" {
			// go-courier/httptransport v1.20.3 调整了
			operator = resp.Header.Get("X-Meta")
		}
		if operator != "" {
			operator = utils.NormOperator(operator)
		}
	}
	if len(operator) == 0 {
		operator = constants.Unknown
	}
	labels := map[string]string{
		constants.OperatorID: operator,
		constants.StatusCode: statusCode,
	}
	// 收集数量
	rt.requestTotalMetric.With(labels).Inc()
	// 收集时间
	duration := time.Since(startAt)
	rt.requestDurationMetric.With(labels).Observe(duration.Seconds())
	return resp, err
}
