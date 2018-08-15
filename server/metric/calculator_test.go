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

import (
	dto "github.com/prometheus/client_model/go"
	"testing"
)

func TestCommonCalculator_Calc(t *testing.T) {
	c := &CommonCalculator{}

	mf := &dto.MetricFamily{}
	mt := dto.MetricType_UNTYPED
	v1 := float64(0)
	v2 := float64(0)
	n := uint64(0)

	if c.Calc(mf) != nil {
		t.Fatalf("TestCommonCalculator_Calc failed")
	}

	mf = &dto.MetricFamily{Type: &mt, Metric: []*dto.Metric{{}}}
	if c.Calc(mf) == nil {
		t.Fatalf("TestCommonCalculator_Calc failed")
	}

	mt = dto.MetricType_GAUGE
	v1 = 1
	v2 = 2
	mf = &dto.MetricFamily{Type: &mt, Metric: []*dto.Metric{
		{Gauge: &dto.Gauge{Value: &v1}}, {Gauge: &dto.Gauge{Value: &v2}}}}
	details := c.Calc(mf)
	if details.Value != 3 {
		t.Fatalf("TestCommonCalculator_Calc failed")
	}

	mt = dto.MetricType_COUNTER
	v1 = 1
	mf = &dto.MetricFamily{Type: &mt, Metric: []*dto.Metric{
		{Counter: &dto.Counter{Value: &v1}}, {Counter: &dto.Counter{Value: &v1}}}}
	details = c.Calc(mf)
	if details.Value != 2 {
		t.Fatalf("TestCommonCalculator_Calc failed")
	}

	mt = dto.MetricType_SUMMARY
	v1 = 3
	n = 2
	mf = &dto.MetricFamily{Type: &mt, Metric: []*dto.Metric{
		{Summary: &dto.Summary{SampleCount: &n, SampleSum: &v1}}, {Summary: &dto.Summary{SampleCount: &n, SampleSum: &v1}}}}
	details = c.Calc(mf)
	if details.Value != v1/float64(n) {
		t.Fatalf("TestCommonCalculator_Calc failed")
	}
}
