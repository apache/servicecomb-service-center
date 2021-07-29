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
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/metrics"
	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

const (
	collectorType       = "TRACING_COLLECTOR"
	fileCollectorPath   = "TRACING_FILE_PATH"
	serverCollectorAddr = "TRACING_SERVER_ADDRESS"
	samplerRate         = "TRACING_SIMPLER_RATE"
	defaultSamplerRate  = 1
	serviceName         = "service-center"
)

func initTracer() {
	collector, err := newCollector()
	if err != nil {
		log.Errorf(err, "new tracing collector failed, use the noop tracer")
		return
	}
	ipPort := metrics.InstanceName()
	recorder := zipkin.NewRecorder(collector, false, ipPort, serviceName)
	tracer, err := zipkin.NewTracer(recorder,
		zipkin.TraceID128Bit(true),
		zipkin.WithSampler(zipkin.NewCountingSampler(GetSamplerRate())))
	if err != nil {
		log.Errorf(err, "new tracer failed")
		return
	}
	opentracing.SetGlobalTracer(tracer)
}

func newCollector() (collector zipkin.Collector, err error) {
	ct := config.GetString("tracing.zipkin.collector.type", "", config.WithENV(collectorType))
	switch ct {
	case "server":
		sa := GetServerEndpoint()
		collector, err = zipkin.NewHTTPCollector(sa + "/api/v1/spans")
		if err != nil {
			return
		}
	case "file":
		fp := GetFilePath(serviceName + ".trace")
		collector, err = NewFileCollector(fp)
		if err != nil {
			return
		}
	default:
		err = fmt.Errorf("unknown tracing collector type '%s'", ct)
	}
	return
}

func ZipkinTracer() opentracing.Tracer {
	once.Do(initTracer)
	return opentracing.GlobalTracer()
}

func GetFilePath(defName string) string {
	path := config.GetString("tracing.zipkin.collector.path", "", config.WithENV(fileCollectorPath))
	if len(path) == 0 {
		wd, _ := os.Getwd()
		return filepath.Join(wd, defName)
	}
	return path
}

func GetServerEndpoint() string {
	sa := config.GetString("tracing.zipkin.collector.endpoint", "", config.WithENV(serverCollectorAddr))
	if len(sa) == 0 {
		sa = "http://127.0.0.1:9411"
	}
	return sa
}

func GetSamplerRate() float64 {
	strRate := config.GetString("tracing.zipkin.sampler.rate", "", config.WithENV(samplerRate))
	rate, err := strconv.ParseFloat(strRate, 64)
	if rate <= 0 || err != nil {
		return defaultSamplerRate
	}
	return rate
}
