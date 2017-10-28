package rest

import (
	roa "github.com/ServiceComb/service-center/pkg/rest"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

func init() {
	// api
	http.Handle("/", roa.GetRouter())

	// prometheus metrics
	http.Handle("/metrics", prometheus.Handler())
}
