//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package maxbody

import (
	"context"
	"fmt"
	. "github.com/ServiceComb/service-center/server/core"
	"github.com/ServiceComb/service-center/util/url"
	"github.com/astaxie/beego"
	"net/http"
	"strings"
)

type MaxBytesHandler struct {
	MaxBytes int64
}

func (h *MaxBytesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	method := r.Method
	w.Header().Add("server", REGISTRY_SERVICE_NAME+"/"+REGISTRY_VERSION)
	switch method {
	case "GET", "POST", "PUT", "DELETE":
	// Only methods in this branch are allowed
	default:
		// unsupported http method, response 405(method not allowed)
		err := fmt.Errorf("Method Not Allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(err.Error()))
		return err
	}

	if !urlvalidator.IsRequestURI(r.RequestURI) {
		err := fmt.Errorf("Invalid Request URI %s", r.RequestURI)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return err
	}
	r.Body = http.MaxBytesReader(w, r.Body, h.MaxBytes)

	addIPToContext(r)
	return nil
}

var maxBodyHandler MaxBytesHandler

func init() {
	maxBodyHandler.MaxBytes = beego.AppConfig.DefaultInt64("max_body_bytes", 40960)
}

func Intercept(w http.ResponseWriter, r *http.Request) error {
	return maxBodyHandler.ServeHTTP(w, r)
}

func GetRealIP(r *http.Request) string {
	addrs := strings.Split(r.RemoteAddr, ":")
	if len(addrs) > 0 {
		return addrs[0]
	}
	return ""
}

func addIPToContext(r *http.Request) {
	terminalIP := GetRealIP(r)
	ctx := r.Context()
	ctx = context.WithValue(ctx, "x-remote-ip", terminalIP)
	request := r.WithContext(ctx)
	*r = *request
}
