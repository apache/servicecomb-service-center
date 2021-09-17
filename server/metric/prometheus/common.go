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
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func getValue(name string, labels prometheus.Labels) float64 {
	f := Family(name)
	if f == nil {
		return 0
	}
	matchAll := len(labels) == 0
	var sum float64
	for _, m := range f.Metric {
		if !matchAll && !MatchLabels(m, labels) {
			continue
		}
		sum += m.GetGauge().GetValue()
	}
	return sum
}

func GaugeValue(name string, labels prometheus.Labels) int64 {
	return int64(getValue(name, labels))
}

func MatchLabels(m *dto.Metric, labels prometheus.Labels) bool {
	count := 0
	for _, label := range m.GetLabel() {
		v, ok := labels[label.GetName()]
		if ok && v != label.GetValue() {
			return false
		}
		if ok {
			count++
		}
	}
	return count == len(labels)
}

func Family(name string) *dto.MetricFamily {
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return nil
	}
	for _, f := range families {
		if f.GetName() == metric.FamilyName+name {
			return f
		}
	}
	return nil
}
