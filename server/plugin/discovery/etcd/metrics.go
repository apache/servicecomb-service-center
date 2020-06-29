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
package etcd

import (
	"github.com/apache/servicecomb-service-center/server/metric"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	cacheSizeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: metric.FamilyName,
			Subsystem: "local",
			Name:      "cache_size_bytes",
			Help:      "Local cache size summary of backend store",
		}, []string{"instance", "resource", "type"})
)

func init() {
	prometheus.MustRegister(cacheSizeGauge)
}

func ReportCacheSize(resource, t string, s int) {
	instance := metric.InstanceName()
	if len(instance) == 0 || len(resource) == 0 {
		// endpoints list will be empty when initializing
		// resource may be empty when report SCHEMA
		return
	}

	cacheSizeGauge.WithLabelValues(instance, resource, t).Set(float64(s))
}
