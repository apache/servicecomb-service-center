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
package cache

import (
	"github.com/apache/servicecomb-service-center/server/metric"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

var (
	eventsCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: "db",
			Name:      "backend_event_total",
			Help:      "Counter of backend events",
		}, []string{"instance", "prefix"})

	eventsLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  metric.FamilyName,
			Subsystem:  "db",
			Name:       "backend_event_durations_microseconds",
			Help:       "Latency of backend events processing",
			Objectives: metric.Pxx,
		}, []string{"instance", "prefix"})
)

func init() {
	prometheus.MustRegister(eventsCounter, eventsLatency)
}

func ReportProcessEventCompleted(prefix string, evts []KvEvent) {
	l := float64(len(evts))
	if l == 0 {
		return
	}
	instance := metric.InstanceName()
	now := time.Now()
	for _, evt := range evts {
		elapsed := float64(now.Sub(evt.CreateAt.Local()).Nanoseconds()) / float64(time.Microsecond)
		eventsLatency.WithLabelValues(instance, prefix).Observe(elapsed)
	}
	eventsCounter.WithLabelValues(instance, prefix).Add(l)
}
