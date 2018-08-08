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
	v := float64(0)
	n := uint64(0)

	if c.Calc(mf) != 0 {
		t.Fatalf("TestCommonCalculator_Calc failed")
	}

	mf = &dto.MetricFamily{Type: &mt, Metric: []*dto.Metric{{}}}
	if c.Calc(mf) != 0 {
		t.Fatalf("TestCommonCalculator_Calc failed")
	}

	mt = dto.MetricType_GAUGE
	v = 1
	mf = &dto.MetricFamily{Type: &mt, Metric: []*dto.Metric{
		{Gauge: &dto.Gauge{Value: &v}}, {Gauge: &dto.Gauge{Value: &v}}}}
	if c.Calc(mf) != 1 {
		t.Fatalf("TestCommonCalculator_Calc failed")
	}

	mt = dto.MetricType_COUNTER
	v = 1
	mf = &dto.MetricFamily{Type: &mt, Metric: []*dto.Metric{
		{Counter: &dto.Counter{Value: &v}}, {Counter: &dto.Counter{Value: &v}}}}
	if c.Calc(mf) != 2 {
		t.Fatalf("TestCommonCalculator_Calc failed")
	}

	mt = dto.MetricType_SUMMARY
	v = 3
	n = 2
	mf = &dto.MetricFamily{Type: &mt, Metric: []*dto.Metric{
		{Summary: &dto.Summary{SampleCount: &n, SampleSum: &v}}, {Summary: &dto.Summary{SampleCount: &n, SampleSum: &v}}}}
	if c.Calc(mf) != v/float64(n) {
		t.Fatalf("TestCommonCalculator_Calc failed")
	}
}
