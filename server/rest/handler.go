package rest

import (
	roa "github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/server/interceptor"
	"net/http"
)

func init() {
	// api
	http.Handle("/", &ServerHandler{})
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
