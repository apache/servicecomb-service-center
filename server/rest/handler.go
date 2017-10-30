package rest

import (
	roa "github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/server/interceptor"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

func init() {
	// api
	http.Handle("/", &ServerHandler{})

	// prometheus metrics
	http.Handle("/metrics", prometheus.Handler())
}

type ServerHandler struct {
}

func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := interceptor.InvokeInterceptors(w, r)
	if err != nil {
		return
	}

	roa.GetRouter().ServeHTTP(w, r)
}
