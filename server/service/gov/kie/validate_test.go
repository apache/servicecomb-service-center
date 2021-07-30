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

package kie

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchGroupValidate(t *testing.T) {
	var marchGroupParam = map[string]interface{}{
		"alias": "matchGroup",
		"matches": []map[string]interface{}{
			{
				"apiPath": map[string]interface{}{},
				"method":  []string{"GET", "POST"},
				"name":    "match1",
			},
			{
				"headers": map[string]interface{}{},
				"method":  []string{"PUT", "DELETE", "PATCH"},
				"name":    "match2",
			},
		},
	}
	err := matchValidate(marchGroupParam)
	assert.Empty(t, err)
}

func TestPolicyValidate(t *testing.T) {
	var bulkheadParam = map[string]interface{}{}
	err := policyValidate(bulkheadParam)
	assert.Empty(t, err)
	bulkheadParam = map[string]interface{}{
		"rules": map[string]interface{}{
			"match": "match",
		}}
	err = policyValidate(bulkheadParam)
	assert.Empty(t, err)
}

func TestBulkheadValidate(t *testing.T) {
	var bulkheadParam = map[string]interface{}{
		"name":               "bulkhead",
		"maxConcurrentCalls": 1000,
		"maxWaitDuration":    "10S",
	}
	err := bulkheadValidate(bulkheadParam)
	assert.Empty(t, err)
}

func TestCircuitBreakerValidate(t *testing.T) {
	var circuitBreakerParam = map[string]interface{}{
		"failureRateThreshold":      float32(50),
		"slowCallRateThreshold":     float32(100),
		"slowCallDurationThreshold": 6000,
		"minimumNumberOfCalls":      100,
		"slidingWindowType":         "count",
		"slidingWindowSize":         100,
	}
	err := circuitBreakerValidate(circuitBreakerParam)
	assert.Empty(t, err)
}
