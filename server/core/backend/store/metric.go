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
package store

import (
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

var (
	cacheSizeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "service_center",
			Subsystem: "local",
			Name:      "cache_size_bytes",
			Help:      "Local cache size summary of backend store",
		}, []string{"instance", "resource", "type"})
)

var (
	instance string
	once     sync.Once
)

func init() {
	prometheus.MustRegister(cacheSizeGauge)
}

func ReportCacheMetrics(resource, t string, obj interface{}) {
	once.Do(func() {
		instance, _ = util.ParseEndpoint(core.Instance.Endpoints[0])
	})
	cacheSizeGauge.WithLabelValues(instance, resource, t).Set(float64(util.Sizeof(obj)))
}
