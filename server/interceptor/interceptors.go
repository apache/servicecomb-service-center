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
	"github.com/ServiceComb/service-center/pkg/util"
	"net/http"
)

var interceptors []*Interception

type InterceptorFunc func(http.ResponseWriter, *http.Request) error

func (f InterceptorFunc) Name() string {
	return util.FuncName(f)
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
	interceptors = make([]*Interception, 0, 10)
}

// InterceptFunc installs a general interceptor.
// This can be applied to any Controller.
// It must have the signature of:
//   func example(c *revel.Controller) revel.Result
func RegisterInterceptFunc(intc InterceptorFunc) {
	interceptors = append(interceptors, &Interception{
		function: intc,
	})

	util.Logger().Infof("Intercept %s", intc.Name())
}

func InvokeInterceptors(w http.ResponseWriter, req *http.Request) error {
	var intc *Interception
	defer func() {
		if itf := recover(); itf != nil {
			name := util.FuncName(intc.function)
			util.Logger().Errorf(nil, "recover from '%s()'! %v", name, itf)
		}
	}()
	for _, intc = range interceptors {
		err := intc.Invoke(w, req)
		if err != nil {
			return err
		}
	}
	return nil
}
