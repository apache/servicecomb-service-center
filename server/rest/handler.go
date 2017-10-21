package rest

import (
	"github.com/ServiceComb/service-center/server/interceptor"
	"net/http"
	"github.com/ServiceComb/service-center/server/rest/controller/v4"
	"github.com/prometheus/client_golang/prometheus"
)

var router http.Handler

func init() {
	router = v4.GetRouter()

	// api
	http.Handle("/", DefaultServerHandler())

	// prometheus metrics
	http.Handle("/metrics", prometheus.Handler())
}

type ServerHandler struct {
}

func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := interceptor.InvokeInterceptors(interceptor.ACCESS_PHASE, w, r)
	if err != nil {
		return
	}
	err = interceptor.InvokeInterceptors(interceptor.FILTER_PHASE, w, r)
	if err != nil {
		return
	}
	err = interceptor.InvokeInterceptors(interceptor.CONTENT_PHASE, w, r)
	if err != nil {
		return
	}
	router.ServeHTTP(w, r)

	interceptor.InvokeInterceptors(interceptor.LOG_PHASE, w, r)
}

var defaultHandler ServerHandler

func DefaultServerHandler() *ServerHandler {
	return &defaultHandler
}
