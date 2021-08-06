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

package prometheus

import (
	"fmt"
	"net/http"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/go-chassis/go-chassis/v2/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

// Vectors is unsafe, so all NewXXXVec funcs should be called during the initialization phase
var Vectors = make(map[string]prometheus.Collector)

func registerMetrics(name string, vec prometheus.Collector) {
	if _, ok := Vectors[name]; ok {
		log.Warn(fmt.Sprintf("found duplicate metrics name[%s], override!", name))
	}
	if err := metrics.GetSystemPrometheusRegistry().Register(vec); err != nil {
		log.Fatal(fmt.Sprintf("register prometheus metrics[%s] failed", name), err)
	}
	Vectors[name] = vec
}

func NewCounterVec(opts prometheus.CounterOpts, labelNames []string) *prometheus.CounterVec {
	name := util.StringJoin([]string{opts.Subsystem, opts.Name}, "_")
	vec := prometheus.NewCounterVec(opts, labelNames)
	registerMetrics(name, vec)
	return vec
}

func NewGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	name := util.StringJoin([]string{opts.Subsystem, opts.Name}, "_")
	vec := prometheus.NewGaugeVec(opts, labelNames)
	registerMetrics(name, vec)
	return vec
}

func NewSummaryVec(opts prometheus.SummaryOpts, labelNames []string) *prometheus.SummaryVec {
	name := util.StringJoin([]string{opts.Subsystem, opts.Name}, "_")
	vec := prometheus.NewSummaryVec(opts, labelNames)
	registerMetrics(name, vec)
	return vec
}

func Gather() ([]*dto.MetricFamily, error) {
	return metrics.GetSystemPrometheusRegistry().Gather()
}

func HTTPHandler() http.Handler {
	return promhttp.HandlerFor(metrics.GetSystemPrometheusRegistry(), promhttp.HandlerOpts{})
}
