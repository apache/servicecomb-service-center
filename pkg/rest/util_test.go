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

package rest_test

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/rest"
)

func TestParseQueries(t *testing.T) {
	type args struct {
		query url.Values
		key   string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{"invalid queries", args{nil, "any"}, map[string]string{}},
		{"invalid queries", args{url.Values{}, "any"}, map[string]string{}},
		{"invalid queries", args{url.Values{"a": {"b:a"}}, "other"}, map[string]string{}},
		{"valid queries", args{url.Values{"a": {""}}, "a"}, map[string]string{"": ""}},
		{"valid queries", args{url.Values{"a": {"b"}}, "a"}, map[string]string{"b": ""}},
		{"valid queries", args{url.Values{"a": {"b:"}}, "a"}, map[string]string{"b": ""}},
		{"valid queries", args{url.Values{"a": {":"}}, "a"}, map[string]string{"": ""}},
		{"valid queries", args{url.Values{"a": {":a"}}, "a"}, map[string]string{"": "a"}},
		{"valid queries", args{url.Values{"a": {"b:a"}}, "a"}, map[string]string{"b": "a"}},
		{"valid queries", args{url.Values{"a": {"b:a", "c:d"}}, "a"}, map[string]string{"b": "a", "c": "d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rest.ParseQueries(tt.args.query, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseQueries() = %v, want %v", got, tt.want)
			}
		})
	}
}
