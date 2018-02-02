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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"os"
	"path/filepath"
	"strings"
)

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

func ZipkinTracer() opentracing.Tracer {
	once.Do(initTracer)
	// use the NOOP tracer if init failed
	return opentracing.GlobalTracer()
}

func GetFilePath(defName string) string {
	path := os.Getenv("TRACING_FILE_PATH")
	if len(path) == 0 {
		wd, _ := os.Getwd()
		return filepath.Join(wd, defName)
	}
	return path
}

func GetServerEndpoint() string {
	sa := os.Getenv("TRACING_SERVER_ADDRESS")
	if len(sa) == 0 {
		sa = "http://127.0.0.1:9411"
	}
	return sa
}
