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
package pzipkin

import (
	"context"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/apache/servicecomb-service-center/server/plugin/registry"
	"github.com/apache/servicecomb-service-center/server/plugin/tracing"
	"net/http"
	"os"
	"testing"
)

func TestZipkin_XBegin(t *testing.T) {
	os.Setenv("TRACING_COLLECTOR", "server")
	core.Instance.HostName, core.Instance.Endpoints = "x", []string{"x"}
	initTracer()

	zk := New().(*Zipkin)
	span := zk.ServerBegin("x", nil)
	if span != nil {
		t.Fatalf("TestZipkin_XBegin failed")
	}
	span = zk.ClientBegin("x", nil)
	if span != nil {
		t.Fatalf("TestZipkin_XBegin failed")
	}

	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1:30100", nil)
	span = zk.ServerBegin("x", req)
	if span == nil {
		t.Fatalf("TestZipkin_XBegin failed")
	}
	zk.ServerEnd(span, 0, "")
	zk.ServerEnd(span, 400, "")

	if zk.ClientBegin("x", req) != nil {
		t.Fatalf("TestZipkin_XBegin failed")
	}

	req = req.WithContext(context.WithValue(req.Context(), tracing.CtxTraceSpan, span))
	span = zk.ClientBegin("x", req)
	if span == nil {
		t.Fatalf("TestZipkin_XBegin failed")
	}

	zk.ClientEnd(span, 0, "")

	span = zk.ClientBegin("x", &tracing.RegistryRequest{
		Ctx:      req.Context(),
		Options:  registry.OpGet(),
		Endpoint: "x",
	})
	if span == nil {
		t.Fatalf("TestZipkin_XBegin failed")
	}

	zk.ClientEnd(span, 0, "")
	zk.ClientEnd(span, 400, "")
}
