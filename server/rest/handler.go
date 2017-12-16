package rest

import (
	roa "github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/pkg/util"
	"github.com/ServiceComb/service-center/server/interceptor"
	"net/http"
	"time"
)

func init() {
	// api
	http.Handle("/", &ServerHandler{})
}

type ServerHandler struct {
}

func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	err := interceptor.InvokeInterceptors(w, r)
	if err != nil {
		return
	}

	roa.GetRouter().ServeHTTP(w, r)

	ReportRequestCompleted(w, r, start)

	util.LogNilOrWarnf(start, "%s %s", r.Method, r.RequestURI)
}
