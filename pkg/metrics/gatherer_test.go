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
package metrics_test

import (
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func TestMetricsGatherer_Collect(t *testing.T) {
	g := metrics.NewGatherer(metrics.Options{})
	err := g.Collect()
	if err != nil {
		t.Fatalf("TestMetricsGatherer_Collect")
	}
}

func TestCollectFamily(t *testing.T) {
	t.Run("not register family name, should return empty", func(t *testing.T) {
		family := metrics.ParseFamily("abc")
		assert.Empty(t, family)

		family = metrics.ParseFamily("")
		assert.Empty(t, family)

		family = metrics.ParseFamily("a_b_c")
		assert.Empty(t, family)
	})

	t.Run("register family name, should return prefix", func(t *testing.T) {
		family := metrics.ParseFamily("service_center_a")
		assert.Equal(t, "service_center", family)

		family = metrics.ParseFamily("test_abc")
		assert.Empty(t, family)

		metrics.CollectFamily("test")

		family = metrics.ParseFamily("test_abc")
		assert.Equal(t, "test", family)
	})
}

func TestRecordName(t *testing.T) {
	t.Run("get metric name without family, should be ok", func(t *testing.T) {
		name := metrics.RecordName("service_center_a")
		assert.Equal(t, "a", name)

		name = metrics.RecordName("sys_abc")
		assert.Empty(t, name)

		metrics.SysMetrics.Put("sys_abc", struct{}{})
		name = metrics.RecordName("sys_abc")
		assert.Equal(t, "sys_abc", name)
	})
}
