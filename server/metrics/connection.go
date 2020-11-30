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

package metrics

import (
	"time"

	"github.com/apache/servicecomb-service-center/pkg/metrics"
	"github.com/apache/servicecomb-service-center/pkg/notify"
	helper "github.com/apache/servicecomb-service-center/pkg/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	notifyCounter = helper.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.FamilyName,
			Subsystem: "notify",
			Name:      "publish_total",
			Help:      "Counter of publishing instance events",
		}, []string{"instance", "source", "status"})

	notifyLatency = helper.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  metrics.FamilyName,
			Subsystem:  "notify",
			Name:       "publish_durations_microseconds",
			Help:       "Latency of publishing instance events",
			Objectives: metrics.Pxx,
		}, []string{"instance", "source", "status"})

	subscriberGauge = helper.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: "notify",
			Name:      "subscriber_total",
			Help:      "Gauge of subscribers",
		}, []string{"instance", "domain", "scheme"})
)

func ReportPublishCompleted(evt notify.Event, err error) {
	instance := metrics.InstanceName()
	elapsed := float64(time.Since(evt.CreateAt()).Nanoseconds()) / float64(time.Microsecond)
	status := success
	if err != nil {
		status = failure
	}
	notifyLatency.WithLabelValues(instance, evt.Type().String(), status).Observe(elapsed)
	notifyCounter.WithLabelValues(instance, evt.Type().String(), status).Inc()
}

func ReportSubscriber(domain, scheme string, n float64) {
	instance := metrics.InstanceName()

	subscriberGauge.WithLabelValues(instance, domain, scheme).Add(n)
}
