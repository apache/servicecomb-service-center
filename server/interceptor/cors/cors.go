package cors

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"errors"
	"github.com/ServiceComb/service-center/util"
)

var cors *CORS

func init() {
	cors = New()
}

type CORS struct {
	allowOrigin      string
	allowMethods     map[string]bool
	allowHeaders     map[string]bool
	allowCredentials bool
	exposeHeaders    string
	maxAge           int
	userHandler      http.Handler
}

func New() *CORS {
	c := new(CORS)
	c.allowOrigin = "*"
	c.allowCredentials = false
	c.allowHeaders = map[string]bool{"origin": true}
	c.allowMethods = map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "UPDATE": true}
	c.maxAge = 1500
	return c
}

func setToArray(set map[string]bool) []string {
	ret := make([]string, 0, len(set))
	for k := range set {
		ret = append(ret, k)
	}
	return ret
}

func (cors *CORS) AllowMethods() []string {
	return setToArray(cors.allowMethods)
}

func (cors *CORS) AllowHeaders() []string {
	return setToArray(cors.allowHeaders)
}

func (cors *CORS) handlePreflightRequest(w http.ResponseWriter, r *http.Request) {
	acrm := r.Header.Get("Access-Control-Request-Method")
	if acrm == "" {
		cors.invalid(w, r)
		util.LOGGER.Warnf(nil,"header 'Access-Control-Request-Method' is empty")
		return
	}
	methods := strings.Split(strings.TrimSpace(acrm), ",")
	for _, m := range methods {
		m = strings.TrimSpace(m)
		if _, ok := cors.allowMethods[m]; !ok {
			cors.invalid(w, r)
			util.LOGGER.Warnf(nil,"only supported methods: %v", cors.allowMethods)
			return
		}
	}
	acrh := r.Header.Get("Access-Control-Request-Headers")
	if acrh != "" {
		headers := strings.Split(strings.TrimSpace(acrh), ",")
		for _, h := range headers {
			h = strings.ToLower(strings.TrimSpace(h))
			if _, ok := cors.allowHeaders[h]; !ok {
				cors.invalid(w, r)
				util.LOGGER.Warnf(nil,"only supported headers: %v", cors.allowHeaders)
				return
			}
		}
	}

	w.Header().Add("Access-Control-Allow-Methods", strings.Join(cors.AllowMethods(), ","))
	w.Header().Add("Access-Control-Allow-Headers", strings.Join(cors.AllowHeaders(), ","))
	w.Header().Add("Access-Control-Max-Age", strconv.Itoa(cors.maxAge))
	cors.addAllowOriginHeader(w, r)
	cors.addAllowCookiesHeader(w, r)
	return
}

func (cors *CORS) invalid(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, "CORS Request Invalid")
	return
}

func (cors *CORS) handleActualRequest(w http.ResponseWriter, r *http.Request) {
	if cors.exposeHeaders != "" {
		w.Header().Add("Access-Control-Expose-Headers", cors.exposeHeaders)
	}
	cors.addAllowOriginHeader(w, r)
	cors.addAllowCookiesHeader(w, r)
	return
}

func (cors *CORS) addAllowOriginHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", cors.allowOrigin)
	return
}

func (cors *CORS) addAllowCookiesHeader(w http.ResponseWriter, r *http.Request) {
	if cors.allowCredentials {
		w.Header().Add("Access-Control-Allow-Credentials", "true")
	}
}

func Intercept(w http.ResponseWriter, r *http.Request) (err error) {
	if origin := r.Header.Get("Origin"); origin == "" {
	} else if r.Method != "OPTIONS" {
		cors.handleActualRequest(w, r)
	} else if acrm := r.Header.Get("Access-Control-Request-Method"); acrm == "" {
		cors.handleActualRequest(w, r)
	} else {
		cors.handlePreflightRequest(w, r)
		err = errors.New("Handle preflight request")
	}
	return
}