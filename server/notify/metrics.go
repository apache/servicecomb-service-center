// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notify

import (
	"github.com/apache/servicecomb-service-center/server/metric"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	success = "SUCCESS"
	failure = "FAILURE"
)

var (
	notifyCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metric.FamilyName,
			Subsystem: "notify",
			Name:      "publish_total",
			Help:      "Counter of publishing instance events",
		}, []string{"instance", "source", "status"})

	notifyLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  metric.FamilyName,
			Subsystem:  "notify",
			Name:       "publish_durations_microseconds",
			Help:       "Latency of publishing instance events",
			Objectives: prometheus.DefObjectives,
		}, []string{"instance", "source", "status"})
)

func init() {
	prometheus.MustRegister(notifyCounter, notifyLatency)
}

func ReportPublishCompleted(source string, err error, start time.Time) {
	instance := metric.InstanceName()
	elapsed := float64(time.Since(start).Nanoseconds()) / float64(time.Microsecond)
	status := success
	if err != nil {
		status = failure
	}
	notifyLatency.WithLabelValues(instance, source, status).Observe(elapsed)
	notifyCounter.WithLabelValues(instance, source, status).Inc()
}
