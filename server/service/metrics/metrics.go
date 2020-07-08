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
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/metric"
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
	domainCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: SubSystem,
			Name:      KeyDomainTotal,
			Help:      "Gauge of domain created in Service Center",
		}, []string{"instance"})

	serviceCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: "db",
			Name:      KeyServiceTotal,
			Help:      "Gauge of microservice created in Service Center",
		}, []string{"instance", "framework", "frameworkVersion", "domain"})

	instanceCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: SubSystem,
			Name:      KeyInstanceTotal,
			Help:      "Gauge of microservice created in Service Center",
		}, []string{"instance", "domain"})

	schemaCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: SubSystem,
			Name:      KeySchemaTotal,
			Help:      "Gauge of schema created in Service Center",
		}, []string{"instance", "domain"})

	frameworkCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: SubSystem,
			Name:      KeyFrameworkTotal,
			Help:      "Gauge of client framework info in Service Center",
		}, metric.ToLabelNames(Framework{}))
)

// Framework return framework info.
type Framework struct {
	DomainName       string `json:"domainName"`
	ProjectName      string `json:"projectName"`
	FrameWork        string `json:"framework"`
	FrameworkVersion string `json:"frameworkVersion"`
}

func init() {
	prometheus.MustRegister(domainCounter, serviceCounter, instanceCounter, schemaCounter, frameworkCounter)
}

func ReportDomains(c float64) {
	instance := metric.InstanceName()
	domainCounter.WithLabelValues(instance).Add(c)
}

func ReportServices(domain, framework, frameworkVersion string, c float64) {
	instance := metric.InstanceName()
	serviceCounter.WithLabelValues(instance, framework, frameworkVersion, domain).Add(c)
}

func ReportInstances(domain string, c float64) {
	instance := metric.InstanceName()
	instanceCounter.WithLabelValues(instance, domain).Add(c)
}

func ReportSchemas(domain string, c float64) {
	instance := metric.InstanceName()
	schemaCounter.WithLabelValues(instance, domain).Add(c)
}

func ReportFramework(domainName, projectName string, framework, frameworkVersion string, c float64) {
	frameworkCounter.WithLabelValues(domainName, projectName, framework, frameworkVersion).Add(c)
}
