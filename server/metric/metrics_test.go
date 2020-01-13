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
	"reflect"
	"testing"
)

func TestDetails_ForEach(t *testing.T) {
	n, v := "name", "value"
	d := NewDetails()
	d.ForEach(func(labels []*dto.LabelPair, v float64) (next bool) {
		t.Fatalf("TestDetails_ForEach failed")
		return true
	})
	d.put([]*dto.LabelPair{{Name: &n, Value: &v}, {Name: &n, Value: &v}}, 1)
	d.ForEach(func(labels []*dto.LabelPair, v float64) (next bool) {
		if len(labels) != 2 || v != 1 {
			t.Fatalf("TestDetails_ForEach failed")
		}
		return true
	})
	d.put([]*dto.LabelPair{{}}, 3)
	d.put([]*dto.LabelPair{{}}, 4)
	d.put(nil, 2)
	l := 0
	d.ForEach(func(labels []*dto.LabelPair, v float64) (next bool) {
		switch {
		case len(labels) == 2:
			if v != 1 {
				t.Fatalf("TestDetails_ForEach failed")
			}
		case len(labels) == 0:
			if v != 2 {
				t.Fatalf("TestDetails_ForEach failed")
			}
		case len(labels) == 1:
			if v != 4 {
				t.Fatalf("TestDetails_ForEach failed")
			}
		default:
			t.Fatalf("TestDetails_ForEach failed")
		}
		l++
		return true
	})
	if l != 3 {
		t.Fatalf("TestDetails_ForEach failed")
	}

	ms := NewMetrics()
	if ms.Summary("x") != 0 {
		t.Fatalf("TestMetrics_Summary failed")
	}
	if ms.Get("x") != nil {
		t.Fatalf("TestMetrics_Details failed")
	}
	find := false
	ms.ForEach(func(k string, v *Details) (next bool) {
		find = true
		return true
	})
	if find {
		t.Fatalf("TestMetrics_ForEach failed")
	}
	ms.put("a", d)
	if ms.Summary("a") != 0 {
		t.Fatalf("TestMetrics_Summary failed")
	}
	if ms.Get("a") != d {
		t.Fatalf("TestMetrics_Get failed")
	}
	ms.ForEach(func(k string, v *Details) (next bool) {
		find = true
		return true
	})
	if !find {
		t.Fatalf("TestMetrics_ForEach failed")
	}
}

func TestToRawData(t *testing.T) {
	type Hello struct {
		A     string  `json:"a"`
		Count float64 `json:"count"`
	}
	result := new(Hello)
	a := "a"
	b := "b"
	labels := []*dto.LabelPair{
		{Name: &a, Value: &b},
	}
	ToRawData(result, labels, 5)

	if result.A != "b" || result.Count != 5 {
		t.Fatalf("To raw data failed, %v", *result)
	}

	data := ToLabelNames(Hello{})
	if !reflect.DeepEqual(data, []string{"a"}) {
		t.Fatalf("to label names failed")
	}
}
