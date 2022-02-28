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

package accesslog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/chain"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/server/handler/accesslog"
	_ "github.com/apache/servicecomb-service-center/test"
)

func TestHandler(t *testing.T) {
	// add white list apis, ignore
	l := log.NewLogger(log.Config{})
	h := accesslog.NewAccessLogHandler(l)
	testAPI := "testAPI"
	h.AddWhiteListAPIs(testAPI)
	if !h.ShouldIgnoreAPI(testAPI) {
		t.Fatalf("Should ignore API: %s", testAPI)
	}

	// handle
	inv := &chain.Invocation{}
	ctx := context.Background()
	inv.Init(ctx, chain.NewChain("c", []chain.Handler{}))
	inv.WithContext(rest.CtxMatchPattern, "/a")
	r, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:80/a", nil)
	w := httptest.NewRecorder()
	inv.WithContext(rest.CtxRequest, r)
	inv.WithContext(rest.CtxResponse, w)
	h.Handle(inv)
}
