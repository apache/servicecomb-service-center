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

package grc_test

import (
	"os"
	"testing"

	"github.com/apache/servicecomb-service-center/server/config"
	"github.com/apache/servicecomb-service-center/server/service/grc"
)

func init() {
	os.Setenv("APP_ROOT", "./../../../etc/")
	config.Init()
	grc.Init()
}

func TestValidatePolicySpec(t *testing.T) {
	type args struct {
		kind string
		spec interface{}
	}

	kindMatchGroup := "match-group"
	kindRetry := "retry"
	kindRateLimiting := "rate-limiting"
	kindLoadbalance := "loadbalance"
	kindCircuitBreaker := "circuit-breaker"
	kindInstanceIsolation := "instance-isolation"
	kindFaultInjection := "fault-injection"
	kindBulkhead := "bulkhead"

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"undefined", args{kind: "undefined", spec: map[string]interface{}{}}, true},

		{kindMatchGroup, args{kind: kindMatchGroup, spec: ""}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"": nil}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": nil}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{}}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{}, "alias": "1"}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{nil}, "alias": "1"}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{},
		}, "alias": "1"}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": ""},
		}, "alias": "1"}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1"},
		}, "alias": "1"}}, false},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1"},
		}, "alias": ""}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1", "apiPath": ""},
		}, "alias": "1"}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1", "apiPath": map[string]interface{}{"prefix": "/"}},
		}, "alias": "1"}}, false},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1", "headers": nil},
		}, "alias": "1"}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1", "headers": map[string]interface{}{}},
		}, "alias": "1"}}, false},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1", "headers": map[string]interface{}{"k": "v"}},
		}, "alias": "1"}}, false},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1", "method": []interface{}{}},
		}, "alias": "1"}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1", "method": []interface{}{1}},
		}, "alias": "1"}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1", "method": []interface{}{""}},
		}, "alias": "1"}}, true},
		{kindMatchGroup, args{kind: kindMatchGroup, spec: map[string]interface{}{"matches": []interface{}{
			map[string]interface{}{"name": "1", "method": []interface{}{"GET"}},
		}, "alias": "1"}}, false},

		{kindRetry, args{kind: kindRetry, spec: ""}, true},
		{kindRetry, args{kind: kindRetry, spec: map[string]interface{}{}}, true},
		{kindRetry, args{kind: kindRetry, spec: map[string]interface{}{"maxAttempts": 3}}, false},

		{kindRateLimiting, args{kind: kindRateLimiting, spec: ""}, true},
		{kindRateLimiting, args{kind: kindRateLimiting, spec: map[string]interface{}{}}, true},
		{kindRateLimiting, args{kind: kindRateLimiting, spec: map[string]interface{}{"rate": 1}}, false},
		{kindRateLimiting, args{kind: kindRateLimiting, spec: map[string]interface{}{"rate": 0.5}}, false},
		{kindRateLimiting, args{kind: kindRateLimiting, spec: map[string]interface{}{"rate": "1"}}, false},

		{kindLoadbalance, args{kind: kindLoadbalance, spec: ""}, true},
		{kindLoadbalance, args{kind: kindLoadbalance, spec: map[string]interface{}{}}, true},
		{kindLoadbalance, args{kind: kindLoadbalance, spec: map[string]interface{}{"rule": 1}}, true},
		{kindLoadbalance, args{kind: kindLoadbalance, spec: map[string]interface{}{"rule": ""}}, true},
		{kindLoadbalance, args{kind: kindLoadbalance, spec: map[string]interface{}{"rule": "1"}}, false},

		{kindCircuitBreaker, args{kind: kindCircuitBreaker, spec: ""}, true},
		{kindCircuitBreaker, args{kind: kindCircuitBreaker, spec: map[string]interface{}{}}, true},
		{kindCircuitBreaker, args{kind: kindCircuitBreaker, spec: map[string]interface{}{"minimumNumberOfCalls": -1}}, true},
		{kindCircuitBreaker, args{kind: kindCircuitBreaker, spec: map[string]interface{}{"minimumNumberOfCalls": 1}}, false},
		{kindCircuitBreaker, args{kind: kindCircuitBreaker, spec: map[string]interface{}{"minimumNumberOfCalls": "1"}}, true},

		{kindInstanceIsolation, args{kind: kindInstanceIsolation, spec: ""}, true},
		{kindInstanceIsolation, args{kind: kindInstanceIsolation, spec: map[string]interface{}{}}, true},
		{kindInstanceIsolation, args{kind: kindInstanceIsolation, spec: map[string]interface{}{"minimumNumberOfCalls": -1}}, false},
		{kindInstanceIsolation, args{kind: kindInstanceIsolation, spec: map[string]interface{}{"minimumNumberOfCalls": 1}}, false},
		{kindInstanceIsolation, args{kind: kindInstanceIsolation, spec: map[string]interface{}{"minimumNumberOfCalls": "1"}}, false},

		{kindFaultInjection, args{kind: kindFaultInjection, spec: ""}, true},
		{kindFaultInjection, args{kind: kindFaultInjection, spec: map[string]interface{}{}}, true},
		{kindFaultInjection, args{kind: kindFaultInjection, spec: map[string]interface{}{"percentage": -1}}, true},
		{kindFaultInjection, args{kind: kindFaultInjection, spec: map[string]interface{}{"percentage": 1}}, false},
		{kindFaultInjection, args{kind: kindFaultInjection, spec: map[string]interface{}{"percentage": 0.5}}, false},
		{kindFaultInjection, args{kind: kindFaultInjection, spec: map[string]interface{}{"percentage": "1"}}, true},

		{kindBulkhead, args{kind: kindBulkhead, spec: ""}, true},
		{kindBulkhead, args{kind: kindBulkhead, spec: map[string]interface{}{}}, true},
		{kindBulkhead, args{kind: kindBulkhead, spec: map[string]interface{}{"maxConcurrentCalls": -1}}, false},
		{kindBulkhead, args{kind: kindBulkhead, spec: map[string]interface{}{"maxConcurrentCalls": 1}}, false},
		{kindBulkhead, args{kind: kindBulkhead, spec: map[string]interface{}{"maxConcurrentCalls": "1"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := grc.ValidatePolicySpec(tt.args.kind, tt.args.spec); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
