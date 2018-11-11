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
	"github.com/apache/servicecomb-service-center/server/metric"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	domainCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: "db",
			Name:      "domain_total",
			Help:      "Gauge of domain created in Service Center",
		}, []string{"instance"})

	serviceCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: "db",
			Name:      "service_total",
			Help:      "Gauge of microservice created in Service Center",
		}, []string{"instance", "framework", "frameworkVersion"})

	instanceCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: "db",
			Name:      "instance_total",
			Help:      "Gauge of microservice created in Service Center",
		}, []string{"instance"})
)

func init() {
	prometheus.MustRegister(domainCounter, serviceCounter, instanceCounter)
}

func ReportDomains(c float64) {
	instance := metric.InstanceName()
	domainCounter.WithLabelValues(instance).Add(c)
}

func ReportServices(framework, frameworkVersion string, c float64) {
	instance := metric.InstanceName()
	serviceCounter.WithLabelValues(instance, framework, frameworkVersion).Add(c)
}

func ReportInstances(c float64) {
	instance := metric.InstanceName()
	instanceCounter.WithLabelValues(instance).Add(c)
}
