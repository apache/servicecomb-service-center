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
)

var (
	DefaultCalculator Calculator = &CommonCalculator{}
)

type Calculator interface {
	Calc(mf *dto.MetricFamily) float64
}

type CommonCalculator struct {
}

// Get value of metricFamily
func (c *CommonCalculator) Calc(mf *dto.MetricFamily) float64 {
	if len(mf.GetMetric()) == 0 {
		return 0
	}

	switch mf.GetType() {
	case dto.MetricType_GAUGE:
		return mf.GetMetric()[0].GetGauge().GetValue()
	case dto.MetricType_COUNTER:
		return metricCounterOf(mf.GetMetric())
	case dto.MetricType_SUMMARY:
		return metricSummaryOf(mf.GetMetric())
	default:
		return 0
	}
}

func metricCounterOf(m []*dto.Metric) float64 {
	var sum float64 = 0
	for _, d := range m {
		sum += d.GetCounter().GetValue()
	}
	return sum
}

func metricSummaryOf(m []*dto.Metric) float64 {
	var (
		count uint64  = 0
		sum   float64 = 0
	)
	for _, d := range m {
		count += d.GetSummary().GetSampleCount()
		sum += d.GetSummary().GetSampleSum()
	}

	if count == 0 {
		return 0
	}

	return sum / float64(count)
}

func RegisterCalculator(c Calculator) {
	DefaultCalculator = c
}

func Calculate(mf *dto.MetricFamily) float64 {
	return DefaultCalculator.Calc(mf)
}
