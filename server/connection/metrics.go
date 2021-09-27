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

package connection

import (
	"time"

	"github.com/apache/servicecomb-service-center/pkg/event"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	success = "SUCCESS"
	failure = "FAILURE"
)

var (
	notifyCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.FamilyName,
			Subsystem: "notify",
			Name:      "publish_total",
			Help:      "Counter of publishing instance events",
		}, []string{"instance", "source", "status"})

	notifyLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  metrics.FamilyName,
			Subsystem:  "notify",
			Name:       "publish_durations_microseconds",
			Help:       "Latency of publishing instance events",
			Objectives: metrics.Pxx,
		}, []string{"instance", "source", "status"})

	pendingGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: "notify",
			Name:      "pending_total",
			Help:      "Counter of pending instance events",
		}, []string{"instance", "source"})

	pendingLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  metrics.FamilyName,
			Subsystem:  "notify",
			Name:       "pending_durations_microseconds",
			Help:       "Latency of pending instance events",
			Objectives: metrics.Pxx,
		}, []string{"instance", "source"})

	subscriberGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: "notify",
			Name:      "subscriber_total",
			Help:      "Gauge of subscribers",
		}, []string{"instance", "domain", "scheme"})
)

func init() {
	prometheus.MustRegister(notifyCounter, notifyLatency, subscriberGauge, pendingLatency, pendingGauge)
}

func ReportPublishCompleted(evt event.Event, err error) {
	instance := metrics.InstanceName()
	elapsed := float64(time.Since(evt.CreateAt()).Nanoseconds()) / float64(time.Microsecond)
	status := success
	if err != nil {
		status = failure
	}
	notifyLatency.WithLabelValues(instance, evt.Type().String(), status).Observe(elapsed)
	notifyCounter.WithLabelValues(instance, evt.Type().String(), status).Inc()
	pendingGauge.WithLabelValues(instance, evt.Type().String()).Dec()
}

func ReportPendingCompleted(evt event.Event) {
	instance := metrics.InstanceName()
	elapsed := float64(time.Since(evt.CreateAt()).Nanoseconds()) / float64(time.Microsecond)
	pendingLatency.WithLabelValues(instance, evt.Type().String()).Observe(elapsed)
	pendingGauge.WithLabelValues(instance, evt.Type().String()).Inc()
}

func ReportSubscriber(domain, scheme string, n float64) {
	instance := metrics.InstanceName()

	subscriberGauge.WithLabelValues(instance, domain, scheme).Add(n)
}
