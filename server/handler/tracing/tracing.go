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
package tracing

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/chain"
	"github.com/apache/incubator-servicecomb-service-center/pkg/rest"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
	"net/http"
)

var tracer opentracing.Tracer

func init() {
	collector, err := zipkin.NewHTTPCollector("http://127.0.0.1:9411/api/v1/spans")
	if err != nil {
		return
	}
	recorder := zipkin.NewRecorder(collector, false, "0.0.0.0:0", core.Service.ServiceName)
	tracer, err = zipkin.NewTracer(recorder, zipkin.TraceID128Bit(true))
}

type TracingHandler struct {
}

func (h *TracingHandler) Handle(i *chain.Invocation) {
	w, request := i.Context().Value(rest.CTX_RESPONSE).(http.ResponseWriter),
		i.Context().Value(rest.CTX_REQUEST).(*http.Request)
	ctx, err := tracer.Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(request.Header))
	switch err {
	case nil:
	case opentracing.ErrSpanContextNotFound:
	default:
	}

	span := tracer.StartSpan("api", ext.RPCServerOption(ctx))
	ext.SpanKindRPCServer.Set(span)

	cb := i.Func
	i.Invoke(func(r chain.Result) {
		cb(r)
		span.SetTag(zipkincore.HTTP_METHOD, request.Method)
		span.SetTag(zipkincore.HTTP_PATH, request.RequestURI)
		span.SetTag(zipkincore.HTTP_STATUS_CODE, w.Header().Get("X-Response-Status"))
		span.SetTag(zipkincore.HTTP_HOST, request.URL.Host)
		span.Finish()
	})
}

func RegisterHandlers() {
	chain.RegisterHandler(rest.SERVER_CHAIN_NAME, &TracingHandler{})
}
