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

package gov_test

import (
	"encoding/json"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/gov"
)

func TestNewInstance3(t *testing.T) {
	b, _ := json.MarshalIndent(&gov.Policy{
		GovernancePolicy: &gov.GovernancePolicy{
			Name: "Traffic2adminAPI",
		},
		Spec: &gov.LBSpec{RetryNext: 3, MarkerName: "traffic2adminAPI"},
	}, "", "  ")
	t.Log(string(b))
}
func TestNewInstance2(t *testing.T) {
	b, _ := json.MarshalIndent(&gov.RateLimiter{
		GovernancePolicy: &gov.GovernancePolicy{
			Name: "limitTraffic2adminAPI",
		},
		Spec: &gov.LimiterSpec{Burst: 10, Rate: 100, MarkerName: "traffic2adminAPI"},
	}, "", "  ")
	t.Log(string(b))
}
func TestNewInstance(t *testing.T) {
	b, _ := json.MarshalIndent(&gov.TrafficMarker{
		GovernancePolicy: &gov.GovernancePolicy{
			Name: "traffic2adminAPI",
			ID:   "",
			Selector: &gov.Selector{
				App:         "default",
				Environment: "development",
			},
		},
		Spec: &gov.MatchSpec{
			TrafficMarkPolicy: "perService",
			MatchPolicies: []*gov.MatchPolicy{
				{
					Headers: map[string]map[string]string{
						"X-User": {"regex": "ja.*"},
					},
					APIPaths: map[string]string{
						"exact": "/metrics",
					},
					Methods: []string{"GET", "POST"},
				},
				{
					APIPaths: map[string]string{
						"exact": "/health",
					},
					Methods: []string{"GET", "POST"},
				},
			},
		},
	}, "", "  ")
	t.Log(string(b))
}
