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
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/tracing"
	mgr "github.com/apache/incubator-servicecomb-service-center/server/plugin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
	"net/http"
	"net/url"
	"sync"
)

var once sync.Once

func init() {
	mgr.RegisterPlugin(mgr.Plugin{mgr.TRACING, "buildin", New})
}

func New() mgr.PluginInstance {
	return &Zipkin{}
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

		wireContext, err := ZipkinTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		switch err {
		case nil:
		case opentracing.ErrSpanContextNotFound:
		default:
			util.Logger().Errorf(err, "tracer extract request failed")
			return nil
		}

		span = ZipkinTracer().StartSpan(operationName, ext.RPCServerOption(wireContext))
		ext.SpanKindRPCServer.Set(span)
		ext.HTTPMethod.Set(span, r.Method)
		ext.HTTPUrl.Set(span, util.ParseRequestURL(r))

		span.SetTag("protocol", "HTTP")
		span.SetTag(zipkincore.HTTP_PATH, r.URL.Path)
		span.SetTag(zipkincore.HTTP_HOST, r.URL.Host)
	default:
		// grpc?
		return nil
	}

	ctx = util.SetContext(ctx, tracing.CTX_TRACE_SPAN, span)
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
		span opentracing.Span
	)
	switch itf.(type) {
	case *http.Request:
		r := itf.(*http.Request)
		ctx := r.Context()

		parentSpan, ok := ctx.Value(tracing.CTX_TRACE_SPAN).(opentracing.Span)
		if !ok {
			return nil
		}
		span = ZipkinTracer().StartSpan(operationName, opentracing.ChildOf(parentSpan.Context()))
		ext.SpanKindRPCClient.Set(span)
		ext.HTTPMethod.Set(span, r.Method)
		ext.HTTPUrl.Set(span, util.ParseRequestURL(r))

		span.SetTag("protocol", "HTTP")
		span.SetTag(zipkincore.HTTP_PATH, r.URL.Path)
		span.SetTag(zipkincore.HTTP_HOST, r.URL.Host)

		carrier := opentracing.HTTPHeadersCarrier(r.Header)

		if err := ZipkinTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			carrier,
		); err != nil {
			util.Logger().Errorf(err, "tracer inject request failed")
		}
	case *tracing.RegistryRequest:
		r := itf.(*tracing.RegistryRequest)
		ctx := r.Ctx

		parentSpan, ok := ctx.Value(tracing.CTX_TRACE_SPAN).(opentracing.Span)
		if !ok {
			return nil
		}

		u, _ := url.Parse(r.Endpoint + "/?" + r.Options.FormatUrlParams())

		span = ZipkinTracer().StartSpan(operationName, opentracing.ChildOf(parentSpan.Context()))
		ext.SpanKindRPCClient.Set(span)
		ext.HTTPMethod.Set(span, r.Options.Action.String())
		ext.HTTPUrl.Set(span, u.String())

		span.SetTag("protocol", "gRPC")
		span.SetTag(zipkincore.HTTP_PATH, u.Path)
		span.SetTag(zipkincore.HTTP_HOST, u.Host)

		carrier := opentracing.HTTPHeadersCarrier{}
		if err := ZipkinTracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			carrier,
		); err != nil {
			util.Logger().Errorf(err, "tracer inject request failed")
		}
		// inject context
		carrier.ForeachKey(func(key, val string) error {
			ctx = util.SetContext(ctx, key, val)
			return nil
		})
	default:
		return nil
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
	if code >= http.StatusBadRequest {
		span.SetTag("error", message)
	}
	ext.HTTPStatusCode.Set(span, uint16(code))
}
