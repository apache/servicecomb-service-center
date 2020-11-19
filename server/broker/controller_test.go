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
package broker_test

import (
	"bytes"
	. "github.com/apache/servicecomb-service-center/server/broker"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	ctrl = &Controller{}
	num  int
)

type mockBrokerHandler struct {
}

func (b *mockBrokerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := ctrl.URLPatterns()[num]
	r.Method = route.Method
	r.URL.RawQuery = ":sha=1"
	route.Func(w, r)
	num++
}

func TestBrokerController_GetHome(t *testing.T) {
	svr := httptest.NewServer(&mockBrokerHandler{})
	defer svr.Close()

	for range ctrl.URLPatterns() {
		http.Post(svr.URL, "application/json", bytes.NewBuffer([]byte("{}")))
	}
}
