//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package rest

import (
	"github.com/ServiceComb/service-center/pkg/chain"
	"github.com/ServiceComb/service-center/pkg/rest"
	"github.com/ServiceComb/service-center/server/core"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"net/http"
	"strconv"
	"time"
)

var (
	incomingRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "service_center",
			Subsystem: "http",
			Name:      "request_total",
			Help:      "Counter of requests received into ROA handler",
		}, []string{"method", "instance", "api"})

	successfulRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "service_center",
			Subsystem: "http",
			Name:      "success_total",
			Help:      "Counter of successful requests processed by ROA handler",
		}, []string{"method", "code", "instance", "api"})

	reqDurations = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "service_center",
			Subsystem: "http",
			Name:      "request_durations_microseconds",
			Help:      "HTTP request latency summary of ROA handler",
		}, []string{"method", "instance", "api"})
)

func init() {
	prometheus.MustRegister(incomingRequests, successfulRequests, reqDurations)

	http.Handle("/metrics", prometheus.Handler())
}

func ReportRequestCompleted(i *chain.Invocation, start time.Time) {
	id := core.Instance.InstanceId
	elapsed := float64(time.Since(start).Nanoseconds()) / 1000
	r := i.Context().Value(rest.CTX_REQUEST).(*http.Request)
	route := i.Context().Value(rest.CTX_MATCH_PATTERN).(string)

	reqDurations.WithLabelValues(r.Method, id, route).Observe(elapsed)

	incomingRequests.WithLabelValues(r.Method, id, route).Inc()

	if success, code := codeOf(r.Header); success {
		successfulRequests.WithLabelValues(r.Method, code, id, route).Inc()
	}
}

func codeOf(h http.Header) (bool, string) {
	statusCode := h.Get("X-Response-Status")
	if statusCode == "" {
		return true, "200"
	}

	if code, _ := strconv.Atoi(statusCode); code >= http.StatusOK && code <= http.StatusAccepted {
		return true, statusCode
	}

	return false, statusCode
}

// Get value of metricFamily
func MetricValueOf(mf *dto.MetricFamily) float64 {
	if len(mf.GetMetric()) == 0 {
		return 0
	}

	switch mf.GetType() {
	case dto.MetricType_GAUGE:
		return mf.GetMetric()[0].GetGauge().GetValue()
	case dto.MetricType_COUNTER:
		return metricCounterOf(mf.GetMetric())
	case dto.MetricType_SUMMARY:
		return metricSummaryOf(mf.GetMetric())
	default:
		return 0
	}
}

func metricCounterOf(m []*dto.Metric) float64 {
	var sum float64 = 0
	for _, d := range m {
		sum += d.GetCounter().GetValue()
	}
	return sum
}

func metricSummaryOf(m []*dto.Metric) float64 {
	var (
		count uint64  = 0
		sum   float64 = 0
	)
	for _, d := range m {
		count += d.GetSummary().GetSampleCount()
		sum += d.GetSummary().GetSampleSum()
	}

	if count == 0 {
		return 0
	}

	return sum / float64(count)
}
