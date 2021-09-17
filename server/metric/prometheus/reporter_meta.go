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
	"github.com/apache/servicecomb-service-center/server/metric"
)

var metaReporter = &MetaReporter{}

type MetricsLabels struct {
	Domain           string `json:"domain,omitempty"`
	Project          string `json:"project,omitempty"`
	Framework        string `json:"framework,omitempty"`
	FrameworkVersion string `json:"frameworkVersion,omitempty"`
}

type MetaReporter struct {
}

func (m *MetaReporter) DomainAdd(delta float64) {
	instance := metric.InstanceName()
	labels := map[string]string{
		"instance": instance,
	}
	domainCounter.With(labels).Add(delta)
}
func (m *MetaReporter) ServiceAdd(delta float64, ml MetricsLabels) {
	instance := metric.InstanceName()
	labels := map[string]string{
		"instance":         instance,
		"framework":        ml.Framework,
		"frameworkVersion": ml.FrameworkVersion,
		"domain":           ml.Domain,
		"project":          ml.Project,
	}
	serviceCounter.With(labels).Add(delta)
}
func (m *MetaReporter) InstanceAdd(delta float64, ml MetricsLabels) {
	instance := metric.InstanceName()
	labels := map[string]string{
		"instance":         instance,
		"framework":        ml.Framework,
		"frameworkVersion": ml.FrameworkVersion,
		"domain":           ml.Domain,
		"project":          ml.Project,
	}
	instanceCounter.With(labels).Add(delta)
}
func (m *MetaReporter) SchemaAdd(delta float64, ml MetricsLabels) {
	instance := metric.InstanceName()
	labels := map[string]string{
		"instance": instance,
		"domain":   ml.Domain,
		"project":  ml.Project,
	}
	schemaCounter.With(labels).Add(delta)
}
func (m *MetaReporter) FrameworkSet(ml MetricsLabels) {
	instance := metric.InstanceName()
	labels := map[string]string{
		"instance":         instance,
		"framework":        ml.Framework,
		"frameworkVersion": ml.FrameworkVersion,
		"domain":           ml.Domain,
		"project":          ml.Project,
	}
	frameworkCounter.With(labels).Set(1)
}

func GetMetaReporter() *MetaReporter {
	return metaReporter
}

func ResetMetaMetrics() {
	domainCounter.Reset()
	serviceCounter.Reset()
	instanceCounter.Reset()
	schemaCounter.Reset()
}
