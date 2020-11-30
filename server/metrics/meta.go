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
	helper "github.com/apache/servicecomb-service-center/pkg/prometheus"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
)

// keys of gauge
const (
	KeyDomainTotal    = "domain_total"
	KeyServiceTotal   = "service_total"
	KeyInstanceTotal  = "instance_total"
	KeySchemaTotal    = "schema_total"
	KeyFrameworkTotal = "framework_total"

	SubSystem = "db"
)

// Key return metrics key
func Key(name string) string {
	return util.StringJoin([]string{SubSystem, name}, "_")
}

var (
	domainCounter = helper.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: SubSystem,
			Name:      KeyDomainTotal,
			Help:      "Gauge of domain created in Service Center",
		}, []string{"instance"})

	serviceCounter = helper.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: "db",
			Name:      KeyServiceTotal,
			Help:      "Gauge of microservice created in Service Center",
		}, []string{"instance", "framework", "frameworkVersion", "domain"})

	instanceCounter = helper.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: SubSystem,
			Name:      KeyInstanceTotal,
			Help:      "Gauge of microservice created in Service Center",
		}, []string{"instance", "domain"})

	schemaCounter = helper.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: SubSystem,
			Name:      KeySchemaTotal,
			Help:      "Gauge of schema created in Service Center",
		}, []string{"instance", "domain"})

	frameworkCounter = helper.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: SubSystem,
			Name:      KeyFrameworkTotal,
			Help:      "Gauge of client framework info in Service Center",
		}, metrics.ToLabelNames(Framework{}))

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

// Framework return framework info.
type Framework struct {
	DomainName       string `json:"domainName"`
	ProjectName      string `json:"projectName"`
	FrameWork        string `json:"framework"`
	FrameworkVersion string `json:"frameworkVersion"`
}

func ReportDomains(c float64) {
	instance := metrics.InstanceName()
	domainCounter.WithLabelValues(instance).Add(c)
}

func ReportServices(domain, framework, frameworkVersion string, c float64) {
	instance := metrics.InstanceName()
	serviceCounter.WithLabelValues(instance, framework, frameworkVersion, domain).Add(c)
}

func ReportInstances(domain string, c float64) {
	instance := metrics.InstanceName()
	instanceCounter.WithLabelValues(instance, domain).Add(c)
}

func ReportSchemas(domain string, c float64) {
	instance := metrics.InstanceName()
	schemaCounter.WithLabelValues(instance, domain).Add(c)
}

func ReportFramework(domainName, projectName string, framework, frameworkVersion string, c float64) {
	frameworkCounter.WithLabelValues(domainName, projectName, framework, frameworkVersion).Add(c)
}

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
