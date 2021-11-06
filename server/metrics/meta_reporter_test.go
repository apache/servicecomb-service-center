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

	_ "github.com/apache/servicecomb-service-center/test"

	"github.com/apache/servicecomb-service-center/datasource"
	promutil "github.com/apache/servicecomb-service-center/pkg/prometheus"
	"github.com/apache/servicecomb-service-center/server/metrics"
	"github.com/apache/servicecomb-service-center/server/plugin/quota"
	"github.com/stretchr/testify/assert"
)

func TestMetaReporter_ServiceUsageSet(t *testing.T) {
	labels := map[string]string{}

	reporter := metrics.MetaReporter{}
	assert.Equal(t, 0/float64(quota.DefaultServiceQuota), promutil.GaugeValue(metrics.KeyServiceUsage, labels))

	reporter.ServiceAdd(1, datasource.MetricsLabels{Domain: "D1", Project: "P1"})
	reporter.ServiceUsageSet()
	assert.Equal(t, 1/float64(quota.DefaultServiceQuota), promutil.GaugeValue(metrics.KeyServiceUsage, labels))

	reporter.ServiceAdd(1, datasource.MetricsLabels{Domain: "D1", Project: "P2"})
	reporter.ServiceAdd(1, datasource.MetricsLabels{Domain: "D2", Project: "P3"})
	reporter.ServiceUsageSet()
	assert.Equal(t, 3/float64(quota.DefaultServiceQuota), promutil.GaugeValue(metrics.KeyServiceUsage, labels))
}

func TestMetaReporter_InstanceUsageSet(t *testing.T) {
	labels := map[string]string{}

	reporter := metrics.MetaReporter{}
	assert.Equal(t, 0/float64(quota.DefaultInstanceQuota), promutil.GaugeValue(metrics.KeyInstanceUsage, labels))

	reporter.InstanceAdd(1, datasource.MetricsLabels{Domain: "D1", Project: "P1"})
	reporter.InstanceUsageSet()
	assert.Equal(t, 1/float64(quota.DefaultInstanceQuota), promutil.GaugeValue(metrics.KeyInstanceUsage, labels))

	reporter.InstanceAdd(1, datasource.MetricsLabels{Domain: "D1", Project: "P2"})
	reporter.InstanceAdd(1, datasource.MetricsLabels{Domain: "D2", Project: "P3"})
	reporter.InstanceUsageSet()
	assert.Equal(t, 3/float64(quota.DefaultInstanceQuota), promutil.GaugeValue(metrics.KeyInstanceUsage, labels))
}
