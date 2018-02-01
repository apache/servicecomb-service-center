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
package buildin

import (
	"context"
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/tracing"
	mgr "github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
	"net/http"
	"os"
	"strings"
	"sync"
)

var once sync.Once

func init() {
	mgr.RegisterPlugin(mgr.Plugin{mgr.TRACING, "buildin", New})
}

func New() mgr.PluginInstance {
	return &Zipkin{}
}

func initTracer() {
	collector, err := newCollector()
	if err != nil {
		util.Logger().Errorf(err, "new tracing collector failed")
		return
	}
	ipPort, _ := util.ParseEndpoint(core.Instance.Endpoints[0])
	recorder := zipkin.NewRecorder(collector, false, ipPort, core.Service.ServiceName)
	tracer, err := zipkin.NewTracer(recorder, zipkin.TraceID128Bit(true))
	if err != nil {
		return
	}
	opentracing.SetGlobalTracer(tracer)
}

func newCollector() (collector zipkin.Collector, err error) {
	ct := strings.TrimSpace(os.Getenv("TRACING_COLLECTOR"))
	switch ct {
	case "server":
		sa := GetServerEndpoint()
		collector, err = zipkin.NewHTTPCollector(sa + "/api/v1/spans")
		if err != nil {
			return
		}
	case "file":
		fp := GetFilePath(core.Service.ServiceName + ".trace")
		collector, err = NewFileCollector(fp)
		if err != nil {
			return
		}
	default:
		err = fmt.Errorf("unknown tracing collector type '%s'", ct)
	}
	return
}

func getTracer() opentracing.Tracer {
	once.Do(initTracer)
	// use the NOOP tracer if init failed
	return opentracing.GlobalTracer()
}

type Zipkin struct {
}

func (zp *Zipkin) ServerBegin(operationName string, itf tracing.Request) tracing.Span {
	var (
		span opentracing.Span
		ctx  context.Context
	)
	switch itf.(type) {
	case *http.Request:
		r := itf.(*http.Request)
		ctx = r.Context()

		wireContext, err := getTracer().Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(r.Header))
		switch err {
		case nil:
		case opentracing.ErrSpanContextNotFound:
		default:
			util.Logger().Errorf(err, "tracer extract request failed")
			return nil
		}

		span = getTracer().StartSpan(operationName, ext.RPCServerOption(wireContext))
		ext.SpanKindRPCServer.Set(span)
		ext.HTTPMethod.Set(span, r.Method)
		ext.HTTPUrl.Set(span, r.URL.String())

		span.SetTag("protocol", "HTTP")
		span.SetTag(zipkincore.HTTP_PATH, r.URL.Path)
		span.SetTag(zipkincore.HTTP_HOST, r.URL.Host)
	default:
		// grpc?
		return nil
	}

	util.SetContext(ctx, tracing.CTX_TRACE_SPAN, span)
	return span
}

func (zp *Zipkin) ServerEnd(itf tracing.Span, code int, message string) {
	span, ok := itf.(opentracing.Span)
	if !ok {
		return
	}
	setResultTags(span, code, message)
	span.Finish()
}

func (zp *Zipkin) ClientBegin(operationName string, itf tracing.Request) tracing.Span {
	var (
		span    opentracing.Span
		ctx     context.Context
		carrier interface{}
	)
	switch itf.(type) {
	case *http.Request:
		r := itf.(*http.Request)
		ctx = r.Context()

		parentSpan, ok := ctx.Value(tracing.CTX_TRACE_SPAN).(opentracing.Span)
		if !ok {
			return nil
		}
		span = getTracer().StartSpan(operationName, opentracing.ChildOf(parentSpan.Context()))
		ext.SpanKindRPCClient.Set(span)
		ext.HTTPMethod.Set(span, r.Method)
		ext.HTTPUrl.Set(span, r.URL.String())

		span.SetTag("protocol", "HTTP")
		span.SetTag(zipkincore.HTTP_PATH, r.URL.Path)
		span.SetTag(zipkincore.HTTP_HOST, r.URL.Host)

		carrier = opentracing.HTTPHeadersCarrier(r.Header)
	default:
		// grpc?
		return nil
	}

	util.SetContext(ctx, tracing.CTX_TRACE_SPAN, span)

	if err := getTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		carrier,
	); err != nil {
		util.Logger().Errorf(err, "tracer inject request failed")
	}

	return span
}

func (zp *Zipkin) ClientEnd(itf tracing.Span, code int, message string) {
	span, ok := itf.(opentracing.Span)
	if !ok {
		return
	}
	setResultTags(span, code, message)
	span.Finish()
}

func setResultTags(span opentracing.Span, code int, message string) {
	result := 0
	if code >= http.StatusBadRequest {
		result = 1
		if len(message) == 0 {
			message = fmt.Sprint(code)
		}
		span.SetTag("error", message)
	}
	span.SetTag("resultCode", code)
	span.SetTag("result", result)
	ext.HTTPStatusCode.Set(span, uint16(code))
}
