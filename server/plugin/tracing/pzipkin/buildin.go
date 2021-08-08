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
	"net/http"
	"net/url"
	"sync"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/plugin"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/plugin/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

var once sync.Once

func init() {
	plugin.RegisterPlugin(plugin.Plugin{Kind: tracing.TRACING, Name: "buildin", New: New})
}

func New() plugin.Instance {
	return &Zipkin{}
}

type Zipkin struct {
}

func (zp *Zipkin) ServerBegin(operationName string, itf interface{}) interface{} {
	var (
		span opentracing.Span
		ctx  context.Context
	)
	switch itf := itf.(type) {
	case *http.Request:
		r := itf
		ctx = r.Context()

		wireContext, err := ZipkinTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		switch err {
		case nil:
		case opentracing.ErrSpanContextNotFound:
		default:
			log.Error("tracer extract request failed", err)
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

	_ = util.SetContext(ctx, tracing.CtxTraceSpan, span)
	return span
}

func (zp *Zipkin) ServerEnd(itf interface{}, code int, message string) {
	span, ok := itf.(opentracing.Span)
	if !ok {
		return
	}
	setResultTags(span, code, message)
	span.Finish()
}

func (zp *Zipkin) ClientBegin(operationName string, itf interface{}) interface{} {
	var (
		span opentracing.Span
	)
	switch r := itf.(type) {
	case *http.Request:
		ctx := r.Context()

		parentSpan, ok := ctx.Value(tracing.CtxTraceSpan).(opentracing.Span)
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
			log.Error("tracer inject request failed", err)
		}
	case *tracing.Request:
		ctx := r.Ctx

		parentSpan, ok := ctx.Value(tracing.CtxTraceSpan).(opentracing.Span)
		if !ok {
			return nil
		}

		u, _ := url.Parse(r.Endpoint + r.URL)

		span = ZipkinTracer().StartSpan(operationName, opentracing.ChildOf(parentSpan.Context()))
		ext.SpanKindRPCClient.Set(span)
		ext.HTTPMethod.Set(span, r.Method)
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
			log.Error("tracer inject request failed", err)
		}
		// inject context
		err := carrier.ForeachKey(func(key, val string) error {
			ctx = util.SetContext(ctx, util.CtxKey(key), val)
			return nil
		})
		if err != nil {
			log.Error("", err)
		}
	default:
		return nil
	}

	return span
}

func (zp *Zipkin) ClientEnd(itf interface{}, code int, message string) {
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
