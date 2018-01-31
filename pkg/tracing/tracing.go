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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
	"net/http"
)

func init() {
	opentracing.GlobalTracer()
	collector, err := zipkin.NewHTTPCollector("http://127.0.0.1:9411/api/v1/spans")
	if err != nil {
		return
	}
	recorder := zipkin.NewRecorder(collector, false, "127.0.0.1:30100", core.Service.ServiceName)
	tracer, err := zipkin.NewTracer(recorder, zipkin.TraceID128Bit(true))
	if err != nil {
		return
	}
	RegisterTracer(tracer)
}

func RegisterTracer(tracer opentracing.Tracer) {
	opentracing.SetGlobalTracer(tracer)
}

func GetTracer() opentracing.Tracer {
	return opentracing.GlobalTracer()
}

func StartServerSpan(operationName string, r *http.Request) opentracing.Span {
	wireContext, err := GetTracer().Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(r.Header))
	switch err {
	case nil:
	case opentracing.ErrSpanContextNotFound:
	default:
	}

	span := GetTracer().StartSpan(operationName, ext.RPCServerOption(wireContext))
	ext.SpanKindRPCServer.Set(span)
	util.SetContext(r.Context(), "x-trace-span", span)

	span.SetTag("protocol", "HTTP")
	span.SetTag(zipkincore.HTTP_METHOD, r.Method)
	span.SetTag(zipkincore.HTTP_PATH, r.URL.Path)
	span.SetTag(zipkincore.HTTP_URL, r.URL.String())
	span.SetTag(zipkincore.HTTP_HOST, r.URL.Host)
	return span
}

func FinishServerSpan(span opentracing.Span, code int, message string) {
	result := "0"
	if code >= http.StatusBadRequest {
		result = "1"
		if len(message) == 0 {
			message = fmt.Sprint(code)
		}
		span.SetTag("error", message)
	}
	span.SetTag("resultCode", code)
	span.SetTag("result", result)
	span.SetTag(zipkincore.HTTP_STATUS_CODE, code)
	span.Finish()
}
