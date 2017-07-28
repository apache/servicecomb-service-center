package access

import (
	"fmt"
	. "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/server/helper"
	"github.com/ServiceComb/service-center/util/url"
	"net/http"
)

func addCommonResponseHeaders(w http.ResponseWriter) {
	w.Header().Add("server", REGISTRY_SERVICE_NAME+"/"+REGISTRY_VERSION)
}

func Intercept(w http.ResponseWriter, r *http.Request) error {
	helper.InitContext(r)

	addCommonResponseHeaders(w)

	if !urlvalidator.IsRequestURI(r.RequestURI) {
		err := fmt.Errorf("Invalid Request URI %s", r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return err
	}
	return nil
}
