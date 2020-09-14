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

package model

//GovernanceFrame is a unified struct
//all governance policy must extend this struct
//Name is the policy name, for example: "rate-limit-payment-api"
//Selector is a description which specify effected scope, it is metadata.
type GovernanceFrame struct {
	Name     string            `json:"name,omitempty"`
	Selector map[string]string `json:"selector,omitempty"`
}

//RateLimiter limit request rate
type RateLimiter struct {
	*GovernanceFrame
	Spec *LimiterSpec `json:"spec,omitempty"`
}
type LimiterSpec struct {
	MarkerName string `json:"match"`
	Rate       int    `json:"rate"`
	Burst      int    `json:"burst"`
}

//LoadBalancer define policy and fault tolerant policy
type LoadBalancer struct {
	*GovernanceFrame
	Spec *LBSpec `json:"spec,omitempty"`
}
type LBSpec struct {
	MarkerName string         `json:"match"`
	RetrySame  int            `json:"retrySame,omitempty"`
	RetryNext  int            `json:"retryNext,omitempty"`
	Bo         *BackOffPolicy `json:"backoff,omitempty"`
}
type BackOffPolicy struct {
	kind            string `json:"kind"`
	InitialInterval int    `json:"initInterval"`
	MaxInterval     int    `json:"maxInterval"`
}

//TrafficMarker marks request, it assign a name to request in runtime
type TrafficMarker struct {
	*GovernanceFrame
	Spec *MatchSpec `json:"spec,omitempty"`
}
type MatchSpec struct {
	MatchPolicies     []*MatchPolicy `json:"matches,omitempty"`
	TrafficMarkPolicy string         `json:"trafficMarkPolicy,omitempty"`
}

//MatchPolicy specify a request mach policy
type MatchPolicy struct {
	Headers  map[string]map[string]string `json:"headers,omitempty"`
	APIPaths map[string]string            `json:"apiPath,omitempty"`
	Methods  []string                     `json:"methods,omitempty"`
}
