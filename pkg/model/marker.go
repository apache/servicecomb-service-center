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

//TrafficMarker marks request, it assign a name to request in runtime
type TrafficMarker struct {
	*GovernancePolicy
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
