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
package interceptor

import (
	"github.com/ServiceComb/service-center/util"
	"net/http"
	"reflect"
	"runtime"
)

type Phase string

const (
	ACCESS_PHASE  Phase = "ACCESS PHASE"
	FILTER_PHASE  Phase = "FILTER PHASE"
	CONTENT_PHASE Phase = "CONTENT PHASE"
	LOG_PHASE     Phase = "LOG PHASE"

	DEFAULT_INTERCEPTION_SIZE = 10
)

var interceptors map[Phase][]*Interception

type InterceptorFunc func(http.ResponseWriter, *http.Request) error

func (f InterceptorFunc) Name() string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

type Interception struct {
	function InterceptorFunc
}

// Invoke performs the given interception.
// val is a pointer to the App Controller.
func (i Interception) Invoke(w http.ResponseWriter, req *http.Request) error {
	return i.function(w, req)
}

func init() {
	interceptors = make(map[Phase][]*Interception)
}

// InterceptFunc installs a general interceptor.
// This can be applied to any Controller.
// It must have the signature of:
//   func example(c *revel.Controller) revel.Result
func InterceptFunc(phase Phase, intc InterceptorFunc) {
	iters, ok := interceptors[phase]
	if !ok {
		iters = make([]*Interception, 0, DEFAULT_INTERCEPTION_SIZE)
	}

	iters = append(iters, &Interception{
		function: intc,
	})

	interceptors[phase] = iters

	util.LOGGER.Warnf(nil, "Intercept %s at %s", intc.Name(), phase)
}

func InvokeInterceptors(phase Phase, w http.ResponseWriter, req *http.Request) error {
	for _, intc := range interceptors[phase] {
		err := intc.Invoke(w, req)
		if err != nil {
			return err
		}
	}
	return nil
}
