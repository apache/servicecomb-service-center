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
	"fmt"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	cacheSizeSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "service_center",
			Subsystem:  "local",
			Name:       "cache_size",
			Help:       "Local cache size summary of backend store",
			Objectives: prometheus.DefObjectives,
		}, []string{"instance", "resource"})
)

func init() {
	prometheus.MustRegister(cacheSizeSummary)
}

func StoreMetric(i *Indexer) {
	instance := fmt.Sprint(core.Instance.Endpoints)
	cacheSizeSummary.WithLabelValues(instance, i.cacheType.String()).Observe(float64(util.Sizeof(i)))
}
