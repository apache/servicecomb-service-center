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
package discovery

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
	eventsCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: "db",
			Name:      "backend_event_total",
			Help:      "Counter of backend events",
		}, []string{"instance"})

	eventsLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  metric.FamilyName,
			Subsystem:  "db",
			Name:       "backend_event_durations_microseconds",
			Help:       "Latency of backend events processing",
			Objectives: prometheus.DefObjectives,
		}, []string{"instance"})
)

func init() {
	prometheus.MustRegister(eventsCounter, eventsLatency)
}

func ReportProcessEventCompleted(evts []KvEvent, start time.Time) {
	l := float64(len(evts))
	if l == 0 {
		return
	}
	instance := metric.InstanceName()
	elapsed := float64(time.Since(start).Nanoseconds()) / float64(time.Microsecond)
	eventsLatency.WithLabelValues(instance).Observe(elapsed / l)
	eventsCounter.WithLabelValues(instance).Add(l)
}
