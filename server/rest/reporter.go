// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"github.com/apache/incubator-servicecomb-service-center/server/metric"
	dto "github.com/prometheus/client_model/go"
)

const (
	httpRequestTotal = "http_request_total"
)

var qpsLabelMap = map[string]int{
	"method":   0,
	"instance": 1,
	"api":      2,
	"domain":   3,
}

type APIReporter struct {
	cache *metric.Details
}

func (r *APIReporter) Report() {
	details := metric.Gatherer.Records.Get(httpRequestTotal)
	defer func() { r.cache = details }()

	if r.cache == nil {
		return
	}
	details.ForEach(func(labels []*dto.LabelPair, v float64) (next bool) {
		old := r.cache.Get(labels)
		queryPerSeconds.WithLabelValues(r.toLabels(labels)...).Set((v - old) / metric.Period.Seconds())
		return true
	})
}

func (r *APIReporter) toLabels(pairs []*dto.LabelPair) (labels []string) {
	labels = make([]string, len(qpsLabelMap))
	for _, pair := range pairs {
		if i, ok := qpsLabelMap[pair.GetName()]; ok {
			labels[i] = pair.GetValue()
		}
	}
	return
}

func init() {
	metric.RegisterReporter("rest", NewAPIReporter())
}

func NewAPIReporter() *APIReporter {
	return &APIReporter{}
}
