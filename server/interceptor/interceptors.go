/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package interceptor

import (
	"fmt"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
)

var interceptors []*Interception

type Intercept func(http.ResponseWriter, *http.Request) error

func (f Intercept) Name() string {
	return util.FuncName(f)
}

type Interception struct {
	function Intercept
}

func (i Interception) Invoke(w http.ResponseWriter, req *http.Request) error {
	return i.function(w, req)
}

func init() {
	interceptors = make([]*Interception, 0, 10)
}

func RegisterInterceptFunc(intc Intercept) {
	interceptors = append(interceptors, &Interception{
		function: intc,
	})

	log.Info(fmt.Sprintf("Intercept %s", intc.Name()))
}

func InvokeInterceptors(w http.ResponseWriter, req *http.Request) (err error) {
	var intc *Interception
	defer func() {
		if itf := recover(); itf != nil {
			log.Panic(itf)

			http.Error(w, fmt.Sprintf("%v", itf), http.StatusInternalServerError)
		}
	}()
	for _, intc = range interceptors {
		err = intc.Invoke(w, req)
		if err != nil {
			return
		}
	}
	return
}
