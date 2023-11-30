package utils

import (
	"github.com/chenniannian90/tools/apm/constants"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func ExposeMetricsHandler(path, address string) {
	if len(path) == 0 {
		path = constants.DefaultMetricsPath
	}
	if len(address) == 0 {
		address = constants.DefaultMetricsAddress
	}
	go func() {
		http.Handle(path, promhttp.Handler())
		if err := http.ListenAndServe(address, nil); err != nil {
			panic(err)
		}
	}()
}
