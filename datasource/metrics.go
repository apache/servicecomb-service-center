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

package datasource

import (
	"github.com/apache/servicecomb-service-center/pkg/metrics"
	helper "github.com/apache/servicecomb-service-center/pkg/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	success = "SUCCESS"
	failure = "FAILURE"
)

var (
	scCounter = helper.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.FamilyName,
			Subsystem: "db",
			Name:      "sc_total",
			Help:      "Counter of the Service Center instance",
		}, []string{"instance"})

	heartbeatCounter = helper.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.FamilyName,
			Subsystem: "db",
			Name:      "heartbeat_total",
			Help:      "Counter of heartbeat renew",
		}, []string{"instance", "status"})

	heartbeatLatency = helper.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  metrics.FamilyName,
			Subsystem:  "db",
			Name:       "heartbeat_durations_microseconds",
			Help:       "Latency of heartbeat renew",
			Objectives: metrics.Pxx,
		}, []string{"instance", "status"})
)

func ReportScInstance() {
	instance := metrics.InstanceName()
	scCounter.WithLabelValues(instance).Add(1)
}

func ReportHeartbeatCompleted(err error, start time.Time) {
	instance := metrics.InstanceName()
	elapsed := float64(time.Since(start).Nanoseconds()) / float64(time.Microsecond)
	status := success
	if err != nil {
		status = failure
	}
	heartbeatLatency.WithLabelValues(instance, status).Observe(elapsed)
	heartbeatCounter.WithLabelValues(instance, status).Inc()
}
