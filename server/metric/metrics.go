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
	"github.com/apache/incubator-servicecomb-service-center/pkg/buffer"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	dto "github.com/prometheus/client_model/go"
	"strings"
)

func NewMetrics() *Metrics {
	return &Metrics{
		mapper: util.NewConcurrentMap(0),
	}
}

func NewDetails() *Details {
	return &Details{
		mapper: util.NewConcurrentMap(0),
		buffer: buffer.NewPool(bufferSize),
	}
}

// Details is the struct to hold the calculated result and index by metric label
type Details struct {
	// Summary is the calculation results of the details
	Summary float64

	mapper *util.ConcurrentMap
	buffer *buffer.Pool
}

// to format 'N1=L1,N2=L2,N3=L3,...'
func (cm *Details) toKey(labels []*dto.LabelPair) string {
	b := cm.buffer.Get()
	for _, label := range labels {
		b.WriteString(label.GetName())
		b.WriteRune('=')
		b.WriteString(label.GetValue())
		b.WriteRune(',')
	}
	key := b.String()
	cm.buffer.Put(b)
	return key
}

func (cm *Details) toLabels(key string) (p []*dto.LabelPair) {
	pairs := strings.Split(key, ",")
	pairs = pairs[:len(pairs)-1]
	p = make([]*dto.LabelPair, 0, len(pairs))
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		p = append(p, &dto.LabelPair{Name: &kv[0], Value: &kv[1]})
	}
	return
}

func (cm *Details) Get(labels []*dto.LabelPair) (val float64) {
	if v, ok := cm.mapper.Get(cm.toKey(labels)); ok {
		val = v.(float64)
	}
	return
}

func (cm *Details) Put(labels []*dto.LabelPair, val float64) {
	cm.mapper.Put(cm.toKey(labels), val)
	return
}

func (cm *Details) ForEach(f func(labels []*dto.LabelPair, v float64) (next bool)) {
	cm.mapper.ForEach(func(item util.MapItem) (next bool) {
		k, _ := item.Key.(string)
		v, _ := item.Value.(float64)
		return f(cm.toLabels(k), v)
	})
}

// Metrics is the struct to hold the Details objects store and index by metric name
type Metrics struct {
	mapper *util.ConcurrentMap
}

func (cm *Metrics) Put(key string, val *Details) {
	cm.mapper.Put(key, val)
}

func (cm *Metrics) Get(key string) (val *Details) {
	if v, ok := cm.mapper.Get(key); ok {
		val = v.(*Details)
	}
	return
}

func (cm *Metrics) ForEach(f func(k string, v *Details) (next bool)) {
	cm.mapper.ForEach(func(item util.MapItem) (next bool) {
		k, _ := item.Key.(string)
		v, _ := item.Value.(*Details)
		return f(k, v)
	})
}

func (cm *Metrics) Summary(key string) (sum float64) {
	if v, ok := cm.mapper.Get(key); ok {
		sum = v.(*Details).Summary
	}
	return
}
