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

package metrics

import (
	dto "github.com/prometheus/client_model/go"
)

var (
	DefaultCalculator Calculator = &CommonCalculator{}
)

// Calculator is the interface to implement customize algorithm of MetricFamily
type Calculator interface {
	Calc(mf *dto.MetricFamily) *Details
}

type CommonCalculator struct {
}

// Get value of metricFamily
func (c *CommonCalculator) Calc(mf *dto.MetricFamily) *Details {
	if len(mf.GetMetric()) == 0 {
		return nil
	}

	details := NewDetails()
	switch mf.GetType() {
	case dto.MetricType_GAUGE:
		metricGaugeOf(details, mf.GetMetric())
	case dto.MetricType_COUNTER:
		metricCounterOf(details, mf.GetMetric())
	case dto.MetricType_SUMMARY:
		metricSummaryOf(details, mf.GetMetric())
	case dto.MetricType_HISTOGRAM:
		metricHistogramOf(details, mf.GetMetric())
	}
	return details
}

func metricGaugeOf(details *Details, m []*dto.Metric) {
	for _, d := range m {
		details.Summary += d.GetGauge().GetValue()
		details.put(d.GetLabel(), d.GetGauge().GetValue())
	}
}

func metricCounterOf(details *Details, m []*dto.Metric) {
	for _, d := range m {
		details.Summary += d.GetCounter().GetValue()
		details.put(d.GetLabel(), d.GetCounter().GetValue())
	}
}

func metricSummaryOf(details *Details, m []*dto.Metric) {
	var (
		count uint64
		sum   float64
	)
	for _, d := range m {
		count += d.GetSummary().GetSampleCount()
		sum += d.GetSummary().GetSampleSum()
		details.put(d.GetLabel(), d.GetSummary().GetSampleSum()/float64(d.GetSummary().GetSampleCount()))
	}

	if count == 0 {
		return
	}

	details.Summary = sum / float64(count)
}

func metricHistogramOf(details *Details, m []*dto.Metric) {
	var (
		count uint64
		sum   float64
	)
	for _, d := range m {
		count += d.GetHistogram().GetSampleCount()
		sum += d.GetHistogram().GetSampleSum()
		details.put(d.GetLabel(), d.GetHistogram().GetSampleSum()/float64(d.GetHistogram().GetSampleCount()))
	}

	if count == 0 {
		return
	}

	details.Summary = sum / float64(count)
}

func RegisterCalculator(c Calculator) {
	DefaultCalculator = c
}

func Calculate(mf *dto.MetricFamily) *Details {
	return DefaultCalculator.Calc(mf)
}
