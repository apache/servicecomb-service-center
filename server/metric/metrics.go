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
package metric

import "github.com/apache/incubator-servicecomb-service-center/pkg/util"

func NewMetrics() *Metrics {
	return &Metrics{
		ConcurrentMap: util.NewConcurrentMap(defaultMetricsSize),
	}
}

type Metrics struct {
	*util.ConcurrentMap
}

func (cm *Metrics) Put(key string, val float64) (old float64) {
	old, _ = cm.ConcurrentMap.Put(key, val).(float64)
	return
}

func (cm *Metrics) Get(key string) (val float64) {
	if v, ok := cm.ConcurrentMap.Get(key); ok {
		val, _ = v.(float64)
	}
	return
}

func (cm *Metrics) Remove(key string) (old float64) {
	old, _ = cm.ConcurrentMap.Remove(key).(float64)
	return
}

func (cm *Metrics) ForEach(f func(k string, v float64) (next bool)) {
	cm.ConcurrentMap.ForEach(func(item util.MapItem) (next bool) {
		k, _ := item.Key.(string)
		v, _ := item.Value.(float64)
		return f(k, v)
	})
}
