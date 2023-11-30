package common

import (
	"github.com/chenniannian90/tools/apm/constants"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ApmRequestTotalMetric    *prometheus.CounterVec
	ApmRequestDurationMetric *prometheus.HistogramVec
)

func init() {
	ApmRequestTotalMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: constants.ApmRequestTotal,
		Help: "Number of request count.",
	}, constants.LabelList)
	ApmRequestDurationMetric = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: constants.ApmRequestDuration,
		Help: "Cost(millisecond) of request",
	}, constants.LabelList)
	prometheus.MustRegister(ApmRequestTotalMetric)
	prometheus.MustRegister(ApmRequestDurationMetric)

}
