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

package prometheus

import (
	"github.com/apache/servicecomb-service-center/server/metrics"
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

	LabelInstance         = "instance"
	LabelFramework        = "framework"
	LabelFrameworkVersion = "frameworkVersion"
	LabelDomain           = "domain"
	LabelProject          = "project"
)

var (
	domainCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: SubSystem,
			Name:      KeyDomainTotal,
			Help:      "Gauge of domain created in Service Center",
		}, []string{LabelInstance})

	serviceCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: "db",
			Name:      KeyServiceTotal,
			Help:      "Gauge of microservice created in Service Center",
		}, []string{LabelInstance, LabelFramework, LabelFrameworkVersion, LabelDomain, LabelProject})

	instanceCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: SubSystem,
			Name:      KeyInstanceTotal,
			Help:      "Gauge of microservice created in Service Center",
		}, []string{LabelInstance, LabelFramework, LabelFrameworkVersion, LabelDomain, LabelProject})

	schemaCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: SubSystem,
			Name:      KeySchemaTotal,
			Help:      "Gauge of schema created in Service Center",
		}, []string{LabelInstance, LabelDomain, LabelProject})

	frameworkCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metrics.FamilyName,
			Subsystem: SubSystem,
			Name:      KeyFrameworkTotal,
			Help:      "Gauge of client framework info in Service Center",
		}, []string{LabelInstance, LabelFramework, LabelFrameworkVersion, LabelDomain, LabelProject})
)

func init() {
	prometheus.MustRegister(domainCounter, serviceCounter, instanceCounter, schemaCounter, frameworkCounter)
}
