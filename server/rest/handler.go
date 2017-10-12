package rest

import (
	"github.com/ServiceComb/service-center/server/interceptor"
	"github.com/ServiceComb/service-center/server/rest/controller/v3"
	"net/http"
)

var router http.Handler

func init() {
	router = v3.GetRouter()

	http.Handle("/", DefaultServerHandler())
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
