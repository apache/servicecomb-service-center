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

	"github.com/apache/servicecomb-service-center/pkg/log"
	metricsvc "github.com/apache/servicecomb-service-center/pkg/metrics"
	promutil "github.com/apache/servicecomb-service-center/pkg/prometheus"
	"github.com/go-chassis/go-chassis/v2/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	SubSystem            = "db"
	KeyServiceTotal      = metricsvc.FamilyName + "_" + SubSystem + "_" + "service_total"
	KeyInstanceTotal     = metricsvc.FamilyName + "_" + SubSystem + "_" + "instance_total"
	KeyServiceUsage      = metricsvc.FamilyName + "_" + SubSystem + "_" + "service_usage"
	KeyInstanceUsage     = metricsvc.FamilyName + "_" + SubSystem + "_" + "instance_usage"
	KeyDomainTotal       = metricsvc.FamilyName + "_" + SubSystem + "_" + "domain_total"
	KeySchemaTotal       = metricsvc.FamilyName + "_" + SubSystem + "_" + "schema_total"
	KeyFrameworkTotal    = metricsvc.FamilyName + "_" + SubSystem + "_" + "framework_total"
	KeyHeartbeatTotal    = metricsvc.FamilyName + "_" + SubSystem + "_" + "heartbeat_total"
	KeyHeartbeatDuration = metricsvc.FamilyName + "_" + SubSystem + "_" + "heartbeat_durations_microseconds"
	KeySCTotal           = metricsvc.FamilyName + "_" + SubSystem + "_" + "sc_total"
)

var metaEnabled = false

func InitMetaMetrics() (err error) {
	defer func() {
		if err != nil {
			log.Error("init metadata metrics failed", err)
		} else {
			metaEnabled = true
		}
	}()
	if err = metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyDomainTotal,
		Help:   "Gauge of domain created in Service Center",
		Labels: []string{"instance"},
	}); err != nil {
		return
	}
	if err = metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyServiceTotal,
		Help:   "Gauge of microservice created in Service Center",
		Labels: []string{"instance", "framework", "frameworkVersion", "domain", "project"},
	}); err != nil {
		return
	}
	if err = metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyInstanceTotal,
		Help:   "Gauge of microservice instance created in Service Center",
		Labels: []string{"instance", "framework", "frameworkVersion", "domain", "project"},
	}); err != nil {
		return
	}
	if err = metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyServiceUsage,
		Help:   "Gauge of microservice usage in Service Center",
		Labels: []string{"instance"},
	}); err != nil {
		return
	}
	if err = metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyInstanceUsage,
		Help:   "Gauge of microservice instance usage in Service Center",
		Labels: []string{"instance"},
	}); err != nil {
		return
	}
	if err = metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeySchemaTotal,
		Help:   "Counter of schema created in Service Center",
		Labels: []string{"instance", "domain", "project"},
	}); err != nil {
		return
	}
	if err = metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeyFrameworkTotal,
		Help:   "Gauge of client framework info in Service Center",
		Labels: []string{"instance", "framework", "frameworkVersion", "domain", "project"},
	}); err != nil {
		return
	}
	if err = metrics.CreateGauge(metrics.GaugeOpts{
		Key:    KeySCTotal,
		Help:   "Counter of the Service Center instance",
		Labels: []string{"instance"},
	}); err != nil {
		return
	}
	if err = metrics.CreateCounter(metrics.CounterOpts{
		Key:    KeyHeartbeatTotal,
		Help:   "Counter of heartbeat renew",
		Labels: []string{"instance", "status"},
	}); err != nil {
		return
	}
	if err = metrics.CreateSummary(metrics.SummaryOpts{
		Key:        KeyHeartbeatDuration,
		Help:       "Latency of heartbeat renew",
		Labels:     []string{"instance", "status"},
		Objectives: metricsvc.Pxx,
	}); err != nil {
		return
	}
	return
}

func GetTotalService(domain, project string) int64 {
	labels := prometheus.Labels{"domain": domain}
	if len(project) > 0 {
		labels["project"] = project
	}
	return int64(promutil.GaugeValue(KeyServiceTotal, labels))
}

func GetTotalInstance(domain, project string) int64 {
	labels := prometheus.Labels{"domain": domain}
	if len(project) > 0 {
		labels["project"] = project
	}
	return int64(promutil.GaugeValue(KeyInstanceTotal, labels))
}

func ReportScInstance() {
	instance := metricsvc.InstanceName()
	labels := map[string]string{"instance": instance}
	if err := metrics.GaugeSet(KeySCTotal, 1, labels); err != nil {
		log.Error("gauge set failed", err)
	}
}

func ReportHeartbeatCompleted(err error, start time.Time) {
	if !metaEnabled {
		return
	}
	instance := metricsvc.InstanceName()
	elapsed := float64(time.Since(start).Nanoseconds()) / float64(time.Microsecond)
	status := success
	if err != nil {
		status = failure
	}
	labels := map[string]string{"instance": instance, "status": status}
	if err := metrics.SummaryObserve(KeyHeartbeatDuration, elapsed, labels); err != nil {
		log.Error("summary observe failed", err)
	}
	if err = metrics.CounterAdd(KeyHeartbeatTotal, 1, labels); err != nil {
		log.Error("counter add failed", err)
	}
}
